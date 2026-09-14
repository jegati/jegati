package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"github.com/jegati/jegati/internal/store"
	"mime"
	"net/http"
	"strconv"
	"time"
)

// This is an authorization credential header, not a URL/body field. Both the
// session capability and challenge are required; only their hashes reach storage.
func arrivalNonce(r *http.Request) (string, bool) {
	values := r.Header.Values("X-Gati-Arrival-Nonce")
	if len(values) != 1 {
		return "", false
	}
	raw, e := base64.RawURLEncoding.DecodeString(values[0])
	if e != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != values[0] {
		return "", false
	}
	return store.Hash(raw), true
}
func (s signalAPI) arrival(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if s.engine == nil {
		writeError(w, 503)
		return
	}
	// A short-lived per-capability counter bounds challenge/claim churn in addition
	// to network/global limits. Its window matches the published network window.
	allowed, e := s.store.Allow(r.Context(), "arrival:"+hash, s.config.Limits.ArrivalRequestsPerSignalWindow, time.Duration(s.config.Limits.NetworkWindowSeconds)*time.Second)
	if e != nil {
		s.fail(w, e)
		return
	}
	if !allowed {
		w.Header().Set("Retry-After", strconv.Itoa(s.config.Limits.NetworkWindowSeconds))
		writeError(w, 429)
		return
	}
	if r.Method == "DELETE" {
		value, e := s.store.RetractArrival(r.Context(), hash)
		if e != nil {
			s.fail(w, e)
			return
		}
		s.respond(w, r, hash, value)
		return
	}
	nonce, ok := arrivalNonce(r)
	if !ok {
		writeError(w, 401)
		return
	}
	if r.URL.Path == "/api/arrival-nonce" {
		if r.ContentLength != 0 {
			writeError(w, 400)
			return
		}
		expiry, e := s.store.IssueArrivalNonce(r.Context(), hash, nonce, int64(s.config.Arrivals.NonceSeconds)*1000)
		if e != nil {
			s.fail(w, e)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{"expires_at": expiry})
		return
	}
	typ, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	var request struct {
		Cell string `json:"cell"`
	}
	if readStrictObject(http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes)), &request, "cell") != nil {
		writeError(w, 400)
		return
	}
	cellID := request.Cell
	cell, e := s.grid.Parse(cellID)
	if e != nil {
		writeError(w, 400)
		return
	}
	value, e := s.store.Status(r.Context(), hash)
	if e != nil {
		s.fail(w, e)
		return
	}
	g, e := s.store.GatheringByID(r.Context(), value.Gathering)
	if e != nil {
		s.fail(w, e)
		return
	}
	destination, e := s.grid.CellAt(g.Intersection.Point)
	rings := s.config.Arrivals.AllowedCellNeighborRings
	if e != nil || abs(cell.X-destination.X) > rings || abs(cell.Y-destination.Y) > rings {
		writeError(w, 409)
		return
	}
	value, e = s.store.ConfirmArrival(r.Context(), hash, nonce, g, int64(s.config.Arrivals.FreshnessMinutes)*60000)
	if e != nil {
		s.fail(w, e)
		return
	}
	if e = s.store.ReconcilePresence(r.Context(), g.ID, s.config.Arrivals.ConfirmationCount, s.config.Limits.CleanupBatchSize, int64(s.config.Arrivals.ConfirmationStabilitySeconds)*1000, int64(s.config.Matching.ReconciliationSeconds)*2000); e != nil {
		s.fail(w, e)
		return
	}
	s.respond(w, r, hash, value)
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
