// Package httpapi exposes the current scaffold's configuration and health only.
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jegati/jegati/internal/config"
)

func Handler(c config.Config) http.Handler {
	data, hash := c.Canonical()
	envelope, _ := json.Marshal(struct {
		SchemaVersion int             `json:"schema_version"`
		Hash          string          `json:"sha256"`
		Config        json.RawMessage `json:"config"`
	}{config.SchemaVersion, hash, data})
	mux := http.NewServeMux()
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
		w.Write([]byte(`{"status":"ok","stage":"scaffold"}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { writeError(w, http.StatusNotFound) })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
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
