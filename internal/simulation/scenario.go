// Package simulation generates synthetic input. It is not imported by the API.
package simulation

import (
	"errors"
	"github.com/jegati/jegati/internal/geography"
	"go.yaml.in/yaml/v3"
	"io"
	"math"
	"math/rand/v2"
	"os"
)

type Scenario struct {
	Behavior         *Behavior `yaml:"behavior,omitempty" json:"behavior,omitempty"`
	Seed             uint64    `yaml:"seed" json:"seed"`
	Population       int       `yaml:"population" json:"population"`
	Distribution     string    `yaml:"distribution" json:"distribution"`
	Availability     []int     `yaml:"availability_minutes" json:"availability_minutes"`
	Radii            []float64 `yaml:"radius_km" json:"radius_km"`
	Cancellation     float64   `yaml:"cancellation_probability" json:"cancellation_probability"`
	Duplicate        float64   `yaml:"duplicate_request_fraction" json:"duplicate_request_fraction"`
	SybilFraction    float64   `yaml:"sybil_fraction" json:"sybil_fraction"`
	SybilCredentials int       `yaml:"sybil_credentials" json:"sybil_credentials"`
}
type Input struct {
	Person    int     `json:"-"`
	Cell      string  `json:"cell"`
	Radius    float64 `json:"radius_km"`
	Minutes   int     `json:"availability_minutes"`
	Cancel    bool    `json:"-"`
	Duplicate bool    `json:"-"`
}

func Load(path string) (Scenario, error) {
	var s Scenario
	f, e := os.Open(path)
	if e != nil {
		return s, e
	}
	defer f.Close()
	dec := yaml.NewDecoder(io.LimitReader(f, 65537))
	dec.KnownFields(true)
	if e = dec.Decode(&s); e != nil {
		return s, e
	}
	if e = dec.Decode(new(any)); e != io.EOF {
		return s, errors.New("one scenario document required")
	}
	return s, s.Validate()
}
func (s Scenario) Validate() error {
	if s.Population < 1 || s.Population > 10000 || s.SybilCredentials < 1 || s.SybilCredentials > 10 || len(s.Availability) == 0 || len(s.Availability) > 32 || len(s.Radii) == 0 || len(s.Radii) > 32 {
		return errors.New("invalid bounded population or choices")
	}
	switch s.Distribution {
	case "clustered", "uniform", "sparse", "single_hotspot":
	default:
		return errors.New("unsupported distribution")
	}
	for _, p := range []float64{s.Cancellation, s.Duplicate, s.SybilFraction} {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return errors.New("invalid probability")
		}
	}
	for _, v := range s.Availability {
		if v < 30 || v > 120 {
			return errors.New("unsupported availability")
		}
	}
	for _, v := range s.Radii {
		if !geography.ValidRadius(v) {
			return errors.New("unsupported radius")
		}
	}
	if s.Behavior != nil {
		return s.Behavior.Validate(s)
	}
	return nil
}
func Generate(s Scenario, g geography.Grid) []Input {
	rng := rand.New(rand.NewPCG(s.Seed, s.Seed^0x67617469))
	out := make([]Input, 0, s.Population)
	for person := 0; person < s.Population; person++ {
		lon := g.West + rng.Float64()*(g.East-g.West)
		lat := g.South + rng.Float64()*(g.North-g.South)
		switch s.Distribution {
		case "single_hotspot":
			lon = 19.818
			lat = 41.327
		case "clustered":
			centers := [][2]float64{{19.818, 41.327}, {19.822, 41.307}, {19.799, 41.335}}
			c := centers[rng.IntN(len(centers))]
			lon = c[0] + (rng.Float64()-.5)*.025
			lat = c[1] + (rng.Float64()-.5)*.02
		case "sparse":
			if person%2 == 0 {
				lon = g.West + .002
			} else {
				lon = g.East - .002
			}
		}
		cell, _ := g.CellAt(geography.Point{lon, lat})
		copies := 1
		if rng.Float64() < s.SybilFraction {
			copies = s.SybilCredentials
		}
		for i := 0; i < copies; i++ {
			out = append(out, Input{person, g.ID(cell), s.Radii[rng.IntN(len(s.Radii))], s.Availability[rng.IntN(len(s.Availability))], rng.Float64() < s.Cancellation, rng.Float64() < s.Duplicate})
		}
	}
	return out
}
