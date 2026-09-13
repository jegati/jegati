// Package config owns the public, versioned functional configuration.
package config

const SchemaVersion = 2

type Config struct {
	Limits         Limits         `yaml:"limits" json:"limits"`
	SchemaVersion  int            `yaml:"schema_version" json:"schema_version"`
	Profile        string         `yaml:"profile" json:"profile"`
	Availability   Availability   `yaml:"availability" json:"availability"`
	Geography      Geography      `yaml:"geography" json:"geography"`
	Matching       Matching       `yaml:"matching" json:"matching"`
	Arrivals       Arrivals       `yaml:"arrivals" json:"arrivals"`
	PublicActivity PublicActivity `yaml:"public_activity" json:"public_activity"`
	Notifications  Notifications  `yaml:"notifications" json:"notifications"`
}

type Availability struct {
	MinimumMinutes int   `yaml:"minimum_minutes" json:"minimum_minutes"`
	ChoicesMinutes []int `yaml:"choices_minutes" json:"choices_minutes"`
	MaximumMinutes int   `yaml:"maximum_minutes" json:"maximum_minutes"`
}

type Geography struct {
	CellSizeMeters        int    `yaml:"cell_size_meters" json:"cell_size_meters"`
	TravelRadiusChoicesKm []int  `yaml:"travel_radius_choices_km" json:"travel_radius_choices_km"`
	CrossingDataset       string `yaml:"crossing_dataset" json:"crossing_dataset"`
}

type Matching struct {
	ActivationCount             int    `yaml:"activation_count" json:"activation_count"`
	ActivationStabilitySeconds  int    `yaml:"activation_stability_seconds" json:"activation_stability_seconds"`
	MaximumDebounceSeconds      int    `yaml:"maximum_debounce_seconds" json:"maximum_debounce_seconds"`
	ReconciliationSeconds       int    `yaml:"reconciliation_seconds" json:"reconciliation_seconds"`
	MinimumRemainingMinutes     int    `yaml:"minimum_remaining_minutes" json:"minimum_remaining_minutes"`
	CrossingIndexBatchSize      int    `yaml:"crossing_index_batch_size" json:"crossing_index_batch_size"`
	CandidateBatchSize          int    `yaml:"candidate_batch_size" json:"candidate_batch_size"`
	DestinationRule             string `yaml:"destination_rule" json:"destination_rule"`
	PreferOpenGatherings        bool   `yaml:"prefer_open_gatherings" json:"prefer_open_gatherings"`
	MaximumGatheringMinutes     int    `yaml:"maximum_gathering_minutes" json:"maximum_gathering_minutes"`
	LateJoinMinRemainingMinutes int    `yaml:"late_join_min_remaining_minutes" json:"late_join_min_remaining_minutes"`
}

type Arrivals struct {
	ConfirmationCount            int `yaml:"confirmation_count" json:"confirmation_count"`
	ConfirmationStabilitySeconds int `yaml:"confirmation_stability_seconds" json:"confirmation_stability_seconds"`
	FreshnessMinutes             int `yaml:"freshness_minutes" json:"freshness_minutes"`
	NonceSeconds                 int `yaml:"nonce_seconds" json:"nonce_seconds"`
	AllowedCellNeighborRings     int `yaml:"allowed_cell_neighbor_rings" json:"allowed_cell_neighbor_rings"`
}

type PublicActivity struct {
	MinimumCount              int   `yaml:"minimum_count" json:"minimum_count"`
	CountBuckets              []int `yaml:"count_buckets" json:"count_buckets"`
	ReleaseSeconds            int   `yaml:"release_seconds" json:"release_seconds"`
	DelayEpochs               int   `yaml:"delay_epochs" json:"delay_epochs"`
	SnapshotRetentionMinutes  int   `yaml:"snapshot_retention_minutes" json:"snapshot_retention_minutes"`
	DailySummaryRetentionDays int   `yaml:"daily_summary_retention_days" json:"daily_summary_retention_days"`
}

type Notifications struct {
	NearbyGatiCount        int `yaml:"nearby_gati_count" json:"nearby_gati_count"`
	NearbyArrivalCount     int `yaml:"nearby_arrival_count" json:"nearby_arrival_count"`
	NearbyRadiusKm         int `yaml:"nearby_radius_km" json:"nearby_radius_km"`
	ForegroundPollSeconds  int `yaml:"foreground_poll_seconds" json:"foreground_poll_seconds"`
	PushMinIntervalSeconds int `yaml:"push_min_interval_seconds" json:"push_min_interval_seconds"`
	PushMaxPerHour         int `yaml:"push_max_per_hour" json:"push_max_per_hour"`
	QueueTtlSeconds        int `yaml:"queue_ttl_seconds" json:"queue_ttl_seconds"`
	AreaFollowMaxHours     int `yaml:"area_follow_max_hours" json:"area_follow_max_hours"`
}

type Limits struct {
	MaxBodyBytes               int `yaml:"max_body_bytes" json:"max_body_bytes"`
	NetworkWindowSeconds       int `yaml:"network_window_seconds" json:"network_window_seconds"`
	RequestsPerNetworkWindow   int `yaml:"requests_per_network_window" json:"requests_per_network_window"`
	NewSignalsPerNetworkWindow int `yaml:"new_signals_per_network_window" json:"new_signals_per_network_window"`
	GlobalWritesPerSecond      int `yaml:"global_writes_per_second" json:"global_writes_per_second"`
	MaxActiveSignals           int `yaml:"max_active_signals" json:"max_active_signals"`
	CleanupBatchSize           int `yaml:"cleanup_batch_size" json:"cleanup_batch_size"`
}
