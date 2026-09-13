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
	mapData, e := os.ReadFile("../../data/tirana/intersections.json")
	if e != nil {
		t.Fatal(e)
	}
	var dataset geography.Dataset
	json.Unmarshal(mapData, &dataset)
	index, e := geography.NewIndex(grid, dataset.Intersections, c.Geography.TravelRadiusChoicesKm)
	if e != nil {
		t.Fatal(e)
	}
	engine := worker.New(backend, index, c)
	handler := Handler(c, backend, nil, engine)
	nonceHeader := ""
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.2:12345"
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		if nonceHeader != "" {
			r.Header.Set("X-Gati-Arrival-Nonce", nonceHeader)
		}
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
		if newcomer.State != "going" || newcomer.Invitation.EndsAt != original.EndsAt || newcomer.Invitation.Intersection.ID != original.Intersection.ID {
			t.Fatal("late admission moved destination/deadline or failed intent")
		}
	}

	// Founders and late joiners use exactly the same challenge and fresh claim.
	w = call("POST", "/api/going", tokens[1], body)
	if w.Code != 200 {
		t.Fatal("founder going failed")
	}
	arrivalCell, _ := grid.CellAt(original.Intersection.Point)
	arrivalBody := `{"cell":"` + grid.ID(arrivalCell) + `"}`
	lastNonce := map[string]string{}
	arrive := func(token string) {
		nonceHeader = newToken()
		lastNonce[token] = nonceHeader
		w := call("POST", "/api/arrival-nonce", token, "")
		if w.Code != 200 {
			t.Fatalf("arrival nonce: %d %s", w.Code, w.Body.String())
		}
		w = call("POST", "/api/arrival", token, `{"cell":"tirana-v1:1000:0:0"}`)
		if w.Code != 409 {
			t.Fatal("remote cell accepted as arrival")
		}
		w = call("POST", "/api/arrival", token, arrivalBody)
		if w.Code != 200 {
			t.Fatalf("arrival: %d %s", w.Code, w.Body.String())
		}
		var first sessionView
		json.Unmarshal(w.Body.Bytes(), &first)
		if first.State != "here" || first.ArrivalUntil == 0 {
			t.Fatal("arrival not recorded")
		}
		w = call("POST", "/api/arrival", token, arrivalBody)
		var again sessionView
		json.Unmarshal(w.Body.Bytes(), &again)
		if w.Code != 200 || again.ArrivalUntil != first.ArrivalUntil {
			t.Fatal("arrival replay renewed freshness")
		}
		nonceHeader = ""
	}
	arrive(tokens[1])
	g, _ := backend.GatheringByID(ctx, original.ID)
	if g.State != "jemi_gati" {
		t.Fatal("one credential or its replay established collective presence")
	}
	arrive(late)
	backend.AdvanceClock(ctx, 9999)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	g, _ = backend.GatheringByID(ctx, original.ID)
	if g.State != "jemi_gati" {
		t.Fatal("presence stability ignored")
	}
	backend.AdvanceClock(ctx, 1)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	g, _ = backend.GatheringByID(ctx, original.ID)
	if g.State != "jemi_ketu" {
		t.Fatal("stable arrivals did not establish JEMI KETU")
	}
	afterPresence := create()
	w = call("POST", "/api/going", afterPresence, body)
	var afterView sessionView
	json.Unmarshal(w.Body.Bytes(), &afterView)
	if w.Code != 200 || afterView.State != "going" || afterView.ArrivalUntil != 0 || afterView.Invitation.State != "jemi_ketu" {
		t.Fatal("joining JEMI KETU failed or automatically counted arrival")
	}
	tokens = append(tokens, afterPresence)
	w = call("DELETE", "/api/arrival", tokens[1], "")
	if w.Code != 200 {
		t.Fatal("retraction failed")
	}
	g, _ = backend.GatheringByID(ctx, original.ID)
	if g.State != "jemi_gati" {
		t.Fatal("retraction did not immediately remove collective presence")
	}
	nonceHeader = lastNonce[tokens[1]]
	w = call("POST", "/api/arrival", tokens[1], arrivalBody)
	if w.Code != 410 {
		t.Fatal("used nonce restored a retracted arrival")
	}
	nonceHeader = ""
	arrive(tokens[1])
	backend.AdvanceClock(ctx, 10000)
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
	}
	g, _ = backend.GatheringByID(ctx, original.ID)
	if g.State != "jemi_ketu" {
		t.Fatal("fresh replacement claims failed new stability")
	}
	backend.AdvanceClock(ctx, 15*60000+1)
	// Reads enforce deadlines even before a scheduled cleanup pass.
	w = call("GET", "/api/signal", tokens[1], "")
	var expired sessionView
	json.Unmarshal(w.Body.Bytes(), &expired)
	if expired.State != "going" || expired.ArrivalUntil != 0 || expired.Invitation.State != "jemi_gati" {
		t.Fatal("stale arrival remained visible before cleanup")
	}
	if e = engine.Step(ctx); e != nil {
		t.Fatal(e)
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
