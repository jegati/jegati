package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestReleaseSuppressionExpiryAndCanaries(t *testing.T) {
	m := New()
	now := time.Unix(600, 0)
	m.now = func() time.Time { return now }
	h := m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	for i := 0; i < 19; i++ {
		r := httptest.NewRequest("POST", "/api/signal?synthetic-secret-canary", strings.NewReader("41.327,19.818"))
		r.Header.Set("Authorization", "Bearer synthetic-secret-canary")
		h.ServeHTTP(httptest.NewRecorder(), r)
	}
	now = now.Add(time.Minute)
	v := m.Snapshot()
	if v.Requests != "suppressed" || v.P95Millis != -1 || v.Failures != "suppressed" {
		t.Fatal("sparse observations published")
	}
	for i := 0; i < 20; i++ {
		m.observe(220*time.Millisecond, 503)
	}
	now = now.Add(time.Minute)
	v = m.Snapshot()
	if v.Requests != "20+" || v.Failures != "20+" || v.P95Millis != 300 {
		t.Fatal("released histogram incorrect", v)
	}
	raw, _ := json.Marshal(v)
	for _, secret := range []string{"synthetic-secret-canary", "41.327", "Authorization", "/api/signal"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("request data entered monitoring")
		}
	}
	now = now.Add(2 * time.Minute)
	if m.Snapshot().Requests != "suppressed" {
		t.Fatal("old request observations retained")
	}
	m.Begin(Matcher)
	now = now.Add(6 * time.Second)
	m.End(Matcher, errors.New("synthetic-secret-canary"))
	raw, _ = json.Marshal(m.Snapshot())
	if strings.Contains(string(raw), "canary") || m.Snapshot().Workers["matcher"].State != "error" {
		t.Fatal("worker error leaked or hidden")
	}
}
func TestConcurrentMonitoringAndSocketBoundary(t *testing.T) {
	m := New()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.observe(time.Millisecond, 200)
				m.Begin(Cleanup)
				m.End(Cleanup, nil)
				m.Snapshot()
			}
		}()
	}
	wg.Wait()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	path := filepath.Join(dir, "ops.sock")
	listener, e := m.Listen(path)
	if e != nil {
		t.Fatal(e)
	}
	defer listener.Close()
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("socket permissions")
	}
	if other, e := m.Listen(path); e == nil {
		other.Close()
		t.Fatal("existing socket replaced")
	}
	server := &http.Server{Handler: m.Handler()}
	go server.Serve(listener)
	defer server.Close()
	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", path)
	}}}
	defer client.CloseIdleConnections()
	for _, p := range []string{"/metrics", "/metrics?token=canary", "/api/signal"} {
		r, e := client.Get("http://local" + p)
		if e != nil {
			t.Fatal(e)
		}
		r.Body.Close()
		want := 404
		if p == "/metrics" {
			want = 200
		}
		if r.StatusCode != want || r.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("monitor route/cache boundary")
		}
	}
}
