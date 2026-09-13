// loadtest is a synthetic-only HTTP load generator for an owned lab process.
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"io"
	"math"
	rand2 "math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type owner struct {
	Base string `json:"base"`
	PID  int    `json:"pid"`
}
type actor struct {
	token, cell string
	accepted    bool
}
type measurement struct {
	mu               sync.Mutex
	Name             string         `json:"name"`
	Started          int64          `json:"started_ms"`
	Seconds          float64        `json:"seconds"`
	Attempted        int            `json:"attempted"`
	GeneratorDropped int            `json:"generator_dropped"`
	Statuses         map[string]int `json:"statuses"`
	Latency          [9]int         `json:"latency_bins"`
	AcceptedLatency  [9]int         `json:"accepted_latency_bins"`
	RejectedLatency  [9]int         `json:"rejected_latency_bins"`
	AcceptedP95      int64          `json:"accepted_p95_upper_ms"`
	RejectedP95      int64          `json:"rejected_p95_upper_ms"`
	P95              int64          `json:"p95_upper_ms"`
}

var bounds = [...]int64{10, 25, 50, 100, 300, 1000, 3000, 10000}

func (m *measurement) record(code int, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Attempted++
	m.Statuses[strconv.Itoa(code)]++
	i := 0
	for i < len(bounds) && d.Milliseconds() > bounds[i] {
		i++
	}
	m.Latency[i]++
	if code >= 200 && code < 300 {
		m.AcceptedLatency[i]++
	} else {
		m.RejectedLatency[i]++
	}
}

type harness struct {
	base    string
	clients []*http.Client
	actors  []actor
	grid    geography.Grid
	rng     *rand2.Rand
}

func (h *harness) call(index int, method, path string, data any) (int, []byte) {
	var body io.Reader
	if data != nil {
		b, _ := json.Marshal(data)
		body = bytes.NewReader(b)
	}
	r, _ := http.NewRequest(method, h.base+path, body)
	if index >= 0 {
		r.Header.Set("Authorization", "Bearer "+h.actors[index].token)
	}
	if data != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	client := h.clients[max(0, index)%len(h.clients)]
	response, e := client.Do(r)
	if e != nil {
		return 0, nil
	}
	defer response.Body.Close()
	b, e := io.ReadAll(io.LimitReader(response.Body, 1000001))
	if e != nil {
		return 0, nil
	}
	return response.StatusCode, b
}
func (h *harness) phase(name string, count, rate int, action func(int) int) *measurement {
	m := &measurement{Name: name, Started: time.Now().UnixMilli(), Statuses: map[string]int{}}
	start := time.Now()
	jobs := make(chan int, 256)
	var wg sync.WaitGroup
	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				at := time.Now()
				code := action(n)
				m.record(code, time.Since(at))
			}
		}()
	}
	tick := time.NewTicker(time.Millisecond)
	sent := 0
	for sent < count {
		<-tick.C
		due := min(count, int(time.Since(start).Seconds()*float64(rate)))
		for sent < due {
			select {
			case jobs <- sent:
			default:
				m.GeneratorDropped++
			}
			sent++
		}
	}
	tick.Stop()
	close(jobs)
	wg.Wait()
	m.Seconds = time.Since(start).Seconds()
	sum := 0
	for i, n := range m.Latency {
		sum += n
		if sum*100 >= m.Attempted*95 {
			if i < len(bounds) {
				m.P95 = bounds[i]
			} else {
				m.P95 = -1
			}
			break
		}
	}
	m.AcceptedP95 = percentile(m.AcceptedLatency)
	m.RejectedP95 = percentile(m.RejectedLatency)
	fmt.Printf("%s: attempted=%d accepted_2xx=%d rejected_429=%d errors=%d generator_dropped=%d p95_upper_ms=%d\n", name, m.Attempted, m.Statuses["200"]+m.Statuses["204"], m.Statuses["429"], m.Statuses["0"]+m.Statuses["500"]+m.Statuses["503"], m.GeneratorDropped, m.P95)
	return m
}
func percentile(bins [9]int) int64 {
	total := 0
	for _, n := range bins {
		total += n
	}
	if total == 0 {
		return 0
	}
	sum := 0
	for i, n := range bins {
		sum += n
		if sum*100 >= total*95 {
			if i < len(bounds) {
				return bounds[i]
			}
			return -1
		}
	}
	return -1
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	lab := flag.String("lab-file", "", "private ownership file created by scripts/lab.py")
	out := flag.String("output", "", "synthetic report path")
	sizes := flag.String("sizes", "1000,10000,100000", "ascending attempted population stages")
	distribution := flag.String("distribution", "uniform", "uniform or hotspot")
	seconds := flag.Int("seconds", 20, "seconds per measured traffic phase")
	rate := flag.Int("create-rate", 300, "attempted creations per second")
	burst := flag.Bool("burst", false, "include 3000 attempted writes/s for 60 seconds")
	flag.Parse()
	if *lab == "" || *out == "" || *seconds < 1 || *seconds > 600 || *rate < 1 || *rate > 1000 || (*distribution != "uniform" && *distribution != "hotspot") {
		return errors.New("invalid owned-lab load arguments")
	}
	raw, e := os.ReadFile(*lab)
	if e != nil {
		return errors.New("owned lab required")
	}
	var o owner
	if json.Unmarshal(raw, &o) != nil || o.PID <= 0 {
		return errors.New("invalid lab owner")
	}
	u, e := url.Parse(o.Base)
	if e != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.RawQuery != "" || u.Path != "" || u.User != nil {
		return errors.New("load target must be an owned literal loopback API")
	}
	exe, e := os.Readlink(fmt.Sprintf("/proc/%d/exe", o.PID))
	if e != nil || filepath.Base(exe) != "gati-lab" {
		return errors.New("refusing unowned/non-lab process")
	}
	stages := []int{}
	for _, value := range strings.Split(*sizes, ",") {
		n, e := strconv.Atoi(value)
		if e != nil || n < 1 || n > 100000 || (len(stages) > 0 && n <= stages[len(stages)-1]) {
			return errors.New("invalid population stages")
		}
		stages = append(stages, n)
	}
	h := &harness{base: o.Base, rng: rand2.New(rand2.NewPCG(42, 17))}
	for i := 0; i < 512; i++ {
		ip := net.ParseIP(fmt.Sprintf("127.20.%d.%d", i/250, i%250+1))
		dial := &net.Dialer{LocalAddr: &net.TCPAddr{IP: ip}, Timeout: 3 * time.Second}
		transport := &http.Transport{Proxy: nil, DialContext: dial.DialContext, MaxIdleConnsPerHost: 4, MaxConnsPerHost: 4, IdleConnTimeout: 30 * time.Second}
		h.clients = append(h.clients, &http.Client{Transport: transport, Timeout: 5 * time.Second})
	}
	defer func() {
		for _, c := range h.clients {
			c.CloseIdleConnections()
		}
	}()
	code, b := h.call(-1, "GET", "/api/config", nil)
	if code != 200 {
		return errors.New("lab config unavailable")
	}
	var cfg struct {
		Config config.Config `json:"config"`
	}
	if json.Unmarshal(b, &cfg) != nil || cfg.Config.Validate(false) != nil {
		return errors.New("invalid production configuration")
	}
	h.grid, _ = geography.NewGrid(cfg.Config.Geography.CellSizeMeters)
	hotspot, _ := h.grid.CellAt(geography.Point{19.818, 41.327})
	var phases []*measurement
	accepted := 0
	stageResults := []map[string]int{}
	for _, size := range stages {
		old := len(h.actors)
		for len(h.actors) < size {
			cell := geography.Cell{X: h.rng.IntN(h.grid.Columns), Y: h.rng.IntN(h.grid.Rows)}
			if *distribution == "hotspot" {
				cell = hotspot
			}
			token := make([]byte, 32)
			if _, e := rand.Read(token); e != nil {
				return e
			}
			h.actors = append(h.actors, actor{token: base64.RawURLEncoding.EncodeToString(token), cell: h.grid.ID(cell)})
		}
		phases = append(phases, h.phase(fmt.Sprintf("create-%d", size), size-old, *rate, func(n int) int {
			i := old + n
			code, _ := h.call(i, "POST", "/api/signals", map[string]any{"cell": h.actors[i].cell, "radius_km": 3, "availability_minutes": 30})
			if code == 200 {
				h.actors[i].accepted = true
			}
			return code
		}))
		accepted = 0
		for _, a := range h.actors {
			if a.accepted {
				accepted++
			}
		}
		stageResults = append(stageResults, map[string]int{"attempted_population": size, "accepted_responses": accepted})
		statusRate := max(10, int(math.Ceil(float64(accepted)/30)))
		phases = append(phases, h.phase(fmt.Sprintf("status-%d", size), statusRate**seconds, statusRate, func(n int) int { code, _ := h.call(n%len(h.actors), "GET", "/api/signal", nil); return code }))
		phases = append(phases, h.phase(fmt.Sprintf("public-origin-%d", size), 1000**seconds, 1000, func(n int) int { code, _ := h.call(-1, "GET", "/api/activity/latest", nil); return code }))
		if accepted < size/2 {
			break
		}
	}
	if *burst {
		phases = append(phases, h.phase("write-burst", 180000, 3000, func(n int) int {
			i := n % len(h.actors)
			code, _ := h.call(i, "POST", "/api/signals", map[string]any{"cell": h.actors[i].cell, "radius_km": 3, "availability_minutes": 30})
			return code
		}))
		time.Sleep(3 * time.Second)
		phases = append(phases, h.phase("post-burst-recovery", 1000, 100, func(n int) int { code, _ := h.call(n%len(h.actors), "GET", "/api/signal", nil); return code }))
	}
	phases = append(phases, h.phase("cancellation", min(1000, len(h.actors)), 200, func(n int) int { code, _ := h.call(n, "DELETE", "/api/signal", nil); return code }))
	report := map[string]any{"synthetic": true, "real_time": true, "distribution": *distribution, "seed": 42, "stages": stageResults, "phases": phases, "latency_upper_bounds_ms": bounds, "same_host_generator": true, "limitations": []string{"Accepted responses are counted; uncertain creations can exist until teardown.", "Production limits are unchanged; 512 loopback source addresses model independent network clients.", "Public reads hit the origin, not a CDN or cache proxy.", "No actual device location or external push provider is used.", "Owned lab teardown removes all synthetic state; this is not an expiry test."}}
	data, _ := json.MarshalIndent(report, "", "  ")
	return os.WriteFile(*out, append(data, '\n'), 0600)
}
