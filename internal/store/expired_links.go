package store

import "context"

// Reuse the existing bounded matching snapshots to discover stale associations.
// The transaction rereads current state, so a concurrent new assignment survives.
// There is no extra participant index, persistent scan cursor or location history.
const expiredGatheringLua = `
local function clearExpiredGathering(key, s, now)
  if s.expires_at <= now or not s._gathering or (s._gathering_until or 0) > now then
    return false
  end
  removeArrival(s)
  s._gathering = nil
  s._gathering_until = nil
  s._arrival_member = nil
  s.arrival_until = nil
  s.state = 'gati'
  redis.call('DEL', 'gati:arrival-nonce:' .. string.sub(key, string.len('gati:s:') + 1))
  redis.call('SET', key, cjson.encode(s), 'KEEPTTL')
  markCell(s.cell)
  return true
end
`

var clearExpiredLinks = newScript(expiredGatheringLua + `
__CLOCK__
local changed = 0
for _, key in ipairs(KEYS) do
  local raw = redis.call('GET', key)
  if raw and clearExpiredGathering(key, cjson.decode(raw), now) then
    changed = changed + 1
  end
end
return changed
`)

func (s *Store) clearExpiredGatheringLinks(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return clearExpiredLinks.Run(ctx, s.Client, keys).Err()
}
