// Offline, synthetic study tool; excluded from application/release entrypoints.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/simulation"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	configPath := flag.String("config", "config/gati.yaml", "current application config")
	studyPath := flag.String("study", "simulation/studies/tirana-day.json", "bounded synthetic assumptions")
	output := flag.String("output", "reports/local/tirana-day.json", "new report file")
	flag.Parse()
	cfg, err := config.Load(*configPath, false)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(*studyPath)
	if err != nil {
		return err
	}
	var study struct {
		Seeds     []uint64                 `json:"seeds"`
		Scenarios []simulation.DayScenario `json:"scenarios"`
	}
	f, err := os.Open(*studyPath)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 65537))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&study); err != nil {
		return err
	}
	if dec.Decode(new(any)) != io.EOF {
		return fmt.Errorf("one study document required")
	}
	if len(study.Seeds) < 1 || len(study.Seeds) > 10 || len(study.Scenarios) < 1 || len(study.Scenarios) > 24 {
		return fmt.Errorf("study matrix exceeds bounds")
	}
	for _, s := range study.Scenarios {
		if err = s.Validate(cfg); err != nil {
			return err
		}
	}
	grid, _ := geography.NewGrid(cfg.Geography.CellSizeMeters)
	mapRaw, err := os.ReadFile("data/tirana/intersections.json")
	if err != nil {
		return err
	}
	var dataset geography.Dataset
	if err = json.Unmarshal(mapRaw, &dataset); err != nil {
		return err
	}
	if dataset.Version != cfg.Geography.IntersectionDataset {
		return fmt.Errorf("map version mismatch")
	}
	fmt.Println("Building the real Tirana geographic index…")
	index, err := geography.NewIndex(grid, dataset.Intersections, cfg.Geography.TravelRadiusChoicesKm)
	if err != nil {
		return err
	}
	_, hash := cfg.Canonical()
	revision, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	status, _ := exec.Command("git", "status", "--porcelain").Output()
	report := map[string]any{"kind": "offline-synthetic-behavioral-model-v1", "config": cfg, "config_sha256": hash, "study_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "map_sha256": fmt.Sprintf("%x", sha256.Sum256(mapRaw)), "source_revision": string(revision), "source_dirty": len(status) > 0, "go": runtime.Version(), "grid": grid, "start_hour": 8, "results": []simulation.DayResult{}}
	// Refuse to overwrite prior evidence. No network, store credentials or endpoints.
	out, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	results := []simulation.DayResult{}
	started := time.Now()
	for _, s := range study.Scenarios {
		for _, seed := range study.Seeds {
			result, err := simulation.RunDay(cfg, index, s, seed)
			if err != nil {
				return err
			}
			results = append(results, result)
			fmt.Printf("%s seed %d: %.0f activated, %.0f confirmed, %.1f hours with confirmed presence, %.0f cells\n", s.Name, seed, result.Metrics["gatherings_activated"], result.Metrics["gatherings_ever_confirmed"], result.Metrics["minutes_any_confirmed"]/60, result.Metrics["confirmed_cells"])
		}
	}
	report["results"] = results
	report["wall_seconds"] = time.Since(started).Seconds()
	report["status"] = "completed"
	enc := json.NewEncoder(out)
	return enc.Encode(report)
}
