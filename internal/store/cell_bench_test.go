package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

// The same benchmark compiles against the baseline revision. Its fallback is
// the former full-snapshot refresh; current code reads only a changed cell.
type cellReader interface {
	CellSnapshot(context.Context, geography.Grid, string, int64, int, int) ([]matching.Signal, error)
}

func BenchmarkChangedCell(b *testing.B) {
	if os.Getenv("GATI_SCALING_BENCH") != "1" {
		b.Skip("use make benchmark-scaling")
	}
	s := connectTest(b)
	ctx := context.Background()
	grid, _ := geography.NewGrid(100)
	const size = 100000
	for _, hotspot := range []bool{false, true} {
		name := "uniform-100000"
		if hotspot {
			name = "hotspot-100000"
		}
		b.Run(name, func(b *testing.B) {
			if s.Client.ZCard(ctx, keyPrefix+"expiry").Val() != 0 {
				b.Fatal("disposable empty store required")
			}
			now, err := s.Now(ctx)
			if err != nil {
				b.Fatal(err)
			}
			cells := map[string]bool{}
			pipe := s.Client.Pipeline()
			wanted := grid.ID(geography.Cell{X: 50, Y: 50})
			expected := 0
			for n := 0; n < size; n++ {
				cell := grid.ID(geography.Cell{X: n % grid.Columns, Y: (n / grid.Columns) % grid.Rows})
				if hotspot {
					cell = wanted
				}
				if cell == wanted {
					expected++
				}
				cells[cell] = true
				hash := fmt.Sprintf("%064x", n)
				raw, _ := json.Marshal(Signal{Cell: cell, RadiusKM: 1, AvailabilityMinutes: 120, CreatedAt: now, ExpiresAt: now + 7200000, State: "gati"})
				pipe.Set(ctx, keyPrefix+"s:"+hash, raw, 2*time.Hour)
				pipe.ZAdd(ctx, keyPrefix+"expiry", redis.Z{Score: float64(now + 7200000), Member: hash + "|" + cell})
				pipe.ZAdd(ctx, keyPrefix+"cell:"+cell, redis.Z{Score: float64(now + 7200000), Member: hash})
				if n%1000 == 999 {
					if _, err := pipe.Exec(ctx); err != nil {
						b.Fatal(err)
					}
				}
			}
			pipe.PExpire(ctx, keyPrefix+"expiry", 2*time.Hour)
			for cell := range cells {
				pipe.PExpire(ctx, keyPrefix+"cell:"+cell, 2*time.Hour)
			}
			if _, err := pipe.Exec(ctx); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				var values []matching.Signal
				if reader, ok := any(s).(cellReader); ok {
					values, err = reader.CellSnapshot(ctx, grid, wanted, now, 1000, size)
				} else {
					var all []matching.Signal
					all, err = s.EligibleSnapshot(ctx, grid, now, 1000, size)
					for _, v := range all {
						if v.Area.Cell == (geography.Cell{X: 50, Y: 50}) {
							values = append(values, v)
						}
					}
				}
				if err != nil || len(values) != expected {
					b.Fatalf("incomplete changed-cell read: count=%d error=%v", len(values), err)
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
			for cell := range cells {
				pipe.Del(ctx, keyPrefix+"cell:"+cell)
			}
			pipe.Del(ctx, keyPrefix+"expiry")
			if _, err := pipe.Exec(ctx); err != nil {
				b.Fatal(err)
			}
		})
	}
}
