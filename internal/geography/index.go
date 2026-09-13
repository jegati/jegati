package geography

import (
	"errors"
	"sort"
)

type Crossing struct {
	ID        string   `json:"id"`
	SourceIDs []string `json:"source_ids"`
	Point     Point    `json:"point"`
	Geometry  []Point  `json:"geometry,omitempty"`
	Label     string   `json:"label"`
}
type Dataset struct {
	Version         string     `json:"version"`
	SourceSHA256    string     `json:"source_sha256"`
	SourceTimestamp string     `json:"source_timestamp"`
	Crossings       []Crossing `json:"crossings"`
}
type ParticipantArea struct {
	Cell     Cell
	RadiusKM int
}
type Index struct {
	Grid      Grid
	Crossings []Crossing
	reachable map[ParticipantArea][]int
}

func NewIndex(g Grid, crossings []Crossing, radii []int) (*Index, error) {
	if len(crossings) > 100000 || len(radii) == 0 || len(radii) > 32 {
		return nil, errors.New("geography index outside bounds")
	}
	cs := append([]Crossing(nil), crossings...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].ID < cs[j].ID })
	for i, c := range cs {
		if !Inside(c.Point) || c.ID == "" || (i > 0 && cs[i-1].ID == c.ID) {
			return nil, errors.New("invalid crossing dataset")
		}
	}
	idx := &Index{g, cs, make(map[ParticipantArea][]int)}
	// Dataset and finite coarse grid bound this offline/startup computation.
	// Request handlers never recompute participant-to-participant distance pairs.
	for y := 0; y < g.Rows; y++ {
		for x := 0; x < g.Columns; x++ {
			for _, r := range radii {
				if r <= 0 || r > 20 {
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
	for _, n := range i.reachable[p] {
		if i.Crossings[n].ID == id {
			return true
		}
	}
	return false
}
func (i *Index) Closest(group []ParticipantArea) (Crossing, bool) {
	if len(group) == 0 {
		return Crossing{}, false
	}
	counts := map[int]int{}
	center := Point{}
	for _, p := range group {
		candidates, ok := i.reachable[p]
		if !ok {
			return Crossing{}, false
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
	for n, c := range i.Crossings {
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
		return Crossing{}, false
	}
	return i.Crossings[best], true
}
