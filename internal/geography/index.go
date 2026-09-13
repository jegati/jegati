package geography

import (
	"errors"
	"sort"
)

type Intersection struct {
	ID        string   `json:"id"`
	SourceIDs []string `json:"source_ids"`
	Point     Point    `json:"point"`
	Geometry  []Point  `json:"geometry,omitempty"`
	Label     string   `json:"label"`
}
type Dataset struct {
	Version         string         `json:"version"`
	SourceSHA256    string         `json:"source_sha256"`
	SourceTimestamp string         `json:"source_timestamp"`
	Intersections   []Intersection `json:"intersections"`
}
type ParticipantArea struct {
	Cell     Cell
	RadiusKM float64
}
type Index struct {
	Grid          Grid
	Intersections []Intersection
	reachable     map[ParticipantArea][]int
}

func NewIndex(g Grid, intersections []Intersection, radii []float64) (*Index, error) {
	if len(intersections) > 100000 || len(radii) == 0 || len(radii) > 32 {
		return nil, errors.New("geography index outside bounds")
	}
	cs := append([]Intersection(nil), intersections...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].ID < cs[j].ID })
	for i, c := range cs {
		if !Inside(c.Point) || c.ID == "" || (i > 0 && cs[i-1].ID == c.ID) {
			return nil, errors.New("invalid intersection dataset")
		}
	}
	idx := &Index{g, cs, make(map[ParticipantArea][]int)}
	// Dataset and finite coarse grid bound this offline/startup computation.
	// Request handlers never recompute participant-to-participant distance pairs.
	for y := 0; y < g.Rows; y++ {
		for x := 0; x < g.Columns; x++ {
			for _, r := range radii {
				if !ValidRadius(r) {
					return nil, errors.New("invalid radius")
				}
				key := ParticipantArea{Cell{x, y}, r}
				ids := []int{}
				for i, c := range cs {
					if g.MaxDistance(key.Cell, c.Point) <= float64(r)*1000 {
						ids = append(ids, i)
					}
				}
				idx.reachable[key] = ids
			}
		}
	}
	return idx, nil
}
func (i *Index) Reachable(p ParticipantArea) []int { return append([]int(nil), i.reachable[p]...) }
func (i *Index) CanReach(p ParticipantArea, id string) bool {
	// Reachable indices follow Intersections' stable ID order. Binary search needs
	// no additional per-cell map and preserves unknown-area/unknown-ID rejection.
	ids := i.reachable[p]
	n := sort.Search(len(ids), func(n int) bool { return i.Intersections[ids[n]].ID >= id })
	return n < len(ids) && i.Intersections[ids[n]].ID == id
}
func (i *Index) Closest(group []ParticipantArea) (Intersection, bool) {
	if len(group) == 0 {
		return Intersection{}, false
	}
	counts := map[int]int{}
	center := Point{}
	for _, p := range group {
		candidates, ok := i.reachable[p]
		if !ok {
			return Intersection{}, false
		}
		for _, n := range candidates {
			counts[n]++
		}
		c := i.Grid.Center(p.Cell)
		center[0] += c[0] / float64(len(group))
		center[1] += c[1] / float64(len(group))
	}
	best := -1
	distance := 0.0
	// Scan stable ID order so equal distances resolve deterministically.
	for n, c := range i.Intersections {
		if counts[n] != len(group) {
			continue
		}
		d := Distance(center, c.Point)
		if best < 0 || d < distance {
			best = n
			distance = d
		}
	}
	if best < 0 {
		return Intersection{}, false
	}
	return i.Intersections[best], true
}
