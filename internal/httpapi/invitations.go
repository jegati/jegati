package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
	"io"
	"mime"
	"net/http"
	"regexp"
	"slices"
)

type sessionView struct {
	store.Signal
	Invitation *store.Gathering `json:"invitation,omitempty"`
}

func (s signalAPI) respond(w http.ResponseWriter, r *http.Request, hash string, value store.Signal) {
	view := sessionView{Signal: value.Public()}
	if s.engine != nil {
		now, e := s.store.Now(r.Context())
		if e != nil {
			s.fail(w, e)
			return
		}
		if value.Gathering != "" && value.GatheringUntil > now {
			g, e := s.store.GatheringByID(r.Context(), value.Gathering)
			if e == nil {
				view.Invitation = &g
			} else if e != store.ErrGone {
				s.fail(w, e)
				return
			}
		}
		if view.Invitation == nil && value.InviteAfter <= now && len(value.Declined) < s.config.Limits.MaxDeclinesPerSignal {
			cell, _ := s.grid.Parse(value.Cell)
			participant := matching.Signal{Area: geography.ParticipantArea{Cell: cell, RadiusKM: value.RadiusKM}, ExpiresAt: value.ExpiresAt, Declined: value.Declined}
			open := []matching.Gathering{}
			for _, g := range s.engine.Open() {
				open = append(open, matching.Gathering{ID: g.ID, Intersection: g.Intersection, EndsAt: g.EndsAt})
			}
			if g, ok := s.engine.Planner.Offer(now, participant, open); ok {
				updated, e := s.store.SetIntent(r.Context(), hash, g.ID, "offer", value.Cell, value.RadiusKM, s.config.Limits.MaxDeclinesPerSignal, int64(s.config.Matching.LateJoinMinRemainingMinutes)*60000, int64(s.config.Matching.InvitationCooldownSeconds)*1000)
				if e == nil {
					view.Signal = updated.Public()
					full, e := s.store.GatheringByID(r.Context(), g.ID)
					if e == nil {
						view.Invitation = &full
					}
				}
			}
		}
		if view.Invitation == nil {
			view.State = "gati"
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(view)
}

var gatheringID = regexp.MustCompile(`^[a-f0-9]{32}$`)

func (s signalAPI) changeIntent(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if s.engine == nil {
		writeError(w, 503)
		return
	}
	typ, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	var id string
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes)))
	opening, e := dec.Token()
	if e != nil || opening != json.Delim('{') {
		writeError(w, 400)
		return
	}
	key, e := dec.Token()
	if e != nil || key != "gathering_id" || dec.Decode(&id) != nil || !gatheringID.MatchString(id) {
		writeError(w, 400)
		return
	}
	closing, e := dec.Token()
	if e != nil || closing != json.Delim('}') || dec.Decode(new(any)) != io.EOF {
		writeError(w, 400)
		return
	}
	value, e := s.store.Status(r.Context(), hash)
	if e != nil {
		s.fail(w, e)
		return
	}
	g, e := s.store.GatheringByID(r.Context(), id)
	if e != nil {
		s.fail(w, e)
		return
	}
	cell, e := s.grid.Parse(value.Cell)
	if e != nil || !s.engine.Planner.Index.CanReach(geography.ParticipantArea{Cell: cell, RadiusKM: value.RadiusKM}, g.Intersection.ID) {
		writeError(w, 409)
		return
	}
	action := "going"
	if r.URL.Path == "/api/decline" {
		action = "decline"
	}
	value, e = s.store.SetIntent(r.Context(), hash, g.ID, action, value.Cell, value.RadiusKM, s.config.Limits.MaxDeclinesPerSignal, int64(s.config.Matching.LateJoinMinRemainingMinutes)*60000, int64(s.config.Matching.InvitationCooldownSeconds)*1000)
	if e != nil {
		s.fail(w, e)
		return
	}
	s.respond(w, r, hash, value)
}

// join permits a notification/map recipient to create willingness and admit it
// in one atomic operation. Merely viewing that notification never calls this.
func (s signalAPI) join(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if s.engine == nil {
		writeError(w, 503)
		return
	}
	typ, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes)))
	opening, e := dec.Token()
	if e != nil || opening != json.Delim('{') {
		writeError(w, 400)
		return
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		key, e := dec.Token()
		if e != nil {
			writeError(w, 400)
			return
		}
		name, ok := key.(string)
		if !ok || fields[name] != nil {
			writeError(w, 400)
			return
		}
		var value json.RawMessage
		if dec.Decode(&value) != nil {
			writeError(w, 400)
			return
		}
		fields[name] = value
	}
	if _, e = dec.Token(); e != nil || dec.Decode(new(any)) != io.EOF {
		writeError(w, 400)
		return
	}
	var id string
	if json.Unmarshal(fields["gathering_id"], &id) != nil || !gatheringID.MatchString(id) {
		writeError(w, 400)
		return
	}
	delete(fields, "gathering_id")
	payload, _ := json.Marshal(fields)
	input, e := readCreate(bytes.NewReader(payload))
	if e != nil {
		writeError(w, 400)
		return
	}
	cell, e := s.grid.Parse(input.Cell)
	if e != nil || !slices.Contains(s.config.Geography.TravelRadiusChoicesKm, input.RadiusKM) || !slices.Contains(s.config.Availability.ChoicesMinutes, input.AvailabilityMinutes) {
		writeError(w, 400)
		return
	}
	g, e := s.store.GatheringByID(r.Context(), id)
	if e != nil {
		s.fail(w, e)
		return
	}
	if !s.engine.Planner.Index.CanReach(geography.ParticipantArea{Cell: cell, RadiusKM: input.RadiusKM}, g.Intersection.ID) {
		writeError(w, 409)
		return
	}
	value, e := s.store.CreateAndJoin(r.Context(), hash, input.Cell, input.RadiusKM, input.AvailabilityMinutes, s.config.Limits.MaxActiveSignals, g, int64(s.config.Matching.LateJoinMinRemainingMinutes)*60000)
	if e != nil {
		s.fail(w, e)
		return
	}
	s.respond(w, r, hash, value)
}
