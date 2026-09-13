// Package httpapi exposes the current scaffold's configuration and health only.
package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
	"net/http"
	"time"

	"github.com/jegati/jegati/internal/config"
)

func Handler(c config.Config, backend *store.Store, roads []byte, engines ...*worker.Engine) http.Handler {
	data, hash := c.Canonical()
	envelope, _ := json.Marshal(struct {
		SchemaVersion int             `json:"schema_version"`
		Hash          string          `json:"sha256"`
		Config        json.RawMessage `json:"config"`
	}{config.SchemaVersion, hash, data})
	mux := http.NewServeMux()
	grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
	signal := signalAPI{config: c, grid: grid, store: backend}
	if len(engines) > 0 {
		signal.engine = engines[0]
	}
	mux.HandleFunc("POST /api/signals", signal.create)
	mux.HandleFunc("GET /api/signal", signal.status)
	mux.HandleFunc("DELETE /api/signal", signal.cancel)
	mux.HandleFunc("POST /api/going", signal.changeIntent)
	mux.HandleFunc("POST /api/join", signal.join)
	mux.HandleFunc("POST /api/decline", signal.changeIntent)
	mux.HandleFunc("GET /api/geography", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(grid)
	})
	if len(roads) > 0 {
		sum := sha256.Sum256(roads)
		etag := fmt.Sprintf("\"%x\"", sum)
		mux.HandleFunc("GET /api/map/roads", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/geo+json")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Header().Set("ETag", etag)
			http.ServeContent(w, r, "roads.geojson", time.Time{}, bytes.NewReader(roads))
		})
	}
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			writeError(w, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(envelope)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok","stage":"willingness"}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { writeError(w, http.StatusNotFound) })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !(r.Method == "POST" && (r.URL.Path == "/api/signals" || r.URL.Path == "/api/going" || r.URL.Path == "/api/decline" || r.URL.Path == "/api/join")) && !(r.Method == "DELETE" && r.URL.Path == "/api/signal") {
			w.Header().Set("Allow", "GET, HEAD")
			writeError(w, http.StatusMethodNotAllowed)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func writeError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": "Kërkesa nuk mund të përpunohet."})
}
