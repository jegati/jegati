package geography

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestImportStreetConnectivity(t *testing.T) {
	// Three legs at node 2 form a T; a split at node 3 is only a bend.
	points := map[int64]osmPoint{1: {41.32, 19.81}, 2: {41.32, 19.82}, 3: {41.32, 19.83}, 4: {41.33, 19.82}, 5: {41.32, 19.84}, 6: {41.32, 19.82}, 7: {41.31, 19.82}}
	way := func(id int64, ns []int64, tags map[string]string) osmElement {
		e := osmElement{Type: "way", ID: id, Nodes: ns, Tags: tags}
		for _, n := range ns {
			e.Geometry = append(e.Geometry, points[n])
		}
		return e
	}
	streets := []osmElement{way(10, []int64{1, 2, 3}, map[string]string{"highway": "residential", "name": "Rruga A"}), way(11, []int64{2, 4}, map[string]string{"highway": "secondary", "name": "Rruga B"}), way(12, []int64{3, 5}, map[string]string{"highway": "residential"})}
	encode := func(elements []osmElement, remark string) []byte {
		raw, _ := json.Marshal(map[string]any{"elements": elements, "remark": remark, "osm3s": map[string]string{"timestamp_osm_base": "2026-09-13T00:00:00Z"}})
		return raw
	}
	ds, roads, err := Import(encode(streets, ""), "test-v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Intersections) != 1 || ds.Intersections[0].ID != "node/2" || len(ds.Intersections[0].SourceIDs) != 3 || len(roads.Features) != 3 {
		t.Fatalf("unexpected junctions: %+v", ds)
	}
	reversed := []osmElement{streets[2], streets[1], streets[0]}
	again, _, err := Import(encode(reversed, ""), "test-v1")
	// Source checksum changes with input order; derived features must not.
	if err != nil || !reflect.DeepEqual(ds.Intersections, again.Intersections) {
		t.Fatal("unstable graph import")
	}
	t.Run("geometry crossing without connectivity", func(t *testing.T) {
		copyStreets := append([]osmElement(nil), streets...)
		copyStreets[1] = way(11, []int64{7, 6, 4}, map[string]string{"highway": "secondary"})
		if _, _, err := Import(encode(copyStreets, ""), "test-v1"); err == nil {
			t.Fatal("invented junction between disconnected roads")
		}
	})
	for _, tags := range []map[string]string{{"highway": "footway", "footway": "crossing"}, {"highway": "service", "service": "driveway"}, {"highway": "residential", "access": "private"}, {"highway": "residential", "foot": "no"}, {"highway": "residential", "bridge": "yes"}, {"highway": "residential", "tunnel": "yes"}, {"highway": "construction"}, {"highway": "motorway"}} {
		t.Run("exclude "+tags["highway"]+tags["access"]+tags["foot"]+tags["bridge"]+tags["tunnel"], func(t *testing.T) {
			copyStreets := append([]osmElement(nil), streets...)
			copyStreets[1] = way(11, []int64{2, 4}, tags)
			if _, _, err := Import(encode(copyStreets, ""), "test-v1"); err == nil {
				t.Fatal("ineligible road formed destination")
			}
		})
	}
	t.Run("duplicate mapped segment cannot add a leg", func(t *testing.T) {
		els := []osmElement{streets[0], way(13, []int64{1, 2, 3}, map[string]string{"highway": "residential"})}
		if _, _, err := Import(encode(els, ""), "test-v1"); err == nil {
			t.Fatal("duplicate segment formed junction")
		}
	})
	t.Run("four-way intersection", func(t *testing.T) {
		els := append(append([]osmElement(nil), streets...), way(14, []int64{2, 7}, map[string]string{"highway": "tertiary"}))
		d, _, err := Import(encode(els, ""), "test-v1")
		if err != nil || len(d.Intersections) != 1 {
			t.Fatal("four-way junction not normalized")
		}
	})
	t.Run("inconsistent shared geometry", func(t *testing.T) {
		els := append([]osmElement(nil), streets...)
		els[1] = way(11, []int64{2, 4}, map[string]string{"highway": "secondary"})
		els[1].Geometry[0].Lat += .001
		if _, _, err := Import(encode(els, ""), "test-v1"); err == nil {
			t.Fatal("accepted inconsistent geometry")
		}
	})
	if _, _, err := Import(encode(streets, "timeout"), "test-v1"); err == nil {
		t.Fatal("accepted incomplete extract")
	}
}
