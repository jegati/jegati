package simulation

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
)

type Report struct {
	Simulation  bool           `json:"simulation"`
	Scenario    Scenario       `json:"scenario"`
	ConfigHash  string         `json:"config_sha256"`
	Grid        geography.Grid `json:"grid"`
	Credentials int            `json:"generated_credentials"`
	Accepted    int            `json:"accepted_credentials"`
	Rejected    int            `json:"rejected_credentials"`
	Cancelled   int            `json:"cancelled"`
	Duplicates  int            `json:"verified_idempotent_retries"`
	Expired     int            `json:"verified_expired"`
	Cells       map[string]int `json:"synthetic_accepted_by_cell"`
}
type Client struct {
	base, control string
	http          *http.Client
}

func NewClient(target, control string) (*Client, error) {
	u, e := url.Parse(target)
	if e != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("simulation requires a plain loopback HTTP origin")
	}
	ip, e := netip.ParseAddr(u.Hostname())
	if e != nil || !ip.IsLoopback() || u.Port() == "" || len(control) != 64 {
		return nil, errors.New("simulation requires literal loopback and dedicated control credential")
	}
	u.Path = ""
	return &Client{u.String(), control, &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("simulation refuses redirects") }}}, nil
}
func (c *Client) do(ctx context.Context, method, path, token string, body any, out any) (int, error) {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	r, e := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(payload))
	if e != nil {
		return 0, e
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	response, e := c.http.Do(r)
	if e != nil {
		return 0, errors.New("simulation HTTP request failed")
	}
	defer response.Body.Close()
	if out != nil && response.StatusCode == 200 {
		if e = json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(out); e != nil {
			return response.StatusCode, errors.New("invalid bounded simulation response")
		}
	}
	return response.StatusCode, nil
}
func (c *Client) Run(ctx context.Context, s Scenario) (Report, error) {
	report := Report{Simulation: true, Scenario: s, Cells: map[string]int{}}
	if e := s.Validate(); e != nil {
		return report, e
	}
	var envelope struct {
		Config config.Config `json:"config"`
		Hash   string        `json:"sha256"`
	}
	code, e := c.do(ctx, "GET", "/api/config", "", nil, &envelope)
	if e != nil || code != 200 || envelope.Config.Profile != "simulation" {
		return report, errors.New("refusing target: simulation profile is required")
	}
	report.ConfigHash = envelope.Hash
	code, e = c.do(ctx, "GET", "/api/simulation/clock", "", nil, nil)
	if e != nil || code != 403 {
		return report, errors.New("simulation control missing authentication enforcement")
	}
	var clock struct {
		Simulation bool  `json:"simulation"`
		Now        int64 `json:"now"`
	}
	code, e = c.do(ctx, "GET", "/api/simulation/clock", c.control, nil, &clock)
	if e != nil || code != 200 || !clock.Simulation {
		return report, errors.New("refusing target: authenticated simulation clock required")
	}
	code, e = c.do(ctx, "GET", "/api/geography", "", nil, &report.Grid)
	if e != nil || code != 200 {
		return report, errors.New("missing simulation geography")
	}
	type accepted struct {
		token     string
		expiry    int64
		cancelled bool
	}
	live := []accepted{}
	for _, input := range Generate(s, report.Grid) {
		report.Credentials++
		raw := make([]byte, 32)
		if _, e = rand.Read(raw); e != nil {
			return report, e
		}
		token := base64.RawURLEncoding.EncodeToString(raw)
		var result struct {
			Expires int64 `json:"expires_at"`
		}
		code, e = c.do(ctx, "POST", "/api/signals", token, input, &result)
		if e != nil {
			return report, e
		}
		if code != 200 {
			if code != 429 && code != 503 {
				return report, errors.New("unexpected simulation admission response")
			}
			report.Rejected++
			continue
		}
		report.Accepted++
		report.Cells[input.Cell]++
		live = append(live, accepted{token: token, expiry: result.Expires})
		if input.Duplicate {
			var second struct {
				Expires int64 `json:"expires_at"`
			}
			code, e = c.do(ctx, "POST", "/api/signals", token, input, &second)
			if e != nil {
				return report, e
			}
			if code == 200 {
				if second.Expires != result.Expires {
					return report, errors.New("retry extended availability")
				}
				report.Duplicates++
			} else if code != 429 {
				return report, errors.New("unexpected retry response")
			}
		}
		if input.Cancel {
			code, e = c.do(ctx, "DELETE", "/api/signal", token, nil, nil)
			if e != nil {
				return report, e
			}
			if code == 204 {
				report.Cancelled++
				live[len(live)-1].cancelled = true
			} else if code != 429 {
				return report, errors.New("unexpected cancel response")
			}
		}
	}
	code, e = c.do(ctx, "POST", "/api/simulation/clock", c.control, map[string]int{"milliseconds": 121 * 60 * 1000}, &clock)
	if e != nil || code != 200 {
		return report, errors.New("cannot advance functional clock")
	}
	for _, record := range live {
		if record.expiry > clock.Now {
			return report, errors.New("advance did not pass signal deadline")
		}
		code, e = c.do(ctx, "GET", "/api/signal", record.token, nil, nil)
		if e != nil || code != 410 {
			return report, errors.New("expired signal remains available or verification was rate limited")
		}
		if !record.cancelled {
			report.Expired++
		}
	}
	return report, nil
}
