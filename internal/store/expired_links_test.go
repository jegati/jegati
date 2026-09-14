package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
)

func shortGathering(t *testing.T, s *Store, lifetime int64) (Gathering, matching.Proposal, geography.Grid) {
	t.Helper()
	p, grid := proposalTest(t, s)
	id := fresh()
	if _, err := s.Reserve(context.Background(), id, "cleanup-test", p, grid, 1, 1000, 1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	g, err := s.Activate(context.Background(), id, "cleanup-test", 1, lifetime)
	if err != nil {
		t.Fatal(err)
	}
	return g, p, grid
}

func TestGatheringCleanupPrunesChurnWithLiveKeeper(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	keeper, _, _ := shortGathering(t, s, 50000)
	for cycle := 0; cycle < 3; cycle++ {
		expired := []Gathering{}
		for i := 0; i < 3; i++ {
			g, _, _ := shortGathering(t, s, 60)
			expired = append(expired, g)
		}
		time.Sleep(90 * time.Millisecond)
		// Native record expiry must not conceal stale membership in a kept-alive index.
		for _, g := range expired {
			if _, err := s.Client.ZScore(ctx, keyPrefix+"gatherings", g.ID).Result(); err != nil {
				t.Fatal("fixture index expired with records", err)
			}
		}
		if lag, err := s.CleanupLag(ctx); err != nil || lag <= 0 {
			t.Fatal("expired gathering lag not observed", lag, err)
		}
		for i := 0; i < 100; i++ {
			n, err := s.Cleanup(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			if n < 1 {
				break
			}
		}
		for _, g := range expired {
			if s.Client.ZScore(ctx, keyPrefix+"gatherings", g.ID).Err() == nil {
				t.Fatal("expired gathering index member retained")
			}
		}
		live, err := s.GatheringByID(ctx, keeper.ID)
		if err != nil || live.EndsAt != keeper.EndsAt || live.Intersection.ID != keeper.Intersection.ID {
			t.Fatal("cleanup changed live keeper", err)
		}
		if s.Client.ZScore(ctx, keyPrefix+"gatherings", keeper.ID).Val() != float64(keeper.EndsAt) {
			t.Fatal("keeper index changed")
		}
	}
}

func TestExpiredGatheringLinksClearedOnReadAndSnapshot(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	g, p, grid := shortGathering(t, s, 100)
	key := func(i int) string { return keyPrefix + "s:" + p.Founders[i].Hash }
	// A nonce is private association metadata and must disappear with the old link.
	s.Client.Set(ctx, keyPrefix+"arrival-nonce:"+p.Founders[0].Hash, "synthetic stale challenge", time.Minute)
	time.Sleep(130 * time.Millisecond)
	assertCleared := func(i int) {
		t.Helper()
		var v Signal
		raw, err := s.Client.Get(ctx, key(i)).Bytes()
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		if v.Gathering != "" || v.GatheringUntil != 0 || v.ArrivalUntil != 0 || v.ArrivalMember != "" || v.State != "gati" {
			t.Fatal("stale association retained", v.State)
		}
		if v.ExpiresAt != p.Founders[i].ExpiresAt {
			t.Fatal("cleanup renewed willingness")
		}
	}
	if _, err := s.Status(ctx, p.Founders[0].Hash); err != nil {
		t.Fatal(err)
	}
	assertCleared(0)
	if s.Client.Exists(ctx, keyPrefix+"arrival-nonce:"+p.Founders[0].Hash).Val() != 0 {
		t.Fatal("stale nonce retained")
	}
	now, _ := s.Now(ctx)
	if _, err := s.CellSnapshot(ctx, grid, grid.ID(p.Founders[1].Area.Cell), now, 500, 1000); err != nil {
		t.Fatal(err)
	}
	assertCleared(1)
	assertCleared(2)
	// A delayed cleanup batch must reread the association, not erase a newer one.
	next := g
	next.ID = fresh()
	next.EndsAt = now + 20000
	encoded, _ := json.Marshal(next)
	s.Client.Set(ctx, keyPrefix+"gathering:"+next.ID, encoded, 20*time.Second)
	if _, err := s.SetIntent(ctx, p.Founders[1].Hash, next.ID, "going", grid.ID(p.Founders[1].Area.Cell), 3, 32, 1000, 1000); err != nil {
		t.Fatal(err)
	}
	if err := s.clearExpiredGatheringLinks(ctx, []string{key(1)}); err != nil {
		t.Fatal(err)
	}
	v, err := s.Status(ctx, p.Founders[1].Hash)
	if err != nil || v.Gathering != next.ID || v.State != "going" || v.ExpiresAt != p.Founders[1].ExpiresAt {
		t.Fatal("new assignment erased by delayed cleanup", err)
	}
}
