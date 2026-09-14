package store

import (
	"context"
	"time"
)

// Bounded oldest-deadline lookup. Return only elapsed time, never index members.
var oldestLag = newScript(`
__CLOCK__
local first = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', now, 'WITHSCORES', 'LIMIT', 0, 1)
if #first == 0 then
  return 0
end
return now - tonumber(first[2])
`)

func (s *Store) PendingLag(ctx context.Context) (time.Duration, error) {
	return s.deadlineLag(ctx, "pending")
}
func (s *Store) CleanupLag(ctx context.Context) (time.Duration, error) {
	return s.deadlineLag(ctx, "expiry")
}
func (s *Store) deadlineLag(ctx context.Context, index string) (time.Duration, error) {
	n, e := oldestLag.Run(ctx, s.Client, []string{keyPrefix + index}).Int64()
	if e != nil {
		return -1, e
	}
	return time.Duration(n) * time.Millisecond, nil
}
