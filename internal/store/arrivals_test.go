package store

import (
	"context"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func TestConcurrentArrivalReplayAndRetraction(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	id := fresh()
	if _, e := s.Reserve(ctx, id, "config", p, grid, 50, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(70 * time.Millisecond)
	g, e := s.Activate(ctx, id, "config", 1000, 60000)
	if e != nil {
		t.Fatal(e)
	}
	hash := p.Founders[0].Hash
	if _, e = s.SetIntent(ctx, hash, id, "going", grid.ID(p.Founders[0].Area.Cell), 3, 32, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	nonce := fresh()
	if _, e = s.IssueArrivalNonce(ctx, hash, nonce, 20000); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 10)
	for n := 0; n < 10; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := s.ConfirmArrival(ctx, hash, nonce, g, 30000); results <- e }()
	}
	wg.Wait()
	close(results)
	for e := range results {
		if e != nil {
			t.Fatal(e)
		}
	}
	count, e := s.Client.ZCard(ctx, keyPrefix+"arrivals:"+id).Result()
	if e != nil || count != 1 {
		t.Fatal("arrival replay inflated distinct credential count")
	}
	if _, e = s.RetractArrival(ctx, hash); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ConfirmArrival(ctx, hash, nonce, g, 30000); e != ErrGone {
		t.Fatal("replay restored retracted arrival")
	}
	if _, e = s.IssueArrivalNonce(ctx, hash, nonce, 20000); e != ErrGone {
		t.Fatal("used nonce renewed by issuance")
	}
	if count, e = s.Client.ZCard(ctx, keyPrefix+"arrivals:"+id).Result(); e != nil || count != 0 {
		t.Fatal("retraction left arrival contribution")
	}
}

// A seeded adversarial sequence checks independent-credential accounting after
// every operation, including retries, withdrawal, wrong nonces and cancellation.
func TestSeededArrivalTransitionModel(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	p, grid := proposalTest(t, s)
	id := fresh()
	if _, e := s.Reserve(ctx, id, "model", p, grid, 20, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(30 * time.Millisecond)
	g, e := s.Activate(ctx, id, "model", 1000, 60000)
	if e != nil {
		t.Fatal(e)
	}
	hash := p.Founders[0].Hash
	if _, e = s.SetIntent(ctx, hash, id, "going", grid.ID(p.Founders[0].Area.Cell), 3, 32, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	rng := rand.New(rand.NewPCG(42, 29))
	here := false
	nonce := ""
	cancelled := false
	for step := 0; step < 250; step++ {
		if step == 200 {
			if e = s.Cancel(ctx, hash); e != nil {
				t.Fatal(e)
			}
			here = false
			cancelled = true
		}
		switch rng.IntN(4) {
		case 0:
			nonce = fresh()
			_, e = s.IssueArrivalNonce(ctx, hash, nonce, 20000)
			if cancelled && e != ErrGone {
				t.Fatal("cancelled nonce issuance accepted")
			}
			if !cancelled && here && e != ErrConflict {
				t.Fatal("duplicate arrival nonce issuance accepted", e)
			}
			if !cancelled && !here && e != nil {
				t.Fatal(e)
			}
		case 1:
			_, e = s.ConfirmArrival(ctx, hash, nonce, g, 30000)
			if cancelled && e != ErrGone {
				t.Fatal("cancelled arrival accepted")
			}
			if e == nil {
				here = true
			}
		case 2:
			_, e = s.RetractArrival(ctx, hash)
			if e == nil {
				here = false
			}
			if cancelled && e != ErrGone {
				t.Fatal("cancelled retraction accepted")
			}
		case 3:
			if _, e = s.ConfirmArrival(ctx, hash, fresh(), g, 30000); e != ErrGone {
				t.Fatal("unknown nonce accepted", e)
			}
		}
		expected := int64(0)
		if here {
			expected = 1
		}
		if count, e := s.Client.ZCard(ctx, keyPrefix+"arrivals:"+id).Result(); e != nil || count != expected {
			t.Fatal("arrival count diverged from model", step, count, e)
		}
		v, e := s.Status(ctx, hash)
		if cancelled {
			if e != ErrGone {
				t.Fatal("cancelled signal restored")
			}
		} else if e != nil || (v.State == "here") != here {
			t.Fatal("private state diverged", step, e)
		}
	}
}
