package store

import (
	"context"
	"encoding/json"
	"github.com/jegati/jegati/internal/geography"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

func TestSnapshotEqualDeadlinesExpiryAndChurn(t *testing.T) {
	s := connectTest(t)
	ctx := context.Background()
	g, _ := geography.NewGrid(100)
	now, err := s.Now(ctx)
	if err != nil {
		t.Fatal(err)
	}
	stable := map[string]bool{}
	add := func(expiry int64) string {
		hash := fresh()
		cell := g.ID(geography.Cell{X: 50, Y: 50})
		raw, _ := json.Marshal(Signal{Cell: cell, RadiusKM: 1, CreatedAt: now, ExpiresAt: expiry, State: "gati"})
		p := s.Client.Pipeline()
		p.Set(ctx, keyPrefix+"s:"+hash, raw, time.Minute)
		p.ZAdd(ctx, keyPrefix+"expiry", redis.Z{Score: float64(expiry), Member: hash + "|" + cell})
		p.PExpire(ctx, keyPrefix+"expiry", time.Minute)
		if _, err := p.Exec(ctx); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { s.Cancel(ctx, hash) })
		return hash
	}
	for n := 0; n < 1500; n++ {
		stable[add(now+60000)] = true
	}
	assigned := add(now + 60000)
	mixed := Signal{Cell: g.ID(geography.Cell{X: 50, Y: 50}), RadiusKM: 1, CreatedAt: now, ExpiresAt: now + 60000, GatheringUntil: now + 60000, Declined: map[string]bool{"synthetic-event": true}}
	raw, _ := json.Marshal(mixed)
	s.Client.Set(ctx, keyPrefix+"s:"+assigned, raw, time.Minute)
	stable[assigned] = true
	expired := add(now)
	missing := add(now + 60000)
	s.Client.Del(ctx, keyPrefix+"s:"+missing)
	t.Cleanup(func() { s.Client.ZRem(ctx, keyPrefix+"expiry", missing+"|"+g.ID(geography.Cell{X: 50, Y: 50})) })
	for _, batch := range []int{1, 127, 1000} {
		values, err := s.EligibleSnapshot(ctx, g, now, batch, 2000)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, v := range values {
			if seen[v.Hash] || v.Hash == expired || v.Hash == missing {
				t.Fatal("duplicate/expired/missing record")
			}
			if v.Assigned != (v.Hash == assigned) || (len(v.Declined) > 0) != (v.Hash == assigned) {
				t.Fatal("decoder carried state between records")
			}
			seen[v.Hash] = true
		}
		for hash := range stable {
			if !seen[hash] {
				t.Fatal("equal-score member missed")
			}
		}
	}
	// Continuously present records must survive concurrent deletion of others.
	volatile := []string{}
	for n := 0; n < 200; n++ {
		volatile = append(volatile, add(now+60000))
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, hash := range volatile {
			s.Cancel(ctx, hash)
		}
	}()
	values, err := s.EligibleSnapshot(ctx, g, now, 10, 2000)
	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v.Hash] {
			t.Fatal("duplicate during churn")
		}
		seen[v.Hash] = true
	}
	for hash := range stable {
		if !seen[hash] {
			t.Fatal("stable record missed during churn")
		}
	}
	if _, err := s.EligibleSnapshot(ctx, g, now, 100, 1); err != ErrCapacity {
		t.Fatal("capacity overflow silently truncated", err)
	}
}
