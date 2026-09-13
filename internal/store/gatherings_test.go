package store

import (
	"context"
	"encoding/json"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"strings"
	"sync"
	"testing"
	"time"
)

func proposalTest(t *testing.T, s *Store) (matching.Proposal, geography.Grid) {
	t.Helper()
	g, _ := geography.NewGrid(1000)
	cell := geography.Cell{X: 5, Y: 5}
	p := matching.Proposal{Crossing: geography.Crossing{ID: fresh(), Point: g.Center(cell)}}
	for n := 0; n < 3; n++ {
		hash := fresh()
		value, e := s.Create(context.Background(), hash, g.ID(cell), 3, 30, time.Minute, 1000)
		if e != nil {
			t.Fatal(e)
		}
		p.Founders = append(p.Founders, matching.Signal{Hash: hash, Area: geography.ParticipantArea{Cell: cell, RadiusKM: 3}, CreatedAt: value.CreatedAt, ExpiresAt: value.ExpiresAt})
		t.Cleanup(func() { s.Cancel(context.Background(), hash) })
	}
	return p, g
}
func TestAtomicReservationActivationAndPrivateFields(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	id := fresh()
	r, e := s.Reserve(ctx, id, "test-config", p, grid, 100, 10000, 1000)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Activate(ctx, id, "test-config", 1000, 60000); e != ErrWait {
		t.Fatalf("activated before stability: %v", e)
	}
	value, e := s.Status(ctx, p.Founders[0].Hash)
	if e != nil {
		t.Fatal(e)
	}
	public, _ := json.Marshal(value.Public())
	if strings.Contains(string(public), "_pending") || strings.Contains(string(public), id) {
		t.Fatal("reservation exposed to client")
	}
	if _, e = s.Reserve(ctx, fresh(), "test-config", p, grid, 100, 10000, 1000); e != ErrConflict {
		t.Fatal("duplicate reservation succeeded")
	}
	time.Sleep(120 * time.Millisecond)
	g, e := s.Activate(ctx, id, "test-config", 1000, 60000)
	if e != nil {
		t.Fatal(e)
	}
	if g.ActivatedAt < r.ReadyAt || g.EndsAt != p.Founders[0].ExpiresAt || g.Crossing.ID != p.Crossing.ID {
		t.Fatal("activation changed deadline/destination or ignored stability")
	}
	again, e := s.Activate(ctx, id, "test-config", 1000, 60000)
	if e != nil || again.EndsAt != g.EndsAt || again.ActivatedAt != g.ActivatedAt {
		t.Fatal("activation replay renewed gathering")
	}
}
func TestCancellationCannotBeReplacedInsideStableCohort(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	id := fresh()
	if _, e := s.Reserve(ctx, id, "config", p, grid, 50, 10000, 1000); e != nil {
		t.Fatal(e)
	}
	if e := s.Cancel(ctx, p.Founders[0].Hash); e != nil {
		t.Fatal(e)
	}
	// Another signal in the same cell does not repair uninterrupted eligibility.
	hash := fresh()
	s.Create(ctx, hash, grid.ID(p.Founders[0].Area.Cell), 3, 30, time.Minute, 1000)
	defer s.Cancel(ctx, hash)
	time.Sleep(70 * time.Millisecond)
	if _, e := s.Activate(ctx, id, "config", 1000, 60000); e != ErrGone {
		t.Fatalf("lost founder did not reset: %v", e)
	}
	value, e := s.Status(ctx, p.Founders[1].Hash)
	if e != nil || value.Pending != "" {
		t.Fatal("failed cohort left surviving founder locked")
	}
}
func TestConcurrentReservationsAndStaleArea(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	changed := p
	changed.Founders = append([]matching.Signal(nil), p.Founders...)
	changed.Founders[0].Area.RadiusKM = 5
	if _, e := s.Reserve(ctx, fresh(), "config", changed, grid, 100, 1000, 1000); e != ErrConflict {
		t.Fatal("stale geographic validation accepted")
	}
	var wg sync.WaitGroup
	results := make(chan error, 10)
	for n := 0; n < 10; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := s.Reserve(ctx, fresh(), "config", p, grid, 100, 1000, 1000)
			results <- e
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else if e != ErrConflict {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatalf("%d concurrent reservations succeeded", success)
	}
}
