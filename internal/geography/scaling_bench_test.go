package geography

import (
	"fmt"
	"testing"
)

func TestIndexedReachabilityMatchesMembership(t *testing.T) {
	g, _ := NewGrid(100)
	p := ParticipantArea{Cell: Cell{50, 50}, RadiusKM: 1}
	i := &Index{Grid: g, reachable: map[ParticipantArea][]int{p: {0, 2, 4}}}
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		i.Intersections = append(i.Intersections, Intersection{ID: id})
	}
	for _, id := range []string{"", "0", "a", "b", "c", "d", "e", "z"} {
		want := id == "a" || id == "c" || id == "e"
		if i.CanReach(p, id) != want {
			t.Fatalf("incorrect membership for %q", id)
		}
		if i.CanReach(ParticipantArea{}, id) {
			t.Fatal("unknown area accepted")
		}
	}
}

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
