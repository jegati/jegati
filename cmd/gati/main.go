package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/httpapi"
	"github.com/jegati/jegati/internal/monitor"
	"github.com/jegati/jegati/internal/notification"
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
	mode := flag.String("mode", "api", "api, config-check, config-show, health-check or monitor-snapshot")
	path := flag.String("config", "config/gati.yaml", "functional configuration path")
	address := flag.String("listen", "127.0.0.1:8080", "HTTP bind address")
	storeAddress := flag.String("store-address", "", "Valkey address (empty disables participant writes)")
	passwordFile := flag.String("store-password-file", "", "mounted service password file")
	intersectionFile := flag.String("intersections", "data/tirana/intersections.json", "versioned public intersection dataset")
	mapFile := flag.String("roads", "data/tirana/roads.geojson", "public road asset")
	pushKeyFile := flag.String("push-key-file", ".runtime/vapid.json", "mounted VAPID service key file (only read when optional push is enabled)")
	monitorSocket := flag.String("monitor-socket", "", "optional owner-only Unix socket for bounded operational summaries")
	role := flag.String("role", "combined", "combined, api or worker process role")
	trustedProxies := flag.String("trusted-proxies", "", "comma-separated exact proxy CIDRs; requires canonical X-Gati-Client-IP")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if *role != "combined" && *role != "api" && *role != "worker" {
		return errors.New("unsupported process role")
	}
	if buildmode.AllowSimulation && (*role != "combined" || *trustedProxies != "") {
		return errors.New("deployment roles/proxies are unavailable in simulation")
	}
	if *mode == "monitor-snapshot" {
		value, err := monitor.ReadSocket(*monitorSocket)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(value)
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
	case "health-check":
	default:
		return errors.New("unsupported mode")
	}
	if err := validateBind(*address, buildmode.AllowSimulation); err != nil {
		return err
	}
	var listener net.Listener
	if *role != "worker" && *mode != "health-check" {
		listener, err = net.Listen("tcp", *address)
		if err != nil {
			return errors.New("cannot bind HTTP listener")
		}
		defer listener.Close()
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
	if *mode == "health-check" {
		if backend == nil {
			return errors.New("health check requires the store")
		}
		if *role == "worker" {
			value, err := monitor.ReadSocket(*monitorSocket)
			if err != nil {
				return err
			}
			worker, ok := value.Workers["matcher"]
			if !ok || worker.State == "error" || worker.LastStartSeconds < 0 || worker.LastStartSeconds > 30 || worker.LastSuccessSeconds > 30 {
				return errors.New("worker unavailable")
			}
			return nil
		}
		client := http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		response, err := client.Get("http://" + *address + "/healthz")
		if err != nil {
			return errors.New("API unavailable")
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return errors.New("API unavailable")
		}
		return nil
	}
	if *role != "combined" && backend == nil {
		return errors.New("separate roles require the store")
	}
	roads, e := os.ReadFile(*mapFile)
	if e != nil {
		return errors.New("cannot read public map asset")
	}
	if c.Notifications.PushEnabled && backend == nil {
		return errors.New("optional push requires the temporary store")
	}
	var metrics *monitor.Registry
	if *monitorSocket != "" {
		metrics = monitor.New()
		socket, e := metrics.Listen(*monitorSocket)
		if e != nil {
			return errors.New("cannot bind monitoring socket")
		}
		defer socket.Close()
		ops := &http.Server{Handler: metrics.Handler(), ReadHeaderTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second, IdleTimeout: 5 * time.Second, ErrorLog: log.New(io.Discard, "", 0)}
		defer ops.Close()
		go ops.Serve(socket)
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
		engine.Monitor = metrics
		engine.Push, e = notification.New(c, backend, *pushKeyFile)
		if e != nil {
			return e
		}
	}
	if engine != nil && engine.Push != nil {
		engine.Push.Monitor = metrics
	}
	handler, e := configureHandler(c, backend, roads, engine)
	if e != nil {
		return e
	}
	handler, e = httpapi.TrustedProxy(handler, *trustedProxies)
	if e != nil {
		return e
	}
	server := &http.Server{Handler: metrics.Wrap(handler), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.New(io.Discard, "", 0)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if engine != nil && backgroundWorkers && *role == "api" {
		go engine.RunView(ctx)
	}
	if engine != nil && backgroundWorkers && *role != "api" {
		go engine.Run(ctx)
		if engine.Push != nil {
			go engine.Push.Run(ctx)
		}
		publisher := worker.NewActivityPublisher(backend, c)
		publisher.Monitor = metrics
		go publisher.Run(ctx)
	}
	if backend != nil && backgroundWorkers && *role != "api" {
		go func() {
			ticker := time.NewTicker(time.Duration(c.Matching.ReconciliationSeconds) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					metrics.Begin(monitor.Cleanup)
					if metrics != nil {
						lag, _ := backend.CleanupLag(ctx)
						metrics.Lag(monitor.Cleanup, lag)
					}
					var cleanupError error
					for i := 0; i < 10; i++ {
						n, e := backend.Cleanup(ctx, c.Limits.CleanupBatchSize)
						cleanupError = e
						if e != nil || n < c.Limits.CleanupBatchSize {
							break
						}
					}
					metrics.End(monitor.Cleanup, cleanupError)
				}
			}
		}()
	}
	if *role == "worker" {
		<-ctx.Done()
		return nil
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
