package store

import (
	"context"
	"encoding/json"
	"errors"
)

var intent = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);local event=redis.call('GET',KEYS[2])
if not raw or not event then return 'GONE' end
local s=cjson.decode(raw);local g=cjson.decode(event);local minimum=now+tonumber(ARGV[2])
if s.expires_at<=now or g.ends_at<=now then return 'GONE' end
if s.cell~=ARGV[3] or s.radius_km~=tonumber(ARGV[4]) then return 'CONFLICT' end
s._declined=s._declined or {}
if ARGV[1]=='decline' then
 if s._declined[g.id] then return cjson.encode(s) end
 if s._gathering~=g.id then return 'CONFLICT' end
 local count=0;for _,v in pairs(s._declined) do count=count+1 end
 if count>=tonumber(ARGV[5]) then return 'CONFLICT' end
 removeArrival(s)
 s._arrival_member=nil;s.arrival_until=nil
 s._declined[g.id]=true;s._gathering=nil;s._gathering_until=nil;s.state='gati';s._invite_after=now+tonumber(ARGV[6])
else
 if ARGV[1]=='going' and (s.state=='going' or s.state=='here') and s._gathering==g.id then return cjson.encode(s) end
 if s.expires_at<minimum or g.ends_at<minimum or s._declined[g.id] then return 'GONE' end
 if ARGV[1]=='offer' and s._gathering and s._gathering~=g.id and (s._gathering_until or 0)>now then return 'CONFLICT' end
 if ARGV[1]=='offer' then
  if (s._invite_after or 0)>now then return 'GONE' end
  local count=0;for _,v in pairs(s._declined) do count=count+1 end
  if count>=tonumber(ARGV[5]) then return 'GONE' end
  if s._gathering~=g.id or (s._gathering_until or 0)<=now or (s.state~='going' and s.state~='here') then s.state='invited' end
 else
  if s._gathering~=g.id and s._arrival_member then removeArrival(s);s._arrival_member=nil;s.arrival_until=nil end
  s.state='going'
 end
 s._gathering=g.id;s._gathering_until=g.ends_at;s._pending=nil;s._pending_until=nil
end
redis.call('SET',KEYS[1],cjson.encode(s),'KEEPTTL')
markCell(s.cell)
return cjson.encode(s)
`)

// SetIntent requires geographic reachability checked against the immutable claim
// and the frozen gathering. Lua revalidates both claims, cutoffs and live state.
func (s *Store) SetIntent(ctx context.Context, hash, gathering, action, cell string, radius float64, maxDeclines int, minimumMS, cooldownMS int64) (Signal, error) {
	if action != "offer" && action != "going" && action != "decline" {
		return Signal{}, errors.New("invalid intent")
	}
	raw, e := intent.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "gathering:" + gathering}, action, minimumMS, cell, radius, maxDeclines, cooldownMS).Text()
	if e != nil {
		return Signal{}, errors.New("intent unavailable")
	}
	if raw == "GONE" {
		return Signal{}, ErrGone
	}
	if raw == "CONFLICT" {
		return Signal{}, ErrConflict
	}
	var value Signal
	e = json.Unmarshal([]byte(raw), &value)
	return value, e
}

var createAndJoin = newScript(`
__CLOCK__
local event=redis.call('GET',KEYS[5]);if not event then return 'GONE' end
local g=cjson.decode(event)
if g.ends_at<=now or g.intersection.id~=ARGV[9] then return 'GONE' end
local previous=redis.call('GET',KEYS[1])
if previous then
 local v=cjson.decode(previous)
 if v._declined and v._declined[g.id] then return 'GONE' end
 if (v.state=='going' or v.state=='here') and v._gathering==g.id and v.expires_at>now and v.cell==ARGV[1] and v.radius_km==tonumber(ARGV[2]) and v.availability_minutes==tonumber(ARGV[3]) then return previous end
end
if g.ends_at<now+tonumber(ARGV[8]) then return 'GONE' end
local function createSignal()
` + createLua + `
end
local raw=createSignal()
if raw=='GONE' or raw=='CONFLICT' or raw=='CAPACITY' then return raw end
local s=cjson.decode(raw)
if s.expires_at<now+tonumber(ARGV[8]) then return 'GONE' end
if s._gathering~=g.id and s._arrival_member then removeArrival(s);s._arrival_member=nil;s.arrival_until=nil end
s._gathering=g.id;s._gathering_until=g.ends_at;s._pending=nil;s._pending_until=nil;s.state='going'
redis.call('SET',KEYS[1],cjson.encode(s),'KEEPTTL')
markCell(s.cell)
return cjson.encode(s)
`)

func (s *Store) CreateAndJoin(ctx context.Context, hash, cell string, radius float64, minutes, capacity int, g Gathering, minimumMS int64) (Signal, error) {
	raw, e := createAndJoin.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "cap:" + hash, keyPrefix + "expiry", keyPrefix + "cell:" + cell, keyPrefix + "gathering:" + g.ID}, cell, radius, minutes, int64(minutes)*60000, capacity, hash+"|"+cell, hash, minimumMS, g.Intersection.ID).Text()
	if e != nil {
		return Signal{}, errors.New("joining unavailable")
	}
	switch raw {
	case "GONE":
		return Signal{}, ErrGone
	case "CONFLICT":
		return Signal{}, ErrConflict
	case "CAPACITY":
		return Signal{}, ErrCapacity
	}
	return decode(raw)
}
