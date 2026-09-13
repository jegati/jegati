//go:build simulation

package main

import (
	"errors"
	"flag"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/httpapi"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
	"net/http"
	"os"
	"strings"
)

var controlFile = flag.String("simulation-control-file", "", "dedicated local simulation control credential file")

func configureHandler(c config.Config, backend *store.Store, roads []byte, engine *worker.Engine) (http.Handler, error) {
	if c.Profile != "simulation" || backend == nil {
		return nil, errors.New("simulation API requires simulation profile and dedicated store")
	}
	raw, err := os.ReadFile(*controlFile)
	token := strings.TrimSpace(string(raw))
	if err != nil || len(token) != 64 {
		return nil, errors.New("simulation control credential file required")
	}
	return httpapi.SimulationHandler(c, backend, roads, token, engine), nil
}

const backgroundWorkers = false
