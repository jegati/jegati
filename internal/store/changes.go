package store

import (
	"context"
	"errors"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
)

// Cell markers contain no participant IDs and expire after ten seconds. The
// finite validated service grid bounds pruning work independently of population.
// The periodic full reconciliation must run at least as often as this lifetime.
const dirtyCellLua = `
local function markCell(cell)
 __CLOCK__
 local key='gati:dirty-cells'
 redis.call('ZREMRANGEBYSCORE',key,'-inf',now-10000)
 redis.call('ZADD',key,now,cell)
 redis.call('PEXPIRE',key,10000)
 redis.call('SET','gati:dirty','1','PX',10000)
end
`

var takeCells = newScript(`
if redis.call('ZCARD',KEYS[1])>tonumber(ARGV[1]) then return redis.error_reply('cell work bound exceeded') end
local cells=redis.call('ZRANGE',KEYS[1],0,-1)
redis.call('DEL',KEYS[1])
return cells
`)

// TakeChangedCells atomically removes the claimed batch BEFORE reading records.
// New changes survive in a new batch. Errors invalidate the worker cache and a
// new lease owner always starts with a full snapshot; no acknowledgement can
// erase a change arriving while the previous batch is processed.
func (s *Store) TakeChangedCells(ctx context.Context, maximum int) ([]string, error) {
	if maximum < 1 {
		return nil, errors.New("invalid cell work bound")
	}
	return takeCells.Run(ctx, s.Client, []string{keyPrefix + "dirty-cells"}, maximum).StringSlice()
}

func (s *Store) CellSnapshot(ctx context.Context, grid geography.Grid, cell string, now int64, batch, maximum int) ([]matching.Signal, error) {
	if _, err := grid.Parse(cell); err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		values, err := s.eligibleSnapshot(ctx, grid, now, batch, maximum, keyPrefix+"cell:"+cell, false)
		if !errors.Is(err, errSnapshotChanged) {
			return values, err
		}
	}
	return nil, errSnapshotChanged
}

var matchingLeaseState = newScript(`
local old=redis.call('GET',KEYS[1]);if old and old~=ARGV[1] then return 0 end
redis.call('SET',KEYS[1],ARGV[1],'PX',ARGV[2])
if old then return 1 end;return 2
`)

func (s *Store) MatchingLeaseState(ctx context.Context, owner string, ttlMS int64) (won, fresh bool, err error) {
	n, e := matchingLeaseState.Run(ctx, s.Client, []string{keyPrefix + "worker:lease"}, owner, ttlMS).Int()
	return n > 0, n == 2, e
}
