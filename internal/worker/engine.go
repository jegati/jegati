// Package worker runs the single-codebase private matcher. API replicas can each
// run an engine; a short store lease elects the writer, and transactions recheck
// conflicts even if a slow worker outlives its lease. Cached views are destinations,
// never member lists or individual signals.
package worker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/monitor"
	"github.com/jegati/jegati/internal/notification"
	"github.com/jegati/jegati/internal/store"
	"sync"
	"time"
)

type Engine struct {
	Monitor          *monitor.Registry
	Push             *notification.Service
	Store            *store.Store
	Planner          matching.Planner
	Config           config.Config
	Hash, owner      string
	step             sync.Mutex
	mu               sync.RWMutex
	open             []store.Gathering
	lastSweep        int64
	offerCursor      string
	offerCursorUntil int64
}

func randomID() string {
	raw := make([]byte, 16)
	if _, e := rand.Read(raw); e != nil {
		panic("operating system randomness unavailable")
	}
	return hex.EncodeToString(raw)
}
func New(s *store.Store, index *geography.Index, c config.Config) *Engine {
	_, hash := c.Canonical()
	return &Engine{Store: s, Planner: matching.Planner{Index: index, Config: c.Matching}, Config: c, Hash: hash, owner: randomID()}
}
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.Config.Matching.MaximumDebounceSeconds) * time.Second)
	defer ticker.Stop()
	for {
		e.Monitor.Begin(monitor.Matcher)
		if e.Monitor != nil {
			lag, _ := e.Store.PendingLag(ctx)
			e.Monitor.Lag(monitor.Matcher, lag)
		}
		err := e.Step(ctx)
		e.Monitor.End(monitor.Matcher, err)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (e *Engine) Open() []store.Gathering {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]store.Gathering(nil), e.open...)
}
func (e *Engine) Refresh(ctx context.Context, now int64) error {
	open, err := e.Store.OpenGatherings(ctx, now, e.Config.Limits.MaxActiveSignals)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.open = open
	e.mu.Unlock()
	return nil
}
func (e *Engine) Step(ctx context.Context) error {
	e.step.Lock()
	defer e.step.Unlock()
	now, err := e.Store.Now(ctx)
	if err != nil {
		e.offerCursor = ""
		e.offerCursorUntil = 0
		return err
	}
	if now >= e.offerCursorUntil {
		e.offerCursor = ""
		e.offerCursorUntil = 0
	}
	if err = e.Refresh(ctx, now); err != nil {
		return err
	}
	leader, err := e.Store.MatchingLease(ctx, e.owner, int64(e.Config.Matching.ReconciliationSeconds)*1000)
	if err != nil || !leader {
		return err
	}
	for _, g := range e.Open() {
		if err = e.Store.ReconcilePresence(ctx, g.ID, e.Config.Arrivals.ConfirmationCount, e.Config.Limits.CleanupBatchSize, int64(e.Config.Arrivals.ConfirmationStabilitySeconds)*1000, int64(e.Config.Matching.ReconciliationSeconds)*2000); err != nil {
			return err
		}
	}
	if err = e.Refresh(ctx, now); err != nil {
		return err
	}
	due, err := e.Store.DueReservations(ctx, now, e.Config.Matching.CandidateBatchSize)
	if err != nil {
		return err
	}
	for _, id := range due {
		_, err = e.Store.Activate(ctx, id, e.Hash, int64(e.Config.Matching.MinimumRemainingMinutes)*60000, int64(e.Config.Matching.MaximumGatheringMinutes)*60000)
		if err != nil && err != store.ErrGone && err != store.ErrWait {
			return err
		}
	}
	if len(due) > 0 {
		if err = e.Refresh(ctx, now); err != nil {
			return err
		}
	}
	dirty, err := e.Store.ConsumeDirty(ctx)
	if err != nil {
		return err
	}
	if !dirty && len(due) == 0 && now-e.lastSweep < int64(e.Config.Matching.ReconciliationSeconds)*1000 {
		return nil
	}
	e.lastSweep = now
	signals, err := e.Store.EligibleSnapshot(ctx, e.Planner.Index.Grid, now, e.Config.Limits.CleanupBatchSize, e.Config.Limits.MaxActiveSignals)
	if err != nil {
		return err
	}
	open := []matching.Gathering{}
	for _, g := range e.Open() {
		open = append(open, matching.Gathering{ID: g.ID, Intersection: g.Intersection, EndsAt: g.EndsAt})
	}
	if err = e.offerWaiting(ctx, now, signals, open); err != nil {
		return err
	}
	pending, err := e.Store.PendingReservations(ctx, e.Config.Limits.MaxActiveSignals)
	if err != nil {
		return err
	}
	// Reserve prospective destinations so waiting compatible users do not found
	// competing gatherings while the original cohort's stability timer is running.
	for _, p := range pending {
		if p.ExpiresAt > now {
			open = append(open, matching.Gathering{ID: p.ID, Intersection: p.Intersection, EndsAt: p.ReadyAt + int64(e.Config.Matching.MinimumRemainingMinutes)*60000})
		}
	}
	for n := 0; n < e.Config.Matching.CandidateBatchSize; n++ {
		proposal, ok := e.Planner.Propose(now, signals, open)
		if !ok {
			break
		}
		id := randomID()
		reservation, err := e.Store.Reserve(ctx, id, e.Hash, proposal, e.Planner.Index.Grid, int64(e.Config.Matching.ActivationStabilitySeconds)*1000, int64(e.Config.Matching.ReconciliationSeconds)*2000, int64(e.Config.Matching.MinimumRemainingMinutes)*60000)
		if err == store.ErrConflict {
			break
		}
		if err != nil {
			return err
		}
		open = append(open, matching.Gathering{ID: id, Intersection: proposal.Intersection, EndsAt: reservation.ReadyAt + int64(e.Config.Matching.MinimumRemainingMinutes)*60000})
	}
	return nil
}
