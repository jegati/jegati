//go:build simulation

package httpapi

import (
	"github.com/jegati/jegati/internal/config"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimulationControlsFailClosed(t *testing.T) {
	c, e := config.Load("../../config/simulation.yaml", true)
	if e != nil {
		t.Fatal(e)
	}
	handler := SimulationHandler(c, nil, nil, strings.Repeat("a", 64))
	for _, origin := range []string{"", "http://hostile.example"} {
		r := httptest.NewRequest("POST", "/api/simulation/clock", strings.NewReader(`{"milliseconds":1800000}`))
		r.RemoteAddr = "127.0.0.1:10000"
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != 403 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("control did not fail closed")
		}
	}
}
