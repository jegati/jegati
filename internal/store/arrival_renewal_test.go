package store

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestArrivalRenewalKeepsOneContributionAndDeadlines(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	g, p, grid := shortGathering(t, s, 50000)
	hash := p.Founders[0].Hash
	if _, err := s.SetIntent(ctx, hash, g.ID, "going", grid.ID(p.Founders[0].Area.Cell), 3, 32, 1000, 1000); err != nil {
		t.Fatal(err)
	}
	nonce := fresh()
	if _, err := s.IssueArrivalNonce(ctx, hash, nonce, 20000); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RenewArrival(ctx, hash, nonce, g, 30000, 30000); err != ErrGone {
		t.Fatal("ordinary challenge accepted as renewal", err)
	}
	before, err := s.ConfirmArrival(ctx, hash, nonce, g, 30000)
	if err != nil {
		t.Fatal(err)
	}
	renew := fresh()
	if _, err = s.IssueArrivalRenewalNonce(ctx, hash, renew, 20000, 1000); err != ErrConflict {
		t.Fatal("early renewal allowed", err)
	}
	// A wide test window avoids sleeps while retaining the production transaction.
	until, err := s.IssueArrivalRenewalNonce(ctx, hash, renew, 40000, 30000)
	if err != nil || until > before.ArrivalUntil {
		t.Fatal("challenge outlived previous confirmation", err)
	}
	if _, err = s.ConfirmArrival(ctx, hash, renew, g, 30000); err != ErrGone {
		t.Fatal("renewal nonce accepted by ordinary arrival", err)
	}
	if err = s.ReconcilePresence(ctx, g.ID, 1, 100, 10000, 1000); err != nil {
		t.Fatal(err)
	}
	cohort, err := s.Client.Get(ctx, keyPrefix+"presence:"+g.ID).Result()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	var wg sync.WaitGroup
	results := make(chan Signal, 10)
	failures := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := s.RenewArrival(ctx, hash, renew, g, 30000, 30000)
			results <- v
			failures <- e
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for e := range failures {
		if e != nil {
			t.Fatal(e)
		}
	}
	var expiry int64
	for v := range results {
		if expiry == 0 {
			expiry = v.ArrivalUntil
		}
		if v.ArrivalUntil != expiry || expiry <= before.ArrivalUntil || v.ArrivalMember != before.ArrivalMember || v.ExpiresAt != before.ExpiresAt || v.GatheringUntil != before.GatheringUntil {
			t.Fatal("renewal replay changed contribution or deadlines")
		}
	}
	if count := s.Client.ZCard(ctx, keyPrefix+"arrivals:"+g.ID).Val(); count != 1 {
		t.Fatal("renewal inflated count", count)
	}
	if err = s.ReconcilePresence(ctx, g.ID, 1, 100, 10000, 1000); err != nil {
		t.Fatal(err)
	}
	if actual := s.Client.Get(ctx, keyPrefix+"presence:"+g.ID).Val(); actual != cohort {
		t.Fatal("uninterrupted renewal reset stability")
	}
	if current, e := s.GatheringByID(ctx, g.ID); e != nil || current.EndsAt != g.EndsAt {
		t.Fatal("renewal extended gathering", e)
	}
	if _, err = s.RetractArrival(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RenewArrival(ctx, hash, renew, g, 30000, 30000); err != ErrGone {
		t.Fatal("replay restored retracted arrival", err)
	}
}

func TestArrivalRenewalCannotRestoreMissingContribution(t *testing.T) {
	for _, change := range []string{"retract", "cancel", "missing-member"} {
		t.Run(change, func(t *testing.T) {
			s := connectTest(t)
			ctx := context.Background()
			g, p, grid := shortGathering(t, s, 50000)
			hash := p.Founders[0].Hash
			if _, e := s.SetIntent(ctx, hash, g.ID, "going", grid.ID(p.Founders[0].Area.Cell), 3, 32, 1000, 1000); e != nil {
				t.Fatal(e)
			}
			nonce := fresh()
			if _, e := s.IssueArrivalNonce(ctx, hash, nonce, 20000); e != nil {
				t.Fatal(e)
			}
			v, e := s.ConfirmArrival(ctx, hash, nonce, g, 30000)
			if e != nil {
				t.Fatal(e)
			}
			nonce = fresh()
			if _, e = s.IssueArrivalRenewalNonce(ctx, hash, nonce, 20000, 30000); e != nil {
				t.Fatal(e)
			}
			switch change {
			case "retract":
				_, e = s.RetractArrival(ctx, hash)
			case "cancel":
				e = s.Cancel(ctx, hash)
			case "missing-member":
				e = s.Client.ZRem(ctx, keyPrefix+"arrivals:"+g.ID, v.ArrivalMember).Err()
			}
			if e != nil {
				t.Fatal(e)
			}
			if _, e = s.RenewArrival(ctx, hash, nonce, g, 30000, 30000); e != ErrConflict && e != ErrGone {
				t.Fatal("renewal restored missing contribution", e)
			}
			if count := s.Client.ZCard(ctx, keyPrefix+"arrivals:"+g.ID).Val(); count != 0 {
				t.Fatal("renewal left a contribution", count)
			}
		})
	}
}

func TestArrivalRenewalCannotBridgeExpiredPresence(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	g, p, grid := shortGathering(t, s, 50000)
	hash := p.Founders[0].Hash
	if _, e := s.SetIntent(ctx, hash, g.ID, "going", grid.ID(p.Founders[0].Area.Cell), 3, 32, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	nonce := fresh()
	if _, e := s.IssueArrivalNonce(ctx, hash, nonce, 20000); e != nil {
		t.Fatal(e)
	}
	if _, e := s.ConfirmArrival(ctx, hash, nonce, g, 100); e != nil {
		t.Fatal(e)
	}
	nonce = fresh()
	if _, e := s.IssueArrivalRenewalNonce(ctx, hash, nonce, 20000, 100); e != nil {
		t.Fatal(e)
	}
	time.Sleep(130 * time.Millisecond)
	if _, e := s.RenewArrival(ctx, hash, nonce, g, 30000, 100); e != ErrGone {
		t.Fatal("expired challenge bridged presence gap", e)
	}
	v, e := s.Status(ctx, hash)
	if e != nil || v.State != "going" || v.ArrivalUntil != 0 {
		t.Fatal("expired arrival restored", e)
	}
	if count := s.Client.ZCard(ctx, keyPrefix+"arrivals:"+g.ID).Val(); count != 0 {
		t.Fatal("expired arrival retained", count)
	}
}
