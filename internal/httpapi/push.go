package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/jegati/jegati/internal/notification"
	"github.com/jegati/jegati/internal/store"
)

func (s signalAPI) pushConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeError(w, 400)
		return
	}
	key := ""
	if s.engine != nil && s.engine.Push != nil {
		key = s.engine.Push.PublicKey()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(struct {
		Enabled   bool   `json:"enabled"`
		PublicKey string `json:"public_key"`
	}{key != "", key})
}
func (s signalAPI) push(w http.ResponseWriter, r *http.Request) {
	hash, ok := s.authorized(w, r)
	if !ok {
		return
	}
	if r.Method == "DELETE" {
		if e := s.store.DropPush(r.Context(), hash); e != nil {
			s.fail(w, e)
			return
		}
		w.WriteHeader(204)
		return
	}
	if r.Method == "GET" {
		b, e := s.store.PushStatus(r.Context(), hash)
		if e != nil && e != store.ErrGone {
			s.fail(w, e)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(struct {
			Enabled   bool   `json:"enabled"`
			Binding   string `json:"binding,omitempty"`
			ExpiresAt int64  `json:"expires_at,omitempty"`
		}{e == nil, b.Binding, b.ExpiresAt})
		return
	}
	if s.engine == nil || s.engine.Push == nil {
		writeError(w, 503)
		return
	}
	typ, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || typ != "application/json" {
		writeError(w, 415)
		return
	}
	input, e := readPush(http.MaxBytesReader(w, r.Body, int64(s.config.Limits.MaxBodyBytes)))
	if e != nil {
		writeError(w, 400)
		return
	}
	// Validation errors are generic and never echo an endpoint or encryption key.
	if e = s.engine.Push.Validate(input); e != nil {
		writeError(w, 400)
		return
	}
	expiry, e := s.engine.Push.Register(r.Context(), hash, input)
	if e != nil {
		s.fail(w, e)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(struct {
		ExpiresAt int64 `json:"expires_at"`
	}{expiry})
}
func readPush(reader io.Reader) (notification.Registration, error) {
	invalid := errors.New("invalid subscription body")
	dec := json.NewDecoder(reader)
	opening, e := dec.Token()
	if e != nil || opening != json.Delim('{') {
		return notification.Registration{}, invalid
	}
	fields := map[string]json.RawMessage{}
	allowed := map[string]bool{"binding": true, "endpoint": true, "p256dh": true, "auth": true, "expires_at": true}
	for dec.More() {
		key, e := dec.Token()
		if e != nil {
			return notification.Registration{}, invalid
		}
		name, ok := key.(string)
		if !ok || !allowed[name] || fields[name] != nil {
			return notification.Registration{}, invalid
		}
		var value json.RawMessage
		if dec.Decode(&value) != nil || string(value) == "null" {
			return notification.Registration{}, invalid
		}
		fields[name] = value
	}
	closing, e := dec.Token()
	if e != nil || closing != json.Delim('}') || dec.Decode(new(any)) != io.EOF || len(fields) != len(allowed) {
		return notification.Registration{}, invalid
	}
	raw, _ := json.Marshal(fields)
	var value notification.Registration
	if json.Unmarshal(raw, &value) != nil {
		return value, invalid
	}
	return value, nil
}
