package simulation

import (
	"context"
	"encoding/json"
	"github.com/jegati/jegati/internal/geography"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDeterministicAndBounded(t *testing.T) {
	s, e := Load("../../simulation/scenarios/tirana-evening.yaml")
	if e != nil {
		t.Fatal(e)
	}
	g, _ := geography.NewGrid(1000)
	a, b := Generate(s, g), Generate(s, g)
	if !reflect.DeepEqual(a, b) || len(a) != 40 {
		t.Fatal("seed is not reproducible")
	}
	for _, input := range a {
		if _, e := g.Parse(input.Cell); e != nil {
			t.Fatal(e)
		}
	}
	s.Seed++
	if reflect.DeepEqual(a, Generate(s, g)) {
		t.Fatal("seed has no effect")
	}
	for _, distribution := range []string{"uniform", "sparse", "single_hotspot"} {
		s.Distribution = distribution
		if len(Generate(s, g)) != 40 {
			t.Fatal("wrong distribution size")
		}
	}
	s.SybilFraction = 1
	s.SybilCredentials = 10
	if len(Generate(s, g)) != 400 {
		t.Fatal("synthetic people were confused with credentials")
	}
	s.Population = 1001
	if s.Validate() == nil {
		t.Fatal("unbounded population accepted")
	}
}
func TestRefusesUnsafeTargets(t *testing.T) {
	for _, target := range []string{"https://example.com", "http://localhost:8082", "http://192.168.1.2:8082", "http://127.0.0.1:8082/api", "http://user:password@127.0.0.1:8082", "http://127.0.0.1:8082/?x=1", "http://127.0.0.1:8082/#fragment"} {
		if _, e := NewClient(target, strings.Repeat("a", 64)); e == nil {
			t.Errorf("accepted %s", target)
		}
	}
}
func TestRefusesProductionBeforeWriting(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
		}
		json.NewEncoder(w).Encode(map[string]any{"config": map[string]string{"profile": "production"}})
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, strings.Repeat("a", 64))
	s, _ := Load("../../simulation/scenarios/tirana-evening.yaml")
	if _, e := client.Run(context.Background(), s); e == nil || writes != 0 {
		t.Fatal("production target was not refused before writes")
	}
}
func TestRefusesRedirect(t *testing.T) {
	reached := false
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, strings.Repeat("a", 64))
	s, _ := Load("../../simulation/scenarios/tirana-evening.yaml")
	if _, e := client.Run(context.Background(), s); e == nil || reached {
		t.Fatal("redirect target reached")
	}
}
