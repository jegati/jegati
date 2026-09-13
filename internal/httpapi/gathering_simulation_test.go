//go:build simulation

package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSimulatedContinuousActivationAndLateAdmission(t *testing.T) {
	if os.Getenv("GATI_SIM_INTEGRATION") != "1" {
		t.Skip("requires isolated simulation store")
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
	// The owning simulation process has stopped; this is a synthetic-only fixture.
	if e = backend.Client.Del(ctx, "gati-sim:worker:lease").Err(); e != nil {
		t.Fatal(e)
	}
	c, e := config.Load("../../config/simulation.yaml", true)
	if e != nil {
		t.Fatal(e)
	}
	grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
	mapData, e := os.ReadFile("../../data/tirana/crossings.json")
	if e != nil {
		t.Fatal(e)
	}
	var dataset geography.Dataset
	json.Unmarshal(mapData, &dataset)
	index, e := geography.NewIndex(grid, dataset.Crossings, c.Geography.TravelRadiusChoicesKm)
	if e != nil {
		t.Fatal(e)
	}
	engine := worker.New(backend, index, c)
	handler := Handler(c, backend, nil, engine)
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.2:12345"
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	newToken := func() string { b := make([]byte, 32); rand.Read(b); return base64.RawURLEncoding.EncodeToString(b) }
	create := func() string {
		token := newToken()
		w := call("POST", "/api/signals", token, `{"cell":"tirana-v1:1000:5:5","radius_km":3,"availability_minutes":30}`)
		if w.Code != 200 {
			t.Fatalf("create: %d %s", w.Code, w.Body.String())
		}
		return token
	}
	tokens := []string{create(), create(), create()}
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	backend.AdvanceClock(ctx, 9999)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	w := call("GET", "/api/signal", tokens[0], "")
	if strings.Contains(w.Body.String(), "invitation") || strings.Contains(w.Body.String(), "_pending") {
		t.Fatal("early invitation or internal reservation leaked")
	}
	backend.AdvanceClock(ctx, 1)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	var founder sessionView
	w = call("GET", "/api/signal", tokens[0], "")
	if json.Unmarshal(w.Body.Bytes(), &founder) != nil || founder.Invitation == nil {
		t.Fatalf("activation missing: %s", w.Body.String())
	}
	if founder.Invitation.ActivatedAt-founder.CreatedAt != 10000 {
		t.Fatal("activation ignored exact stability deadline")
	}
	original := *founder.Invitation
	late := create()
	var newcomer sessionView
	w = call("GET", "/api/signal", late, "")
	json.Unmarshal(w.Body.Bytes(), &newcomer)
	if newcomer.Invitation == nil || newcomer.Invitation.ID != original.ID {
		t.Fatal("late willingness did not receive current gathering")
	}
	body := `{"gathering_id":"` + original.ID + `"}`
	direct := newToken()
	directBody := `{"cell":"tirana-v1:1000:5:5","radius_km":3,"availability_minutes":30,"gathering_id":"` + original.ID + `"}`
	w = call("POST", "/api/join", direct, directBody)
	if w.Code != 200 {
		t.Fatalf("atomic direct join: %d %s", w.Code, w.Body.String())
	}
	var directView sessionView
	json.Unmarshal(w.Body.Bytes(), &directView)
	if directView.State != "going" || directView.Invitation == nil || directView.Invitation.ID != original.ID {
		t.Fatal("direct recipient not atomically admitted")
	}
	tokens = append(tokens, direct)

	for n := 0; n < 2; n++ {
		w = call("POST", "/api/going", late, body)
		if w.Code != 200 {
			t.Fatalf("late admission: %d", w.Code)
		}
		json.Unmarshal(w.Body.Bytes(), &newcomer)
		if newcomer.State != "going" || newcomer.Invitation.EndsAt != original.EndsAt || newcomer.Invitation.Crossing.ID != original.Crossing.ID {
			t.Fatal("late admission moved destination/deadline or failed intent")
		}
	}
	w = call("POST", "/api/decline", tokens[0], body)
	if w.Code != 200 {
		t.Fatalf("decline: %d", w.Code)
	}
	var declined sessionView
	json.Unmarshal(w.Body.Bytes(), &declined)
	if declined.State != "gati" || declined.Invitation != nil || declined.ExpiresAt != founder.ExpiresAt {
		t.Fatal("decline ended willingness or repeated invitation")
	}
	w = call("GET", "/api/signal", tokens[0], "")
	if strings.Contains(w.Body.String(), "invitation") {
		t.Fatal("declined gathering offered again")
	}
	now, _ := backend.Now(ctx)
	backend.AdvanceClock(ctx, original.EndsAt-now-5*60000+1)
	w = call("POST", "/api/join", direct, directBody)
	if w.Code != 200 {
		t.Fatal("existing admission retry failed at cutoff")
	}
	denied := newToken()
	w = call("POST", "/api/join", denied, directBody)
	if w.Code != 410 {
		t.Fatal("late direct join accepted")
	}
	w = call("GET", "/api/signal", denied, "")
	if w.Code != 410 {
		t.Fatal("failed direct join left unintended willingness")
	}
	tooLate := create()
	w = call("POST", "/api/going", tooLate, body)
	if w.Code != 410 {
		t.Fatalf("late cutoff not enforced: %d", w.Code)
	}
	backend.AdvanceClock(ctx, 121*60000)
	engine.Step(ctx)
	for _, token := range append(tokens, late, tooLate) {
		w = call("GET", "/api/signal", token, "")
		if w.Code != 410 {
			t.Fatal("expired participation readable")
		}
	}
}
