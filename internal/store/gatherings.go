package store

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

type Reservation struct {
	ID           string                 `json:"id"`
	Intersection geography.Intersection `json:"intersection"`
	Founders     []string               `json:"founders"`
	ReadyAt      int64                  `json:"ready_at"`
	ExpiresAt    int64                  `json:"expires_at"`
	ConfigHash   string                 `json:"config_sha256"`
}
type Gathering struct {
	ID           string                 `json:"id"`
	Intersection geography.Intersection `json:"intersection"`
	ActivatedAt  int64                  `json:"activated_at"`
	EndsAt       int64                  `json:"ends_at"`
	State        string                 `json:"state"`
	ConfigHash   string                 `json:"config_sha256"`
}

var readClock = newScript(`__CLOCK__; return now`)

func (s *Store) Now(ctx context.Context) (int64, error) {
	return readClock.Run(ctx, s.Client, nil).Int64()
}

var reserve = newScript(`
__CLOCK__
if redis.call('EXISTS',KEYS[1])==1 or redis.call('EXISTS',KEYS[3])==1 or redis.call('EXISTS',KEYS[4])==1 then return 'CONFLICT' end
local proposal=cjson.decode(ARGV[1])
local ready=now+tonumber(ARGV[2]);local expiry=ready+tonumber(ARGV[3]);local minimum=ready+tonumber(ARGV[4])
local records={};local expected=cjson.decode(ARGV[5])
for n,hash in ipairs(proposal.founders) do
 local raw=redis.call('GET','gati:s:'..hash);if not raw then return 'CONFLICT' end
 local v=cjson.decode(raw)
 if v.cell~=expected[n].cell or v.radius_km~=expected[n].radius or v.created_at~=expected[n].created or v.expires_at~=expected[n].expires or v.expires_at<minimum or (v._pending_until or 0)>now or (v._gathering_until or 0)>now then return 'CONFLICT' end
 table.insert(records,v)
end
proposal.ready_at=ready;proposal.expires_at=expiry
redis.call('SET',KEYS[4],proposal.id,'PX',expiry-now)
redis.call('SET',KEYS[1],cjson.encode(proposal),'PX',expiry-now)
redis.call('ZADD',KEYS[2],ready,proposal.id)
if redis.call('PTTL',KEYS[2])<expiry-now then redis.call('PEXPIRE',KEYS[2],expiry-now) end
for n,hash in ipairs(proposal.founders) do
 local v=records[n];v._pending=proposal.id;v._pending_until=expiry
 redis.call('SET','gati:s:'..hash,cjson.encode(v),'KEEPTTL')
end
return cjson.encode(proposal)
`)

// Reserve takes immutable signal area/radius claims already checked by the planner.
// Every founder is re-read atomically; one credential can reserve only one group.
func (s *Store) Reserve(ctx context.Context, id, configHash string, p matching.Proposal, grid geography.Grid, stabilityMS, recoveryMS, minimumMS int64) (Reservation, error) {
	var result Reservation
	if len(p.Founders) == 0 || len(p.Founders) > 500 || stabilityMS < 1 || recoveryMS < 1 || minimumMS < 1 {
		return result, errors.New("reservation outside transaction bounds")
	}
	proposal := Reservation{ID: id, Intersection: p.Intersection, ConfigHash: configHash}
	seen := map[string]bool{}
	for _, founder := range p.Founders {
		if seen[founder.Hash] {
			return result, ErrConflict
		}
		seen[founder.Hash] = true
		proposal.Founders = append(proposal.Founders, founder.Hash)
	}
	expected := []map[string]any{}
	for _, f := range p.Founders {
		expected = append(expected, map[string]any{"cell": grid.ID(f.Area.Cell), "radius": f.Area.RadiusKM, "created": f.CreatedAt, "expires": f.ExpiresAt})
	}
	raw, _ := json.Marshal(proposal)
	claims, _ := json.Marshal(expected)
	value, e := reserve.Run(ctx, s.Client, []string{keyPrefix + "pending:" + id, keyPrefix + "pending", keyPrefix + "gathering:" + id, keyPrefix + "destination:" + p.Intersection.ID}, string(raw), stabilityMS, recoveryMS, minimumMS, string(claims)).Text()
	if e != nil {
		return result, errors.New("reservation failed")
	}
	if value == "CONFLICT" {
		return result, ErrConflict
	}
	e = json.Unmarshal([]byte(value), &result)
	return result, e
}

var activate = newScript(`
__CLOCK__
local existing=redis.call('GET',KEYS[3]);if existing then return existing end
local raw=redis.call('GET',KEYS[1]);if not raw then redis.call('ZREM',KEYS[2],ARGV[1]);return 'GONE' end
local p=cjson.decode(raw);if now<p.ready_at then return 'WAIT' end
local records={};local valid=now<p.expires_at and p.config_sha256==ARGV[4]
local ends=now+tonumber(ARGV[3]);local minimum=now+tonumber(ARGV[2])
for _,hash in ipairs(p.founders) do
 local value=redis.call('GET','gati:s:'..hash)
 if value then
  local v=cjson.decode(value);table.insert(records,{hash=hash,value=v})
  if v._pending~=p.id or v.expires_at<minimum or (v._gathering_until or 0)>now then valid=false end
  if v.expires_at<ends then ends=v.expires_at end
 else valid=false end
end
if not valid then
 for _,entry in ipairs(records) do
  local v=entry.value;if v._pending==p.id then v._pending=nil;v._pending_until=nil;redis.call('SET','gati:s:'..entry.hash,cjson.encode(v),'KEEPTTL') end
 end
 redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[1]);if redis.call('GET','gati:destination:'..p.intersection.id)==p.id then redis.call('DEL','gati:destination:'..p.intersection.id) end;return 'GONE'
end
local g={id=p.id,intersection=p.intersection,activated_at=now,ends_at=ends,state='jemi_gati',config_sha256=p.config_sha256}
redis.call('SET',KEYS[3],cjson.encode(g),'PX',ends-now)
redis.call('SET','gati:destination:'..g.intersection.id,g.id,'PX',ends-now)
redis.call('ZADD',KEYS[4],ends,g.id)
if redis.call('PTTL',KEYS[4])<ends-now then redis.call('PEXPIRE',KEYS[4],ends-now) end
for _,entry in ipairs(records) do
 local v=entry.value;v._pending=nil;v._pending_until=nil;v._gathering=g.id;v._gathering_until=ends;v.state='invited'
 redis.call('SET','gati:s:'..entry.hash,cjson.encode(v),'KEEPTTL')
end
redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[1])
return cjson.encode(g)
`)
var ErrWait = errors.New("stability interval incomplete")

func (s *Store) Activate(ctx context.Context, id, configHash string, minimumMS, maximumMS int64) (Gathering, error) {
	var g Gathering
	raw, e := activate.Run(ctx, s.Client, []string{keyPrefix + "pending:" + id, keyPrefix + "pending", keyPrefix + "gathering:" + id, keyPrefix + "gatherings"}, id, minimumMS, maximumMS, configHash).Text()
	if e != nil {
		return g, errors.New("activation failed")
	}
	if raw == "GONE" {
		return g, ErrGone
	}
	if raw == "WAIT" {
		return g, ErrWait
	}
	e = json.Unmarshal([]byte(raw), &g)
	return g, e
}
func (s *Store) DueReservations(ctx context.Context, now int64, batch int) ([]string, error) {
	return s.Client.ZRangeByScore(ctx, keyPrefix+"pending", &redis.ZRangeBy{Min: "-inf", Max: strconv.FormatInt(now, 10), Count: int64(batch)}).Result()
}

// EligibleSnapshot is private worker input, read in bounded pages. Concurrent
// changes can cause a missed candidate until reconciliation; atomic transitions
// prevent those stale snapshots from activating unavailable founders.
func (s *Store) EligibleSnapshot(ctx context.Context, grid geography.Grid, now int64, batch, maximum int) ([]matching.Signal, error) {
	result := []matching.Signal{}
	for offset := 0; offset < maximum; offset += batch {
		members, e := s.Client.ZRangeByScore(ctx, keyPrefix+"expiry", &redis.ZRangeBy{Min: strconv.FormatInt(now, 10), Max: "+inf", Offset: int64(offset), Count: int64(batch)}).Result()
		if e != nil {
			return nil, e
		}
		pipe := s.Client.Pipeline()
		commands := make([]*redis.StringCmd, 0, len(members))
		hashes := []string{}
		for _, member := range members {
			hash, _, ok := strings.Cut(member, "|")
			if !ok {
				continue
			}
			hashes = append(hashes, hash)
			commands = append(commands, pipe.Get(ctx, keyPrefix+"s:"+hash))
		}
		_, e = pipe.Exec(ctx)
		if e != nil && e != redis.Nil {
			return nil, e
		}
		for n, command := range commands {
			raw, e := command.Result()
			if e == redis.Nil {
				continue
			}
			if e != nil {
				return nil, e
			}
			v, e := decode(raw)
			if e != nil {
				return nil, e
			}
			cell, e := grid.Parse(v.Cell)
			if e != nil {
				return nil, e
			}
			result = append(result, matching.Signal{Hash: hashes[n], Area: geography.ParticipantArea{Cell: cell, RadiusKM: v.RadiusKM}, CreatedAt: v.CreatedAt, ExpiresAt: v.ExpiresAt, Assigned: v.GatheringUntil > now, ReservedUntil: v.PendingUntil, Declined: v.Declined})
		}
		if len(members) < batch {
			break
		}
	}
	return result, nil
}

// OpenGatherings returns private worker input, never a public aggregate endpoint.
func (s *Store) OpenGatherings(ctx context.Context, now int64, maximum int) ([]Gathering, error) {
	ids, e := s.Client.ZRangeByScore(ctx, keyPrefix+"gatherings", &redis.ZRangeBy{Min: strconv.FormatInt(now+1, 10), Max: "+inf", Count: int64(maximum)}).Result()
	if e != nil {
		return nil, e
	}
	out := []Gathering{}
	for offset := 0; offset < len(ids); offset += 500 {
		end := min(offset+500, len(ids))
		pipe := s.Client.Pipeline()
		commands := []*redis.StringCmd{}
		for _, id := range ids[offset:end] {
			commands = append(commands, pipe.Get(ctx, keyPrefix+"gathering:"+id))
		}
		_, e = pipe.Exec(ctx)
		if e != nil && e != redis.Nil {
			return nil, e
		}
		for _, command := range commands {
			raw, e := command.Result()
			if e == redis.Nil {
				continue
			}
			if e != nil {
				return nil, e
			}
			var g Gathering
			if e = json.Unmarshal([]byte(raw), &g); e != nil {
				return nil, e
			}
			if g.EndsAt > now {
				out = append(out, g)
			}
		}
	}
	return out, nil
}
func (s *Store) PendingReservations(ctx context.Context, maximum int) ([]Reservation, error) {
	ids, e := s.Client.ZRangeByScore(ctx, keyPrefix+"pending", &redis.ZRangeBy{Min: "-inf", Max: "+inf", Count: int64(maximum)}).Result()
	if e != nil {
		return nil, e
	}
	out := []Reservation{}
	for _, id := range ids {
		raw, e := s.Client.Get(ctx, keyPrefix+"pending:"+id).Result()
		if e == redis.Nil {
			continue
		}
		if e != nil {
			return nil, e
		}
		var p Reservation
		if e = json.Unmarshal([]byte(raw), &p); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, nil
}

var readGathering = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 'GONE' end
local g=cjson.decode(raw);if g.ends_at<=now then return 'GONE' end
if g.state=='jemi_ketu' and g._presence_threshold and redis.call('ZCOUNT','gati:arrivals:'..g.id,'('..now,'+inf')<g._presence_threshold then g.state='jemi_gati';raw=cjson.encode(g);redis.call('SET',KEYS[1],raw,'KEEPTTL');redis.call('DEL','gati:presence:'..g.id) end
return raw
`)

func (s *Store) GatheringByID(ctx context.Context, id string) (Gathering, error) {
	var g Gathering
	raw, e := readGathering.Run(ctx, s.Client, []string{keyPrefix + "gathering:" + id}).Text()
	if e != nil {
		return g, errors.New("gathering unavailable")
	}
	if raw == "GONE" {
		return g, ErrGone
	}
	e = json.Unmarshal([]byte(raw), &g)
	return g, e
}

var lease = newScript(`
local old=redis.call('GET',KEYS[1]);if old and old~=ARGV[1] then return 0 end
redis.call('SET',KEYS[1],ARGV[1],'PX',ARGV[2]);return 1
`)

func (s *Store) MatchingLease(ctx context.Context, owner string, ttlMS int64) (bool, error) {
	n, e := lease.Run(ctx, s.Client, []string{keyPrefix + "worker:lease"}, owner, ttlMS).Int()
	return n == 1, e
}

var consumeDirty = newScript(`local v=redis.call('GET',KEYS[1]);redis.call('DEL',KEYS[1]);if v then return 1 end;return 0`)

func (s *Store) ConsumeDirty(ctx context.Context) (bool, error) {
	n, e := consumeDirty.Run(ctx, s.Client, []string{keyPrefix + "dirty"}).Int()
	return n == 1, e
}
