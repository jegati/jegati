package httpapi

import (
	"encoding/json"
	"mime"
	"net/http"
	"slices"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
)

// preview discloses a published gathering's destination only after explicit coarse
// eligibility checking. It creates no signal, reservation, intent or preview record.
func (s signalAPI) preview(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if s.engine == nil {
		writeError(w, 503)
		return
	}
	typ, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	id, input, e := readJoin(http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes)))
	if e != nil {
		writeError(w, 400)
		return
	}
	cell, e := s.grid.Parse(input.Cell)
	if e != nil || !slices.Contains(s.config.Geography.TravelRadiusChoicesKm, input.RadiusKM) || !slices.Contains(s.config.Availability.ChoicesMinutes, input.AvailabilityMinutes) {
		writeError(w, 400)
		return
	}
	// A preview must resolve public discovery, not probe arbitrary private IDs.
	data, now, e := s.store.LatestActivity(r.Context(), s.config)
	if e != nil {
		s.fail(w, e)
		return
	}
	var release activity.Release
	if len(data) == 0 || json.Unmarshal(data, &release) != nil {
		writeError(w, 409)
		return
	}
	found := false
	for _, g := range release.Gatherings {
		if g.ID == id {
			found = true
			break
		}
	}
	if !found {
		writeError(w, 409)
		return
	}
	value, e := s.store.Status(r.Context(), hash)
	if e != nil && e != store.ErrGone {
		s.fail(w, e)
		return
	}
	minimum := int64(s.config.Matching.LateJoinMinRemainingMinutes) * 60000
	if e == nil && (value.Cell != input.Cell || value.RadiusKM != input.RadiusKM || value.AvailabilityMinutes != input.AvailabilityMinutes || value.ExpiresAt < now+minimum || value.Declined[id]) {
		writeError(w, 409)
		return
	}
	g, e := s.store.GatheringByID(r.Context(), id)
	if e == store.ErrGone {
		writeError(w, 409)
		return
	}
	if e != nil {
		s.fail(w, e)
		return
	}
	if g.EndsAt <= now+minimum || !s.engine.Planner.Index.CanReach(geography.ParticipantArea{Cell: cell, RadiusKM: input.RadiusKM}, g.Intersection.ID) {
		writeError(w, 409)
		return
	}
	expires := min(now+int64(s.config.Geography.LocationFixMaxAgeSeconds)*1000, g.EndsAt-minimum, release.ExpiresAt)
	if value.ExpiresAt > 0 {
		expires = min(expires, value.ExpiresAt)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(struct {
		Invitation store.Gathering `json:"invitation"`
		ExpiresAt  int64           `json:"preview_expires_at"`
	}{g, expires})
}
