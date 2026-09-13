package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// PushBinding stores encrypted transport material, never a plaintext endpoint or
// application capability. Ciphertext is authenticated with the signal hash as AAD
// by the notification service. This record is also the bounded coalescing outbox.
type PushBinding struct {
	Binding    string `json:"binding"`
	Digest     string `json:"digest"`
	Ciphertext string `json:"ciphertext"`
	ExpiresAt  int64  `json:"expires_at"`
}
type PushJob struct {
	PushBinding
	Claim string `json:"claim"`
	Until int64  `json:"until"`
}
type PushPolicy struct {
	PollMS, QueueMS, GapMS, RetryMS, LeaseMS, MinimumMS int64
	MaxAttempts                                         int
}

var registerPush = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 'GONE' end
local s=cjson.decode(raw);if s.expires_at<=now then return 'GONE' end
local old=redis.call('GET',KEYS[2]);local b=cjson.decode(ARGV[1])
if old then
 local previous=cjson.decode(old)
 if previous.expires_at>now then
  if previous.digest~=b.digest or previous.binding~=b.binding then return 'CONFLICT' end
  return tostring(previous.expires_at)
 end
end
b.expires_at=math.min(s.expires_at,b.expires_at)
if b.expires_at<=now then return 'GONE' end
redis.call('SET',KEYS[2],cjson.encode(b),'PX',b.expires_at-now)
redis.call('ZADD',KEYS[3],now,ARGV[2])
if redis.call('PTTL',KEYS[3])<b.expires_at-now then redis.call('PEXPIRE',KEYS[3],b.expires_at-now) end
return tostring(b.expires_at)
`)

func (s *Store) RegisterPush(ctx context.Context, hash string, binding PushBinding) (int64, error) {
	raw, _ := json.Marshal(binding)
	result, e := registerPush.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "push:" + hash, keyPrefix + "push-due"}, string(raw), hash).Text()
	if e != nil {
		return 0, errors.New("notification registration unavailable")
	}
	if result == "GONE" {
		return 0, ErrGone
	}
	if result == "CONFLICT" {
		return 0, ErrConflict
	}
	return strconv.ParseInt(result, 10, 64)
}

var dropPush = newScript(`redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[1]);return 1`)

func (s *Store) DropPush(ctx context.Context, hash string) error {
	if e := dropPush.Run(ctx, s.Client, []string{keyPrefix + "push:" + hash, keyPrefix + "push-due"}, hash).Err(); e != nil {
		return errors.New("notification removal unavailable")
	}
	return nil
}

// PushDue reads a bounded scheduling index. ClaimPush removes dead members even
// when native key TTL removed their records first. The index also has a hard TTL.
func (s *Store) PushDue(ctx context.Context, maximum int) ([]string, error) {
	now, e := s.Now(ctx)
	if e != nil {
		return nil, e
	}
	return s.Client.ZRangeByScore(ctx, keyPrefix+"push-due", &redis.ZRangeBy{Min: "-inf", Max: strconv.FormatInt(now, 10), Count: int64(maximum)}).Result()
}

var claimPush = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);local signal=redis.call('GET',KEYS[2])
local function drop()
 redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[3],ARGV[1]);return ''
end
if not raw or not signal then return drop() end
local b=cjson.decode(raw);local s=cjson.decode(signal)
if b.expires_at<=now or s.expires_at<=now then return drop() end
local function save(next)
 redis.call('SET',KEYS[1],cjson.encode(b),'KEEPTTL')
 redis.call('ZADD',KEYS[3],math.min(next,b.expires_at),ARGV[1])
end
if (b.lease_until or 0)>now then save(b.lease_until);return '' end
local current='';local deadline=b.expires_at
if s._gathering and (s._gathering_until or 0)>now then
 local event=redis.call('GET','gati:gathering:'..s._gathering)
 if event then
  local g=cjson.decode(event)
  if g.ends_at>now and (s.state=='going' or s.state=='here' or (g.ends_at>=now+tonumber(ARGV[9]) and s.expires_at>=now+tonumber(ARGV[9]))) then
   local state=g.state
   if state=='jemi_ketu' and g._presence_threshold and redis.call('ZCOUNT','gati:arrivals:'..g.id,'('..now,'+inf')<g._presence_threshold then state='jemi_gati' end
   current=g.id..':'..state;deadline=math.min(deadline,g.ends_at)
  end
 end
end
-- Keep only the latest relevant state; never a transition history. This bounded
-- reconciler can coalesce short-lived transitions between checks.
if b.seen~=current then
 b.seen=current;b.pending=nil;b.claim=nil;b.lease_until=nil
 if current~='' then b.pending=current;b.until_at=math.min(deadline,now+tonumber(ARGV[4]));b.attempts=0;b.retry_at=now end
end
if not b.pending or b.until_at<=now or (b.attempts or 0)>=tonumber(ARGV[8]) then
 b.pending=nil;save(now+tonumber(ARGV[3]));return ''
end
if (b.retry_at or 0)>now then save(b.retry_at);return '' end
if b.attempts==0 then
 local gap=tonumber(redis.call('GET',KEYS[4]) or '0')
 if gap>now then save(math.min(gap,now+tonumber(ARGV[3])));return '' end
 redis.call('SET',KEYS[4],now+tonumber(ARGV[5]),'PX',s.expires_at-now)
end
b.attempts=b.attempts+1;b.claim=ARGV[2];b.lease_until=now+tonumber(ARGV[7]);b.retry_at=now+tonumber(ARGV[6])*2^(b.attempts-1)
save(b.lease_until)
return cjson.encode({binding=b.binding,digest=b.digest,ciphertext=b.ciphertext,expires_at=b.expires_at,claim=b.claim,["until"]=math.min(b.until_at,deadline)})
`)

func (s *Store) ClaimPush(ctx context.Context, hash, claim string, p PushPolicy) (*PushJob, error) {
	raw, e := claimPush.Run(ctx, s.Client, []string{keyPrefix + "push:" + hash, keyPrefix + "s:" + hash, keyPrefix + "push-due", keyPrefix + "push-gap:" + hash}, hash, claim, p.PollMS, p.QueueMS, p.GapMS, p.RetryMS, p.LeaseMS, p.MaxAttempts, p.MinimumMS).Text()
	if e != nil {
		return nil, errors.New("notification claim unavailable")
	}
	if raw == "" {
		return nil, nil
	}
	var job PushJob
	if json.Unmarshal([]byte(raw), &job) != nil {
		return nil, errors.New("invalid notification claim")
	}
	return &job, nil
}

var finishPush = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);if not raw then return 0 end
local b=cjson.decode(raw)
if b.expires_at<=now then redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[1]);return 0 end
if b.claim~=ARGV[2] then return 0 end
if ARGV[3]=='gone' then redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[1]);return 1 end
b.claim=nil;b.lease_until=nil
if ARGV[3]=='sent' then b.pending=nil end
redis.call('SET',KEYS[1],cjson.encode(b),'KEEPTTL')
redis.call('ZADD',KEYS[2],math.min(b.expires_at,math.max(now+tonumber(ARGV[4]),b.retry_at or 0)),ARGV[1]);return 1
`)

// Acknowledgement cannot overwrite a replacement binding or another worker claim.
func (s *Store) FinishPush(ctx context.Context, hash, claim, outcome string, pollMS int64) error {
	if outcome != "sent" && outcome != "gone" && outcome != "retry" {
		return errors.New("invalid delivery outcome")
	}
	if e := finishPush.Run(ctx, s.Client, []string{keyPrefix + "push:" + hash, keyPrefix + "push-due"}, hash, claim, outcome, pollMS).Err(); e != nil {
		return errors.New("notification acknowledgement unavailable")
	}
	return nil
}

var pushStatus = newScript(`
__CLOCK__
local raw=redis.call('GET',KEYS[1]);local signal=redis.call('GET',KEYS[2])
if not raw or not signal then return '' end
local b=cjson.decode(raw);local s=cjson.decode(signal)
if b.expires_at<=now or s.expires_at<=now then return '' end
return cjson.encode({binding=b.binding,expires_at=b.expires_at})
`)

func (s *Store) PushStatus(ctx context.Context, hash string) (PushBinding, error) {
	raw, e := pushStatus.Run(ctx, s.Client, []string{keyPrefix + "push:" + hash, keyPrefix + "s:" + hash}).Text()
	if e != nil {
		return PushBinding{}, errors.New("notification status unavailable")
	}
	if raw == "" {
		return PushBinding{}, ErrGone
	}
	var b PushBinding
	e = json.Unmarshal([]byte(raw), &b)
	return b, e
}
