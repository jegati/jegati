package geography

import (
	"math"
	"testing"
)

func FuzzGrid(f *testing.F) {
	f.Add(100, 19.818, 41.327, "tirana-v1:100:55:55")
	f.Add(5000, West, South, "bad")
	f.Add(100, math.NaN(), math.Inf(1), "")
	f.Fuzz(func(t *testing.T, size int, x, y float64, id string) {
		if len(id) > 2048 {
			return
		}
		g, e := NewGrid(size)
		if e != nil {
			return
		}
		p := Point{x, y}
		cell, e := g.CellAt(p)
		if e == nil {
			if !Inside(p) {
				t.Fatal("invalid coordinate accepted")
			}
			parsed, e := g.Parse(g.ID(cell))
			if e != nil || parsed != cell {
				t.Fatal("cell roundtrip")
			}
			if g.MaxDistance(cell, p) < Distance(g.Center(cell), p) {
				t.Fatal("conservative bound violated")
			}
		}
		if parsed, e := g.Parse(id); e == nil && g.ID(parsed) != id {
			t.Fatal("noncanonical cell accepted")
		}
	})
}
