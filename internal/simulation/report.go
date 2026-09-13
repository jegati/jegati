package simulation

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/jegati/jegati/internal/geography"
)

// Dashboard contains only synthetic totals on the public Tirana road extract.
// It is a standalone local artifact; no participant endpoint can serve it.
func WriteReport(directory, roadsPath string, r Report) error {
	if e := os.MkdirAll(directory, 0700); e != nil {
		return e
	}
	data, _ := json.MarshalIndent(r, "", "  ")
	if e := os.WriteFile(filepath.Join(directory, "report.json"), append(data, '\n'), 0600); e != nil {
		return e
	}
	var roads struct {
		Features []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	raw, e := os.ReadFile(roadsPath)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(raw, &roads); e != nil {
		return e
	}
	var drawing strings.Builder
	project := func(lon, lat float64) (float64, float64) {
		return (lon - r.Grid.West) / (r.Grid.East - r.Grid.West) * 900, 700 - (lat-r.Grid.South)/(r.Grid.North-r.Grid.South)*700
	}
	for _, feature := range roads.Features {
		var path strings.Builder
		for i, p := range feature.Geometry.Coordinates {
			if len(p) != 2 {
				continue
			}
			x, y := project(p[0], p[1])
			command := "L"
			if i == 0 {
				command = "M"
			}
			fmt.Fprintf(&path, "%s%.1f %.1f", command, x, y)
		}
		fmt.Fprintf(&drawing, `<path d="%s" fill="none" stroke="#c3c8bf" stroke-width=".7"/>`, path.String())
	}
	grid, _ := geography.NewGrid(r.Grid.SizeMeters)
	for cell, n := range r.Cells {
		c, e := grid.Parse(cell)
		if e != nil {
			return e
		}
		center := grid.Center(c)
		x, y := project(center[0], center[1])
		fmt.Fprintf(&drawing, `<circle cx="%.1f" cy="%.1f" r="17" fill="#963754"/><text x="%.1f" y="%.1f" text-anchor="middle" fill="white" font-size="12">%d</text>`, x, y, x, y+4, n)
	}
	page := `<!doctype html><html lang="sq"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>GATI — Simulim</title><style>body{font:16px system-ui;background:#fff8f3;color:#352c31;max-width:960px;margin:2rem auto;padding:1rem}svg{width:100%;background:#ecefe6}strong{color:#963754}</style><h1>🦩 SIMULIM — Tiranë</h1><strong>Vetëm të dhëna sintetike. Kjo nuk është hartë e pjesëmarrjes reale.</strong><p>Fara: {{.Seed}} · Persona sintetikë: {{.People}} · Kredenciale: {{.Credentials}} · Pranuar: {{.Accepted}} · Refuzuar: {{.Rejected}}</p><p>Numrat janë sinjale sintetike të pranuara gjatë provës, jo persona ose prani aktuale. Ora u çua përpara për të kontrolluar skadimin.</p><svg viewBox="0 0 900 700" role="img" aria-label="Harta e sinjaleve sintetike në Tiranë">{{.Drawing}}</svg><p>© OpenStreetMap · ODbL. Konfigurimi: {{.Hash}}</p><p>Aktivizimi, mbërritjet dhe njoftimet ende nuk simulohen.</p></html>`
	t, e := template.New("report").Parse(page)
	if e != nil {
		return e
	}
	f, e := os.Create(filepath.Join(directory, "index.html"))
	if e != nil {
		return e
	}
	defer f.Close()
	return t.Execute(f, struct {
		Seed                                    uint64
		People, Credentials, Accepted, Rejected int
		Hash                                    string
		Drawing                                 template.HTML
	}{r.Scenario.Seed, r.Scenario.Population, r.Credentials, r.Accepted, r.Rejected, r.ConfigHash, template.HTML(drawing.String())})
}
