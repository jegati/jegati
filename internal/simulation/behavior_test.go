package simulation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/jegati/jegati/internal/geography"
)

func TestPopulationDeterministicAndOrdered(t *testing.T) {
	s, err := Load("../../simulation/scenarios/tirana-population.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if s.Population != 3000 {
		t.Fatal("main scenario must contain 3000 people")
	}
	g, _ := geography.NewGrid(1000)
	a, b := population(s, g), population(s, g)
	if len(a) != 3000 || !reflect.DeepEqual(a, b) {
		t.Fatal("population not reproducible")
	}
	for i, v := range a {
		if _, err := g.Parse(v.Cell); err != nil {
			t.Fatal(err)
		}
		if i > 0 && a[i-1].Enroll >= v.Enroll {
			t.Fatal("enrollment hash tie")
		}
		if v.Group < 0 || v.Group >= s.Behavior.Networks {
			t.Fatal("network group out of range")
		}
	}
	s.Seed++
	if reflect.DeepEqual(a, population(s, g)) {
		t.Fatal("seed has no effect")
	}
	s.Population = 10001
	if s.Validate() == nil {
		t.Fatal("accepted oversized population")
	}
}
func TestBehaviorRejectsInvalidAssumptions(t *testing.T) {
	for name, change := range map[string]func(*Scenario){"overlap": func(s *Scenario) { s.Behavior.Accept = .9; s.Behavior.Decline = .2 }, "weights": func(s *Scenario) { s.Behavior.RadiusWeights = []float64{1} }, "networks": func(s *Scenario) { s.Behavior.Networks = 0 }, "duration": func(s *Scenario) { s.Behavior.DurationMinutes = 30 }, "speed": func(s *Scenario) { s.Behavior.Speed = [2]float64{0, 1} }, "infinite work": func(s *Scenario) { s.Behavior.SnapshotSeconds = 1 }} {
		t.Run(name, func(t *testing.T) {
			s, _ := Load("../../simulation/scenarios/tirana-population.yaml")
			change(&s)
			if s.Validate() == nil {
				t.Fatal("invalid behavior accepted")
			}
		})
	}
}
func TestPopulationRefusesProductionBeforeWrites(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
		}
		json.NewEncoder(w).Encode(map[string]any{"config": map[string]string{"profile": "production"}})
	}))
	defer server.Close()
	c, _ := NewClient(server.URL, strings.Repeat("a", 64))
	s, _ := Load("../../simulation/scenarios/tirana-population.yaml")
	if _, err := c.RunPopulation(context.Background(), s); err == nil || writes != 0 {
		t.Fatal("population target not refused")
	}
}
func TestEmptyPercentiles(t *testing.T) {
	if percentile(nil, .5) != nil {
		t.Fatal("invented empty percentile")
	}
	if *percentile([]int64{4, 1, 3, 2}, .95) != 4 {
		t.Fatal("wrong measured percentile")
	}
}
