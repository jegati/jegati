package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jegati/jegati/internal/buildmode"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/httpapi"
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
	server := &http.Server{Handler: httpapi.Handler(c), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.New(io.Discard, "", 0)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	fmt.Fprintln(os.Stdout, "GATI API scaffold started; no participant endpoints are implemented.")
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
