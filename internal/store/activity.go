package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/redis/go-redis/v9"
)

// CaptureActivity observes all states over a bounded interval, not an atomic point
// in time. ZSCAN includes continuously present index members; churn may be sampled
// on either side of a change. Each credential is read once and all its metrics use
// that same record. No capture state is persisted. Expired rows are filtered at end.
func (s *Store) CaptureActivity(ctx context.Context, c config.Config) (activity.Release, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.PublicActivity.CaptureMaxSeconds)*time.Second)
	defer cancel()
	start, err := s.Now(ctx)
	if err != nil {
		return activity.Release{}, nil, err
	}
	records := []activity.Participant{}
	seen := map[string]bool{}
	cursor := uint64(0)
	work := 0
	batch := c.Limits.CleanupBatchSize
	maximum := c.Limits.MaxActiveSignals
	for {
		entries, next, e := s.Client.ZScan(ctx, keyPrefix+"expiry", cursor, "*", int64(batch)).Result()
		if e != nil {
			return activity.Release{}, nil, e
		}
		work += len(entries) / 2
		if work > maximum*4+batch {
			return activity.Release{}, nil, errors.New("activity capture work exceeded")
		}
		hashes := []string{}
		for i := 0; i < len(entries); i += 2 {
			hash, _, ok := strings.Cut(entries[i], "|")
			if !ok {
				return activity.Release{}, nil, errors.New("invalid activity index")
			}
			if seen[hash] {
				continue
			}
			seen[hash] = true
			hashes = append(hashes, hash)
			if len(seen) > maximum {
				return activity.Release{}, nil, ErrCapacity
			}
		}
		// SCAN's COUNT is a hint; bound actual GET pipeline sizes separately.
		for offset := 0; offset < len(hashes); offset += batch {
			pipe := s.Client.Pipeline()
			commands := []*redis.StringCmd{}
			end := min(offset+batch, len(hashes))
			for _, hash := range hashes[offset:end] {
				commands = append(commands, pipe.Get(ctx, keyPrefix+"s:"+hash))
			}
			_, e = pipe.Exec(ctx)
			if e != nil && e != redis.Nil {
				return activity.Release{}, nil, e
			}
			for n, cmd := range commands {
				raw, e := cmd.Result()
				if e == redis.Nil {
					continue
				}
				if e != nil {
					return activity.Release{}, nil, e
				}
				v, e := decode(raw)
				if e != nil {
					return activity.Release{}, nil, e
				}
				records = append(records, activity.Participant{Hash: hashes[offset+n], Cell: v.Cell, Gathering: v.Gathering, State: v.State, ExpiresAt: v.ExpiresAt, GatheringUntil: v.GatheringUntil, ArrivalUntil: v.ArrivalUntil})
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	now, e := s.Now(ctx)
	if e != nil {
		return activity.Release{}, nil, e
	}
	// Request one sentinel beyond capacity; truncation is never a complete release.
	groups, e := s.OpenGatherings(ctx, now, maximum+1)
	if e != nil {
		return activity.Release{}, nil, e
	}
	if len(groups) > maximum {
		return activity.Release{}, nil, ErrCapacity
	}
	grid, _ := geography.NewGrid(c.PublicActivity.AreaSizeMeters)
	events := []activity.Gathering{}
	for _, g := range groups {
		cell, e := grid.CellAt(g.Intersection.Point)
		if e != nil {
			return activity.Release{}, nil, e
		}
		events = append(events, activity.Gathering{ID: g.ID, Cell: grid.ID(cell), State: g.State, EndsAt: g.EndsAt})
	}
	end, e := s.Now(ctx)
	if e != nil {
		return activity.Release{}, nil, e
	}
	return activity.Build(c, start, end, records, events)
}

func activityKey(hash string, epoch int64) string {
	return keyPrefix + "activity:" + hash + ":" + strconv.FormatInt(epoch, 10)
}
func (s *Store) ActivityLease(ctx context.Context, owner string, ttlMS int64) (bool, error) {
	n, e := lease.Run(ctx, s.Client, []string{keyPrefix + "activity:lease"}, owner, ttlMS).Int()
	return n == 1, e
}
func (s *Store) HasActivityEpoch(ctx context.Context, hash string, epoch int64) (bool, error) {
	n, e := s.Client.Exists(ctx, activityKey(hash, epoch)).Result()
	return n != 0, e
}

var stageActivity = newScript(`
__CLOCK__
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
  return 0
end
local expires = tonumber(ARGV[3])
local release = tonumber(ARGV[4])
if now >= release or expires <= release then
  return 0
end
local result = redis.call('SET', KEYS[2], ARGV[2], 'PX', expires - now, 'NX')
if result then
  return 1
end
return 0
`)

func (s *Store) StageActivity(ctx context.Context, owner string, c config.Config, r activity.Release, data []byte) (bool, error) {
	// Validate bytes against the typed cell-only contract before they enter the public store.
	expected, e := json.Marshal(r)
	if e != nil || string(expected) != string(data) || len(data) > c.PublicActivity.MaxSnapshotBytes {
		return false, errors.New("invalid activity document")
	}
	_, hash := c.Canonical()
	if r.ConfigHash != hash {
		return false, errors.New("activity configuration mismatch")
	}
	epoch := r.ObservedFrom / (int64(c.PublicActivity.ReleaseSeconds) * 1000)
	n, e := stageActivity.Run(ctx, s.Client, []string{keyPrefix + "activity:lease", activityKey(hash, epoch)}, owner, string(data), r.ExpiresAt, r.ReleaseAt).Int()
	return n == 1, e
}

var latestActivity = newScript(`
__CLOCK__
for _, key in ipairs(KEYS) do
  local raw = redis.call('GET', key)
  if raw then
    local v = cjson.decode(raw)
    if v.release_at <= now and v.expires_at > now then
      return { now, raw }
    end
  end
end
return { now, '' }
`)

// LatestActivity reads only bounded precomputed aggregate documents, never signals.
func (s *Store) LatestActivity(ctx context.Context, c config.Config) ([]byte, int64, error) {
	now, e := s.Now(ctx)
	if e != nil {
		return nil, 0, e
	}
	_, hash := c.Canonical()
	interval := int64(c.PublicActivity.ReleaseSeconds) * 1000
	current := now/interval - int64(c.PublicActivity.DelayEpochs) - 1
	keys := []string{}
	for i := int64(0); i <= int64(c.PublicActivity.SnapshotRetentionMinutes)*60000/interval; i++ {
		keys = append(keys, activityKey(hash, current-i))
	}
	values, e := latestActivity.Run(ctx, s.Client, keys).Slice()
	if e != nil {
		return nil, 0, e
	}
	return []byte(values[1].(string)), values[0].(int64), nil
}
