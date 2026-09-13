package notification

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
)

type outageClient struct {
	mu                  sync.Mutex
	active, peak, calls int
}

func (f *outageClient) Do(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	f.active++
	f.calls++
	f.peak = max(f.peak, f.active)
	f.mu.Unlock()
	defer func() { f.mu.Lock(); f.active--; f.mu.Unlock() }()
	select {
	case <-r.Context().Done():
		return nil, r.Context().Err()
	case <-time.After(10 * time.Millisecond):
	}
	return &http.Response{StatusCode: 503, Header: http.Header{"Retry-After": []string{"999999999"}}, Body: io.NopCloser(strings.NewReader("synthetic outage"))}, nil
}
func TestOutageFanoutIsBoundedAcrossCompetingWorkers(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("make test-store")
	}
	s, _, registration := setup(t)
	ctx := context.Background()
	password, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential unavailable")
	}
	backend, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(password)))
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Client.Close()
	s.Store = backend
	client := &outageClient{}
	s.client = client
	grid, _ := geography.NewGrid(100)
	cell := geography.Cell{X: 55, Y: 55}
	proposal := matching.Proposal{Intersection: geography.Intersection{ID: registration.Binding, Point: grid.Center(cell)}}
	for i := 0; i < 100; i++ {
		_, _, r := setup(t)
		hash := r.Binding
		v, e := backend.Create(ctx, hash, grid.ID(cell), 3, 30, 10*time.Minute, 1000)
		if e != nil {
			t.Fatal(e)
		}
		r.ExpiresAt = v.ExpiresAt
		if _, e = s.Register(ctx, hash, r); e != nil {
			t.Fatal(e)
		}
		proposal.Founders = append(proposal.Founders, matching.Signal{Hash: hash, Area: geography.ParticipantArea{Cell: cell, RadiusKM: 3}, CreatedAt: v.CreatedAt, ExpiresAt: v.ExpiresAt})
	}
	// Cancel before closing this owned connection.
	defer func() {
		for _, f := range proposal.Founders {
			backend.Cancel(ctx, f.Hash)
		}
	}()
	id := registration.Binding[:32]
	if _, e = backend.Reserve(ctx, id, "outage-test", proposal, grid, 20, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(30 * time.Millisecond)
	if _, e = backend.Activate(ctx, id, "outage-test", 1000, 600000); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- s.Step(ctx) }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	if client.calls != 100 || client.peak > 2*s.config.Notifications.PushConcurrency {
		t.Fatalf("fanout calls=%d peak=%d", client.calls, client.peak)
	}
	if e = s.Step(ctx); e != nil {
		t.Fatal(e)
	}
	if client.calls != 100 {
		t.Fatal("provider outage caused immediate retry storm")
	}
	for _, f := range proposal.Founders {
		if e = backend.Cancel(ctx, f.Hash); e != nil {
			t.Fatal(e)
		}
	}
	if e = s.Step(ctx); e != nil {
		t.Fatal(e)
	}
	if client.calls != 100 {
		t.Fatal("cancelled subscriptions reached provider")
	}
}
