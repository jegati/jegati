package httpapi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

func activityHandler(c config.Config, s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			writeError(w, 400)
			return
		}
		if s == nil {
			writeError(w, 503)
			return
		}
		data, now, e := s.LatestActivity(r.Context(), c)
		if e != nil {
			writeError(w, 503)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if len(data) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var release activity.Release
		if json.Unmarshal(data, &release) != nil || release.ExpiresAt <= now {
			writeError(w, 503)
			return
		}
		// Never permit shared caching of a request carrying an authorization credential.
		if r.Header.Get("Authorization") == "" {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, must-revalidate", min(int64(30), int64(c.PublicActivity.ReleaseSeconds), (release.ExpiresAt-now)/1000)))
		}
		etag := fmt.Sprintf("\"%x\"", sha256.Sum256(data))
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if r.Method != http.MethodHead {
			w.Write(data)
		}
	}
}
