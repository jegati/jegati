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
	crossing := geography.Crossing{ID: "near", Point: center}
	other := geography.Crossing{ID: "far", Point: geography.Point{center[0] + .01, center[1]}}
	index, e := geography.NewIndex(g, []geography.Crossing{other, crossing}, []int{1, 3, 5})
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
func TestNearestCrossingAndOldestBoundedFounders(t *testing.T) {
	p, signals, now := fixture(t)
	proposal, ok := p.Propose(now, signals, nil)
	if !ok || proposal.Crossing.ID != "near" || len(proposal.Founders) != 3 {
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
	open := []Gathering{{"existing", proposal.Crossing, now + 20*60000}}
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
func TestMissingCommonCrossingAndReservations(t *testing.T) {
	p, signals, now := fixture(t)
	signals[0].Assigned = true
	signals[1].ReservedUntil = now + 10000
	signals[2].ExpiresAt = now
	if _, ok := p.Propose(now, signals, nil); ok {
		t.Fatal("unavailable signals counted")
	}
	p.Index, _ = geography.NewIndex(p.Index.Grid, nil, []int{1, 3, 5})
	for n := range signals {
		signals[n].Assigned = false
		signals[n].ReservedUntil = 0
		signals[n].ExpiresAt = now + 30*60000
	}
	if _, ok := p.Propose(now, signals, nil); ok {
		t.Fatal("invented meeting point without mapped crossing")
	}
}
