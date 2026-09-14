package httpapi

import (
	"encoding/json"
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
			Revision  int64  `json:"revision"`
		}{e == nil, b.Binding, b.ExpiresAt, b.Revision})
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
	var request notification.Registration
	err := readStrictObject(reader, &request, "revision", "binding", "endpoint", "p256dh", "auth", "expires_at")
	return request, err
}
