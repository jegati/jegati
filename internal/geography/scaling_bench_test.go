package geography

import (
	"fmt"
	"testing"
)

// A bounded synthetic index isolates lookup from the map-import/startup cost.
func BenchmarkCanReach(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			g, _ := NewGrid(100)
			p := ParticipantArea{Cell: Cell{50, 50}, RadiusKM: 3}
			idx := &Index{Grid: g, reachable: map[ParticipantArea][]int{}}
			for n := 0; n < size; n++ {
				idx.Intersections = append(idx.Intersections, Intersection{ID: fmt.Sprintf("road-%05d", n)})
				idx.reachable[p] = append(idx.reachable[p], n)
			}
			last := idx.Intersections[size-1].ID
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				if !idx.CanReach(p, last) || idx.CanReach(p, "road-missing") {
					b.Fatal("incorrect reachability")
				}
			}
		})
	}
}
