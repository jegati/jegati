//go:build simulation

package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
)

func TestSimulatedArrivalRenewal(t *testing.T) {
	if os.Getenv("GATI_SIM_INTEGRATION") != "1" {
		t.Skip("requires isolated simulation store")
	}
	ctx := context.Background()
	password, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal(e)
	}
	s, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(password)))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Client.Close()
	c, e := config.Load("../../config/simulation.yaml", true)
	if e != nil {
		t.Fatal(e)
	}
	grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
	cell := geography.Cell{X: 5, Y: 5}
	token := func() string {
		b := make([]byte, 32)
		if _, e := rand.Read(b); e != nil {
			t.Fatal(e)
		}
		return base64.RawURLEncoding.EncodeToString(b)
	}
	id := strings.Repeat("9", 32)
	p := matching.Proposal{Intersection: geography.Intersection{ID: "node/renewal-test", Point: grid.Center(cell)}}
	var capability, hash string
	for i := 0; i < 3; i++ {
		capability = token()
		raw, _ := base64.RawURLEncoding.DecodeString(capability)
		hash = store.Hash(raw)
		v, e := s.Create(ctx, hash, grid.ID(cell), 3, 30, 30*time.Minute, 1000)
		if e != nil {
			t.Fatal(e)
		}
		p.Founders = append(p.Founders, matching.Signal{Hash: hash, Area: geography.ParticipantArea{Cell: cell, RadiusKM: 3}, CreatedAt: v.CreatedAt, ExpiresAt: v.ExpiresAt})
		founderHash := hash
		defer s.Cancel(ctx, founderHash)
	}
	// The actual transaction and API are used; only map choice/population and time
	// are synthetic. No production test controls are introduced.
	if _, e = s.Reserve(ctx, id, "renewal-test", p, grid, 10000, 20000, 1000); e != nil {
		t.Fatal(e)
	}
	advance := func(ms int64) {
		t.Helper()
		if _, e := s.AdvanceClock(ctx, ms); e != nil {
			t.Fatal(e)
		}
	}
	advance(10000)
	g, e := s.Activate(ctx, id, "renewal-test", 1000, 3600000)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SetIntent(ctx, hash, id, "going", grid.ID(cell), 3, 32, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	index, e := geography.NewIndex(grid, []geography.Intersection{p.Intersection}, c.Geography.TravelRadiusChoicesKm)
	if e != nil {
		t.Fatal(e)
	}
	handler := Handler(c, s, nil, worker.New(s, index, c))
	nonce := token()
	call := func(path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.8:12345"
		r.Header.Set("Authorization", "Bearer "+capability)
		r.Header.Set("X-Gati-Arrival-Nonce", nonce)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s: %d want %d: %s", path, w.Code, want, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private response cacheable")
		}
		if w.Code == 200 && (path == "/api/arrival" || path == "/api/arrival-renewal") {
			assertSessionContract(t, w.Body.Bytes())
		}
		return w
	}
	body := `{"cell":"` + grid.ID(cell) + `"}`
	call("/api/arrival-nonce", "", 200)
	call("/api/arrival", body, 200)
	before, e := s.Status(ctx, hash)
	if e != nil {
		t.Fatal(e)
	}
	nonce = token()
	advance(13*60000 - 1)
	call("/api/arrival-renewal-nonce", "", 409)
	advance(1)
	w := call("/api/arrival-renewal-nonce", "", 200)
	var challenge struct {
		ExpiresAt int64 `json:"expires_at"`
	}
	if e = json.Unmarshal(w.Body.Bytes(), &challenge); e != nil {
		t.Fatal(e)
	}
	if challenge.ExpiresAt > before.ArrivalUntil {
		t.Fatal("renewal challenge outlived old confirmation")
	}
	call("/api/arrival", body, 410)
	call("/api/arrival-renewal", `{"cell":"tirana-v1:1000:0:0"}`, 409)
	unchanged, e := s.Status(ctx, hash)
	if e != nil || unchanged.ArrivalUntil != before.ArrivalUntil {
		t.Fatal("wrong location changed presence", e)
	}
	call("/api/arrival-renewal", `{"cell":"`+grid.ID(cell)+`","latitude":41.3}`, 400)
	call("/api/arrival-renewal", body, 200)
	renewed, e := s.Status(ctx, hash)
	if e != nil {
		t.Fatal(e)
	}
	if renewed.ArrivalUntil != before.ArrivalUntil+13*60000 || renewed.ExpiresAt != before.ExpiresAt || renewed.GatheringUntil != g.EndsAt || renewed.ArrivalMember != before.ArrivalMember {
		t.Fatal("renewal changed original deadlines or contribution")
	}
	advance(1000)
	call("/api/arrival-renewal", body, 200)
	replay, e := s.Status(ctx, hash)
	if e != nil || replay.ArrivalUntil != renewed.ArrivalUntil {
		t.Fatal("retry extended freshness", e)
	}
	// Second renewal occurs after the admission cutoff, but this participant is
	// already admitted. It is capped by the original gathering/session deadline.
	advance(13*60000 - 1000)
	nonce = token()
	call("/api/arrival-renewal-nonce", "", 200)
	call("/api/arrival-renewal", body, 200)
	capped, e := s.Status(ctx, hash)
	if e != nil || capped.ArrivalUntil != g.EndsAt {
		t.Fatal("renewal not capped at original deadline", e)
	}
	if count := s.Client.ZCard(ctx, "gati-sim:arrivals:"+id).Val(); count != 1 {
		t.Fatal("renewal inflated count", count)
	}
	nonce = token()
	call("/api/arrival-renewal-nonce", "", 409)
	now, e := s.Now(ctx)
	if e != nil {
		t.Fatal(e)
	}
	advance(g.EndsAt - now)
	call("/api/arrival-renewal", body, 410)
}
