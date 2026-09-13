package geography

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type osmPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
type osmElement struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Lat      float64           `json:"lat"`
	Lon      float64           `json:"lon"`
	Nodes    []int64           `json:"nodes"`
	Geometry []osmPoint        `json:"geometry"`
	Tags     map[string]string `json:"tags"`
}
type Feature struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	Geometry   LineGeometry      `json:"geometry"`
}
type LineGeometry struct {
	Type        string  `json:"type"`
	Coordinates []Point `json:"coordinates"`
}
type Roads struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

func forbidden(tags map[string]string) bool {
	for _, key := range []string{"access", "foot"} {
		switch tags[key] {
		case "no", "private":
			return true
		}
	}
	if tags["crossing"] == "no" || tags["railway"] != "" {
		return true
	}
	for key, value := range tags {
		if value != "" && value != "no" && (key == "construction" || key == "disused" || key == "abandoned" || strings.HasPrefix(key, "construction:") || strings.HasPrefix(key, "disused:") || strings.HasPrefix(key, "abandoned:")) {
			return true
		}
	}
	switch tags["highway"] {
	case "construction", "proposed", "abandoned", "disused":
		return true
	}
	return false
}
func shortName(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > 120 {
		r = r[:120]
	}
	return string(r)
}

// Import consumes a bounded public Overpass extract, not an HTTP request. Retain
// only road geometry/names and crossing features in derived application assets.
func Import(raw []byte, version string) (Dataset, Roads, error) {
	var input struct {
		Elements []osmElement `json:"elements"`
		Remark   string       `json:"remark"`
		Meta     struct {
			Timestamp string `json:"timestamp_osm_base"`
		} `json:"osm3s"`
	}
	if len(raw) > 64<<20 {
		return Dataset{}, Roads{}, errors.New("map source exceeds size bound")
	}
	if err := json.Unmarshal(raw, &input); err != nil || input.Remark != "" || len(input.Elements) == 0 || input.Meta.Timestamp == "" {
		return Dataset{}, Roads{}, errors.New("invalid, empty or incomplete map extract")
	}
	sort.Slice(input.Elements, func(i, j int) bool {
		a, b := input.Elements[i], input.Elements[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		return a.ID < b.ID
	})
	nodes := map[int64]*Crossing{}
	roadNames := map[int64]string{}
	seen := map[string]bool{}
	blockedNodes := map[int64]bool{}
	for _, e := range input.Elements {
		if e.Type == "way" && forbidden(e.Tags) {
			for _, n := range e.Nodes {
				blockedNodes[n] = true
			}
		}
	}
	roads := Roads{Type: "FeatureCollection", Features: []Feature{}}
	for _, e := range input.Elements {
		id := fmt.Sprintf("%s/%d", e.Type, e.ID)
		if seen[id] {
			return Dataset{}, Roads{}, errors.New("duplicate map element")
		}
		seen[id] = true
		if e.ID <= 0 || forbidden(e.Tags) || (e.Type == "node" && blockedNodes[e.ID]) {
			continue
		}
		if e.Type == "node" && e.Tags["highway"] == "crossing" && Inside(Point{e.Lon, e.Lat}) {
			nodes[e.ID] = &Crossing{ID: id, SourceIDs: []string{id}, Point: Point{e.Lon, e.Lat}, Label: "Vendkalim këmbësorësh"}
		}
		if e.Type == "way" && e.Tags["highway"] != "" {
			name := shortName(e.Tags["name"])
			if name != "" {
				for _, n := range e.Nodes {
					if roadNames[n] == "" {
						roadNames[n] = name
					}
				}
			}
			part := 0
			coordinates := []Point{}
			flush := func() {
				if len(coordinates) >= 2 {
					roads.Features = append(roads.Features, Feature{"Feature", fmt.Sprintf("%s/%d", id, part), map[string]string{"name": name, "highway": e.Tags["highway"]}, LineGeometry{"LineString", coordinates}})
					part++
				}
				coordinates = []Point{}
			}
			for _, p := range e.Geometry {
				q := Point{p.Lon, p.Lat}
				if Inside(q) {
					coordinates = append(coordinates, q)
				} else {
					flush()
				}
			}
			flush()
		}
	}
	ways := []Crossing{}
	for _, e := range input.Elements {
		if e.Type != "way" || e.Tags["highway"] != "footway" || e.Tags["footway"] != "crossing" || forbidden(e.Tags) || len(e.Geometry) < 2 {
			continue
		}
		geom := []Point{}
		valid := true
		for _, p := range e.Geometry {
			q := Point{p.Lon, p.Lat}
			if !Inside(q) {
				valid = false
				break
			}
			geom = append(geom, q)
		}
		if !valid {
			continue
		}
		id := fmt.Sprintf("way/%d", e.ID)
		linked := false
		for _, n := range e.Nodes {
			if c := nodes[n]; c != nil {
				linked = true
				c.SourceIDs = append(c.SourceIDs, id)
				if len(c.Geometry) == 0 {
					c.Geometry = geom
				}
			}
		}
		if linked {
			continue
		} // Node and crossing-way representations become one record.
		midpoint := Point{}
		for _, p := range geom {
			midpoint[0] += p[0] / float64(len(geom))
			midpoint[1] += p[1] / float64(len(geom))
		}
		ways = append(ways, Crossing{ID: id, SourceIDs: []string{id}, Point: midpoint, Geometry: geom, Label: "Vendkalim këmbësorësh"})
	}
	for id, c := range nodes {
		if name := roadNames[id]; name != "" {
			c.Label += " — " + name
		}
		ways = append(ways, *c)
	}
	sort.Slice(ways, func(i, j int) bool { return ways[i].ID < ways[j].ID })
	if len(ways) == 0 {
		return Dataset{}, Roads{}, errors.New("extract has no eligible crossings")
	}
	sum := sha256.Sum256(raw)
	return Dataset{version, hex.EncodeToString(sum[:]), input.Meta.Timestamp, ways}, roads, nil
}
