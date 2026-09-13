package store

import (
	"context"
	"encoding/json"
	"errors"
)

// The client chooses a random nonce and authenticates issuance and confirmation
// with it. Only its hash is stored. Issuance retries preserve the first deadline.
var issueNonce = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 'GONE' end
local s=cjson.decode(raw)
if s.expires_at<=now or not s._gathering or (s._gathering_until or 0)<=now then return 'GONE' end
if s.arrival_until and s.arrival_until<=now then removeArrival(s);s.arrival_until=nil;s._arrival_member=nil;if s.state=='here' then s.state='going' end;redis.call('SET',KEYS[1],cjson.encode(s),'KEEPTTL') end
if s.state~='going' or (s.arrival_until or 0)>now then return 'CONFLICT' end
local event=redis.call('GET','gati:gathering:'..s._gathering);if not event or cjson.decode(event).ends_at<=now then return 'GONE' end
local old=redis.call('GET',KEYS[2])
if old then
 local nonce=cjson.decode(old)
 if nonce.hash==ARGV[1] then if nonce.gathering~=s._gathering or nonce.used or nonce.expires_at<=now then return 'GONE' end;return old end
end
local expiry=math.min(now+tonumber(ARGV[2]),s.expires_at,s._gathering_until)
local nonce=cjson.encode({hash=ARGV[1],gathering=s._gathering,expires_at=expiry,used=false})
redis.call('SET',KEYS[2],nonce,'PX',expiry-now)
return nonce
`)

func (s *Store) IssueArrivalNonce(ctx context.Context, hash, nonceHash string, ttlMS int64) (int64, error) {
	raw, e := issueNonce.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "arrival-nonce:" + hash}, nonceHash, ttlMS).Text()
	if e != nil {
		return 0, errors.New("arrival challenge unavailable")
	}
	if raw == "GONE" {
		return 0, ErrGone
	}
	if raw == "CONFLICT" {
		return 0, ErrConflict
	}
	var nonce struct {
		Expires int64 `json:"expires_at"`
	}
	e = json.Unmarshal([]byte(raw), &nonce)
	return nonce.Expires, e
}

var confirmArrival = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);local challenge=redis.call('GET',KEYS[2]);local event=redis.call('GET',KEYS[3])
if not raw or not challenge or not event then return 'GONE' end
local s=cjson.decode(raw);local nonce=cjson.decode(challenge);local g=cjson.decode(event)
if s.expires_at<=now or g.ends_at<=now or nonce.expires_at<=now or nonce.hash~=ARGV[1] or nonce.gathering~=g.id or s._gathering~=g.id or g.crossing.id~=ARGV[3] then return 'GONE' end
if nonce.used then if s.state=='here' and (s.arrival_until or 0)>now and s._arrival_member==ARGV[4]..'|'..ARGV[1] then return raw end;return 'GONE' end
if s.state~='going' or (s.arrival_until or 0)>now then return 'CONFLICT' end
if s._arrival_member then redis.call('ZREM',KEYS[4],s._arrival_member) end
local expiry=math.min(now+tonumber(ARGV[2]),s.expires_at,g.ends_at)
s.arrival_until=expiry;s._arrival_member=ARGV[4]..'|'..ARGV[1];s.state='here';nonce.used=true
redis.call('SET',KEYS[2],cjson.encode(nonce),'KEEPTTL')
redis.call('SET',KEYS[1],cjson.encode(s),'KEEPTTL')
redis.call('ZADD',KEYS[4],expiry,s._arrival_member)
if redis.call('PTTL',KEYS[4])<g.ends_at-now then redis.call('PEXPIRE',KEYS[4],g.ends_at-now) end
return cjson.encode(s)
`)

func (s *Store) ConfirmArrival(ctx context.Context, hash, nonceHash string, g Gathering, freshnessMS int64) (Signal, error) {
	raw, e := confirmArrival.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "arrival-nonce:" + hash, keyPrefix + "gathering:" + g.ID, keyPrefix + "arrivals:" + g.ID}, nonceHash, freshnessMS, g.Crossing.ID, hash).Text()
	if e != nil {
		return Signal{}, errors.New("arrival confirmation unavailable")
	}
	if raw == "GONE" {
		return Signal{}, ErrGone
	}
	if raw == "CONFLICT" {
		return Signal{}, ErrConflict
	}
	return decode(raw)
}

var retractArrival = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 'GONE' end
local s=cjson.decode(raw);if s.expires_at<=now then return 'GONE' end
removeArrival(s)
s._arrival_member=nil;s.arrival_until=nil
if s.state=='here' then s.state='going' end
redis.call('SET',KEYS[1],cjson.encode(s),'KEEPTTL')
return cjson.encode(s)
`)

func (s *Store) RetractArrival(ctx context.Context, hash string) (Signal, error) {
	raw, e := retractArrival.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash}).Text()
	if e != nil {
		return Signal{}, errors.New("arrival retraction unavailable")
	}
	if raw == "GONE" {
		return Signal{}, ErrGone
	}
	return decode(raw)
}

// Stable presence follows the same exact temporary arrival members throughout
// the interval. Retraction/reconfirmation creates a different nonce member and
// cannot repair a broken stability interval by replacing an old claim.
var reconcilePresence = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 0 end
local g=cjson.decode(raw);if g.ends_at<=now then return 0 end
local expired=redis.call('ZRANGEBYSCORE',KEYS[2],'-inf',now,'LIMIT',0,ARGV[3])
for _,member in ipairs(expired) do
 redis.call('ZREM',KEYS[2],member)
 local split=string.find(member,'|',1,true)
 if split then
  local key='gati:s:'..string.sub(member,1,split-1);local value=redis.call('GET',key)
  if value then local s=cjson.decode(value);if s._arrival_member==member then s._arrival_member=nil;s.arrival_until=nil;if s.state=='here' then s.state='going' end;redis.call('SET',key,cjson.encode(s),'KEEPTTL') end end
 end
end
g._presence_threshold=tonumber(ARGV[1])
local count=redis.call('ZCOUNT',KEYS[2],'('..now,'+inf')
local cohort=redis.call('GET',KEYS[3]);local valid=false;local pending=nil
if cohort then
 pending=cjson.decode(cohort);valid=true
 for _,member in ipairs(pending.members) do local expiry=redis.call('ZSCORE',KEYS[2],member);if not expiry or tonumber(expiry)<=now then valid=false;break end end
end
if count<tonumber(ARGV[1]) then
 redis.call('DEL',KEYS[3]);g.state='jemi_gati'
elseif g.state=='jemi_ketu' then
 -- Already achieved. Enough current claims keep it true; loss below threshold
 -- resets it, and a future rise must complete a new stable interval.
elseif valid and now>=pending.ready_at then
 g.state='jemi_ketu';redis.call('DEL',KEYS[3])
elseif not valid then
 local members=redis.call('ZRANGEBYSCORE',KEYS[2],'('..now,'+inf','LIMIT',0,ARGV[1])
 local value=cjson.encode({members=members,ready_at=now+tonumber(ARGV[2])})
 redis.call('SET',KEYS[3],value,'PX',math.min(g.ends_at-now,tonumber(ARGV[2])+tonumber(ARGV[4])))
end
redis.call('SET',KEYS[1],cjson.encode(g),'KEEPTTL')
return 1
`)

func (s *Store) ReconcilePresence(ctx context.Context, id string, threshold, batch int, stabilityMS, recoveryMS int64) error {
	return reconcilePresence.Run(ctx, s.Client, []string{keyPrefix + "gathering:" + id, keyPrefix + "arrivals:" + id, keyPrefix + "presence:" + id}, threshold, stabilityMS, batch, recoveryMS).Err()
}
