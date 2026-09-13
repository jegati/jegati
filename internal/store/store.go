// Package store owns bounded ephemeral state in Valkey. It never logs requests.
package store

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type silentLogger struct{}

func (silentLogger) Printf(context.Context, string, ...interface{}) {}
func init()                                                         { redis.SetLogger(silentLogger{}) }

var ErrGone = errors.New("session expired or absent")
var ErrConflict = errors.New("capability already used")
var ErrCapacity = errors.New("admission capacity reached")

type Signal struct {
	ArrivalUntil        int64           `json:"arrival_until,omitempty"`
	ArrivalMember       string          `json:"_arrival_member,omitempty"`
	InviteAfter         int64           `json:"_invite_after,omitempty"`
	Pending             string          `json:"_pending,omitempty"`
	PendingUntil        int64           `json:"_pending_until,omitempty"`
	Gathering           string          `json:"_gathering,omitempty"`
	GatheringUntil      int64           `json:"_gathering_until,omitempty"`
	Declined            map[string]bool `json:"_declined,omitempty"`
	Cell                string          `json:"cell"`
	RadiusKM            float64         `json:"radius_km"`
	AvailabilityMinutes int             `json:"availability_minutes"`
	CreatedAt           int64           `json:"created_at"`
	ExpiresAt           int64           `json:"expires_at"`
	State               string          `json:"state"`
}
type Store struct{ Client *redis.Client }

func Connect(ctx context.Context, address, username, password string) (*Store, error) {

	c := redis.NewClient(&redis.Options{Addr: address, Username: username, Password: password, Protocol: 2, DisableIdentity: true, PoolSize: 32, MaxActiveConns: 32, PoolTimeout: time.Second, MaxRetries: -1, DialTimeout: 2 * time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second, ContextTimeoutEnabled: true})
	if err := c.Ping(ctx).Err(); err != nil {
		c.Close()
		return nil, errors.New("temporary store unavailable")
	}
	backend := &Store{c}
	if err := backend.initializeClock(ctx); err != nil {
		c.Close()
		return nil, err
	}
	return backend, nil
}
func Hash(capability []byte) string {
	sum := sha256.Sum256(capability)
	return hex.EncodeToString(sum[:])
}
func decode(raw string) (Signal, error) {
	var s Signal
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, errors.New("invalid stored session")
	}
	return s, nil
}

const createLua = `
__CLOCK__
local old=redis.call('GET',KEYS[1])
if old then
 local s=cjson.decode(old)
 if s.expires_at<=now then return 'GONE' end
 if s.cell==ARGV[1] and s.radius_km==tonumber(ARGV[2]) and s.availability_minutes==tonumber(ARGV[3]) then return old end
 return 'CONFLICT'
end
if redis.call('EXISTS',KEYS[2])==1 then return 'GONE' end
if redis.call('ZCARD',KEYS[3])>=tonumber(ARGV[5]) then return 'CAPACITY' end
local expiry=now+tonumber(ARGV[4])
local s=cjson.encode({cell=ARGV[1],radius_km=tonumber(ARGV[2]),availability_minutes=tonumber(ARGV[3]),created_at=now,expires_at=expiry,state='gati'})
redis.call('SET',KEYS[1],s,'PX',ARGV[4])
redis.call('SET',KEYS[2],'used','PX',ARGV[4])
redis.call('ZADD',KEYS[3],expiry,ARGV[6])
redis.call('ZADD',KEYS[4],expiry,ARGV[7])
for n=3,4 do if redis.call('PTTL',KEYS[n])<tonumber(ARGV[4]) then redis.call('PEXPIRE',KEYS[n],ARGV[4]) end end
redis.call('SET','gati:dirty','1','PX',10000)
return s
`

var create = newScript(createLua)

func (s *Store) Create(ctx context.Context, hash, cell string, radius float64, minutes int, ttl time.Duration, capacity int) (Signal, error) {
	raw, err := create.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "cap:" + hash, keyPrefix + "expiry", keyPrefix + "cell:" + cell}, cell, radius, minutes, ttl.Milliseconds(), capacity, hash+"|"+cell, hash).Text()
	if err != nil {
		return Signal{}, errors.New("temporary store write failed")
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

var status = newScript(`
local raw=redis.call('GET',KEYS[1]);if not raw then return 'GONE' end
__CLOCK__
local s=cjson.decode(raw)
if s.expires_at<=now then return 'GONE' end
if s.arrival_until and s.arrival_until<=now then removeArrival(s);s.arrival_until=nil;s._arrival_member=nil;if s.state=='here' then s.state='going' end;raw=cjson.encode(s);redis.call('SET',KEYS[1],raw,'KEEPTTL') end
return raw
`)

func (s *Store) Status(ctx context.Context, hash string) (Signal, error) {
	raw, err := status.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash}).Text()
	if err != nil {
		return Signal{}, errors.New("temporary store read failed")
	}
	if raw == "GONE" {
		return Signal{}, ErrGone
	}
	return decode(raw)
}

var cancel = newScript(`
local raw=redis.call('GET',KEYS[1]);if not raw then return 0 end
local s=cjson.decode(raw)
removeArrival(s)
redis.call('DEL','gati:arrival-nonce:'..ARGV[1])
redis.call('DEL',KEYS[1])
redis.call('ZREM',KEYS[2],ARGV[1]..'|'..s.cell)
redis.call('ZREM','gati:cell:'..s.cell,ARGV[1])
redis.call('SET','gati:dirty','1','PX',10000)
return 1
`)

func (s *Store) Cancel(ctx context.Context, hash string) error {
	if err := cancel.Run(ctx, s.Client, []string{keyPrefix + "s:" + hash, keyPrefix + "expiry"}, hash).Err(); err != nil {
		return errors.New("temporary store cancel failed")
	}
	return nil
}

var cleanup = newScript(`
__CLOCK__
local expired=redis.call('ZRANGEBYSCORE',KEYS[1],'-inf',now,'LIMIT',0,ARGV[1])
for _,member in ipairs(expired) do
 local split=string.find(member,'|',1,true)
 if split then
  local hash=string.sub(member,1,split-1);local cell=string.sub(member,split+1)
  redis.call('ZREM','gati:cell:'..cell,hash)
 end
 redis.call('ZREM',KEYS[1],member)
end
return #expired
`)

func (s *Store) Cleanup(ctx context.Context, batch int) (int, error) {
	return cleanup.Run(ctx, s.Client, []string{keyPrefix + "expiry"}, batch).Int()
}

var limit = newScript(`
local n=redis.call('INCR',KEYS[1]);if n==1 then redis.call('PEXPIRE',KEYS[1],ARGV[2]) end
if n>tonumber(ARGV[1]) then return 0 end
return 1
`)

// NetworkKey uses a shared, expiring rotation secret. No address is stored and the
// resulting key is never linked to an individual session record. It remains a
// sensitive short-lived pseudonym, not anonymized data.
func (s *Store) NetworkKey(ctx context.Context, address string) (string, error) {
	epoch := time.Now().Unix() / 600
	key := keyPrefix + "rate:secret:" + strconv.FormatInt(epoch, 10)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	ttl := time.Until(time.Unix((epoch+1)*600, 0))
	if ttl <= 0 {
		ttl = time.Millisecond
	}
	if err := s.Client.SetNX(ctx, key, b, ttl).Err(); err != nil {
		return "", errors.New("limiter unavailable")
	}
	secret, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		return "", errors.New("limiter unavailable")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(address))
	return strconv.FormatInt(epoch, 10) + ":" + hex.EncodeToString(mac.Sum(nil)), nil
}
func (s *Store) Allow(ctx context.Context, key string, count int, ttl time.Duration) (bool, error) {
	v, err := limit.Run(ctx, s.Client, []string{keyPrefix + "rate:" + key}, count, ttl.Milliseconds()).Int()
	return v == 1, err
}

func newScript(source string) *redis.Script {
	source = removeArrivalLua + source
	source = strings.ReplaceAll(source, "__CLOCK__", clockLua)
	return redis.NewScript(strings.ReplaceAll(source, "gati:", keyPrefix))
}

// Public removes private matching bookkeeping from the own-session API contract.
func (s Signal) Public() Signal {
	s.ArrivalMember = ""
	s.InviteAfter = 0
	s.Pending = ""
	s.PendingUntil = 0
	s.Gathering = ""
	s.GatheringUntil = 0
	s.Declined = nil
	return s
}

// Removal updates the threshold immediately; a replacement claim cannot hide a
// below-threshold interval between worker ticks. No public counts are returned.
const removeArrivalLua = `
local function removeArrival(s)
 if not s._arrival_member or not s._gathering then return end
 __CLOCK__
 local index='gati:arrivals:'..s._gathering
 redis.call('ZREM',index,s._arrival_member)
 local key='gati:gathering:'..s._gathering;local raw=redis.call('GET',key)
 if raw then
  local g=cjson.decode(raw)
  if g._presence_threshold and redis.call('ZCOUNT',index,'('..now,'+inf')<g._presence_threshold then
   g.state='jemi_gati';redis.call('SET',key,cjson.encode(g),'KEEPTTL');redis.call('DEL','gati:presence:'..g.id)
  end
 end
end
`
