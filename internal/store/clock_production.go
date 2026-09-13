//go:build !simulation

package store

import "context"

const keyPrefix = "gati:"
const clockLua = `local clock=redis.call('TIME'); local now=tonumber(clock[1])*1000+math.floor(tonumber(clock[2])/1000)`

func (s *Store) initializeClock(context.Context) error { return nil }
