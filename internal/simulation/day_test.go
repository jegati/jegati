package simulation

import (
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"math"
	"math/rand/v2"
	"reflect"
	"testing"
)

func dayFixture(t *testing.T) (config.Config, *geography.Index, DayScenario) {
	t.Helper()
	c, err := config.Load("../../config/gati.yaml", false)
	if err != nil {
		t.Fatal(err)
	}
	grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
	idx, err := geography.NewIndex(grid, []geography.Intersection{{ID: "synthetic", Point: geography.Point{19.818, 41.327}}}, []float64{3})
	if err != nil {
		t.Fatal(err)
	}
	return c, idx, DayScenario{Name: "test", Population: 1000, WillingFraction: .2, ReactiveFraction: .25, RadiusKM: 3, Distribution: "clustered", AvailabilityMinutes: 60, AcceptFraction: 1, ArriveFraction: 1, RenewalFraction: 1, DwellMinutes: 30, MapCheckMinutes: 60, WalkMetersSecond: 2, Detour: 1, Hours: 2, StepSeconds: 30}
}
func TestDayModelBehavior(t *testing.T) {
	c, idx, s := dayFixture(t)
	run := func(s DayScenario) DayResult {
		t.Helper()
		r, e := RunDay(c, idx, s, 42)
		if e != nil {
			t.Fatal(e)
		}
		return r
	}
	r := run(s)
	if !reflect.DeepEqual(r, run(s)) {
		t.Fatal("fixed seed is not reproducible")
	}
	if r.Metrics["gatherings_activated"] == 0 || r.Metrics["arrival_events"] == 0 || r.Metrics["gatherings_ever_confirmed"] == 0 {
		t.Fatal("positive control never gathers", r.Metrics)
	}
	if r.Metrics["unique_synthetic_arrivals"] > float64(s.Population) || r.Metrics["minutes_any_confirmed"] > float64(s.Hours*60) {
		t.Fatal("double-counted population or time")
	}
	for _, g := range r.Gatherings {
		if g.EndMinute-g.StartMinute > float64(c.Matching.MaximumGatheringMinutes) || g.ConfirmedMinutes > g.EndMinute-g.StartMinute {
			t.Fatal("late joins extended the frozen deadline")
		}
	}
	for _, f := range r.Frames {
		if f.Fresh > f.Going || f.Going > f.Active || f.Active > s.Population {
			t.Fatal("invalid funnel")
		}
		for _, g := range f.Public {
			if g.Going != 0 && g.Going < c.PublicActivity.MinimumCount || g.Here != 0 && g.Here < c.PublicActivity.MinimumCount {
				t.Fatal("unsuppressed public counts")
			}
		}
	}
	none := s
	none.WillingFraction = 0
	none.ReactiveFraction = 1
	empty := run(none)
	if empty.Metrics["reactive_sessions"] != 0 || empty.Metrics["gatherings_activated"] != 0 {
		t.Fatal("reactors created gatherings without a visible seed")
	}
	noArrival := s
	noArrival.ArriveFraction = 0
	absent := run(noArrival)
	if absent.Metrics["gatherings_activated"] == 0 || absent.Metrics["gatherings_ever_confirmed"] != 0 {
		t.Fatal("willingness counted as presence")
	}
	noReaction := s
	noReaction.ReactiveFraction = 0
	baseline := run(noReaction)
	if baseline.Metrics["reactive_sessions"] != 0 || baseline.Metrics["baseline_person_minutes"] != r.Metrics["baseline_person_minutes"] {
		t.Fatal("reaction changed the exogenous availability process")
	}
	// Nothing is publicly released if the member suppression threshold cannot be met.
	hidden := c
	hidden.PublicActivity.MinimumCount = 10000
	hiddenResult, e := RunDay(hidden, idx, s, 42)
	if e != nil {
		t.Fatal(e)
	}
	if hiddenResult.Metrics["reactive_sessions"] != 0 {
		t.Fatal("reactors discovered hidden gatherings")
	}
}
func TestDayStudyValidation(t *testing.T) {
	c, _, s := dayFixture(t)
	for _, change := range []func(*DayScenario){func(s *DayScenario) { s.Population = 50001 }, func(s *DayScenario) { s.StepSeconds = 0 }, func(s *DayScenario) { s.WillingFraction = 1 }, func(s *DayScenario) { s.ReactiveFraction = math.NaN() }, func(s *DayScenario) { s.AvailabilityMinutes = 15 }, func(s *DayScenario) { s.WalkMetersSecond = math.Inf(1) }, func(s *DayScenario) { s.RadiusKM = 2 }} {
		v := s
		change(&v)
		if v.Validate(c) == nil {
			t.Fatal("invalid assumptions accepted")
		}
	}
}
func TestDayStationaryIdleDistribution(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 43))
	total := int64(0)
	const n = 100000
	for i := 0; i < n; i++ {
		d := idleDelay(rng, .05, 3600000, 30000)
		if d <= 0 || d%30000 != 0 {
			t.Fatal("bad idle interval")
		}
		total += d
	}
	observed := 3600000 / (3600000 + float64(total)/n)
	if math.Abs(observed-.05) > .001 {
		t.Fatal("x percent is not concurrent availability", observed)
	}
}
