package worker

import (
	"context"
	"github.com/jegati/jegati/internal/monitor"
	"time"
)

// RunView refreshes destinations for API-only replicas. It never acquires the
// matching lease or retains a participant working set.
func (e *Engine) RunView(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.Config.Matching.MaximumDebounceSeconds) * time.Second)
	defer ticker.Stop()
	for {
		e.Monitor.Begin(monitor.View)
		now, err := e.Store.Now(ctx)
		if err == nil {
			err = e.Refresh(ctx, now)
		}
		e.Monitor.End(monitor.View, err)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
