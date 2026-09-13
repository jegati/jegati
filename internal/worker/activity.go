package worker

import (
	"context"
	"encoding/json"
	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/monitor"
	"sync"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

// ActivityPublisher has its own lease and loop so capture never holds the matcher
// mutex. Only safe aggregates are staged; the API enforces their release deadline.
type ActivityPublisher struct {
	Monitor *monitor.Registry
	Store   *store.Store
	Config  config.Config
	owner   string
	mu      sync.Mutex
}

func NewActivityPublisher(s *store.Store, c config.Config) *ActivityPublisher {
	return &ActivityPublisher{Store: s, Config: c, owner: randomID()}
}
func (p *ActivityPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.Config.Matching.ReconciliationSeconds) * time.Second)
	defer ticker.Stop()
	firstEpoch := int64(-1)
	for {
		p.Monitor.Begin(monitor.Publisher)
		if p.Monitor != nil {
			// Missing expected publication is distinct from a successful idle tick.
			data, now, err := p.Store.LatestActivity(ctx, p.Config)
			lag := time.Duration(-1)
			if err == nil {
				interval := int64(p.Config.PublicActivity.ReleaseSeconds) * 1000
				if firstEpoch < 0 {
					firstEpoch = now / interval
				}
				expected := (now/interval - int64(p.Config.PublicActivity.DelayEpochs) - 1) * interval
				var release activity.Release
				if len(data) > 0 && json.Unmarshal(data, &release) == nil {
					lag = time.Duration(max(0, expected-release.ObservedFrom)) * time.Millisecond
				} else if expected < firstEpoch*interval {
					lag = 0 // No release from this process lifetime is due yet.
				} else {
					lag = time.Duration(interval) * time.Millisecond
				}
			}
			p.Monitor.Lag(monitor.Publisher, lag)
		}
		err := p.Step(ctx)
		p.Monitor.End(monitor.Publisher, err)
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
