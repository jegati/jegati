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
	ticker := time.NewTicker(time.Duration(min(p.Config.Matching.ReconciliationSeconds, max(1, p.Config.PublicActivity.ReleaseSeconds/10))) * time.Second)
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
				lag = publicationLag(now, firstEpoch, interval, int64(p.Config.PublicActivity.DelayEpochs), data)
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
	interval := int64(p.Config.PublicActivity.ReleaseSeconds) * 1000
	// Leave the full capture budget before the fixed publication boundary.
	// Never start a capture that is guaranteed to miss its zero-delay release.
	if p.Config.PublicActivity.DelayEpochs == 0 && interval-now%interval <= int64(p.Config.PublicActivity.CaptureMaxSeconds)*1000 {
		return nil
	}
	epoch := now / interval
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

// Missing publication must accumulate lag even when intervals are shorter than
// the alert threshold; a single interval would hide a permanently empty map.
func publicationLag(now, firstEpoch, interval, delay int64, data []byte) time.Duration {
	expected := (now/interval - delay - 1) * interval
	var release activity.Release
	if len(data) > 0 && json.Unmarshal(data, &release) == nil {
		return time.Duration(max(0, expected-release.ObservedFrom)) * time.Millisecond
	}
	if expected < firstEpoch*interval {
		return 0
	}
	return time.Duration(expected-firstEpoch*interval+interval) * time.Millisecond
}
