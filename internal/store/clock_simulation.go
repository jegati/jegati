//go:build simulation

package store

import (
	"context"
	"errors"
	"time"
)

// Separate ACL namespace prevents a simulation build from sharing the production
// store role. Virtual time is frozen between explicit advances. Real native TTLs
// remain a maximum lifetime, independently of paused functional time.
const keyPrefix = "gati-sim:"
const clockLua = `
local value = redis.call('GET', 'gati:simulation:clock')
if not value then
  return redis.error_reply('simulation clock absent')
end
local now = tonumber(value)
`
const SimulationEpoch int64 = 1800000000000

func (s *Store) initializeClock(ctx context.Context) error {
	if err := s.Client.SetNX(ctx, keyPrefix+"simulation:clock", SimulationEpoch, 2*time.Hour).Err(); err != nil {
		return errors.New("simulation requires its dedicated store credential and namespace")
	}
	return nil
}

var advanceClock = newScript(`
__CLOCK__
local delta = tonumber(ARGV[1])
if delta < 0 or delta > 86400000 or now + delta > 1800604800000 then
  return redis.error_reply('invalid advance')
end
now = now + delta
redis.call('SET', KEYS[1], now, 'KEEPTTL')
return now
`)

func (s *Store) AdvanceClock(ctx context.Context, milliseconds int64) (int64, error) {
	if milliseconds < 0 || milliseconds > 86400000 {
		return 0, errors.New("invalid advance")
	}
	return advanceClock.Run(ctx, s.Client, []string{keyPrefix + "simulation:clock"}, milliseconds).Int64()
}
