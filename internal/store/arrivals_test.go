package store

import (
	"context"
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
