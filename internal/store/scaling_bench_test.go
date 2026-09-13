package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/geography"
	"github.com/redis/go-redis/v9"
)

// This fixture bypasses HTTP/admission to isolate traversal cost. Only the owned
// disposable benchmark store is permitted by the runner. No raw records are output.
func BenchmarkEligibleSnapshot(b *testing.B) {
	if os.Getenv("GATI_SCALING_BENCH") != "1" {
		b.Skip("use make benchmark-scaling")
	}
	s := connectTest(b)
	ctx := context.Background()
	grid, _ := geography.NewGrid(100)
	for _, size := range []int{10000, 100000, 1000000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			if s.Client.ZCard(ctx, keyPrefix+"expiry").Val() != 0 {
				b.Fatal("benchmark requires an empty disposable index")
			}
			now, err := s.Now(ctx)
			if err != nil {
				b.Fatal(err)
			}
			pipe := s.Client.Pipeline()
			for n := 0; n < size; n++ {
				hash := fmt.Sprintf("%064x", n)
				cell := grid.ID(geography.Cell{X: n % grid.Columns, Y: (n / grid.Columns) % grid.Rows})
				// Deliberate equal-score runs exercise pagination ties.
				v := Signal{Cell: cell, RadiusKM: 1, AvailabilityMinutes: 120, CreatedAt: now, ExpiresAt: now + 7200000, State: "gati"}
				raw, _ := json.Marshal(v)
				pipe.Set(ctx, keyPrefix+"s:"+hash, raw, 2*time.Hour)
				pipe.ZAdd(ctx, keyPrefix+"expiry", redis.Z{Score: float64(v.ExpiresAt), Member: hash + "|" + cell})
				if n%1000 == 999 {
					if _, err := pipe.Exec(ctx); err != nil {
						b.Fatal(err)
					}
				}
			}
			if _, err := pipe.Exec(ctx); err != nil {
				b.Fatal(err)
			}
			s.Client.PExpire(ctx, keyPrefix+"expiry", 2*time.Hour)
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				v, err := s.EligibleSnapshot(ctx, grid, now, 1000, size)
				if err != nil || len(v) != size {
					b.Fatalf("incomplete snapshot: count=%d error=%v", len(v), err)
				}
			}
			b.StopTimer()
			for n := 0; n < size; n++ {
				pipe.Del(ctx, keyPrefix+"s:"+fmt.Sprintf("%064x", n))
				if n%1000 == 999 {
					if _, err := pipe.Exec(ctx); err != nil {
						b.Fatal(err)
					}
				}
			}
			pipe.Del(ctx, keyPrefix+"expiry")
			if _, err := pipe.Exec(ctx); err != nil {
				b.Fatal(err)
			}
		})
	}
}
