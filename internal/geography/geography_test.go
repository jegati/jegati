package geography

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestGridRoundtripAndBounds(t *testing.T) {
	for _, size := range []int{500, 1000, 5000} {
		g, e := NewGrid(size)
		if e != nil {
			t.Fatal(e)
		}
		for y := 0; y < g.Rows; y++ {
			for x := 0; x < g.Columns; x++ {
				c := Cell{x, y}
				got, e := g.Parse(g.ID(c))
				if e != nil || got != c {
					t.Fatal("cell identity")
				}
				if Inside(g.Center(c)) {
					got, e = g.CellAt(g.Center(c))
					if e != nil || got != c {
						t.Fatal("cell center roundtrip")
					}
				}
			}
		}
		for _, p := range []Point{{West - 0.00001, South}, {East, South}, {West, North}, {math.NaN(), South}, {West, math.Inf(1)}} {
			if _, e := g.CellAt(p); e == nil {
				t.Fatal("invalid point accepted")
			}
		}
		for _, id := range []string{"tirana-v1:1000:-1:0", "tirana-v1:1000:01:1", "tirana-v1:1000:999:999", "tirana-v1:5:0:0", "tirana-v2:1000:0:0", "name"} {
			if _, e := g.Parse(id); e == nil {
				t.Fatalf("accepted %s", id)
			}
		}
	}
}
func TestConservativeRadiusAndNearestSharedIntersection(t *testing.T) {
	g, _ := NewGrid(1000)
	cell := Cell{4, 4}
	center := g.Center(cell)
	candidates := []Intersection{{ID: "node/2", Point: Point{center[0] + 0.001, center[1]}}, {ID: "node/1", Point: center}, {ID: "node/3", Point: Point{center[0] + 0.01, center[1]}}}
	idx, e := NewIndex(g, candidates, []int{1, 3})
	if e != nil {
		t.Fatal(e)
	}
	chosen, ok := idx.Closest([]ParticipantArea{{cell, 1}, {cell, 3}})
	if !ok || chosen.ID != "node/1" {
		t.Fatal("closest shared intersection not selected")
	}
	if idx.CanReach(ParticipantArea{cell, 1}, "node/3") {
		t.Fatal("used center-only distance")
	}
	for _, intersection := range candidates {
		bound := g.MaxDistance(cell, intersection.Point)
		for x := 0; x <= 10; x++ {
			for y := 0; y <= 10; y++ {
				p := Point{g.West + (float64(cell.X)+float64(x)/10)*g.LonStep, g.South + (float64(cell.Y)+float64(y)/10)*g.LatStep}
				if Distance(p, intersection.Point) > bound {
					t.Fatal("bound underestimated sample")
				}
			}
		}
	}
	if _, ok := idx.Closest(nil); ok {
		t.Fatal("empty group matched")
	}
	if _, ok := idx.Closest([]ParticipantArea{{cell, 1}, {Cell{g.Columns - 1, g.Rows - 1}, 1}}); ok {
		t.Fatal("incompatible group matched")
	}
	empty, _ := NewIndex(g, nil, []int{1})
	if _, ok := empty.Closest([]ParticipantArea{{cell, 1}}); ok {
		t.Fatal("invented missing intersection")
	}
	// Exact same public map point: stable identifier resolves the tie.
	idx, _ = NewIndex(g, []Intersection{{ID: "node/b", Point: center}, {ID: "node/a", Point: center}}, []int{1})
	chosen, _ = idx.Closest([]ParticipantArea{{cell, 1}})
	if chosen.ID != "node/a" {
		t.Fatal("unstable tie")
	}
}
func TestCheckedInDataset(t *testing.T) {
	raw, e := os.ReadFile("../../data/tirana/intersections.json")
	if e != nil {
		t.Fatal(e)
	}
	var data Dataset
	if e = json.Unmarshal(raw, &data); e != nil {
		t.Fatal(e)
	}
	if len(data.Intersections) == 0 || len(data.SourceSHA256) != 64 || data.Version != "tirana-intersections-v1" {
		t.Fatal("invalid provenance")
	}
	g, _ := NewGrid(1000)
	idx, e := NewIndex(g, data.Intersections, []int{1, 3, 5})
	if e != nil {
		t.Fatal(e)
	}
	center, _ := g.CellAt(Point{19.818, 41.327})
	if len(idx.Reachable(ParticipantArea{center, 3})) == 0 {
		t.Fatal("central Tirana fixture has no reachable intersection")
	}
}
