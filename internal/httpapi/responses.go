package httpapi

import (
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
)

// Own-session responses are authenticated and never shared-cacheable. Enumerate
// their wire fields here: storage bookkeeping must not become an API by accident.
// Public cell aggregates have a separate contract in internal/activity.
type sessionView struct {
	ArrivalUntil        int64           `json:"arrival_until,omitempty"`
	Cell                string          `json:"cell"`
	RadiusKM            float64         `json:"radius_km"`
	AvailabilityMinutes int             `json:"availability_minutes"`
	CreatedAt           int64           `json:"created_at"`
	ExpiresAt           int64           `json:"expires_at"`
	State               string          `json:"state"`
	Invitation          *invitationView `json:"invitation,omitempty"`
}

type invitationView struct {
	ID           string           `json:"id"`
	Intersection intersectionView `json:"intersection"`
	ActivatedAt  int64            `json:"activated_at"`
	EndsAt       int64            `json:"ends_at"`
	State        string           `json:"state"`
	ConfigHash   string           `json:"config_sha256"`
}

// These coordinates describe an imported map landmark, never a participant.
// This private destination view is not used by public activity responses.
type intersectionView struct {
	ID        string            `json:"id"`
	SourceIDs []string          `json:"source_ids"`
	Point     geography.Point   `json:"point"`
	Geometry  []geography.Point `json:"geometry,omitempty"`
	Label     string            `json:"label"`
}

func ownSession(s store.Signal) sessionView {
	return sessionView{
		ArrivalUntil: s.ArrivalUntil, Cell: s.Cell, RadiusKM: s.RadiusKM,
		AvailabilityMinutes: s.AvailabilityMinutes, CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt, State: s.State,
	}
}

func privateInvitation(g store.Gathering) *invitationView {
	return &invitationView{
		ID: g.ID, ActivatedAt: g.ActivatedAt, EndsAt: g.EndsAt,
		State: g.State, ConfigHash: g.ConfigHash,
		Intersection: intersectionView{
			ID: g.Intersection.ID, SourceIDs: g.Intersection.SourceIDs,
			Point: g.Intersection.Point, Geometry: g.Intersection.Geometry,
			Label: g.Intersection.Label,
		},
	}
}
