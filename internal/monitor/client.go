package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

// ReadSocket lets an operator use docker exec without exposing metrics over TCP
// or mounting a writable disk directory into the participant process.
func ReadSocket(path string) (Snapshot, error) {
	var value Snapshot
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: time.Second}).DialContext(ctx, "unix", path)
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Get("http://local/metrics")
	if err != nil {
		return value, errors.New("private monitoring unavailable")
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 8193))
	if err != nil || response.StatusCode != 200 || len(data) > 8192 || json.Unmarshal(data, &value) != nil {
		return Snapshot{}, errors.New("invalid private monitoring response")
	}
	if value.Version != 2 || len(value.Workers) > int(taskCount) {
		return Snapshot{}, errors.New("unsupported private monitoring response")
	}
	for name, worker := range value.Workers {
		known := false
		for _, expected := range taskNames {
			if name == expected {
				known = true
			}
		}
		if !known || (worker.State != "ok" && worker.State != "error" && worker.State != "running") {
			return Snapshot{}, errors.New("invalid worker summary")
		}
	}
	return value, nil
}
