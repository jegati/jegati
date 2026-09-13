package simulation

import (
	"container/heap"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	random "math/rand/v2"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
)

type view struct {
	store.Signal
	Invitation *store.Gathering `json:"invitation,omitempty"`
}
type actor struct {
	position                  geography.Point
	spec                      actorSpec
	client                    *Client
	token                     string
	rng                       *random.Rand
	accepted, active, invited bool
	went, arrived             bool
	state                     view
	seen                      map[string]bool
	firstSeen                 int64
}
type GatheringObservation struct {
	Alias        string                 `json:"alias"`
	Intersection geography.Intersection `json:"intersection"`
	ActivatedAt  int64                  `json:"activated_at"`
	EndsAt       int64                  `json:"ends_at"`
	FirstSeen    int64                  `json:"first_seen"`
	HereSeen     int64                  `json:"here_first_seen,omitempty"`
	State        string                 `json:"state"`
}
type Frame struct {
	At         int64                  `json:"at"`
	Cells      map[string]int         `json:"cells"`
	Going      int                    `json:"going"`
	Here       int                    `json:"here"`
	Gatherings []GatheringObservation `json:"gatherings"`
	Counts     map[string]int         `json:"counts"`
}
type PopulationReport struct {
	Status             string                    `json:"status"`
	Failure            string                    `json:"failure,omitempty"`
	Simulation         bool                      `json:"simulation"`
	Scenario           Scenario                  `json:"scenario"`
	Config             config.Config             `json:"effective_config"`
	ConfigHash         string                    `json:"config_sha256"`
	InputHash          string                    `json:"input_sha256"`
	Dataset            string                    `json:"dataset"`
	DatasetHash        string                    `json:"dataset_sha256"`
	SourceRevision     string                    `json:"source_revision"`
	SourceDirty        bool                      `json:"source_dirty"`
	Grid               geography.Grid            `json:"grid"`
	StartedAt          int64                     `json:"started_at"`
	Counts             map[string]int            `json:"counts"`
	ByRadius           map[string]map[string]int `json:"by_radius"`
	ByCell             map[string]map[string]int `json:"by_cell"`
	Gatherings         []GatheringObservation    `json:"gatherings"`
	Frames             []Frame                   `json:"frames"`
	WaitMedianMS       *int64                    `json:"invitation_wait_median_ms"`
	WaitP95MS          *int64                    `json:"invitation_wait_p95_ms"`
	JourneyMedianMS    *int64                    `json:"journey_median_ms"`
	MaxDriverHeapBytes uint64                    `json:"max_sampled_driver_heap_bytes"`
	WallSeconds        float64                   `json:"wall_seconds"`
	Limitations        []string                  `json:"limitations"`
}
type populationRun struct {
	control         *Client
	config          config.Config
	report          PopulationReport
	actors          []actor
	queue           events
	order           int
	now             int64
	end             int64
	gatherings      map[string]*GatheringObservation
	lastWrite       time.Time
	waits, journeys []int64
}

func percentile(v []int64, p float64) *int64 {
	if len(v) == 0 {
		return nil
	}
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	x := v[int(math.Ceil(p*float64(len(v))))-1]
	return &x
}
func (r *populationRun) add(at int64, kind string, actorID int, gathering string) {
	if at <= r.end {
		r.queue.add(event{at: at, kind: kind, actor: actorID, gathering: gathering}, &r.order)
	}
}
func (r *populationRun) call(ctx context.Context, a *actor, method, path string, body any, headers map[string]string, out any) (int, error) {
	if method == "POST" || method == "DELETE" {
		gap := time.Second / time.Duration(r.report.Scenario.Behavior.WritesPerSecond)
		wait := time.Until(r.lastWrite.Add(gap))
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return 0, ctx.Err()
			case <-timer.C:
			}
		}
		r.lastWrite = time.Now()
	}
	code, err := a.client.doHeaders(ctx, method, path, a.token, body, headers, out)
	r.report.Counts["http_requests"]++
	if code == 429 {
		r.report.Counts["rate_limited_requests"]++
	}
	if err != nil {
		return code, err
	}
	if code >= 500 {
		return code, fmt.Errorf("simulation API failure on %s (%d)", path, code)
	}
	return code, nil
}
func (r *populationRun) advance(ctx context.Context, to int64, step bool) error {
	var result struct {
		Now int64 `json:"now"`
	}
	code, err := r.control.do(ctx, "POST", "/api/simulation/clock", r.control.control, map[string]any{"milliseconds": to - r.now, "run_worker": step}, &result)
	if err != nil || code != 200 || result.Now != to {
		return errors.New("simulation clock/worker did not reach scheduled time")
	}
	r.now = to
	return nil
}
func (r *populationRun) observe(a *actor, v view) error {
	a.state = v
	if v.State == "going" || v.State == "here" {
		a.went = true
	}
	if v.State == "here" {
		a.arrived = true
	}
	if g := v.Invitation; g != nil {
		known := r.gatherings[g.ID]
		if known == nil {
			known = &GatheringObservation{Alias: fmt.Sprintf("takim-%03d", len(r.gatherings)+1), Intersection: g.Intersection, ActivatedAt: g.ActivatedAt, EndsAt: g.EndsAt, FirstSeen: r.now, State: g.State}
			r.gatherings[g.ID] = known
			r.report.Counts["gatherings_observed"]++
		}
		if known.Intersection.ID != g.Intersection.ID || known.EndsAt != g.EndsAt {
			return errors.New("activated destination/deadline changed")
		}
		if !a.invited {
			a.invited = true
			a.firstSeen = r.now
			r.waits = append(r.waits, r.now-v.CreatedAt)
			r.report.Counts["credentials_invited"]++
		}
		known.State = g.State
		if g.State == "jemi_ketu" && known.HereSeen == 0 {
			known.HereSeen = r.now
			r.report.Counts["gatherings_confirmed_observed"]++
		}
	}
	return nil
}
func (r *populationRun) offerResponse(i int) {
	a := &r.actors[i]
	g := a.state.Invitation
	if g == nil || a.seen[g.ID] || a.state.State == "going" || a.state.State == "here" {
		return
	}
	a.seen[g.ID] = true
	r.report.Counts["invitations_observed"]++
	v := a.rng.Float64()
	kind := "ignore"
	b := r.report.Scenario.Behavior
	if v < b.Accept {
		kind = "going"
	} else if v < b.Accept+b.Decline {
		kind = "decline"
	}
	if kind == "ignore" {
		r.report.Counts["ignored_invitations"]++
		return
	}
	r.add(r.now+delay(a.rng, b.ResponseSeconds), kind, i, g.ID)
}
func (r *populationRun) journey(i int, g *store.Gathering) {
	a := &r.actors[i]
	b := r.report.Scenario.Behavior
	if a.rng.Float64() >= b.Arrive {
		r.report.Counts["planned_no_shows"]++
		return
	}
	distance := geography.Distance(a.position, g.Intersection.Point)
	duration := int64(math.Ceil(distance*ranged(a.rng, b.Detour)/ranged(a.rng, b.Speed))) * 1000
	r.journeys = append(r.journeys, duration)
	if r.now+max(int64(1000), duration) > r.end {
		r.report.Counts["journeys_beyond_horizon"]++
		return
	}
	r.add(r.now+max(int64(1000), duration), "arrive", i, g.ID)
}
func (r *populationRun) enroll(ctx context.Context, i int) error {
	a := &r.actors[i]
	var v view
	var code int
	var err error
	// Synthetic discovery uses only gatherings already observed by another actor.
	var target *GatheringObservation
	targetID := ""
	best := math.Inf(1)
	if a.rng.Float64() < r.report.Scenario.Behavior.LateJoin {
		ids := []string{}
		for id := range r.gatherings {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return r.gatherings[ids[i]].Alias < r.gatherings[ids[j]].Alias })
		for _, id := range ids {
			g := r.gatherings[id]
			cell, _ := r.report.Grid.Parse(a.spec.Cell)
			d := r.report.Grid.MaxDistance(cell, g.Intersection.Point)
			if g.EndsAt > r.now && d <= a.spec.Radius*1000 && d < best {
				target = g
				targetID = id
				best = d
			}
		}
	}
	if target != nil {
		r.report.Counts["direct_join_attempts"]++
		code, err = r.call(ctx, a, "POST", "/api/join", map[string]any{"gathering_id": targetID, "cell": a.spec.Cell, "radius_km": a.spec.Radius, "availability_minutes": a.spec.Minutes}, nil, &v)
		if err != nil {
			return err
		}
		if code == 200 {
			r.report.Counts["direct_joins_accepted"]++
			r.report.Counts["going_responses_accepted"]++
			if v.Invitation != nil && v.Invitation.State == "jemi_ketu" {
				r.report.Counts["joins_after_jemi_ketu"]++
			}
		} else {
			r.report.Counts["direct_joins_rejected"]++
		}
	}
	if target == nil || code != 200 {
		code, err = r.call(ctx, a, "POST", "/api/signals", a.spec.Input, nil, &v)
	}
	if err != nil {
		return err
	}
	if code != 200 {
		if code != 429 && code != 503 {
			return fmt.Errorf("unexpected enrollment status %d", code)
		}
		r.report.Counts["credentials_rejected"]++
		return nil
	}
	a.accepted = true
	a.active = true
	r.report.Counts["credentials_accepted"]++
	if err = r.observe(a, v); err != nil {
		return err
	}
	if a.spec.Duplicate {
		var retry view
		code, err = r.call(ctx, a, "POST", "/api/signals", a.spec.Input, nil, &retry)
		if err != nil {
			return err
		}
		if code == 200 {
			if retry.ExpiresAt != v.ExpiresAt {
				return errors.New("retry extended availability")
			}
			r.report.Counts["idempotent_retries_verified"]++
		} else if code != 429 {
			return errors.New("unexpected creation retry status")
		}
	}
	if a.spec.Cancel {
		at := r.now + int64(float64(v.ExpiresAt-r.now)*(.2+a.rng.Float64()*.6))
		r.add(at, "cancel", i, "")
	}
	r.add(v.ExpiresAt, "expire", i, "")
	if v.State == "going" && v.Invitation != nil {
		r.journey(i, v.Invitation)
	} else {
		r.offerResponse(i)
	}
	r.pollLater(i)
	return nil
}
func (r *populationRun) pollLater(i int) {
	a := &r.actors[i]
	next := int64(float64(r.config.Notifications.ForegroundPollSeconds*1000) * (1 + a.rng.Float64()*.2))
	b := r.report.Scenario.Behavior
	if a.rng.Float64() < b.Offline {
		next += delay(a.rng, b.OfflineSeconds)
		r.report.Counts["offline_intervals"]++
	}
	r.add(r.now+next, "poll", i, "")
}
func (r *populationRun) act(ctx context.Context, e event) error {
	a := &r.actors[e.actor]
	if e.kind == "enroll" {
		return r.enroll(ctx, e.actor)
	}
	if !a.active {
		if e.kind == "arrive" {
			r.report.Counts["journeys_ended_after_session"]++
		}
		return nil
	}
	var v view
	switch e.kind {
	case "poll":
		code, err := r.call(ctx, a, "GET", "/api/signal", nil, nil, &v)
		if err != nil {
			return err
		}
		if code == 410 {
			a.active = false
			r.report.Counts["expired_verified"]++
			return nil
		}
		if code == 200 {
			if err = r.observe(a, v); err != nil {
				return err
			}
			r.offerResponse(e.actor)
		} else if code != 429 {
			return fmt.Errorf("unexpected poll %d", code)
		}
		r.pollLater(e.actor)
	case "expire":
		code, err := r.call(ctx, a, "GET", "/api/signal", nil, nil, nil)
		if err != nil {
			return err
		}
		if code == 429 {
			r.add(r.now+60000, "expire", e.actor, "")
			return nil
		}
		if code != 410 {
			return errors.New("session survives its server deadline")
		}
		a.active = false
		r.report.Counts["expired_verified"]++
	case "cancel":
		code, err := r.call(ctx, a, "DELETE", "/api/signal", nil, nil, nil)
		if err != nil {
			return err
		}
		if code == 204 || code == 410 {
			a.active = false
			r.report.Counts["cancelled"]++
		} else if code == 429 {
			r.add(r.now+60000, "cancel", e.actor, "")
		} else {
			return fmt.Errorf("unexpected cancellation %d", code)
		}
	case "going", "decline":
		if a.state.Invitation == nil || a.state.Invitation.ID != e.gathering {
			return nil
		}
		code, err := r.call(ctx, a, "POST", "/api/"+e.kind, map[string]string{"gathering_id": e.gathering}, nil, &v)
		if err != nil {
			return err
		}
		if code == 200 {
			r.report.Counts[e.kind+"_responses_accepted"]++
			if e.kind == "decline" && v.ExpiresAt != a.state.ExpiresAt {
				return errors.New("decline changed willingness deadline")
			}
			if e.kind == "going" && v.Invitation != nil && v.CreatedAt > v.Invitation.ActivatedAt {
				r.report.Counts["late_going_accepted"]++
			}
			if err = r.observe(a, v); err != nil {
				return err
			}
			if e.kind == "going" && v.Invitation != nil {
				r.journey(e.actor, v.Invitation)
			}
		} else if code == 410 || code == 409 || code == 429 {
			r.report.Counts[e.kind+"_responses_rejected"]++
		} else {
			return fmt.Errorf("unexpected intent %d", code)
		}
	case "arrive":
		if a.state.Invitation == nil || a.state.Invitation.ID != e.gathering {
			r.report.Counts["journeys_obsolete"]++
			return nil
		}
		a.position = a.state.Invitation.Intersection.Point
		r.report.Counts["arrival_attempts"]++
		nonce, err := token()
		if err != nil {
			return err
		}
		headers := map[string]string{"X-Gati-Arrival-Nonce": nonce}
		code, err := r.call(ctx, a, "POST", "/api/arrival-nonce", nil, headers, nil)
		if err != nil {
			return err
		}
		if code != 200 {
			if code != 410 && code != 409 && code != 429 {
				return fmt.Errorf("unexpected nonce %d", code)
			}
			r.report.Counts["arrivals_rejected"]++
			return nil
		}
		cell, _ := r.report.Grid.CellAt(a.state.Invitation.Intersection.Point)
		code, err = r.call(ctx, a, "POST", "/api/arrival", map[string]string{"cell": r.report.Grid.ID(cell)}, headers, &v)
		if err != nil {
			return err
		}
		if code == 200 {
			r.report.Counts["arrivals_accepted"]++
			if err = r.observe(a, v); err != nil {
				return err
			}
			if a.spec.Duplicate {
				var again view
				code, err = r.call(ctx, a, "POST", "/api/arrival", map[string]string{"cell": r.report.Grid.ID(cell)}, headers, &again)
				if err != nil {
					return err
				}
				if code == 200 {
					if again.ArrivalUntil != v.ArrivalUntil {
						return errors.New("arrival replay renewed freshness")
					}
					r.report.Counts["arrival_replays_verified"]++
				} else if code != 429 {
					return errors.New("unexpected arrival replay rejection")
				}
			}
			if a.rng.Float64() < r.report.Scenario.Behavior.Retract {
				r.add(r.now+60000, "retract", e.actor, e.gathering)
			}
		} else if code == 410 || code == 409 || code == 429 {
			r.report.Counts["arrivals_rejected"]++
		} else {
			return fmt.Errorf("unexpected arrival %d", code)
		}
	case "retract":
		if a.state.Invitation == nil || a.state.Invitation.ID != e.gathering {
			return nil
		}
		code, err := r.call(ctx, a, "DELETE", "/api/arrival", nil, nil, &v)
		if err != nil {
			return err
		}
		if code == 200 {
			r.report.Counts["arrivals_retracted"]++
			return r.observe(a, v)
		}
		if code != 410 && code != 429 && code != 409 {
			return fmt.Errorf("unexpected retraction %d", code)
		}
	}
	return nil
}
func token() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func (r *populationRun) frame() {
	f := Frame{At: r.now, Cells: map[string]int{}, Counts: map[string]int{}, Gatherings: []GatheringObservation{}}
	for k, v := range r.report.Counts {
		f.Counts[k] = v
	}
	for _, a := range r.actors {
		if a.active && a.state.ExpiresAt > r.now {
			f.Cells[a.spec.Cell]++
			if a.state.Invitation != nil && a.state.Invitation.EndsAt > r.now {
				if a.state.State == "going" || a.state.State == "here" {
					f.Going++
				}
				if a.state.State == "here" && a.state.ArrivalUntil > r.now {
					f.Here++
				}
			}
		}
	}
	for _, g := range r.gatherings {
		if g.EndsAt > r.now {
			f.Gatherings = append(f.Gatherings, *g)
		}
	}
	sort.Slice(f.Gatherings, func(i, j int) bool { return f.Gatherings[i].Alias < f.Gatherings[j].Alias })
	r.report.Frames = append(r.report.Frames, f)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	r.report.MaxDriverHeapBytes = max(r.report.MaxDriverHeapBytes, m.HeapAlloc)
}

func (c *Client) RunPopulation(ctx context.Context, s Scenario) (PopulationReport, error) {
	started := time.Now()
	r := &populationRun{control: c, gatherings: map[string]*GatheringObservation{}, report: PopulationReport{Status: "running", Simulation: true, Scenario: s, Counts: map[string]int{}, ByRadius: map[string]map[string]int{}, ByCell: map[string]map[string]int{}, Limitations: []string{"Synthetic assumptions, not real Tirana demographics or human usability evidence.", "Bulk actors use coarse API claims; device UI is tested separately with mocked geolocation.", "Journey durations approximate geographic distance, not pedestrian routes.", "Gathering state is observed through participant responses; short unobserved transitions can be missed.", "Direct-join discovery is injected from synthetic observers, not implemented public maps or notifications.", "Rate windows and native TTL use real time; this is not a 100k capacity benchmark.", "Known colluding-input privacy inference and location spoofing remain possible."}}}
	if s.Behavior == nil {
		return r.report, errors.New("population behavior required")
	}
	if err := s.Validate(); err != nil {
		return r.report, err
	}
	var envelope struct {
		Config config.Config `json:"config"`
		Hash   string        `json:"sha256"`
	}
	code, err := c.do(ctx, "GET", "/api/config", "", nil, &envelope)
	if err != nil || code != 200 || envelope.Config.Profile != "simulation" {
		return r.report, errors.New("refusing non-simulation target")
	}
	if err = envelope.Config.Validate(true); err != nil {
		return r.report, err
	}
	r.config = envelope.Config
	r.report.Config = envelope.Config
	r.report.ConfigHash = envelope.Hash
	if s.Behavior.WritesPerSecond > r.config.Limits.GlobalWritesPerSecond/2 {
		return r.report, errors.New("simulation pacing must leave half the global write budget free")
	}
	for _, v := range s.Radii {
		if !slices.Contains(r.config.Geography.TravelRadiusChoicesKm, v) {
			return r.report, errors.New("scenario radius absent from selected application config")
		}
	}
	for _, v := range s.Availability {
		if !slices.Contains(r.config.Availability.ChoicesMinutes, v) {
			return r.report, errors.New("scenario availability absent from selected application config")
		}
	}
	code, err = c.do(ctx, "GET", "/api/simulation/clock", "", nil, nil)
	if err != nil || code != 403 {
		return r.report, errors.New("simulation control lacks authentication")
	}
	var clock struct {
		Now        int64 `json:"now"`
		Simulation bool  `json:"simulation"`
	}
	code, err = c.do(ctx, "GET", "/api/simulation/clock", c.control, nil, &clock)
	if err != nil || code != 200 || !clock.Simulation {
		return r.report, errors.New("authenticated clock missing")
	}
	r.now = clock.Now
	r.report.StartedAt = r.now
	r.end = r.now + int64(s.Behavior.DurationMinutes+1)*60000
	code, err = c.do(ctx, "GET", "/api/geography", "", nil, &r.report.Grid)
	if err != nil || code != 200 {
		return r.report, errors.New("geography unavailable")
	}
	raw, err := os.ReadFile("data/tirana/intersections.json")
	if err != nil {
		return r.report, err
	}
	sum := sha256.Sum256(raw)
	r.report.DatasetHash = hex.EncodeToString(sum[:])
	r.report.Dataset = r.config.Geography.IntersectionDataset
	var dataset geography.Dataset
	if err = json.Unmarshal(raw, &dataset); err != nil || dataset.Version != r.report.Dataset {
		return r.report, errors.New("local report dataset differs from selected config")
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, v := range info.Settings {
			if v.Key == "vcs.revision" {
				r.report.SourceRevision = v.Value
			}
			if v.Key == "vcs.modified" {
				r.report.SourceDirty = v.Value == "true"
			}
		}
	}
	inputs := population(s, r.report.Grid)
	raw, _ = json.Marshal(inputs)
	sum = sha256.Sum256(raw)
	r.report.InputHash = hex.EncodeToString(sum[:])
	r.report.Counts["synthetic_people"] = s.Population
	r.report.Counts["generated_credentials"] = len(inputs)
	clients := make([]*Client, s.Behavior.Networks)
	for i := range clients {
		copy := *c
		dialer := &net.Dialer{LocalAddr: &net.TCPAddr{IP: net.IPv4(127, 1, byte(i/250), byte(i%250+1))}, Timeout: 5 * time.Second}
		transport := &http.Transport{DialContext: dialer.DialContext, MaxIdleConns: 2, MaxIdleConnsPerHost: 2}
		copy.http = &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: c.http.CheckRedirect}
		defer transport.CloseIdleConnections()
		clients[i] = &copy
	}
	for i, input := range inputs {
		secret, err := token()
		if err != nil {
			return r.report, err
		}
		r.actors = append(r.actors, actor{spec: input, position: input.Origin, token: secret, client: clients[input.Group], rng: random.New(random.NewPCG(s.Seed^uint64(i+1), 0x6163746f72)), seen: map[string]bool{}})
		r.add(r.now+input.Enroll, "enroll", i, "")
	}
	r.add(r.now, "worker", 0, "")
	r.add(r.now, "frame", 0, "")
	for r.queue.Len() > 0 {
		e := heap.Pop(&r.queue).(event)
		if e.at != r.now || e.kind == "worker" {
			if err = r.advance(ctx, e.at, e.kind == "worker"); err != nil {
				return r.report, err
			}
		}
		switch e.kind {
		case "worker":
			r.add(r.now+int64(r.config.Matching.MaximumDebounceSeconds)*1000, "worker", 0, "")
		case "frame":
			r.frame()
			if (r.now-r.report.StartedAt)%1800000 == 0 {
				fmt.Printf("Simulation minute %d: accepted %d, invited %d, arrivals %d.\n", (r.now-r.report.StartedAt)/60000, r.report.Counts["credentials_accepted"], r.report.Counts["credentials_invited"], r.report.Counts["arrivals_accepted"])
			}
			r.add(r.now+int64(s.Behavior.SnapshotSeconds)*1000, "frame", 0, "")
		default:
			if err = r.act(ctx, e); err != nil {
				return r.report, err
			}
		}
	}
	// Final authenticated reads enforce expiry even if a prior expiry check was throttled.
	acceptedPeople, invitedPeople, arrivedPeople := map[int]bool{}, map[int]bool{}, map[int]bool{}
	for i := range r.actors {
		a := &r.actors[i]
		if a.active {
			code, err := r.call(ctx, a, "GET", "/api/signal", nil, nil, nil)
			if err != nil || code != 410 {
				return r.report, errors.New("final logical expiry could not be verified")
			}
			a.active = false
			r.report.Counts["expired_verified"]++
		}
		if a.accepted {
			acceptedPeople[a.spec.Person] = true
		}
		if a.invited {
			invitedPeople[a.spec.Person] = true
		}
		if a.went {
			r.report.Counts["credentials_went"]++
		}
		if a.arrived {
			arrivedPeople[a.spec.Person] = true
			r.report.Counts["credentials_arrived"]++
		}
		if a.accepted && !a.invited {
			r.report.Counts["accepted_never_invited"]++
		}
		for _, pair := range []struct {
			table map[string]map[string]int
			key   string
		}{{r.report.ByRadius, fmt.Sprint(a.spec.Radius)}, {r.report.ByCell, a.spec.Cell}} {
			if pair.table[pair.key] == nil {
				pair.table[pair.key] = map[string]int{}
			}
			pair.table[pair.key]["generated"]++
			if a.accepted {
				pair.table[pair.key]["accepted"]++
			}
			if a.invited {
				pair.table[pair.key]["invited"]++
			}
		}
	}
	r.report.Counts["synthetic_people_accepted"] = len(acceptedPeople)
	r.report.Counts["synthetic_people_invited"] = len(invitedPeople)
	r.report.Counts["synthetic_people_arrived"] = len(arrivedPeople)
	// Diagnose geometry alone; this does not assert that a compatible cohort existed.
	for _, a := range r.actors {
		exact, coarse := false, false
		cell, _ := r.report.Grid.Parse(a.spec.Cell)
		for _, point := range dataset.Intersections {
			if !exact && geography.Distance(a.spec.Origin, point.Point) <= a.spec.Radius*1000 {
				exact = true
			}
			if !coarse && r.report.Grid.MaxDistance(cell, point.Point) <= a.spec.Radius*1000 {
				coarse = true
			}
			if exact && coarse {
				break
			}
		}
		bucket := r.report.ByRadius[fmt.Sprint(a.spec.Radius)]
		if exact {
			bucket["exact_start_has_destination"]++
		}
		if coarse {
			bucket["coarse_start_has_destination"]++
		}
		if exact && !coarse {
			bucket["coarsening_excludes_all_start_destinations"]++
		}
	}
	r.frame()
	for _, g := range r.gatherings {
		r.report.Gatherings = append(r.report.Gatherings, *g)
	}
	sort.Slice(r.report.Gatherings, func(i, j int) bool { return r.report.Gatherings[i].Alias < r.report.Gatherings[j].Alias })
	r.report.WaitMedianMS = percentile(r.waits, .5)
	r.report.WaitP95MS = percentile(r.waits, .95)
	r.report.JourneyMedianMS = percentile(r.journeys, .5)
	r.report.Status = "completed"
	r.report.WallSeconds = time.Since(started).Seconds()
	return r.report, nil
}
