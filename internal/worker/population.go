package worker

import (
	"context"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
)

// Only the active matching owner keeps this private working set. It is rebuilt
// at least every reconciliation interval (<=10s), discarded on error/ownership
// change, and never exposed to HTTP handlers or operational monitoring.
type population struct {
	cells                map[geography.Cell][]matching.Signal
	fullAt, nextDeadline int64
}

func (e *Engine) populationSnapshot(ctx context.Context, now int64) ([]matching.Signal, error) {
	g := e.Planner.Index.Grid
	changed, err := e.Store.TakeChangedCells(ctx, g.Columns*g.Rows)
	if err != nil {
		return nil, err
	}
	if e.population == nil || now-e.population.fullAt >= int64(e.Config.Matching.ReconciliationSeconds)*1000 {
		values, err := e.Store.EligibleSnapshot(ctx, g, now, e.Config.Limits.CleanupBatchSize, e.Config.Limits.MaxActiveSignals)
		if err != nil {
			return nil, err
		}
		p := &population{cells: map[geography.Cell][]matching.Signal{}, fullAt: now}
		for _, v := range values {
			p.cells[v.Area.Cell] = append(p.cells[v.Area.Cell], v)
		}
		e.population = p
	} else {
		for _, id := range changed {
			cell, err := g.Parse(id)
			if err != nil {
				return nil, err
			}
			values, err := e.Store.CellSnapshot(ctx, g, id, now, e.Config.Limits.CleanupBatchSize, e.Config.Limits.MaxActiveSignals)
			if err != nil {
				return nil, err
			}
			e.population.cells[cell] = values
		}
	}
	count := 0
	for _, values := range e.population.cells {
		count += len(values)
	}
	if count > e.Config.Limits.MaxActiveSignals {
		return nil, store.ErrCapacity
	}
	return e.population.live(now, int64(e.Config.Matching.MinimumRemainingMinutes)*60000), nil
}

func (p *population) live(now, minimumMS int64) []matching.Signal {
	count := 0
	for _, values := range p.cells {
		count += len(values)
	}
	result := make([]matching.Signal, 0, count)
	p.nextDeadline = 0
	for cell, values := range p.cells {
		live := values[:0]
		for _, v := range values {
			if v.ExpiresAt <= now {
				continue
			}
			v.Assigned = v.AssignedUntil > now
			for _, deadline := range []int64{v.ExpiresAt, v.ExpiresAt - minimumMS + 1, v.ReservedUntil, v.AssignedUntil} {
				if deadline > now && (p.nextDeadline == 0 || deadline < p.nextDeadline) {
					p.nextDeadline = deadline
				}
			}
			live = append(live, v)
			result = append(result, v)
		}
		// Remove references in the unused capacity as well as logical membership.
		clear(values[len(live):])
		if len(live) == 0 {
			delete(p.cells, cell)
		} else {
			p.cells[cell] = live
		}
	}
	return result
}
