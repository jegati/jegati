package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const MaxBytes = 64 * 1024

// Load rejects omissions, unknown fields, duplicate keys, aliases, nulls and extra
// documents. The allowSimulation argument must come from the build, not an env var.
func Load(path string, allowSimulation bool) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, errors.New("cannot open configuration")
	}
	defer f.Close()
	return Decode(f, allowSimulation)
}

func Decode(r io.Reader, allowSimulation bool) (Config, error) {
	var c Config
	data, err := io.ReadAll(io.LimitReader(r, MaxBytes+1))
	if err != nil || len(data) > MaxBytes {
		return c, errors.New("configuration unreadable or too large")
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return c, errors.New("invalid YAML configuration")
	}
	if len(root.Content) != 1 {
		return c, errors.New("configuration must be one mapping")
	}
	if err := shape(root.Content[0], reflect.TypeOf(c), "config"); err != nil {
		return c, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&c); err != nil {
		return Config{}, errors.New("invalid configuration field or type")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Config{}, errors.New("multiple configuration documents are forbidden")
	}
	if err := c.Validate(allowSimulation); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Strict shapes also require explicitly supplied zero/false values. Errors never
// echo raw input values, since a misconfigured file could accidentally contain secrets.
func shape(n *yaml.Node, t reflect.Type, path string) error {
	if n.Kind == yaml.AliasNode || n.Tag == "!!null" {
		return fmt.Errorf("%s: aliases/nulls forbidden", path)
	}
	if t.Kind() != reflect.Struct {
		if t.Kind() == reflect.Slice {
			if n.Kind != yaml.SequenceNode {
				return fmt.Errorf("%s: expected sequence", path)
			}
			for _, child := range n.Content {
				if err := shape(child, t.Elem(), path); err != nil {
					return err
				}
			}
		} else {
			want := "!!str"
			if t.Kind() == reflect.Int {
				want = "!!int"
			}
			if t.Kind() == reflect.Bool {
				want = "!!bool"
			}
			if n.Kind != yaml.ScalarNode || n.Tag != want {
				return fmt.Errorf("%s: invalid type", path)
			}
		}
		return nil
	}
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("%s: expected mapping", path)
	}
	fields := map[string]reflect.Type{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fields[f.Tag.Get("yaml")] = f.Type
	}
	seen := map[string]bool{}
	for i := 0; i < len(n.Content); i += 2 {
		key := n.Content[i].Value
		ft, ok := fields[key]
		if !ok || seen[key] {
			return fmt.Errorf("%s: unknown or duplicate field", path)
		}
		seen[key] = true
		if err := shape(n.Content[i+1], ft, path+"."+key); err != nil {
			return err
		}
	}
	for key := range fields {
		if !seen[key] {
			return fmt.Errorf("%s.%s: required", path, key)
		}
	}
	return nil
}

func (c Config) Validate(allowSimulation bool) error {
	if c.SchemaVersion != SchemaVersion {
		return errors.New("unsupported configuration schema")
	}
	if c.Profile != "production" && c.Profile != "simulation" {
		return errors.New("unsupported profile")
	}
	if c.Profile == "simulation" && !allowSimulation {
		return errors.New("simulation profile forbidden by this build")
	}
	// Bound all integers before multiplication; prevent overflow and runaway configuration.
	var positive func(reflect.Value, string) error
	positive = func(v reflect.Value, path string) error {
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if err := positive(v.Field(i), path+"."+v.Type().Field(i).Tag.Get("yaml")); err != nil {
					return err
				}
			}
		case reflect.Slice:
			if v.Len() == 0 || v.Len() > 32 {
				return fmt.Errorf("%s: list size outside bounds", path)
			}
			for i := 0; i < v.Len(); i++ {
				if err := positive(v.Index(i), path); err != nil {
					return err
				}
			}
		case reflect.Int:
			min := int64(1)
			if strings.HasSuffix(path, ".allowed_cell_neighbor_rings") {
				min = 0
			}
			if v.Int() < min || v.Int() > 1000000 {
				return fmt.Errorf("%s: integer outside bounds", path)
			}
		}
		return nil
	}
	if err := positive(reflect.ValueOf(c), "config"); err != nil {
		return err
	}
	ascending := func(a []int) bool {
		for i := 1; i < len(a); i++ {
			if a[i] <= a[i-1] {
				return false
			}
		}
		return true
	}
	a := c.Availability
	m := c.Matching
	r := c.Arrivals
	p := c.PublicActivity
	n := c.Notifications
	g := c.Geography
	if a.MinimumMinutes < 30 || a.MaximumMinutes > 120 || a.MinimumMinutes > a.MaximumMinutes || !ascending(a.ChoicesMinutes) || a.ChoicesMinutes[0] != a.MinimumMinutes || a.ChoicesMinutes[len(a.ChoicesMinutes)-1] != a.MaximumMinutes {
		return errors.New("availability choices must span minimum/maximum within 30..120 minutes")
	}
	if !ascending(g.TravelRadiusChoicesKm) || g.TravelRadiusChoicesKm[len(g.TravelRadiusChoicesKm)-1] > 20 || g.CellSizeMeters < 500 || g.CellSizeMeters > 5000 {
		return errors.New("geographic configuration outside supported bounds")
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`).MatchString(g.CrossingDataset) {
		return errors.New("invalid crossing dataset identifier")
	}
	if m.DestinationRule != "nearest_eligible_crosswalk_to_coarse_group_center" || !m.PreferOpenGatherings {
		return errors.New("unsupported destination or admission policy")
	}
	if c.Limits.ArrivalRequestsPerSignalWindow > 60 {
		return errors.New("arrival request bound exceeded")
	}
	if c.Limits.MaxDeclinesPerSignal > 128 || m.InvitationCooldownSeconds > 300 || (c.Profile == "production" && m.InvitationCooldownSeconds < 30) {
		return errors.New("invitation lifecycle bounds violated")
	}
	if m.ActivationCount > 500 || m.ActivationCount > c.Limits.MaxActiveSignals {
		return errors.New("activation threshold exceeds bounded founding transaction or capacity")
	}
	if m.MaximumGatheringMinutes > 60 || m.MinimumRemainingMinutes > m.MaximumGatheringMinutes || m.LateJoinMinRemainingMinutes > m.MinimumRemainingMinutes {
		return errors.New("inconsistent gathering deadlines")
	}
	if m.MaximumDebounceSeconds > 5 || m.ReconciliationSeconds > 10 || m.ReconciliationSeconds < m.MaximumDebounceSeconds {
		return errors.New("continuous matching processing bounds exceeded")
	}
	if m.MinimumRemainingMinutes*60+m.ActivationStabilitySeconds+m.ReconciliationSeconds >= a.MinimumMinutes*60 {
		return errors.New("minimum availability cannot accommodate matching stability")
	}
	if r.ConfirmationCount > 500 {
		return errors.New("arrival threshold exceeds bounded confirmation cohort")
	}
	if r.FreshnessMinutes > 15 || r.NonceSeconds > 120 || r.NonceSeconds >= r.FreshnessMinutes*60 || r.ConfirmationStabilitySeconds >= r.FreshnessMinutes*60 || r.ConfirmationStabilitySeconds >= m.LateJoinMinRemainingMinutes*60 || r.AllowedCellNeighborRings > 1 {
		return errors.New("arrival lifetime or geography outside bounds")
	}
	if !ascending(p.CountBuckets) || p.CountBuckets[0] != p.MinimumCount {
		return errors.New("public buckets must be increasing and begin at public minimum")
	}
	if !slices.Contains(p.CountBuckets, n.NearbyGatiCount) || !slices.Contains(p.CountBuckets, n.NearbyArrivalCount) {
		return errors.New("nearby alert counts must be public bucket boundaries")
	}
	if p.SnapshotRetentionMinutes > 15 || p.DailySummaryRetentionDays > 30 || p.ReleaseSeconds*(p.DelayEpochs+1) >= p.SnapshotRetentionMinutes*60 {
		return errors.New("public retention must cover delayed releases within privacy caps")
	}
	if n.AreaFollowMaxHours > 24 || n.QueueTtlSeconds > 300 || n.NearbyRadiusKm > 20 || n.ForegroundPollSeconds > 60 || n.PushMinIntervalSeconds < n.QueueTtlSeconds || n.PushMaxPerHour > 60 {
		return errors.New("notification configuration outside bounds")
	}
	if c.Profile == "production" && (m.ActivationCount < 10 || r.ConfirmationCount < 10 || p.MinimumCount < 20 || p.ReleaseSeconds < 300) {
		return errors.New("production privacy floors violated")
	}
	if c.Limits.MaxBodyBytes > 4096 || c.Limits.NetworkWindowSeconds > 600 || c.Limits.CleanupBatchSize > 1000 || c.Limits.NewSignalsPerNetworkWindow > c.Limits.RequestsPerNetworkWindow {
		return errors.New("abuse/cleanup configuration outside bounds")
	}
	return nil
}

// Canonical returns deterministic bytes for publishing and hashing, independent
// of YAML formatting/comments. There are no secret fields in Config.
func (c Config) Canonical() ([]byte, string) {
	data, err := json.Marshal(c)
	if err != nil {
		panic("configuration must be JSON serializable")
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:])
}
