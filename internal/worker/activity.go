package worker

import (
	"context"
	"sync"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

// ActivityPublisher has its own lease and loop so capture never holds the matcher
// mutex. Only safe aggregates are staged; the API enforces their release deadline.
type ActivityPublisher struct {
	Store  *store.Store
	Config config.Config
	owner  string
	mu     sync.Mutex
}

func NewActivityPublisher(s *store.Store, c config.Config) *ActivityPublisher {
	return &ActivityPublisher{Store: s, Config: c, owner: randomID()}
}
func (p *ActivityPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.Config.Matching.ReconciliationSeconds) * time.Second)
	defer ticker.Stop()
	for {
		_ = p.Step(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (p *ActivityPublisher) Step(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	now, e := p.Store.Now(ctx)
	if e != nil {
		return e
	}
	_, hash := p.Config.Canonical()
	epoch := now / (int64(p.Config.PublicActivity.ReleaseSeconds) * 1000)
	exists, e := p.Store.HasActivityEpoch(ctx, hash, epoch)
	if e != nil || exists {
		return e
	}
	leader, e := p.Store.ActivityLease(ctx, p.owner, int64(p.Config.PublicActivity.CaptureMaxSeconds+5)*1000)
	if e != nil || !leader {
		return e
	}
	r, data, e := p.Store.CaptureActivity(ctx, p.Config)
	if e != nil {
		return e
	}
	_, e = p.Store.StageActivity(ctx, p.owner, p.Config, r, data)
	return e
}
