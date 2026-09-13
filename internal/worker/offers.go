package worker

import (
	"context"
	"slices"
	"strings"

	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
)

// offerWaiting gives background sessions the same offers as foreground polling.
// CandidateBatchSize bounds offer attempts per reconciliation; a rotating cursor
// prevents one stable prefix from starving other sessions. The cursor is private,
// process-local and reset on an empty snapshot. Delivery preference is irrelevant.
func (e *Engine) offerWaiting(ctx context.Context, now int64, signals []matching.Signal, open []matching.Gathering) error {
	if len(signals) == 0 {
		e.offerCursor = ""
		return nil
	}
	if len(open) == 0 {
		e.offerCursor = ""
		return nil
	}
	slices.SortFunc(signals, func(a, b matching.Signal) int { return strings.Compare(a.Hash, b.Hash) })
	start, _ := slices.BinarySearchFunc(signals, e.offerCursor, func(a matching.Signal, b string) int { return strings.Compare(a.Hash, b) })
	if start < len(signals) && signals[start].Hash == e.offerCursor {
		start++
	}
	for n := 0; n < min(len(signals), e.Config.Matching.CandidateBatchSize); n++ {
		i := (start + n) % len(signals)
		candidate := signals[i]
		e.offerCursor = candidate.Hash
		e.offerCursorUntil = candidate.ExpiresAt
		if candidate.Assigned || candidate.ReservedUntil > now {
			continue
		}
		g, ok := e.Planner.Offer(now, candidate, open)
		if !ok {
			continue
		}
		value, err := e.Store.SetIntent(ctx, candidate.Hash, g.ID, "offer", e.Planner.Index.Grid.ID(candidate.Area.Cell), candidate.Area.RadiusKM, e.Config.Limits.MaxDeclinesPerSignal, int64(e.Config.Matching.LateJoinMinRemainingMinutes)*60000, int64(e.Config.Matching.InvitationCooldownSeconds)*1000)
		if err == store.ErrGone || err == store.ErrConflict {
			continue
		}
		if err != nil {
			return err
		}
		signals[i].Assigned = value.GatheringUntil > now
	}
	return nil
}
