//go:build simulation

package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
	"io"
	"net"
	"net/http"
)

// The simulation wrapper is absent from production builds. It rejects browsers'
// cross-origin requests, non-loopback peers and callers without the run secret.
func SimulationHandler(c config.Config, backend *store.Store, roads []byte, control string, engines ...*worker.Engine) http.Handler {
	normal := Handler(c, backend, roads, engines...)
	expected := sha256.Sum256([]byte("Bearer " + control))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/simulation/clock" {
			normal.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		actual := sha256.Sum256([]byte(r.Header.Get("Authorization")))
		if c.Profile != "simulation" || backend == nil || err != nil || !ip.IsLoopback() || r.Header.Get("Origin") != "" || r.URL.RawQuery != "" || subtle.ConstantTimeCompare(expected[:], actual[:]) != 1 {
			writeError(w, 403)
			return
		}
		delta := int64(0)
		runWorker := true
		if r.Method == "POST" {
			var value struct {
				Milliseconds int64 `json:"milliseconds"`
				RunWorker    *bool `json:"run_worker,omitempty"`
			}
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128))
			dec.DisallowUnknownFields()
			if dec.Decode(&value) != nil || value.Milliseconds < 0 || value.Milliseconds > 86400000 {
				writeError(w, 400)
				return
			}
			if dec.Decode(new(any)) != io.EOF {
				writeError(w, 400)
				return
			}
			delta = value.Milliseconds
			if value.RunWorker != nil {
				runWorker = *value.RunWorker
			}
		} else if r.Method != "GET" {
			writeError(w, 405)
			return
		}
		now, err := backend.AdvanceClock(r.Context(), delta)
		if err != nil {
			writeError(w, 503)
			return
		}
		if runWorker {
			for i := 0; i < 100; i++ {
				n, e := backend.Cleanup(r.Context(), c.Limits.CleanupBatchSize)
				if e != nil {
					writeError(w, 503)
					return
				}
				if n < c.Limits.CleanupBatchSize {
					break
				}
			}
		}
		if runWorker && len(engines) > 0 && engines[0] != nil {
			if e := engines[0].Step(r.Context()); e != nil {
				writeError(w, 503)
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"simulation": true, "now": now})
	})
}
