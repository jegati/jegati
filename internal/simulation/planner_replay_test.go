package simulation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
)

type replayCheckpoint struct {
	Minute    int    `json:"minute"`
	Live      int    `json:"live"`
	Proposals int    `json:"proposals"`
	Offered   int    `json:"offered"`
	Decisions string `json:"decisions_sha256"`
}

type plannerReplay struct {
	Config      string             `json:"config_sha256"`
	Dataset     string             `json:"dataset_sha256"`
	Inputs      string             `json:"inputs_sha256"`
	Population  int                `json:"population"`
	CellMeters  int                `json:"cell_meters"`
	Checkpoints []replayCheckpoint `json:"checkpoints"`
}

// This fixture never sends requests or issues credentials. Fixed participant and
// gathering labels exist only in the test binary; production crypto/rand, worker
// ownership and the ordinary population simulation remain untouched.
func TestTiranaPlannerReplay(t *testing.T) {
	if os.Getenv("GATI_PLANNER_REPLAY") != "1" {
		t.Skip("use make test-planner-replay (full 100 m Tirana index)")
	}
	cfg, err := config.Load("../../config/gati.yaml", false)
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := Load("../../simulation/scenarios/tirana-population.yaml")
	if err != nil {
		t.Fatal(err)
	}
	grid, err := geography.NewGrid(cfg.Geography.CellSizeMeters)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../data/tirana/intersections.json")
	if err != nil {
		t.Fatal(err)
	}
	var dataset geography.Dataset
	if err := json.Unmarshal(raw, &dataset); err != nil {
		t.Fatal(err)
	}
	index, err := geography.NewIndex(grid, dataset.Intersections, cfg.Geography.TravelRadiusChoicesKm)
	if err != nil {
		t.Fatal(err)
	}
	inputs := population(scenario, grid)
	if len(inputs) != 3000 {
		t.Fatal("fixture requires 3000 synthetic participants")
	}
	_, configHash := cfg.Canonical()
	encoded, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	result := plannerReplay{Config: configHash, Dataset: fmt.Sprintf("%x", sha256.Sum256(raw)), Inputs: fmt.Sprintf("%x", sha256.Sum256(encoded)), Population: len(inputs), CellMeters: cfg.Geography.CellSizeMeters}
	planner := matching.Planner{Index: index, Config: cfg.Matching}
	const epoch int64 = 1800000000000
	for _, minute := range []int{1, 5, 15, 30, 60, 90, 120, 180} {
		now := epoch + int64(minute)*60000
		var signals []matching.Signal
		for i, input := range inputs {
			created := epoch + input.Enroll
			expires := created + int64(input.Minutes)*60000
			// A reproducible cancellation snapshot, not the API response model.
			if created > now || expires <= now || (input.Cancel && now >= created+(expires-created)/2) {
				continue
			}
			cell, err := grid.Parse(input.Cell)
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, matching.Signal{Hash: fmt.Sprintf("fixture-%04d", i), Area: geography.ParticipantArea{Cell: cell, RadiusKM: input.Radius}, CreatedAt: created, ExpiresAt: expires})
		}
		first := replayDecisions(t, planner, now, minute, signals)
		shuffled := slices.Clone(signals)
		rng := rand.New(rand.NewPCG(scenario.Seed, uint64(minute)))
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		if second := replayDecisions(t, planner, now, minute, shuffled); first != second {
			t.Fatalf("input ordering changed decisions at minute %d", minute)
		}
		result.Checkpoints = append(result.Checkpoints, first)
	}
	got, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/tirana-planner-replay.json")
	if err != nil {
		t.Fatalf("read reviewed replay fixture: %v\n%s", err, got)
	}
	var expected plannerReplay
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("planner replay changed; review config/map/behavior before updating the fixture:\n%s", got)
	}
	t.Logf("3000 participants, %d m cells, %d mapped crossroads: %d snapshots match the reviewed fixture and shuffled-input replay", grid.SizeMeters, len(dataset.Intersections), len(result.Checkpoints))
}

func replayDecisions(t *testing.T, planner matching.Planner, now int64, minute int, input []matching.Signal) replayCheckpoint {
	t.Helper()
	signals := slices.Clone(input)
	result := replayCheckpoint{Minute: minute, Live: len(signals)}
	var open []matching.Gathering
	digest := sha256.New()
	// Three proposals per snapshot bound work while exercising preference for an
	// existing gathering. This is a planner regression, not an end-to-end run.
	for n := 0; n < 3; n++ {
		proposal, ok := planner.Propose(now, signals, open)
		if !ok {
			break
		}
		if len(proposal.Founders) != planner.Config.ActivationCount {
			t.Fatal("activation cohort size changed")
		}
		deadline := planner.Deadline(now, proposal.Founders)
		id := fmt.Sprintf("fixture-gathering-%d", n)
		fmt.Fprintf(digest, "proposal %s %s %d\n", id, proposal.Intersection.ID, deadline)
		founders := map[string]bool{}
		for _, founder := range proposal.Founders {
			if founders[founder.Hash] || !planner.Index.CanReach(founder.Area, proposal.Intersection.ID) || founder.ExpiresAt < deadline {
				t.Fatal("duplicate, unreachable or expired founder")
			}
			founders[founder.Hash] = true
			fmt.Fprintf(digest, "founder %s\n", founder.Hash)
		}
		for i := range signals {
			if founders[signals[i].Hash] {
				signals[i].Assigned = true
				signals[i].AssignedUntil = deadline
			}
		}
		open = append(open, matching.Gathering{ID: id, Intersection: proposal.Intersection, EndsAt: deadline})
		result.Proposals++
	}
	// Stable rendering of decisions prevents map/slice iteration order from
	// masquerading as a product behavior difference.
	slices.SortFunc(signals, func(a, b matching.Signal) int {
		if a.Hash < b.Hash {
			return -1
		}
		if a.Hash > b.Hash {
			return 1
		}
		return 0
	})
	for _, signal := range signals {
		if offer, ok := planner.Offer(now, signal, open); ok {
			if !planner.Index.CanReach(signal.Area, offer.Intersection.ID) {
				t.Fatal("unreachable offer")
			}
			fmt.Fprintf(digest, "offer %s %s\n", signal.Hash, offer.ID)
			result.Offered++
			// Declining one invitation must exclude it on the next lookup.
			signal.Declined = map[string]bool{offer.ID: true}
			next, found := planner.Offer(now, signal, open)
			if found && next.ID == offer.ID {
				t.Fatal("declined invitation offered again")
			}
			fmt.Fprintf(digest, "after-decline %s %s %t\n", signal.Hash, next.ID, found)
		}
	}
	result.Decisions = fmt.Sprintf("%x", digest.Sum(nil))
	return result
}
