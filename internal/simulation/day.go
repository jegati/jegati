package simulation

// A synthetic, offline behavioral model. It reuses the production geographic
// planner and aggregate builder, but does not emulate HTTP, Valkey or real people.
import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
)

type DayScenario struct {
	Name                string  `json:"name"`
	Population          int     `json:"population"`
	WillingFraction     float64 `json:"willing_fraction"`
	ReactiveFraction    float64 `json:"reactive_fraction"`
	RadiusKM            float64 `json:"radius_km"`
	Distribution        string  `json:"distribution"`
	AvailabilityMinutes int     `json:"availability_minutes"`
	AcceptFraction      float64 `json:"accept_fraction"`
	ArriveFraction      float64 `json:"arrive_fraction"`
	RenewalFraction     float64 `json:"renewal_fraction"`
	DwellMinutes        int     `json:"dwell_minutes"`
	MapCheckMinutes     int     `json:"map_check_minutes"`
	WalkMetersSecond    float64 `json:"walk_meters_second"`
	Detour              float64 `json:"detour"`
	Hours               int     `json:"hours"`
	StepSeconds         int     `json:"step_seconds"`
}

func (s DayScenario) Validate(c config.Config) error {
	if s.Name == "" || s.Population < 1 || s.Population > 50000 || s.Hours < 1 || s.Hours > 24 || !slices.Contains([]int{10, 15, 30}, s.StepSeconds) || s.MapCheckMinutes < 1 || s.MapCheckMinutes > 120 || s.DwellMinutes < 1 || s.DwellMinutes > 120 {
		return errors.New("invalid bounded day study")
	}
	if !slices.Contains(c.Availability.ChoicesMinutes, s.AvailabilityMinutes) || !slices.Contains(c.Geography.TravelRadiusChoicesKm, s.RadiusKM) {
		return errors.New("study choices absent from app config")
	}
	if s.Distribution != "uniform" && s.Distribution != "clustered" {
		return errors.New("invalid distribution")
	}
	for _, p := range []float64{s.WillingFraction, s.ReactiveFraction, s.AcceptFraction, s.ArriveFraction, s.RenewalFraction} {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return errors.New("invalid probability")
		}
	}
	if c.PublicActivity.DelayEpochs != 0 || c.PublicActivity.ReleaseSeconds > s.StepSeconds {
		return errors.New("day model requires zero extra delay and publication no slower than its time step")
	}
	if s.WillingFraction >= 1 || !(s.WalkMetersSecond >= 0.5 && s.WalkMetersSecond <= 2) || !(s.Detour >= 1 && s.Detour <= 3) {
		return errors.New("invalid timing assumptions")
	}
	return nil
}

type DayFrame struct {
	Minute    float64          `json:"minute"`
	Baseline  int              `json:"baseline"`
	Active    int              `json:"active"`
	Going     int              `json:"going"`
	Fresh     int              `json:"fresh"`
	Open      int              `json:"open"`
	Confirmed int              `json:"confirmed"`
	Public    []activity.Event `json:"public"`
	Areas     []activity.Area  `json:"areas"`
}
type DayGathering struct {
	ID               string  `json:"id"`
	Cell             string  `json:"cell"`
	StartMinute      float64 `json:"start_minute"`
	EndMinute        float64 `json:"end_minute"`
	ConfirmedMinutes float64 `json:"confirmed_minutes"`
	PeakFresh        int     `json:"peak_fresh"`
}
type DayResult struct {
	Scenario   DayScenario        `json:"scenario"`
	Seed       uint64             `json:"seed"`
	Metrics    map[string]float64 `json:"metrics"`
	Frames     []DayFrame         `json:"frames"`
	Gatherings []DayGathering     `json:"gatherings"`
}
type dayActor struct {
	signal                                              matching.Signal
	baseline                                            *rand.Rand
	response                                            *rand.Rand
	nextStart, freeUntil                                int64
	group                                               int // -1 means unassigned
	state                                               string
	responseAt, arrivalAt, leaveAt, freshUntil, renewAt int64
	mapPhase                                            int64
	reactive                                            bool
}
type dayGathering struct {
	gathering      matching.Gathering
	founders       []int
	readyAt        int64
	active, failed bool
	sufficientAt   int64
	confirmed      bool
	report         DayGathering
}

func idleDelay(r *rand.Rand, p float64, duration, step int64) int64 {
	if p == 0 {
		return math.MaxInt64 / 4
	}
	// Geometric idle time (minimum one step), with E[idle]=D*(1-p)/p.
	q := float64(step) / float64(duration) * p / (1 - p)
	if q >= 1 {
		return step
	}
	return (int64(math.Floor(math.Log1p(-r.Float64())/math.Log1p(-q))) + 1) * step
}
func roundUp(v, step int64) int64 { return ((v + step - 1) / step) * step }

// RunDay never opens a network connection or uses a live store. Times are rounded
// up to the study step; this is an approximation of asynchronous application time.
func RunDay(c config.Config, index *geography.Index, s DayScenario, seed uint64) (DayResult, error) {
	out := DayResult{Scenario: s, Seed: seed, Metrics: map[string]float64{}, Frames: []DayFrame{}, Gatherings: []DayGathering{}}
	if err := s.Validate(c); err != nil {
		return out, err
	}
	planner := matching.Planner{Index: index, Config: c.Matching}
	grid := index.Grid
	publicGrid, err := geography.NewGrid(c.PublicActivity.AreaSizeMeters)
	if err != nil {
		return out, err
	}
	step := int64(s.StepSeconds) * 1000
	duration := int64(s.AvailabilityMinutes) * 60000
	horizon := int64(s.Hours) * 3600000
	mapPeriod := int64(s.MapCheckMinutes) * 60000
	actors := make([]dayActor, s.Population)
	for i := range actors {
		placement := rand.New(rand.NewPCG(seed, uint64(i)+0x706c616365))
		p := geography.Point{grid.West + placement.Float64()*(grid.East-grid.West), grid.South + placement.Float64()*(grid.North-grid.South)}
		if s.Distribution == "clustered" {
			centers := []geography.Point{{19.818, 41.327}, {19.822, 41.307}, {19.799, 41.335}}
			center := centers[placement.IntN(len(centers))]
			p = geography.Point{center[0] + (placement.Float64()-.5)*.025, center[1] + (placement.Float64()-.5)*.02}
		}
		cell, _ := grid.CellAt(p)
		a := dayActor{signal: matching.Signal{Hash: fmt.Sprintf("synthetic-%06d", i), Area: geography.ParticipantArea{Cell: cell, RadiusKM: s.RadiusKM}}, baseline: rand.New(rand.NewPCG(seed, uint64(i)+0x62617365)), response: rand.New(rand.NewPCG(seed, uint64(i)+0x72657370)), group: -1, mapPhase: int64(placement.IntN(int(mapPeriod/step))) * step}
		if a.baseline.Float64() < s.WillingFraction {
			a.freeUntil = int64(a.baseline.IntN(int(duration/step))+1) * step
			a.signal.CreatedAt = a.freeUntil - duration
			a.signal.ExpiresAt = a.freeUntil
			a.state = "willing"
			out.Metrics["baseline_sessions"]++
		}
		a.nextStart = a.freeUntil + idleDelay(a.baseline, s.WillingFraction, duration, step)
		actors[i] = a
	}
	var groups []dayGathering
	// Kept separately so modeled actor discovery consults released, suppressed data.
	var released, staged activity.Release
	visible := map[string]bool{}
	activatedCells, confirmedCells := map[string]bool{}, map[string]bool{}
	uniqueArrival := map[int]bool{}
	openList := func(now int64) []matching.Gathering {
		result := []matching.Gathering{}
		for _, g := range groups {
			if g.active && g.gathering.EndsAt > now {
				result = append(result, g.gathering)
			}
		}
		return result
	}
	groupIndex := func(id string) int {
		for i := len(groups) - 1; i >= 0; i-- {
			if groups[i].gathering.ID == id {
				return i
			}
		}
		return -1
	}
	assign := func(a *dayActor, gid int, now int64, direct bool) {
		a.group = gid
		a.state = "invited"
		a.signal.Assigned = true
		a.signal.AssignedUntil = groups[gid].gathering.EndsAt
		a.responseAt = now + roundUp(int64(5+a.response.IntN(56))*1000, step)
		a.reactive = direct
		out.Metrics["invitations"]++
	}
	startSession := func(a *dayActor, now, until int64) {
		a.signal.CreatedAt = now
		a.signal.ExpiresAt = until
		a.signal.Assigned = false
		a.signal.ReservedUntil = 0
		a.signal.Declined = map[string]bool{}
		a.group = -1
		a.state = "willing"
		a.freshUntil = 0
		a.reactive = false
	}
	for now := int64(0); now < horizon; now += step {
		if staged.ReleaseAt <= now && staged.Version != 0 {
			released = staged
		}
		visible = map[string]bool{}
		if now < released.ExpiresAt {
			for _, g := range released.Gatherings {
				if g.EndsAt > now {
					visible[g.ID] = true
				}
			}
		}
		for i := range actors {
			a := &actors[i]
			if now >= a.nextStart {
				a.freeUntil = now + duration
				a.nextStart = a.freeUntil + idleDelay(a.baseline, s.WillingFraction, duration, step)
				if a.signal.ExpiresAt <= now {
					startSession(a, now, a.freeUntil)
					out.Metrics["baseline_sessions"]++
				}
			}
			if a.signal.ExpiresAt <= now {
				a.group = -1
				a.state = ""
				a.signal.Assigned = false
				continue
			}
			if a.group >= 0 && groups[a.group].gathering.EndsAt <= now {
				a.group = -1
				a.state = "willing"
				a.signal.Assigned = false
				a.freshUntil = 0
			}
			if a.group < 0 || !groups[a.group].active {
				continue
			}
			g := &groups[a.group]
			if a.state == "invited" && a.responseAt <= now {
				minimum := now + int64(c.Matching.LateJoinMinRemainingMinutes)*60000
				if a.signal.ExpiresAt < minimum || g.gathering.EndsAt < minimum {
					a.responseAt = math.MaxInt64
					out.Metrics["admission_cutoff_missed"]++
					continue
				}
				if a.reactive || a.response.Float64() < s.AcceptFraction {
					a.state = "going"
					out.Metrics["going_decisions"]++
					a.arrivalAt = math.MaxInt64
					if a.response.Float64() < s.ArriveFraction {
						distance := geography.Distance(grid.Center(a.signal.Area.Cell), g.gathering.Intersection.Point)
						a.arrivalAt = now + max(step, roundUp(int64(math.Ceil(distance*s.Detour/s.WalkMetersSecond))*1000, step))
					}
				} else {
					if a.signal.Declined == nil {
						a.signal.Declined = map[string]bool{}
					}
					a.signal.Declined[g.gathering.ID] = true
					a.group = -1
					a.signal.Assigned = false
					a.state = "willing"
					out.Metrics["declines"]++
				}
			}
			if a.state == "going" && a.arrivalAt <= now && now < g.gathering.EndsAt {
				a.state = "here"
				a.arrivalAt = math.MaxInt64
				a.leaveAt = min(now+int64(s.DwellMinutes)*60000, a.signal.ExpiresAt, g.gathering.EndsAt)
				a.freshUntil = min(now+int64(c.Arrivals.FreshnessMinutes)*60000, a.leaveAt)
				a.renewAt = max(now+step, a.freshUntil-int64(c.Arrivals.RenewalWindowSeconds)*1000)
				out.Metrics["arrival_events"]++
				uniqueArrival[i] = true
			}
			if a.state == "here" {
				if now >= a.leaveAt {
					a.state = "going"
					a.freshUntil = 0
					out.Metrics["departures"]++
				} else if now >= a.renewAt && now < a.freshUntil && a.freshUntil < a.leaveAt {
					if a.response.Float64() < s.RenewalFraction {
						a.freshUntil = min(now+int64(c.Arrivals.FreshnessMinutes)*60000, a.leaveAt)
						out.Metrics["renewals"]++
						a.renewAt = max(now+step, a.freshUntil-int64(c.Arrivals.RenewalWindowSeconds)*1000)
					} else {
						a.renewAt = math.MaxInt64
					}
				}
				if now >= a.freshUntil {
					a.state = "going"
				}
			}
		}
		// Finish stable reservations. Founder expiry is rechecked at activation.
		for gi := range groups {
			g := &groups[gi]
			if g.active || g.failed || now < g.readyAt {
				continue
			}
			founders := []matching.Signal{}
			for _, ai := range g.founders {
				a := &actors[ai]
				if a.group == gi && a.signal.ExpiresAt >= now+int64(c.Matching.MinimumRemainingMinutes)*60000 {
					founders = append(founders, a.signal)
				}
			}
			if len(founders) != c.Matching.ActivationCount {
				g.failed = true
				for _, ai := range g.founders {
					if actors[ai].group == gi {
						actors[ai].group = -1
						actors[ai].signal.ReservedUntil = 0
					}
				}
				out.Metrics["failed_reservations"]++
				continue
			}
			g.active = true
			g.gathering.EndsAt = planner.Deadline(now, founders)
			pc, _ := publicGrid.CellAt(g.gathering.Intersection.Point)
			g.report = DayGathering{ID: g.gathering.ID, Cell: publicGrid.ID(pc), StartMinute: float64(now) / 60000, EndMinute: float64(g.gathering.EndsAt) / 60000}
			activatedCells[g.report.Cell] = true
			for _, ai := range g.founders {
				assign(&actors[ai], gi, now, false)
			}
		}
		open := openList(now)
		for i := range actors {
			a := &actors[i]
			if a.signal.ExpiresAt <= now {
				if now%mapPeriod != a.mapPhase {
					continue
				}
				hypothetical := a.signal
				hypothetical.ExpiresAt = now + duration
				hypothetical.Declined = nil
				candidates := []matching.Gathering{}
				for _, g := range open {
					if visible[g.ID] {
						candidates = append(candidates, g)
					}
				}
				offer, ok := planner.Offer(now, hypothetical, candidates)
				if !ok {
					continue
				}
				out.Metrics["eligible_map_exposures"]++
				if a.response.Float64() >= s.ReactiveFraction {
					continue
				}
				startSession(a, now, now+duration)
				out.Metrics["reactive_sessions"]++
				assign(a, groupIndex(offer.ID), now, true)
			} else if a.group < 0 {
				if offer, ok := planner.Offer(now, a.signal, open); ok {
					assign(a, groupIndex(offer.ID), now, false)
				}
			}
		}
		// The actual production planner handles whole-cell reachability, oldest cohorts,
		// nearest mapped crossroad and preference for open/pending gatherings.
		signals := []matching.Signal{}
		for i := range actors {
			if actors[i].signal.ExpiresAt > now {
				signals = append(signals, actors[i].signal)
			}
		}
		for _, g := range groups {
			if !g.active && !g.failed {
				open = append(open, g.gathering)
			}
		}
		for n := 0; n < c.Matching.CandidateBatchSize; n++ {
			proposal, ok := planner.Propose(now, signals, open)
			if !ok {
				break
			}
			ready := now + max(step, roundUp(int64(c.Matching.ActivationStabilitySeconds)*1000, step))
			g := dayGathering{gathering: matching.Gathering{ID: fmt.Sprintf("model-%05d", len(groups)), Intersection: proposal.Intersection, EndsAt: ready + int64(c.Matching.MinimumRemainingMinutes)*60000}, readyAt: ready}
			selected := map[string]bool{}
			for _, f := range proposal.Founders {
				selected[f.Hash] = true
			}
			for i := range actors {
				if selected[actors[i].signal.Hash] {
					actors[i].signal.ReservedUntil = ready + step
					actors[i].group = len(groups)
					g.founders = append(g.founders, i)
				}
			}
			for i := range signals {
				if selected[signals[i].Hash] {
					signals[i].ReservedUntil = ready + step
				}
			}
			groups = append(groups, g)
			open = append(open, g.gathering)
		}
		frame := DayFrame{Minute: float64(now) / 60000}
		participants := []activity.Participant{}
		fresh := make([]int, len(groups))
		for i := range actors {
			a := &actors[i]
			if now < a.freeUntil {
				frame.Baseline++
			}
			if a.signal.ExpiresAt <= now {
				continue
			}
			frame.Active++
			p := activity.Participant{Hash: a.signal.Hash, Cell: grid.ID(a.signal.Area.Cell), ExpiresAt: a.signal.ExpiresAt, State: a.state}
			if a.group >= 0 && groups[a.group].active && groups[a.group].gathering.EndsAt > now {
				p.Gathering = groups[a.group].gathering.ID
				p.GatheringUntil = groups[a.group].gathering.EndsAt
				p.ArrivalUntil = a.freshUntil
				if a.state == "going" || a.state == "here" {
					frame.Going++
				}
				if a.state == "here" && a.freshUntil > now {
					frame.Fresh++
					fresh[a.group]++
				}
			}
			participants = append(participants, p)
		}
		publicGroups := []activity.Gathering{}
		for gi := range groups {
			g := &groups[gi]
			if !g.active || g.gathering.EndsAt <= now {
				continue
			}
			frame.Open++
			if fresh[gi] >= c.Arrivals.ConfirmationCount {
				if g.sufficientAt == 0 {
					g.sufficientAt = now
				}
				g.confirmed = now >= g.sufficientAt+int64(c.Arrivals.ConfirmationStabilitySeconds)*1000
			} else {
				g.sufficientAt = 0
				g.confirmed = false
			}
			state := "jemi_gati"
			if g.confirmed {
				frame.Confirmed++
				g.report.ConfirmedMinutes += float64(step) / 60000
				confirmedCells[g.report.Cell] = true
				state = "jemi_ketu"
			}
			g.report.PeakFresh = max(g.report.PeakFresh, fresh[gi])
			publicGroups = append(publicGroups, activity.Gathering{ID: g.gathering.ID, Cell: g.report.Cell, EndsAt: g.gathering.EndsAt, State: state})
		}
		staged, _, err = activity.Build(c, now, now, participants, publicGroups)
		if err != nil {
			return out, err
		}
		if frame.Open > 0 {
			out.Metrics["minutes_any_open"] += float64(step) / 60000
		}
		if frame.Confirmed > 0 {
			out.Metrics["minutes_any_confirmed"] += float64(step) / 60000
		}
		out.Metrics["gathering_minutes"] += float64(frame.Open) * float64(step) / 60000
		out.Metrics["confirmed_gathering_minutes"] += float64(frame.Confirmed) * float64(step) / 60000
		out.Metrics["peak_simultaneous_confirmed"] = max(out.Metrics["peak_simultaneous_confirmed"], float64(frame.Confirmed))
		out.Metrics["peak_simultaneous_open"] = max(out.Metrics["peak_simultaneous_open"], float64(frame.Open))
		out.Metrics["baseline_person_minutes"] += float64(frame.Baseline) * float64(step) / 60000
		out.Metrics["active_person_minutes"] += float64(frame.Active) * float64(step) / 60000
		if now%300000 == 0 {
			if now < released.ExpiresAt {
				frame.Public = released.Gatherings
				frame.Areas = released.Areas
			}
			out.Frames = append(out.Frames, frame)
		}
	}
	for _, g := range groups {
		if g.active {
			out.Gatherings = append(out.Gatherings, g.report)
			if g.report.ConfirmedMinutes > 0 {
				out.Metrics["gatherings_ever_confirmed"]++
			}
		}
	}
	out.Metrics["gatherings_activated"] = float64(len(out.Gatherings))
	out.Metrics["activated_cells"] = float64(len(activatedCells))
	out.Metrics["confirmed_cells"] = float64(len(confirmedCells))
	out.Metrics["unique_synthetic_arrivals"] = float64(len(uniqueArrival))
	out.Metrics["mean_baseline_active"] = out.Metrics["baseline_person_minutes"] / float64(s.Hours*60)
	out.Metrics["mean_total_active"] = out.Metrics["active_person_minutes"] / float64(s.Hours*60)
	return out, nil
}
