package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jegati/jegati/internal/config"
)

func TestPublicConfigIsCanonicalAndNoTracking(t *testing.T) {
	c, err := config.Load("../../config/gati.yaml", false)
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(c, nil, nil)
	for _, target := range []string{"/api/config", "/healthz", "/missing"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
		if w.Header().Get("Set-Cookie") != "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("unexpected cookie/CORS")
		}
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatal("missing response protections")
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/config", nil))
	var envelope struct {
		Hash   string          `json:"sha256"`
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(envelope.Config)
	if envelope.Hash != hex.EncodeToString(sum[:]) {
		t.Fatal("published configuration hash mismatch")
	}
}
func TestNoParticipantOrSimulationEndpoints(t *testing.T) {
	c, err := config.Load("../../config/gati.yaml", false)
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(c, nil, nil)
	for _, target := range []string{"/api/signals", "/api/members", "/api/admin", "/api/simulation/clock"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", target, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/config", strings.NewReader(`{"activation_count":1}`)))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatal("config is writable")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/config?private=secret", nil))
	if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "secret") {
		t.Fatal("query rejection leaked input")
	}
}
