package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
)

// Response fields are intentionally independent of storage structs. Adding a
// storage field must never silently expand the browser's data contract.
func assertJSONFields(t *testing.T, value map[string]json.RawMessage, required, optional []string) {
	t.Helper()
	for _, key := range required {
		if _, ok := value[key]; !ok {
			t.Fatalf("missing response field %q", key)
		}
	}
	for key := range value {
		if !slices.Contains(required, key) && !slices.Contains(optional, key) {
			t.Fatalf("unexpected response field %q", key)
		}
	}
}

func TestPrivateInvitationWireContract(t *testing.T) {
	g := store.Gathering{
		ID: "synthetic-gathering", ActivatedAt: 1000, EndsAt: 90000,
		State: "jemi_gati", ConfigHash: "synthetic-config",
		Intersection: geography.Intersection{
			ID: "osm-junction", SourceIDs: []string{"osm-source"},
			Point: geography.Point{19.82, 41.32}, Label: "Kryqëzim",
		},
	}
	for _, geometry := range [][]geography.Point{nil, {{19.82, 41.32}, {19.821, 41.321}}} {
		g.Intersection.Geometry = geometry
		encoded, err := json.Marshal(privateInvitation(g))
		if err != nil {
			t.Fatal(err)
		}
		assertInvitationContract(t, encoded)
		var decoded store.Gathering
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, g) {
			t.Fatal("private map landmark or gathering contract changed")
		}
	}
}

func assertSessionContract(t *testing.T, body []byte) {
	t.Helper()
	var value map[string]json.RawMessage
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	assertJSONFields(t, value, []string{"cell", "radius_km", "availability_minutes", "created_at", "expires_at", "state"}, []string{"arrival_until", "invitation"})
	if raw, ok := value["invitation"]; ok {
		assertInvitationContract(t, raw)
	}
}

func assertInvitationContract(t *testing.T, body []byte) {
	t.Helper()
	var value map[string]json.RawMessage
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	assertJSONFields(t, value, []string{"id", "intersection", "activated_at", "ends_at", "state", "config_sha256"}, nil)
	var intersection map[string]json.RawMessage
	if err := json.Unmarshal(value["intersection"], &intersection); err != nil {
		t.Fatal(err)
	}
	assertJSONFields(t, intersection, []string{"id", "source_ids", "point", "label"}, []string{"geometry"})
}

func TestOwnSessionResponseAllowlist(t *testing.T) {
	for _, state := range []string{"gati", "invited", "going", "here"} {
		t.Run(state, func(t *testing.T) {
			signal := store.Signal{
				Cell: "tirana-v1:100:55:55", RadiusKM: 3, AvailabilityMinutes: 30,
				CreatedAt: 1000, ExpiresAt: 1801000, State: state,
				PushRevision: 999, ArrivalMember: "private-arrival-member",
				InviteAfter: 1234, Pending: "private-reservation", PendingUntil: 2345,
				Gathering: "private-link", GatheringUntil: 3456,
				Declined: map[string]bool{"private-decline": true},
			}
			if state == "here" {
				signal.ArrivalUntil = 45000
			}
			response := httptest.NewRecorder()
			signalAPI{}.respond(response, httptest.NewRequest("GET", "/api/signal", nil), "private-capability-hash", signal)
			assertSessionContract(t, response.Body.Bytes())
			if strings.Contains(response.Body.String(), "private-") {
				t.Fatal("private bookkeeping leaked")
			}
			var got map[string]json.RawMessage
			json.Unmarshal(response.Body.Bytes(), &got)
			_, hasArrival := got["arrival_until"]
			if hasArrival != (state == "here") {
				t.Fatal("arrival omission contract changed")
			}
			var decoded store.Signal
			json.Unmarshal(response.Body.Bytes(), &decoded)
			if decoded.State != state || decoded.ExpiresAt != signal.ExpiresAt || decoded.Cell != signal.Cell {
				t.Fatal("own-session values changed")
			}
		})
	}
}
