package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func source(t *testing.T) string {
	t.Helper()
	b, e := os.ReadFile("../config/testdata/default.yaml")
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestProductionDefaultsAndCanonicalHash(t *testing.T) {
	text := source(t)
	c, err := Decode(strings.NewReader(text), false)
	if err != nil {
		t.Fatal(err)
	}
	if c.Availability.MinimumMinutes != 30 || c.Matching.ActivationCount != 20 || c.Arrivals.ConfirmationCount != 10 {
		t.Fatal("unexpected functional defaults")
	}
	data, hash := c.Canonical()
	other, err := Decode(strings.NewReader("# extra comment\n"+text+"\n"), false)
	if err != nil {
		t.Fatal(err)
	}
	otherData, otherHash := other.Canonical()
	if !bytes.Equal(data, otherData) || hash != otherHash || len(hash) != 64 {
		t.Fatal("canonical representation depends on YAML formatting")
	}
}
func TestInvalidConfigurations(t *testing.T) {
	original := source(t)
	tests := []struct{ name, old, new string }{
		{"unknown field", "profile: production", "profile: production\npassword: do-not-print-me"},
		{"duplicate", "profile: production", "profile: production\nprofile: production"},
		{"missing explicit zero", "  allowed_cell_neighbor_rings: 0\n", ""},
		{"missing section", "availability:", "missing_availability:"},
		{"null", "activation_count: 20", "activation_count: null"},
		{"quoted number", "activation_count: 20", "activation_count: \"20\""},
		{"unsupported schema", "schema_version: 7", "schema_version: 99"},
		{"simulation build guard", "profile: production", "profile: simulation"},
		{"unknown profile", "profile: production", "profile: public"},
		{"minimum under 30", "minimum_minutes: 30", "minimum_minutes: 15"},
		{"long retention", "maximum_minutes: 120", "maximum_minutes: 121"},
		{"choice out of bounds", "[30, 60, 90, 120]", "[30, 60, 90, 121]"},
		{"choice reordered", "[30, 60, 90, 120]", "[30, 90, 60, 120]"},
		{"empty list", "[30, 60, 90, 120]", "[]"},
		{"fine geography", "cell_size_meters: 1000", "cell_size_meters: 10"},
		{"radius", "[1, 3, 5]", "[1, 3, 25]"},
		{"location accuracy beyond coarse cell", "location_max_accuracy_meters: 100", "location_max_accuracy_meters: 501"},
		{"location age too large", "location_fix_max_age_seconds: 60", "location_fix_max_age_seconds: 121"},
		{"location age too small", "location_fix_max_age_seconds: 60", "location_fix_max_age_seconds: 4"},
		{"dataset traversal", "tirana-intersections-v1", "../../secret"},
		{"unbounded founding transaction", "activation_count: 20", "activation_count: 501"},
		{"unbounded decline links", "max_declines_per_signal: 32", "max_declines_per_signal: 129"},
		{"low activation", "activation_count: 20", "activation_count: 2"},
		{"low arrival", "confirmation_count: 10", "confirmation_count: 2"},
		{"low public count", "minimum_count: 20", "minimum_count: 2"},
		{"public cells too precise", "area_size_meters: 1000", "area_size_meters: 100"},
		{"public cells not aligned", "area_size_meters: 1000", "area_size_meters: 1500"},
		{"capture too long", "capture_max_seconds: 30", "capture_max_seconds: 31"},
		{"snapshot too small", "max_snapshot_bytes: 1000000", "max_snapshot_bytes: 100"},
		{"public frequency", "release_seconds: 300", "release_seconds: 10"},
		{"no publication delay", "delay_epochs: 1", "delay_epochs: 0"},
		{"excessive publication delay", "delay_epochs: 1", "delay_epochs: 5"},
		{"bucket ordering", "[20, 50, 100, 250, 500, 1000]", "[20, 100, 50]"},
		{"private alert threshold", "nearby_gati_count: 50", "nearby_gati_count: 49"},
		{"arrival alert threshold", "nearby_arrival_count: 50", "nearby_arrival_count: 49"},
		{"gathering retention", "maximum_gathering_minutes: 60", "maximum_gathering_minutes: 61"},
		{"unusable minimum duration", "minimum_remaining_minutes: 15", "minimum_remaining_minutes: 30"},
		{"late joining timing", "late_join_min_remaining_minutes: 5", "late_join_min_remaining_minutes: 20"},
		{"continuous processing", "reconciliation_seconds: 10", "reconciliation_seconds: 60"},
		{"arrival lifetime", "freshness_minutes: 15", "freshness_minutes: 16"},
		{"nonce lifetime", "nonce_seconds: 120", "nonce_seconds: 121"},
		{"arrival stability", "confirmation_stability_seconds: 10", "confirmation_stability_seconds: 300"},
		{"presence area", "allowed_cell_neighbor_rings: 0", "allowed_cell_neighbor_rings: 2"},
		{"public retention", "snapshot_retention_minutes: 15", "snapshot_retention_minutes: 16"},
		{"summary retention", "daily_summary_retention_days: 30", "daily_summary_retention_days: 31"},
		{"follow retention", "area_follow_max_hours: 24", "area_follow_max_hours: 25"},
		{"queue retention", "queue_ttl_seconds: 300", "queue_ttl_seconds: 301"},
		{"disabled preference", "prefer_open_gatherings: true", "prefer_open_gatherings: false"},
		{"integer overflow", "activation_count: 20", "activation_count: 9223372036854775807"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(original, tt.old) {
				t.Fatal("test substitution missing")
			}
			_, err := Decode(strings.NewReader(strings.Replace(original, tt.old, tt.new, 1)), false)
			if err == nil {
				t.Fatal("accepted unsafe/invalid configuration")
			}
			if strings.Contains(err.Error(), "do-not-print-me") {
				t.Fatal("input leaked into error")
			}
		})
	}
	for _, text := range []string{"", "[]", original + "---\nprofile: production\n", strings.Repeat("x", MaxBytes+1), strings.Replace(original, "minimum_minutes: 30", "minimum_minutes: &m 30", 1) + "alias: *m\n"} {
		if _, err := Decode(strings.NewReader(text), false); err == nil {
			t.Fatal("accepted malformed document")
		}
	}
}
func TestSimulationAndIndependentThresholdConfiguration(t *testing.T) {
	c, err := Load("../../config/simulation.yaml", true)
	if err != nil {
		t.Fatal(err)
	}
	if c.Matching.ActivationCount != 3 || c.Arrivals.ConfirmationCount != 2 {
		t.Fatal("simulation defaults ignored")
	}
	if _, err := Load("../../config/simulation.yaml", false); err == nil {
		t.Fatal("production accepted simulation profile")
	}
	c, err = Decode(strings.NewReader(strings.Replace(source(t), "activation_count: 20", "activation_count: 40", 1)), false)
	if err != nil {
		t.Fatal(err)
	}
	if c.Matching.ActivationCount != 40 || c.Arrivals.ConfirmationCount != 10 || c.Notifications.NearbyGatiCount != 50 {
		t.Fatal("thresholds are coupled")
	}
}
func FuzzDecode(f *testing.F) {
	b, _ := os.ReadFile("../config/testdata/default.yaml")
	f.Add(string(b))
	f.Add("")
	f.Add("profile: null")
	f.Fuzz(func(t *testing.T, text string) {
		c, err := Decode(strings.NewReader(text), false)
		if err == nil {
			if err := c.Validate(false); err != nil {
				t.Fatal(err)
			}
			c.Canonical()
		}
	})
}

func TestFractionalRadiiAndFineGrid(t *testing.T) {
	text := strings.ReplaceAll(source(t), "[1, 3, 5]", "[0.1, 0.5, 1, 3]")
	text = strings.ReplaceAll(text, "cell_size_meters: 1000", "cell_size_meters: 100")
	text = strings.ReplaceAll(text, "location_max_accuracy_meters: 100", "location_max_accuracy_meters: 50")
	if _, err := Decode(strings.NewReader(text), false); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"[0.1, 0.1]", "[0.15, 1]", "[.nan, 1]", "[0, 1]", "[\"0.1\", 1]"} {
		if _, err := Decode(strings.NewReader(strings.ReplaceAll(text, "[0.1, 0.5, 1, 3]", bad)), false); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}
