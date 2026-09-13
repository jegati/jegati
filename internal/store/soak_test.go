package store

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
	"time"
)

// This real-time churn test uses short synthetic TTLs at the storage boundary.
// It does not lower the public API's 30-minute availability floor.
func TestLifecycleSoak(t *testing.T) {
	seconds, e := strconv.Atoi(os.Getenv("GATI_SOAK_SECONDS"))
	if e != nil {
		t.Skip("make test-soak")
	}
	if seconds < 5 || seconds > 86400 {
		t.Fatal("soak seconds outside 5..86400")
	}
	s := connectTest(t)
	ctx := context.Background()
	rng := rand.New(rand.NewPCG(42, 19))
	keeper := fresh()
	cell := "tirana-v1:100:55:55"
	value, e := s.Create(ctx, keeper, cell, 3, 30, time.Duration(seconds+60)*time.Second, 10000)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Cancel(ctx, keeper)
	if _, e = s.RegisterPush(ctx, keeper, PushBinding{Binding: fresh(), Digest: fresh(), Ciphertext: "synthetic", ExpiresAt: value.ExpiresAt}); e != nil {
		t.Fatal(e)
	}
	start := time.Now()
	cycles, created, cancelled := 0, 0, 0
	maxIndex := int64(0)
	for time.Since(start) < time.Duration(seconds)*time.Second {
		hashes := []string{}
		for i := 0; i < 100; i++ {
			hash := fresh()
			hashes = append(hashes, hash)
			v, e := s.Create(ctx, hash, cell, 3, 30, 1200*time.Millisecond, 10000)
			if e != nil {
				t.Fatal(e)
			}
			created++
			if i%2 == 0 {
				if _, e = s.RegisterPush(ctx, hash, PushBinding{Binding: fresh(), Digest: fresh(), Ciphertext: "synthetic", ExpiresAt: v.ExpiresAt}); e != nil {
					t.Fatal(e)
				}
			}
			if rng.IntN(3) == 0 {
				if e = s.Cancel(ctx, hash); e != nil {
					t.Fatal(e)
				}
				cancelled++
				if _, e = s.Create(ctx, hash, cell, 3, 30, time.Second, 10000); e != ErrGone {
					t.Fatal("cancelled credential resurrected")
				}
			}
		}
		if n := s.Client.ZCard(ctx, keyPrefix+"expiry").Val(); n > maxIndex {
			maxIndex = n
		}
		time.Sleep(1250 * time.Millisecond)
		for _, hash := range hashes {
			if _, e = s.Status(ctx, hash); e != ErrGone {
				t.Fatal("expired signal readable")
			}
		}
		if _, e = s.Cleanup(ctx, 1000); e != nil {
			t.Fatal(e)
		}
		for _, key := range []string{"expiry", "cell:" + cell, "push-due"} {
			if n, e := s.Client.ZCard(ctx, keyPrefix+key).Result(); e != nil || n != 1 {
				t.Fatal("index accumulated expired members", key, n, e)
			}
		}
		cycles++
	}
	result := map[string]any{"synthetic": true, "real_time": true, "seconds": time.Since(start).Seconds(), "cycles": cycles, "created": created, "cancelled": cancelled, "peak_expiry_members": maxIndex, "remaining_expected_keeper": 1, "short_fixture_ttl_ms": 1200}
	if path := os.Getenv("GATI_SOAK_REPORT"); path != "" {
		raw, _ := json.MarshalIndent(result, "", "  ")
		if e = os.WriteFile(path, append(raw, '\n'), 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("soak: %d cycles, %d creations, %d cancellations; indexes returned to baseline", cycles, created, cancelled)
}
