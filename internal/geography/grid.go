// Package geography handles public map features and coarse cells, not device history.
package geography

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

// Fixed Tirana service grid, WGS84 longitude/latitude. The grid version is part
// of cell IDs. A different service region must get a new public grid definition.
const GridVersion = "tirana-v1"
const West = 19.75
const South = 41.28
const East = 19.90
const North = 41.38
const ReferenceLat = 41.33
const degrees = math.Pi / 180

type Point = orb.Point

type Grid struct {
	Version    string  `json:"version"`
	West       float64 `json:"west"`
	South      float64 `json:"south"`
	East       float64 `json:"east"`
	North      float64 `json:"north"`
	LatStep    float64 `json:"lat_step"`
	LonStep    float64 `json:"lon_step"`
	SizeMeters int     `json:"size_meters"`
	Columns    int     `json:"columns"`
	Rows       int     `json:"rows"`
}
type Cell struct{ X, Y int }

func NewGrid(size int) (Grid, error) {
	if size < 500 || size > 5000 {
		return Grid{}, errors.New("unsupported grid size")
	}
	dy := float64(size) / (orb.EarthRadius * degrees)
	dx := dy / math.Cos(ReferenceLat*degrees)
	return Grid{GridVersion, West, South, East, North, dy, dx, size, int(math.Ceil((East - West) / dx)), int(math.Ceil((North - South) / dy))}, nil
}
func Inside(p Point) bool {
	return finite(p[0]) && finite(p[1]) && p[0] >= West && p[0] < East && p[1] >= South && p[1] < North
}
func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }
func (g Grid) CellAt(p Point) (Cell, error) {
	if !Inside(p) {
		return Cell{}, errors.New("outside supported geography")
	}
	return Cell{int(math.Floor((p[0] - g.West) / g.LonStep)), int(math.Floor((p[1] - g.South) / g.LatStep))}, nil
}
func (g Grid) ID(c Cell) string { return fmt.Sprintf("%s:%d:%d:%d", g.Version, g.SizeMeters, c.X, c.Y) }
func (g Grid) Parse(id string) (Cell, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 4 || parts[0] != g.Version {
		return Cell{}, errors.New("invalid cell")
	}
	size, e1 := strconv.Atoi(parts[1])
	x, e2 := strconv.Atoi(parts[2])
	y, e3 := strconv.Atoi(parts[3])
	c := Cell{x, y}
	if e1 != nil || e2 != nil || e3 != nil || size != g.SizeMeters || x < 0 || y < 0 || x >= g.Columns || y >= g.Rows || g.ID(c) != id {
		return Cell{}, errors.New("invalid cell")
	}
	return c, nil
}
func (g Grid) Center(c Cell) Point {
	return Point{g.West + (float64(c.X)+0.5)*g.LonStep, g.South + (float64(c.Y)+0.5)*g.LatStep}
}
func (g Grid) Corners(c Cell) []Point {
	x := g.West + float64(c.X)*g.LonStep
	y := g.South + float64(c.Y)*g.LatStep
	return []Point{{x, y}, {x + g.LonStep, y}, {x + g.LonStep, y + g.LatStep}, {x, y + g.LatStep}, {x, y}}
}
func Distance(a, b Point) float64 { return geo.DistanceHaversine(a, b) }

// A geodesic path from cell center to any point in this small rectangle is no
// longer than a straight path in local latitude/longitude coordinates. Use a 6,400 km curvature upper bound
// and scale the library's spherical distance conservatively for WGS84.
// This intentionally misses some reachable boundary cases; it never treats a
// center-only distance as sufficient for all positions represented by a cell.
func (g Grid) MaxDistance(c Cell, p Point) float64 {
	minLat := g.South + float64(c.Y)*g.LatStep
	radius := 6400000 * degrees * math.Hypot(g.LatStep/2, g.LonStep/2*math.Cos(minLat*degrees))
	return Distance(g.Center(c), p)*(6400000/orb.EarthRadius) + radius
}
