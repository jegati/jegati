package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jegati/jegati/internal/buildmode"
	"github.com/jegati/jegati/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	mode := flag.String("mode", "api", "api, config-check or config-show")
	path := flag.String("config", "config/gati.yaml", "functional configuration path")
	address := flag.String("listen", "127.0.0.1:8080", "HTTP bind address")
	storeAddress := flag.String("store-address", "", "Valkey address (empty disables participant writes)")
	passwordFile := flag.String("store-password-file", "", "mounted service password file")
	intersectionFile := flag.String("intersections", "data/tirana/intersections.json", "versioned public intersection dataset")
	mapFile := flag.String("roads", "data/tirana/roads.geojson", "public road asset")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	c, err := config.Load(*path, buildmode.AllowSimulation)
	if err != nil {
		return err
	}
	switch *mode {
	case "config-check":
		_, hash := c.Canonical()
		fmt.Printf("Configuration valid (schema %d, sha256 %s)\n", config.SchemaVersion, hash)
		return nil
	case "config-show":
		data, _ := c.Canonical()
		fmt.Println(string(data))
		return nil
	case "api":
	default:
		return errors.New("unsupported mode")
	}
	if err := validateBind(*address, buildmode.AllowSimulation); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return errors.New("cannot bind HTTP listener")
	}
	var backend *store.Store
	if *storeAddress != "" {
		password, e := os.ReadFile(*passwordFile)
		if e != nil {
			return errors.New("cannot read store credential file")
		}
		backend, e = store.Connect(context.Background(), *storeAddress, "app", strings.TrimSpace(string(password)))
		if e != nil {
			return e
		}
		defer backend.Client.Close()
	}
	roads, e := os.ReadFile(*mapFile)
	if e != nil {
		return errors.New("cannot read public map asset")
	}
	var engine *worker.Engine
	if backend != nil {
		raw, e := os.ReadFile(*intersectionFile)
		if e != nil {
			return errors.New("cannot read intersection dataset")
		}
		var dataset geography.Dataset
		if json.Unmarshal(raw, &dataset) != nil || dataset.Version != c.Geography.IntersectionDataset {
			return errors.New("intersection dataset version mismatch")
		}
		grid, _ := geography.NewGrid(c.Geography.CellSizeMeters)
		index, e := geography.NewIndex(grid, dataset.Intersections, c.Geography.TravelRadiusChoicesKm)
		if e != nil {
			return e
		}
		engine = worker.New(backend, index, c)
	}
	handler, e := configureHandler(c, backend, roads, engine)
	if e != nil {
		return e
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.New(io.Discard, "", 0)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if engine != nil {
		go engine.Run(ctx)
	}
	if backend != nil {
		go func() {
			ticker := time.NewTicker(time.Duration(c.Matching.ReconciliationSeconds) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					for i := 0; i < 10; i++ {
						n, e := backend.Cleanup(ctx, c.Limits.CleanupBatchSize)
						if e != nil || n < c.Limits.CleanupBatchSize {
							break
						}
					}
				}
			}
		}()
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	fmt.Fprintln(os.Stdout, "GATI API started.")
	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.New("HTTP service failed")
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			server.Close()
			return errors.New("HTTP shutdown timed out")
		}
		return nil
	}
}
func validateBind(address string, simulation bool) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return errors.New("invalid listen address")
	}
	if simulation {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return errors.New("simulation builds require a literal loopback bind address")
		}
	}
	return nil
}
