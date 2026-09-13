//go:build simulation

package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
)

// This is a counterexample reproduction, NOT a passing privacy assurance test.
// It runs only on explicit request, writes synthetic evidence and makes the
// make privacy-probe target fail its privacy gate after collecting that evidence.
func TestKnownColludingInvitationInference(t *testing.T) {
	if os.Getenv("GATI_SIM_INTEGRATION") != "1" || os.Getenv("GATI_PRIVACY_PROBE") != "1" {
		t.Skip("explicit isolated privacy probe only")
	}
	ctx := context.Background()
	raw, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal(e)
	}
	backend, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(raw)))
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Client.Close()
	backend.Client.Del(ctx, "gati-sim:worker:lease")
	// Production thresholds/limits with the isolated store's controlled clock.
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	_, configHash := c.Canonical()
	grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
	raw, e = os.ReadFile("../../data/tirana/crossings.json")
	if e != nil {
		t.Fatal(e)
	}
	var dataset geography.Dataset
	if json.Unmarshal(raw, &dataset) != nil {
		t.Fatal("invalid fixture")
	}
	index, e := geography.NewIndex(grid, dataset.Crossings, c.Geography.TravelRadiusChoicesKm)
	if e != nil {
		t.Fatal(e)
	}
	known := geography.ParticipantArea{Cell: geography.Cell{X: 5, Y: 5}, RadiusKM: 3}
	group := []geography.ParticipantArea{}
	for n := 1; n < c.Matching.ActivationCount; n++ {
		group = append(group, known)
	}
	possible := map[string]map[string]bool{}
	compatible := map[string]bool{}
	targetByCrossing := map[string]geography.ParticipantArea{}
	for y := 0; y < grid.Rows; y++ {
		for x := 0; x < grid.Columns; x++ {
			for _, radius := range c.Geography.TravelRadiusChoicesKm {
				target := geography.ParticipantArea{Cell: geography.Cell{X: x, Y: y}, RadiusKM: radius}
				crossing, ok := index.Closest(append(group, target))
				if !ok {
					continue
				}
				cell := grid.ID(target.Cell)
				compatible[cell] = true
				if possible[crossing.ID] == nil {
					possible[crossing.ID] = map[string]bool{}
					targetByCrossing[crossing.ID] = target
				}
				possible[crossing.ID][cell] = true
			}
		}
	}
	ids := []string{}
	for id := range possible {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if len(possible[ids[i]]) != len(possible[ids[j]]) {
			return len(possible[ids[i]]) < len(possible[ids[j]])
		}
		return ids[i] < ids[j]
	})
	if len(ids) == 0 {
		t.Fatal("no compatible synthetic target")
	}
	selected := ids[0]
	target := targetByCrossing[selected]
	engine := worker.New(backend, index, c)
	handler := Handler(c, backend, nil, engine)
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.3:12345"
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	create := func(area geography.ParticipantArea) string {
		raw := make([]byte, 32)
		rand.Read(raw)
		token := base64.RawURLEncoding.EncodeToString(raw)
		body, _ := json.Marshal(map[string]any{"cell": grid.ID(area.Cell), "radius_km": area.RadiusKM, "availability_minutes": 30})
		w := call("POST", "/api/signals", token, string(body))
		if w.Code != 200 {
			t.Fatalf("synthetic enrollment rejected: %d", w.Code)
		}
		return token
	}
	attacker := []string{}
	for n := 1; n < c.Matching.ActivationCount; n++ {
		attacker = append(attacker, create(known))
	}
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	before := call("GET", "/api/signal", attacker[0], "")
	var beforeView sessionView
	json.Unmarshal(before.Body.Bytes(), &beforeView)
	if beforeView.Invitation != nil {
		t.Fatal("fixture contaminated by existing gathering")
	}
	targetToken := create(target)
	defer call("DELETE", "/api/signal", targetToken, "")
	for _, token := range attacker {
		defer call("DELETE", "/api/signal", token, "")
	}
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	backend.AdvanceClock(ctx, int64(c.Matching.ActivationStabilitySeconds)*1000)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	after := call("GET", "/api/signal", attacker[0], "")
	var afterView sessionView
	json.Unmarshal(after.Body.Bytes(), &afterView)
	if afterView.Invitation == nil || afterView.Invitation.Crossing.ID != selected {
		t.Fatal("observed invitation did not reproduce selection model")
	}
	cells := []string{}
	for cell := range possible[selected] {
		cells = append(cells, cell)
	}
	sort.Strings(cells)
	report := map[string]any{
		"status": "privacy_gate_failed", "synthetic_only": true, "config_sha256": configHash, "dataset": dataset.Version,
		"activation_threshold": c.Matching.ActivationCount, "controlled_credentials": len(attacker), "controlled_networks": 1,
		"invitation_without_target": false, "invitation_with_target": true, "compatible_target_cells_before_crossing": len(compatible),
		"possible_target_cells_after_crossing": cells, "actual_synthetic_target_cell": grid.ID(target.Cell), "observed_crossing": selected,
		"assumptions": []string{"Observer controls all but one founding credential and knows its own coarse inputs.", "Exactly one additional synthetic signal exists in this isolated candidate neighborhood.", "Target radius is unknown; all configured radii are included in the inference set.", "No exact target GPS, account identity, database contents, or public aggregate endpoint is observed."},
		"finding":     "The ordinary anonymous invitation reveals the extra signal's participation and narrows its private coarse cell under these assumptions. It does not reveal a civil identity or exact GPS by itself.",
	}
	output := "../../reports/local/privacy-probe.json"
	os.MkdirAll(filepath.Dir(output), 0700)
	data, _ := json.MarshalIndent(report, "", "  ")
	if e = os.WriteFile(output, append(data, '\n'), 0600); e != nil {
		t.Fatal(e)
	}
	t.Logf("KNOWN PRIVACY FAILURE reproduced: %d compatible cells narrowed to %d. See reports/local/privacy-probe.json; this is not a privacy pass.", len(compatible), len(cells))
}
