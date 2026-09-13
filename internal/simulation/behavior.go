package simulation

import (
	"container/heap"
	"errors"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/jegati/jegati/internal/geography"
)

type Cluster struct {
	Center   geography.Point `yaml:"center" json:"center"`
	SpreadKM float64         `yaml:"spread_km" json:"spread_km"`
	Weight   float64         `yaml:"weight" json:"weight"`
}
type Behavior struct {
	EnrollmentMinutes   int        `yaml:"enrollment_minutes" json:"enrollment_minutes"`
	DurationMinutes     int        `yaml:"duration_minutes" json:"duration_minutes"`
	Networks            int        `yaml:"network_groups" json:"network_groups"`
	WritesPerSecond     int        `yaml:"writes_per_second" json:"writes_per_second"`
	SnapshotSeconds     int        `yaml:"snapshot_seconds" json:"snapshot_seconds"`
	Accept              float64    `yaml:"accept_probability" json:"accept_probability"`
	Decline             float64    `yaml:"decline_probability" json:"decline_probability"`
	Arrive              float64    `yaml:"arrival_probability_given_going" json:"arrival_probability_given_going"`
	LateJoin            float64    `yaml:"late_join_probability" json:"late_join_probability"`
	Retract             float64    `yaml:"retract_probability" json:"retract_probability"`
	Offline             float64    `yaml:"offline_probability" json:"offline_probability"`
	OfflineSeconds      [2]int     `yaml:"offline_seconds" json:"offline_seconds"`
	ResponseSeconds     [2]int     `yaml:"response_seconds" json:"response_seconds"`
	Speed               [2]float64 `yaml:"speed_meters_per_second" json:"speed_meters_per_second"`
	Detour              [2]float64 `yaml:"detour_multiplier" json:"detour_multiplier"`
	AvailabilityWeights []float64  `yaml:"availability_weights" json:"availability_weights"`
	RadiusWeights       []float64  `yaml:"radius_weights" json:"radius_weights"`
	Clusters            []Cluster  `yaml:"clusters" json:"clusters"`
}

func probability(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 1 }
func weights(v []float64, n int) bool {
	if len(v) != n {
		return false
	}
	sum := 0.
	for _, w := range v {
		if !probability(w) {
			return false
		}
		sum += w
	}
	return math.Abs(sum-1) < 1e-9
}
func (b Behavior) Validate(s Scenario) error {
	if b.EnrollmentMinutes < 1 || b.DurationMinutes < b.EnrollmentMinutes || b.DurationMinutes > 240 || b.Networks < 1 || b.Networks > 512 || b.WritesPerSecond < 1 || b.WritesPerSecond > 250 || b.SnapshotSeconds < 30 || b.SnapshotSeconds > 300 {
		return errors.New("behavior work/timing bounds invalid")
	}
	maxAvailability := 0
	for _, v := range s.Availability {
		maxAvailability = max(maxAvailability, v)
	}
	if b.DurationMinutes < b.EnrollmentMinutes+maxAvailability {
		return errors.New("duration must cover enrollment plus longest availability")
	}
	for _, v := range []float64{b.Accept, b.Decline, b.Arrive, b.LateJoin, b.Retract, b.Offline} {
		if !probability(v) {
			return errors.New("invalid behavior probability")
		}
	}
	if b.Accept+b.Decline > 1 || !weights(b.AvailabilityWeights, len(s.Availability)) || !weights(b.RadiusWeights, len(s.Radii)) {
		return errors.New("invalid behavior weights")
	}
	for _, v := range [][2]int{b.ResponseSeconds, b.OfflineSeconds} {
		if v[0] < 0 || v[1] < v[0] || v[1] > 3600 {
			return errors.New("invalid delay range")
		}
	}
	for _, v := range [][2]float64{b.Speed, b.Detour} {
		if v[0] <= 0 || v[1] < v[0] || v[1] > 5 || math.IsNaN(v[0]) || math.IsNaN(v[1]) {
			return errors.New("invalid journey range")
		}
	}
	if b.Detour[0] < 1 {
		return errors.New("detour cannot shorten straight-line distance")
	}
	if len(b.Clusters) > 8 {
		return errors.New("too many synthetic clusters")
	}
	if len(b.Clusters) > 0 {
		w := []float64{}
		for _, c := range b.Clusters {
			if !geography.Inside(c.Center) || c.SpreadKM < 0 || c.SpreadKM > 5 || math.IsNaN(c.SpreadKM) {
				return errors.New("invalid synthetic cluster")
			}
			w = append(w, c.Weight)
		}
		if !weights(w, len(w)) {
			return errors.New("invalid cluster weights")
		}
	}
	if s.Population*s.SybilCredentials > 30000 {
		return errors.New("credential work bound exceeded")
	}
	return nil
}
func sample(r *rand.Rand, w []float64) int {
	v := r.Float64()
	sum := 0.
	for i, x := range w {
		sum += x
		if v < sum {
			return i
		}
	}
	return len(w) - 1
}
func ranged(r *rand.Rand, v [2]float64) float64 { return v[0] + r.Float64()*(v[1]-v[0]) }
func delay(r *rand.Rand, v [2]int) int64        { return int64(v[0]+r.IntN(v[1]-v[0]+1)) * 1000 }

type actorSpec struct {
	Input
	Origin geography.Point `json:"synthetic_origin"`
	Enroll int64           `json:"enroll_ms"`
	Group  int             `json:"network_group"`
}

func population(s Scenario, g geography.Grid) []actorSpec {
	rng := rand.New(rand.NewPCG(s.Seed, s.Seed^0x706f70756c617469))
	b := s.Behavior
	out := []actorSpec{}
	for person := 0; person < s.Population; person++ {
		p := geography.Point{g.West + rng.Float64()*(g.East-g.West), g.South + rng.Float64()*(g.North-g.South)}
		if len(b.Clusters) > 0 && s.Distribution == "clustered" {
			w := []float64{}
			for _, c := range b.Clusters {
				w = append(w, c.Weight)
			}
			c := b.Clusters[sample(rng, w)]
			p = geography.Point{c.Center[0] + (rng.Float64()-.5)*2*c.SpreadKM/84, pLatitude(c.Center[1], rng.Float64(), c.SpreadKM)}
		}
		if s.Distribution == "single_hotspot" {
			p = geography.Point{19.818, 41.327}
		}
		if s.Distribution == "sparse" {
			if person%2 == 0 {
				p[0] = g.West + .002
			} else {
				p[0] = g.East - .002
			}
		}
		p[0] = math.Max(g.West, math.Min(g.East-1e-8, p[0]))
		p[1] = math.Max(g.South, math.Min(g.North-1e-8, p[1]))
		cell, _ := g.CellAt(p)
		copies := 1
		if rng.Float64() < s.SybilFraction {
			copies = s.SybilCredentials
		}
		enroll := int64(rng.IntN(b.EnrollmentMinutes * 60000))
		for n := 0; n < copies; n++ {
			out = append(out, actorSpec{Input{person, g.ID(cell), s.Radii[sample(rng, b.RadiusWeights)], s.Availability[sample(rng, b.AvailabilityWeights)], rng.Float64() < s.Cancellation, rng.Float64() < s.Duplicate}, p, enroll, person % b.Networks})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Enroll < out[j].Enroll })
	for i := 1; i < len(out); i++ {
		out[i].Enroll = max(out[i].Enroll, out[i-1].Enroll+1)
	}
	return out
}
func pLatitude(center, random, spread float64) float64 { return center + (random-.5)*2*spread/111 }

type event struct {
	at        int64
	order     int
	kind      string
	actor     int
	gathering string
}
type events []event

func (q events) Len() int { return len(q) }
func (q events) Less(i, j int) bool {
	if q[i].at != q[j].at {
		return q[i].at < q[j].at
	}
	return q[i].order < q[j].order
}
func (q events) Swap(i, j int)            { q[i], q[j] = q[j], q[i] }
func (q *events) Push(v any)              { *q = append(*q, v.(event)) }
func (q *events) Pop() any                { v := (*q)[len(*q)-1]; *q = (*q)[:len(*q)-1]; return v }
func (q *events) add(e event, order *int) { e.order = *order; *order++; heap.Push(q, e) }
