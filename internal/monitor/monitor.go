// Package monitor keeps bounded operational summaries, never request metadata.
// The fixed publication policy is a privacy bound, not formal anonymity.
package monitor

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const WindowSeconds = 60
const MinimumSamples = 20

type Task int

const (
	Matcher Task = iota
	Publisher
	Cleanup
	Push
	taskCount
)

var taskNames = [taskCount]string{"matcher", "publisher", "cleanup", "push"}
var latencyBounds = [...]int64{10, 50, 100, 300, 1000, 3000, 10000}

type window struct {
	epoch         int64
	count, failed int
	latency       [8]int
}
type taskState struct {
	enabled                    bool
	started, finished, success time.Time
	failed                     bool
}
type Registry struct {
	mu                sync.Mutex
	current, previous window
	tasks             [taskCount]taskState
	now               func() time.Time
}

func New() *Registry { return &Registry{now: time.Now} }
func (m *Registry) Begin(task Task) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task].enabled = true
	m.tasks[task].started = m.now()
}
func (m *Registry) End(task Task, err error) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v := &m.tasks[task]
	v.finished = m.now()
	v.failed = err != nil
	if err == nil {
		v.success = v.finished
	}
}
func (m *Registry) rotate(now time.Time) {
	epoch := now.Unix() / WindowSeconds
	if epoch != m.current.epoch {
		if epoch == m.current.epoch+1 {
			m.previous = m.current
		} else {
			m.previous = window{}
		}
		m.current = window{epoch: epoch}
	}
}
func (m *Registry) observe(duration time.Duration, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rotate(m.now())
	m.current.count++
	if status >= 500 {
		m.current.failed++
	}
	i := 0
	for i < len(latencyBounds) && duration.Milliseconds() > latencyBounds[i] {
		i++
	}
	m.current.latency[i]++
}

type response struct {
	http.ResponseWriter
	status int
}

func (w *response) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *response) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.ResponseWriter.Write(data)
}
func (w *response) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (m *Registry) Wrap(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		v := &response{ResponseWriter: w}
		defer func() { m.observe(time.Since(start), v.status) }()
		next.ServeHTTP(v, r)
	})
}
func bucket(n int) string {
	bounds := []struct {
		n int
		s string
	}{{20, "20+"}, {50, "50+"}, {100, "100+"}, {250, "250+"}, {500, "500+"}, {1000, "1000+"}, {5000, "5000+"}, {10000, "10000+"}}
	value := "suppressed"
	for _, b := range bounds {
		if n >= b.n {
			value = b.s
		}
	}
	return value
}
func age(now, then time.Time) int64 {
	if then.IsZero() {
		return -1
	}
	return max(0, int64(now.Sub(then).Seconds())/5*5)
}

type TaskView struct {
	State              string `json:"state"`
	LastStartSeconds   int64  `json:"last_start_seconds"`
	LastFinishSeconds  int64  `json:"last_finish_seconds"`
	LastSuccessSeconds int64  `json:"last_success_seconds"`
}
type Snapshot struct {
	Version          int                 `json:"version"`
	WindowSeconds    int                 `json:"window_seconds"`
	RetentionSeconds int                 `json:"retention_seconds"`
	MinimumSamples   int                 `json:"minimum_samples"`
	ObservedEpoch    int64               `json:"observed_epoch"`
	Requests         string              `json:"requests"`
	Failures         string              `json:"failures"`
	P95Millis        int64               `json:"p95_upper_ms"`
	Workers          map[string]TaskView `json:"workers"`
}

func (m *Registry) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	m.rotate(now)
	v := m.previous
	s := Snapshot{Version: 1, WindowSeconds: WindowSeconds, RetentionSeconds: 2 * WindowSeconds, MinimumSamples: MinimumSamples, ObservedEpoch: v.epoch, Requests: bucket(v.count), Failures: bucket(v.failed), P95Millis: -1, Workers: map[string]TaskView{}}
	if v.count >= MinimumSamples {
		n := 0
		for i, count := range v.latency {
			n += count
			if n*100 >= v.count*95 {
				if i < len(latencyBounds) {
					s.P95Millis = latencyBounds[i]
				} else {
					s.P95Millis = 0
				}
				break
			}
		}
	}
	for i, t := range m.tasks {
		if !t.enabled {
			continue
		}
		state := "ok"
		if t.finished.Before(t.started) {
			state = "running"
		} else if t.failed {
			state = "error"
		}
		s.Workers[taskNames[i]] = TaskView{state, age(now, t.started), age(now, t.finished), age(now, t.success)}
	}
	return s
}

// Listen refuses to replace existing files and exposes no TCP listener. Its parent
// must be a private directory; the socket itself is owner-only. No response logs.
func (m *Registry) Listen(path string) (net.Listener, error) {
	parent, e := os.Stat(filepath.Dir(path))
	if e != nil || !parent.IsDir() || parent.Mode().Perm()&0077 != 0 {
		return nil, errors.New("monitoring requires a private directory")
	}
	listener, e := net.Listen("unix", path)
	if e != nil {
		return nil, e
	}
	if e = os.Chmod(path, 0600); e != nil {
		listener.Close()
		return nil, e
	}
	return listener, nil
}
func (m *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" || r.URL.Path != "/metrics" || r.URL.RawQuery != "" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.Snapshot())
	})
}
