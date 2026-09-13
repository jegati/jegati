package worker

import (
	"context"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWorkingSetDropsExpiryAndReevaluatesAssignment(t *testing.T) {
	c := geography.Cell{X: 50, Y: 50}
	values := []matching.Signal{{Hash: "expired", ExpiresAt: 100}, {Hash: "live", ExpiresAt: 1000, Assigned: true, AssignedUntil: 90}, {Hash: "reserved", ExpiresAt: 1200, ReservedUntil: 300}}
	p := &population{cells: map[geography.Cell][]matching.Signal{c: values}}
	live := p.live(100, 500)
	if len(live) != 2 || live[0].Assigned || p.nextDeadline != 300 {
		t.Fatal("deadline transition missed")
	}
	if values[2].Hash != "" {
		t.Fatal("expired buffer references retained")
	}
	p.live(1300, 500)
	if len(p.cells) != 0 || p.nextDeadline != 0 {
		t.Fatal("expired working set retained")
	}
}

func TestChangedCellsRefreshAndFullReconciliation(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("use make test-store")
	}
	password, err := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	s, err := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(password)))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Client.Close()
	cfg, err := config.Load("../../config/gati.yaml", false)
	if err != nil {
		t.Fatal(err)
	}
	g, _ := geography.NewGrid(cfg.Geography.CellSizeMeters)
	idx, err := geography.NewIndex(g, []geography.Intersection{{ID: "synthetic-road", Point: g.Center(geography.Cell{X: 50, Y: 50})}}, cfg.Geography.TravelRadiusChoicesKm)
	if err != nil {
		t.Fatal(err)
	}
	e := New(s, idx, cfg)
	first, second := randomID(), randomID()
	for n, hash := range []string{first, second} {
		if _, err := s.Create(ctx, hash, g.ID(geography.Cell{X: 50 + n, Y: 50}), 1, 30, 30*time.Minute, cfg.Limits.MaxActiveSignals); err != nil {
			t.Fatal(err)
		}
		defer s.Cancel(ctx, hash)
	}
	now, _ := s.Now(ctx)
	before, err := e.populationSnapshot(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	contains := func(values []matching.Signal, hash string) bool {
		for _, v := range values {
			if v.Hash == hash {
				return true
			}
		}
		return false
	}
	if !contains(before, first) || !contains(before, second) {
		t.Fatal("initial population missing")
	}
	if err := s.Cancel(ctx, first); err != nil {
		t.Fatal(err)
	}
	updated, err := e.populationSnapshot(ctx, now+1)
	if err != nil {
		t.Fatal(err)
	}
	if contains(updated, first) || !contains(updated, second) || e.population.fullAt != now {
		t.Fatal("incremental cancellation lost or full scan used")
	}
	// Simulate loss of a dirty batch: periodic reconciliation still repairs state.
	s.Cancel(ctx, second)
	s.TakeChangedCells(ctx, g.Columns*g.Rows)
	refreshed, err := e.populationSnapshot(ctx, now+int64(cfg.Matching.ReconciliationSeconds)*1000)
	if err != nil {
		t.Fatal(err)
	}
	if contains(refreshed, second) {
		t.Fatal("reconciliation did not repair a lost batch")
	}
}
