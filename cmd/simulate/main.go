package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/simulation"
	"os"
	"strings"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	target := flag.String("target", "http://127.0.0.1:8082", "isolated literal loopback simulation origin")
	control := flag.String("control-file", ".runtime/simulation/control-password", "simulation credential file")
	scenario := flag.String("scenario", "simulation/scenarios/tirana-evening.yaml", "scenario YAML")
	output := flag.String("output", "reports/local/simulation", "synthetic report directory")
	seed := flag.Uint64("seed", 42, "deterministic seed override")
	flag.Parse()
	s, e := simulation.Load(*scenario)
	if e != nil {
		return e
	}
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			s.Seed = *seed
		}
	})
	raw, e := os.ReadFile(*control)
	if e != nil {
		return fmt.Errorf("cannot read simulation control file")
	}
	c, e := simulation.NewClient(*target, strings.TrimSpace(string(raw)))
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	if s.Behavior != nil {
		report, err := c.RunPopulation(ctx, s)
		if err != nil {
			report.Status = "failed"
			report.Failure = err.Error()
			_ = simulation.WritePopulationReport(*output, "data/tirana/roads.geojson", report)
			return err
		}
		if err = simulation.WritePopulationReport(*output, "data/tirana/roads.geojson", report); err != nil {
			return err
		}
		fmt.Printf("Population simulation: %d synthetic people; %d accepted credentials, %d invitations, %d arrivals, %d gatherings observed. %.1f wall seconds.\nReport: %s/index.html\n", s.Population, report.Counts["credentials_accepted"], report.Counts["credentials_invited"], report.Counts["arrivals_accepted"], report.Counts["gatherings_observed"], report.WallSeconds, *output)
		return nil
	}
	report, e := c.Run(ctx, s)
	if e != nil {
		return e
	}
	if e = simulation.WriteReport(*output, "data/tirana/roads.geojson", report); e != nil {
		return e
	}
	fmt.Printf("Synthetic people: %d; credentials: %d; accepted: %d; rejected: %d; expiry verified: %d.\nReport: %s/index.html\n", s.Population, report.Credentials, report.Accepted, report.Rejected, report.Expired, *output)
	return nil
}
