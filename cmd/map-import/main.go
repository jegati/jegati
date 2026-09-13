// map-import is an offline developer command, not an application endpoint.
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jegati/jegati/internal/geography"
)

func main() {
	source := flag.String("source", "data/tirana/source/overpass.json.gz", "gzip public source extract")
	output := flag.String("out", "data/tirana", "output directory")
	version := flag.String("version", "tirana-crossings-v1", "public dataset version")
	flag.Parse()
	if err := run(*source, *output, *version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(source, output, version string) error {
	f, e := os.Open(source)
	if e != nil {
		return e
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		return e
	}
	defer z.Close()
	data, e := io.ReadAll(io.LimitReader(z, 64<<20+1))
	if e != nil {
		return e
	}
	dataset, roads, e := geography.Import(data, version)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(output, 0755); e != nil {
		return e
	}
	for _, item := range []struct {
		name  string
		value any
	}{{"crossings.json", dataset}, {"roads.geojson", roads}} {
		b, e := json.Marshal(item.value)
		if e != nil {
			return e
		}
		b = append(b, '\n')
		if e = os.WriteFile(filepath.Join(output, item.name), b, 0644); e != nil {
			return e
		}
	}
	fmt.Printf("Imported %d crossings and %d road segments; source sha256 %s\n", len(dataset.Crossings), len(roads.Features), dataset.SourceSHA256)
	return nil
}
