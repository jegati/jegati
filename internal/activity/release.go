// Package activity defines the cell-only public release contract. Its deterministic
// buckets reduce direct disclosure; they do not guarantee resistance to inference.
package activity

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
)

const Version = 1

// Input records are private, bounded capture data. Never serialize them publicly.
type Participant struct {
	Hash, Cell, Gathering, State            string
	ExpiresAt, GatheringUntil, ArrivalUntil int64
}
type Gathering struct {
	ID, Cell, State string
	EndsAt          int64
}
type Area struct {
	Cell    string `json:"cell"`
	Willing int    `json:"willing,omitempty"`
}
type Event struct {
	ID     string `json:"id"`
	Cell   string `json:"cell"`
	EndsAt int64  `json:"ends_at"`
	State  string `json:"state"`
	Going  int    `json:"going,omitempty"`
	Here   int    `json:"here,omitempty"`
}
type Release struct {
	Version       int            `json:"version"`
	ID            string         `json:"id"`
	ConfigHash    string         `json:"config_sha256"`
	MapVersion    string         `json:"map_version"`
	ObservedFrom  int64          `json:"observed_from"`
	ObservedUntil int64          `json:"observed_until"`
	ReleaseAt     int64          `json:"release_at"`
	ExpiresAt     int64          `json:"expires_at"`
	Grid          geography.Grid `json:"grid"`
	Areas         []Area         `json:"areas"`
	Gatherings    []Event        `json:"gatherings"`
}

func Bucket(n int, bounds []int) int {
	result := 0
	for _, b := range bounds {
		if n < b {
			break
		}
		result = b
	}
	return result
}

// PublicCell uses integer cell indices, avoiding floating-point edge ambiguity.
func PublicCell(private, public geography.Grid, id string) (string, error) {
	c, err := private.Parse(id)
	if err != nil || public.SizeMeters < private.SizeMeters || public.SizeMeters%private.SizeMeters != 0 {
		return "", errors.New("incompatible activity cell")
	}
	ratio := public.SizeMeters / private.SizeMeters
	return public.ID(geography.Cell{X: c.X / ratio, Y: c.Y / ratio}), nil
}

// Build creates one release from a bounded observation interval. Every retained
// contribution must still meet deadlines at the end of capture. The reader must
// supply at most one observed state per credential; duplicate input fails closed.
func Build(c config.Config, start, end int64, participants []Participant, gatherings []Gathering) (Release, []byte, error) {
	fail := func() (Release, []byte, error) {
		return Release{}, nil, errors.New("invalid or incomplete activity capture")
	}
	p := c.PublicActivity
	if start < 0 || end < start || end-start > int64(p.CaptureMaxSeconds)*1000 || len(participants) > c.Limits.MaxActiveSignals || len(gatherings) > c.Limits.MaxActiveSignals {
		return fail()
	}
	private, err := geography.NewGrid(c.Geography.CellSizeMeters)
	if err != nil {
		return fail()
	}
	grid, err := geography.NewGrid(p.AreaSizeMeters)
	if err != nil {
		return fail()
	}
	_, hash := c.Canonical()
	epoch := start / (int64(p.ReleaseSeconds) * 1000)
	r := Release{Version: Version, ID: fmt.Sprintf("%s-%d", hash, epoch), ConfigHash: hash, MapVersion: c.Geography.IntersectionDataset, ObservedFrom: start, ObservedUntil: end,
		ReleaseAt: (epoch + 1 + int64(p.DelayEpochs)) * int64(p.ReleaseSeconds) * 1000, ExpiresAt: start + int64(p.SnapshotRetentionMinutes)*60000, Grid: grid, Areas: []Area{}, Gatherings: []Event{}}
	if end >= r.ReleaseAt || r.ReleaseAt >= r.ExpiresAt {
		return fail()
	}
	events := map[string]Gathering{}
	members := map[string]int{}
	going := map[string]int{}
	here := map[string]int{}
	for _, g := range gatherings {
		if _, err := grid.Parse(g.Cell); err != nil || g.ID == "" {
			return fail()
		}
		if _, exists := events[g.ID]; exists {
			return fail()
		}
		if g.EndsAt > end {
			events[g.ID] = g
		}
	}
	willing := map[string]int{}
	seen := map[string]bool{}
	for _, v := range participants {
		if v.Hash == "" || seen[v.Hash] {
			return fail()
		}
		seen[v.Hash] = true
		cell, err := PublicCell(private, grid, v.Cell)
		if err != nil {
			return fail()
		}
		if v.ExpiresAt <= end {
			continue
		}
		willing[cell]++
		if _, ok := events[v.Gathering]; !ok || v.GatheringUntil <= end {
			continue
		}
		if v.State != "invited" && v.State != "going" && v.State != "here" {
			continue
		}
		members[v.Gathering]++
		if v.State == "going" || v.State == "here" {
			going[v.Gathering]++
		}
		if v.State == "here" && v.ArrivalUntil > end {
			here[v.Gathering]++
		}
	}
	for cell, n := range willing {
		if b := Bucket(n, p.CountBuckets); b > 0 {
			r.Areas = append(r.Areas, Area{cell, b})
		}
	}
	for id, g := range events {
		if members[id] < p.MinimumCount {
			continue
		}
		v := Event{ID: id, Cell: g.Cell, EndsAt: g.EndsAt, State: "jemi_gati", Going: Bucket(going[id], p.CountBuckets), Here: Bucket(here[id], p.CountBuckets)}
		if v.Here > 0 && g.State == "jemi_ketu" {
			v.State = "jemi_ketu"
		}
		r.Gatherings = append(r.Gatherings, v)
	}
	sort.Slice(r.Areas, func(i, j int) bool { return r.Areas[i].Cell < r.Areas[j].Cell })
	sort.Slice(r.Gatherings, func(i, j int) bool { return r.Gatherings[i].ID < r.Gatherings[j].ID })
	data, err := json.Marshal(r)
	if err != nil || len(data) > p.MaxSnapshotBytes {
		return fail()
	}
	return r, data, nil
}
