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

// eligibleStreet restricts destinations to mapped public street junctions. Road
// paths, driveways, ramps and grade-separated roads are not gathering landmarks.
func eligibleStreet(tags map[string]string) bool {
	if forbidden(tags) || (tags["bridge"] != "" && tags["bridge"] != "no") ||
		(tags["tunnel"] != "" && tags["tunnel"] != "no") || tags["motorroad"] == "yes" || tags["area"] == "yes" {
		return false
	}
	switch tags["highway"] {
	case "primary", "secondary", "tertiary", "unclassified", "residential", "living_street", "pedestrian":
		return true
	}
	return false
}

// Import consumes a bounded archived public extract, never a participant query.
// A junction has at least three distinct adjacent street nodes. Shared node IDs,
// not intersecting geometry or split-way counts, establish road connectivity.
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
	type junction struct {
		point          Point
		neighbors      map[int64]bool
		sources, names map[string]bool
	}
	nodes := map[int64]*junction{}
	blocked := map[int64]bool{}
	seen := map[string]bool{}
	roads := Roads{Type: "FeatureCollection", Features: []Feature{}}
	for _, e := range input.Elements {
		id := fmt.Sprintf("%s/%d", e.Type, e.ID)
		if seen[id] {
			return Dataset{}, Roads{}, errors.New("duplicate map element")
		}
		seen[id] = true
		if e.ID <= 0 {
			return Dataset{}, Roads{}, errors.New("invalid map element ID")
		}
		if e.Type == "node" && forbidden(e.Tags) {
			blocked[e.ID] = true
		}
		// Conservatively omit a junction touching an explicitly prohibited street.
		if e.Type == "way" && (forbidden(e.Tags) || e.Tags["highway"] == "motorway" || e.Tags["highway"] == "trunk" || e.Tags["motorroad"] == "yes" || (e.Tags["bridge"] != "" && e.Tags["bridge"] != "no") || (e.Tags["tunnel"] != "" && e.Tags["tunnel"] != "no")) {
			for _, n := range e.Nodes {
				blocked[n] = true
			}
		}
		if e.Type != "way" || e.Tags["highway"] == "" || forbidden(e.Tags) {
			continue
		}
		name := shortName(e.Tags["name"])
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
		if !eligibleStreet(e.Tags) {
			continue
		}
		if len(e.Nodes) < 2 || len(e.Nodes) != len(e.Geometry) {
			return Dataset{}, Roads{}, errors.New("street node/geometry mismatch")
		}
		for i, n := range e.Nodes {
			if n <= 0 {
				return Dataset{}, Roads{}, errors.New("invalid street node ID")
			}
			point := Point{e.Geometry[i].Lon, e.Geometry[i].Lat}
			if !Inside(point) {
				continue
			}
			j := nodes[n]
			if j == nil {
				j = &junction{point, map[int64]bool{}, map[string]bool{}, map[string]bool{}}
				nodes[n] = j
			}
			if j.point != point {
				return Dataset{}, Roads{}, errors.New("inconsistent shared street node geometry")
			}
			for _, adjacent := range []int{i - 1, i + 1} {
				if adjacent >= 0 && adjacent < len(e.Nodes) && e.Nodes[adjacent] != n {
					j.neighbors[e.Nodes[adjacent]] = true
				}
			}
			j.sources[id] = true
			if name != "" {
				j.names[name] = true
			}
		}
	}
	intersections := []Intersection{}
	for n, j := range nodes {
		if blocked[n] || len(j.neighbors) < 3 {
			continue
		}
		id := fmt.Sprintf("node/%d", n)
		sources := []string{id}
		names := []string{}
		for source := range j.sources {
			sources = append(sources, source)
		}
		sort.Strings(sources)
		for name := range j.names {
			names = append(names, name)
		}
		sort.Strings(names)
		label := "Kryqëzim rrugësh"
		if len(names) > 0 {
			label += " — " + shortName(strings.Join(names, " / "))
		}
		intersections = append(intersections, Intersection{ID: id, SourceIDs: sources, Point: j.point, Label: label})
	}
	sort.Slice(intersections, func(i, j int) bool { return intersections[i].ID < intersections[j].ID })
	if len(intersections) == 0 {
		return Dataset{}, Roads{}, errors.New("extract has no eligible intersections")
	}
	sum := sha256.Sum256(raw)
	return Dataset{version, hex.EncodeToString(sum[:]), input.Meta.Timestamp, intersections}, roads, nil
}
