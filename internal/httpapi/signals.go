package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/jegati/jegati/internal/worker"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/store"
)

type signalAPI struct {
	engine *worker.Engine
	config config.Config
	grid   geography.Grid
	store  *store.Store
}
type createRequest struct {
	Cell                string `json:"cell"`
	RadiusKM            int    `json:"radius_km"`
	AvailabilityMinutes int    `json:"availability_minutes"`
}

func capability(r *http.Request) (string, error) {
	header := r.Header.Values("Authorization")
	if len(header) != 1 || !strings.HasPrefix(header[0], "Bearer ") {
		return "", errors.New("invalid capability")
	}
	token := strings.TrimPrefix(header[0], "Bearer ")
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != token {
		return "", errors.New("invalid capability")
	}
	return store.Hash(raw), nil
}
func (s signalAPI) authorized(w http.ResponseWriter, r *http.Request) (string, bool) {
	if r.URL.RawQuery != "" {
		writeError(w, 400)
		return "", false
	}
	hash, err := capability(r)
	if err != nil {
		writeError(w, 401)
		return "", false
	}
	if s.store == nil {
		writeError(w, 503)
		return "", false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, e := url.Parse(origin)
		if e != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host != r.Host {
			writeError(w, 403)
			return "", false
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		writeError(w, 400)
		return "", false
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		writeError(w, 400)
		return "", false
	}
	ip = ip.Unmap()
	// Ignore X-Forwarded-For from clients. Trusted edge forwarding is later scope.
	network := ip.String()
	if ip.Is6() {
		network = netip.PrefixFrom(ip, 56).Masked().String()
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	key, err := s.store.NetworkKey(ctx, network)
	if err != nil {
		writeError(w, 503)
		return "", false
	}
	ok, err := s.store.Allow(ctx, "requests:"+key, s.config.Limits.RequestsPerNetworkWindow, time.Duration(s.config.Limits.NetworkWindowSeconds)*time.Second)
	if err != nil {
		writeError(w, 503)
		return "", false
	}
	if !ok {
		w.Header().Set("Retry-After", "60")
		writeError(w, 429)
		return "", false
	}
	if r.Method == "POST" || r.Method == "DELETE" {
		ok, err = s.store.Allow(ctx, "writes", s.config.Limits.GlobalWritesPerSecond, time.Second)
		if err != nil {
			writeError(w, 503)
			return "", false
		}
		if !ok {
			w.Header().Set("Retry-After", "1")
			writeError(w, 429)
			return "", false
		}
	}
	if r.Method == "POST" && (r.URL.Path == "/api/signals" || r.URL.Path == "/api/join") {
		ok, err = s.store.Allow(ctx, "creates:"+key, s.config.Limits.NewSignalsPerNetworkWindow, time.Duration(s.config.Limits.NetworkWindowSeconds)*time.Second)
		if err != nil {
			writeError(w, 503)
			return "", false
		}
		if !ok {
			w.Header().Set("Retry-After", "60")
			writeError(w, 429)
			return "", false
		}
	}
	return hash, true
}
func (s signalAPI) create(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	typ, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes))
	request, err := readCreate(r.Body)
	if err != nil {
		writeError(w, 400)
		return
	}
	if _, err := s.grid.Parse(request.Cell); err != nil || !slices.Contains(s.config.Geography.TravelRadiusChoicesKm, request.RadiusKM) || !slices.Contains(s.config.Availability.ChoicesMinutes, request.AvailabilityMinutes) {
		writeError(w, 400)
		return
	}
	signal, err := s.store.Create(r.Context(), hash, request.Cell, request.RadiusKM, request.AvailabilityMinutes, time.Duration(request.AvailabilityMinutes)*time.Minute, s.config.Limits.MaxActiveSignals)
	if err != nil {
		s.fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	s.respond(w, r, hash, signal)
}
func readCreate(body io.Reader) (createRequest, error) {
	var request createRequest
	dec := json.NewDecoder(body)
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return request, errors.New("object required")
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return request, err
		}
		name, ok := key.(string)
		if !ok {
			return request, errors.New("invalid field")
		}
		if _, duplicate := fields[name]; duplicate {
			return request, errors.New("duplicate field")
		}
		if name != "cell" && name != "radius_km" && name != "availability_minutes" {
			return request, errors.New("unknown field")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return request, err
		}
		if string(raw) == "null" {
			return request, errors.New("null forbidden")
		}
		fields[name] = raw
	}
	if _, err := dec.Token(); err != nil {
		return request, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return request, errors.New("trailing input")
	}
	if len(fields) != 3 {
		return request, errors.New("missing fields")
	}
	data, _ := json.Marshal(fields)
	if err := json.Unmarshal(data, &request); err != nil {
		return request, err
	}
	return request, nil
}
func (s signalAPI) status(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	value, err := s.store.Status(r.Context(), hash)
	if err != nil {
		s.fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	s.respond(w, r, hash, value)
}
func (s signalAPI) cancel(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if err := s.store.Cancel(r.Context(), hash); err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s signalAPI) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrGone):
		writeError(w, 410)
	case errors.Is(err, store.ErrConflict):
		writeError(w, 409)
	case errors.Is(err, store.ErrCapacity):
		writeError(w, 503)
	default:
		writeError(w, 503)
	}
}
