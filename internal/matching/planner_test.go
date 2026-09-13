package matching

import (
	"fmt"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"testing"
)

func fixture(t *testing.T) (Planner, []Signal, int64) {
	t.Helper()
	g, _ := geography.NewGrid(1000)
	c, _ := g.CellAt(geography.Point{19.818, 41.327})
	center := g.Center(c)
	intersection := geography.Intersection{ID: "near", Point: center}
	other := geography.Intersection{ID: "far", Point: geography.Point{center[0] + .01, center[1]}}
	index, e := geography.NewIndex(g, []geography.Intersection{other, intersection}, []float64{1, 3, 5})
	if e != nil {
		t.Fatal(e)
	}
	cfg, e := config.Load("../../config/simulation.yaml", true)
	if e != nil {
		t.Fatal(e)
	}
	now := int64(1234567890123)
	signals := []Signal{}
	for n := 0; n < 5; n++ {
		signals = append(signals, Signal{Hash: fmt.Sprint(n), Area: geography.ParticipantArea{Cell: c, RadiusKM: 3}, CreatedAt: now - int64(5-n)*1000, ExpiresAt: now + 30*60000})
	}
	return Planner{index, cfg.Matching}, signals, now
}
func TestNearestIntersectionAndOldestBoundedFounders(t *testing.T) {
	p, signals, now := fixture(t)
	proposal, ok := p.Propose(now, signals, nil)
	if !ok || proposal.Intersection.ID != "near" || len(proposal.Founders) != 3 {
		t.Fatal("missing automatic bounded formation")
	}
	for n, s := range proposal.Founders {
		if s.Hash != fmt.Sprint(n) {
			t.Fatal("oldest eligible signals not selected")
		}
	}
	if p.Deadline(now, proposal.Founders) != signals[0].ExpiresAt {
		t.Fatal("founder deadline not respected")
	}
}
func TestContinuousAvailabilityAtUnalignedTime(t *testing.T) {
	p, signals, now := fixture(t)
	if _, ok := p.Propose(now, signals, nil); !ok {
		t.Fatal("30-minute willingness did not immediately qualify")
	}
	for n := range signals {
		signals[n].ExpiresAt = now + 15*60000
	}
	if _, ok := p.Propose(now, signals, nil); !ok {
		t.Fatal("exact remaining boundary should qualify")
	}
	if _, ok := p.Propose(now+1, signals, nil); ok {
		t.Fatal("insufficient remaining availability counted")
	}
}
func TestOpenGatheringLateOfferAndDecline(t *testing.T) {
	p, signals, now := fixture(t)
	proposal, _ := p.Propose(now, signals, nil)
	open := []Gathering{{"existing", proposal.Intersection, now + 20*60000}}
	if _, ok := p.Propose(now, signals, open); ok {
		t.Fatal("formed competitor despite compatible open gathering")
	}
	for _, s := range signals {
		if g, ok := p.Offer(now, s, open); !ok || g.ID != "existing" {
			t.Fatal("new willingness cannot join current gathering")
		}
	}
	s := signals[0]
	s.Declined = map[string]bool{"existing": true}
	if _, ok := p.Offer(now, s, open); ok {
		t.Fatal("declined invitation offered again")
	}
	if s.ExpiresAt != signals[0].ExpiresAt {
		t.Fatal("decline changed willingness")
	}
	if _, ok := p.Offer(open[0].EndsAt-5*60000+1, signals[0], open); ok {
		t.Fatal("offered after cutoff")
	}
}
func TestMissingCommonIntersectionAndReservations(t *testing.T) {
	p, signals, now := fixture(t)
	signals[0].Assigned = true
	signals[1].ReservedUntil = now + 10000
	signals[2].ExpiresAt = now
	if _, ok := p.Propose(now, signals, nil); ok {
		t.Fatal("unavailable signals counted")
	}
	p.Index, _ = geography.NewIndex(p.Index.Grid, nil, []float64{1, 3, 5})
	for n := range signals {
		signals[n].Assigned = false
		signals[n].ReservedUntil = 0
		signals[n].ExpiresAt = now + 30*60000
	}
	if _, ok := p.Propose(now, signals, nil); ok {
		t.Fatal("invented meeting point without mapped intersection")
	}
}

// An independent small-cohort oracle enumerates triples and tests each map point,
// without using the reachability index or planner's candidate-ranking algorithm.
func TestPlannerAgainstExhaustiveSmallCohortOracle(t *testing.T) {
	p, _, now := fixture(t)
	for trial := 0; trial < 40; trial++ {
		input := []Signal{}
		for n := 0; n < 6; n++ {
			cell := geography.Cell{X: (trial + n*3) % 10, Y: (trial*3 + n) % 9}
			radius := []float64{.1, .5, 1, 3}[(trial+n)%4]
			input = append(input, Signal{Hash: fmt.Sprint(n), Area: geography.ParticipantArea{Cell: cell, RadiusKM: radius}, CreatedAt: now - int64(n+1), ExpiresAt: now + int64(10+(trial+n)%3*10)*60000})
		}
		// Include short radii in the public static index for this oracle fixture.
		index, err := geography.NewIndex(p.Index.Grid, p.Index.Intersections, []float64{.1, .5, 1, 3})
		if err != nil {
			t.Fatal(err)
		}
		p.Index = index
		exists := false
		for a := 0; a < 6; a++ {
			for b := a + 1; b < 6; b++ {
				for c := b + 1; c < 6; c++ {
					for _, point := range p.Index.Intersections {
						ok := true
						for _, member := range []Signal{input[a], input[b], input[c]} {
							if member.ExpiresAt < now+int64(p.Config.MinimumRemainingMinutes)*60000 || p.Index.Grid.MaxDistance(member.Area.Cell, point.Point) > member.Area.RadiusKM*1000 {
								ok = false
							}
						}
						exists = exists || ok
					}
				}
			}
		}
		proposal, ok := p.Propose(now, input, nil)
		if ok != exists {
			t.Fatalf("trial %d: oracle availability differs from planner", trial)
		}
		if ok {
			for _, member := range proposal.Founders {
				if p.Index.Grid.MaxDistance(member.Area.Cell, proposal.Intersection.Point) > member.Area.RadiusKM*1000 {
					t.Fatal("planner violated a founding radius")
				}
			}
		}
	}
}
