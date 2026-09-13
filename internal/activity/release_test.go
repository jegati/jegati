package activity

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
)

func fixture(t *testing.T) (config.Config, []Participant, []Gathering) {
	t.Helper()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	v := []Participant{}
	for i := 0; i < 60; i++ {
		v = append(v, Participant{Hash: fmt.Sprint(i), Cell: "tirana-v1:100:42:42", Gathering: "opaque-event", State: "invited", ExpiresAt: 2000000, GatheringUntil: 2000000})
	}
	g := []Gathering{{ID: "opaque-event", Cell: "tirana-v1:1000:5:5", State: "jemi_ketu", EndsAt: 2000000}}
	return c, v, g
}
func TestCountSemanticsAndCellOnlySerialization(t *testing.T) {
	c, v, g := fixture(t)
	for i := 0; i < 50; i++ {
		v[i].State = "going"
	}
	for i := 0; i < 20; i++ {
		v[i].State = "here"
		v[i].ArrivalUntil = 1000000
	}
	r, data, e := Build(c, 600000, 601000, v, g)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Areas) != 1 || r.Areas[0].Willing != 50 || r.Areas[0].Cell != "tirana-v1:1000:4:4" || len(r.Gatherings) != 1 || r.Gatherings[0].Going != 50 || r.Gatherings[0].Here != 20 || r.Gatherings[0].State != "jemi_ketu" {
		t.Fatalf("bad counts: %+v", r)
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(data, &fields)
	for _, forbidden := range []string{`"intersection":`, `"point":`, `"label":`, `"latitude":`, `"longitude":`, "hash\"", "tirana-v1:100:42:42", "members", "remainder"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("public private field: %s", forbidden)
		}
	}
	if r.ReleaseAt != 1200000 || r.ExpiresAt != 1500000 {
		t.Fatal("wrong release/retention clock")
	}
	// Expiration removes willingness; declining removes gathering membership only;
	// a stale arrival retains going but cannot imply current presence.
	v[0].ExpiresAt = 601000
	v[1].State = "gati"
	v[1].Gathering = ""
	v[2].ArrivalUntil = 601000
	r, _, e = Build(c, 600000, 601000, v, g)
	if e != nil {
		t.Fatal(e)
	}
	if r.Areas[0].Willing != 50 || r.Gatherings[0].Going != 20 || r.Gatherings[0].Here != 0 || r.Gatherings[0].State != "jemi_gati" {
		t.Fatal("expiry or state semantics wrong")
	}
}
func TestSuppressionThresholdsAndAcceptedInference(t *testing.T) {
	c, v, g := fixture(t)
	for _, n := range []int{0, 1, 19, 20, 49, 50, 59} {
		r, data, e := Build(c, 600000, 600000, v[:n], g)
		if e != nil {
			t.Fatal(e)
		}
		if n < 20 {
			if len(r.Areas) != 0 || len(r.Gatherings) != 0 {
				t.Fatal("small group exposed")
			}
		} else if r.Areas[0].Willing != Bucket(n, c.PublicActivity.CountBuckets) {
			t.Fatal("wrong bucket")
		}
		if strings.Contains(string(data), "\"here\":") || strings.Contains(string(data), "\"going\":") {
			t.Fatal("small counts serialized")
		}
	}
	// Deliberately records an accepted counterexample, not an anonymity test pass:
	// 49 controlled inputs + one otherwise isolated target change the public bucket.
	before, _, _ := Build(c, 600000, 600000, v[:49], g)
	after, _, _ := Build(c, 900000, 900000, v[:50], g)
	if before.Areas[0].Willing != 20 || after.Areas[0].Willing != 50 {
		t.Fatal("inference fixture changed; review documented limitation")
	}
}
func TestCaptureFailureAndDeterminism(t *testing.T) {
	c, v, g := fixture(t)
	_, data, e := Build(c, 600000, 600000, v, g)
	if e != nil {
		t.Fatal(e)
	}
	for i, j := 0, len(v)-1; i < j; i, j = i+1, j-1 {
		v[i], v[j] = v[j], v[i]
	}
	_, other, e := Build(c, 600000, 600000, v, g)
	if e != nil || string(data) != string(other) {
		t.Fatal("order changed canonical bytes")
	}
	if _, _, e = Build(c, 600000, 631000, v, g); e == nil {
		t.Fatal("unbounded capture accepted")
	}
	if _, _, e = Build(c, 600000, 600000, append(v, v[0]), g); e == nil {
		t.Fatal("duplicate credential accepted")
	}
	c.PublicActivity.MaxSnapshotBytes = 1
	if _, _, e = Build(c, 600000, 600000, v, g); e == nil {
		t.Fatal("oversize accepted")
	}
}
func TestPrivateCellsMapIntoPublicCells(t *testing.T) {
	private, _ := geography.NewGrid(100)
	public, _ := geography.NewGrid(1000)
	for y := 0; y < private.Rows; y++ {
		for x := 0; x < private.Columns; x++ {
			id, e := PublicCell(private, public, private.ID(geography.Cell{X: x, Y: y}))
			if e != nil {
				t.Fatal(e)
			}
			cell, e := public.Parse(id)
			if e != nil || cell.X != x/10 || cell.Y != y/10 {
				t.Fatal("public mapping boundary")
			}
		}
	}
}

func TestNestedCountsRetainDocumentedInferenceWithoutExactRemainder(t *testing.T) {
	c, v, g := fixture(t)
	v = v[:50]
	for i := range v {
		v[i].State = "here"
		v[i].ArrivalUntil = 2000000
	}
	v[49].State = "going"
	v[49].ArrivalUntil = 0
	r, data, e := Build(c, 600000, 600000, v, g)
	if e != nil {
		t.Fatal(e)
	}
	// An observer with knowledge of its own 49 arrived credentials may infer the
	// additional going contribution. User acceptance permits these bucketed views.
	if r.Gatherings[0].Going != 50 || r.Gatherings[0].Here != 20 {
		t.Fatal("nested bucket fixture changed")
	}
	if strings.Contains(string(data), `"remainder"`) || strings.Contains(string(data), `"here":49`) {
		t.Fatal("exact complementary count exposed")
	}
}

func TestExpiredPresenceCannotRetainStrongerPrivateThresholdLabel(t *testing.T) {
	c, v, g := fixture(t)
	c.Arrivals.ConfirmationCount = 50
	for i := range v {
		v[i].State = "here"
		v[i].ArrivalUntil = 2000000
	}
	for i := 20; i < len(v); i++ {
		v[i].ArrivalUntil = 600000
	}
	r, _, e := Build(c, 600000, 601000, v, g)
	if e != nil {
		t.Fatal(e)
	}
	if r.Gatherings[0].Here != 20 || r.Gatherings[0].State != "jemi_gati" {
		t.Fatal("stale private state falsely retained confirmed public presence")
	}
}
