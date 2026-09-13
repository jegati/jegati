package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/notification"
	"github.com/jegati/jegati/internal/store"
	"github.com/jegati/jegati/internal/worker"
)

func TestPushBodyIsStrict(t *testing.T) {
	valid := `{"revision":0,"binding":"b","endpoint":"e","p256dh":"p","auth":"a","expires_at":123}`
	if _, e := readPush(strings.NewReader(valid)); e != nil {
		t.Fatal(e)
	}
	for _, body := range []string{valid + `{}`, strings.Replace(valid, `"revision":0,`, ``, 1), strings.Replace(valid, `"auth":"a"`, `"auth":"a","auth":"b"`, 1), strings.Replace(valid, `"auth":"a"`, `"latitude":41.3`, 1), strings.Replace(valid, `"auth":"a"`, `"auth":null`, 1), strings.Replace(valid, `"expires_at":123`, `"expires_at":"123"`, 1), `[]`} {
		if _, e := readPush(strings.NewReader(body)); e == nil {
			t.Fatal("non-strict push body accepted")
		}
	}
}
func TestRealPushAPIOptInAndOptOut(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("real temporary store: make test-store")
	}
	ctx := context.Background()
	secret, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential unavailable")
	}
	backend, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(secret)))
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Client.Close()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	c.Notifications.PushEnabled = true
	c.Notifications.PushContact = "https://project.example.org/security"
	keys, e := notification.GenerateKeys()
	if e != nil {
		t.Fatal(e)
	}
	rawKeys, _ := json.Marshal(keys)
	keyPath := filepath.Join(t.TempDir(), "keys.json")
	if e = os.WriteFile(keyPath, rawKeys, 0600); e != nil {
		t.Fatal(e)
	}
	service, e := notification.New(c, backend, keyPath)
	if e != nil {
		t.Fatal(e)
	}
	engine := worker.New(backend, nil, c)
	engine.Push = service
	handler := Handler(c, backend, nil, engine)
	raw := make([]byte, 32)
	rand.Read(raw)
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := store.Hash(raw)
	signal, e := backend.Create(ctx, hash, "tirana-v1:100:55:55", 3, 30, time.Minute, 1000)
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Cancel(ctx, hash)
	call := func(method, path, auth, body, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.6:43210"
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", auth)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	w := call("GET", "/api/push-config", "", "", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), keys.Private) || !strings.Contains(w.Body.String(), keys.Public) {
		t.Fatal("push configuration leaked or omitted key")
	}
	b := notification.Registration{Binding: hex.EncodeToString(raw), Endpoint: "https://fcm.googleapis.com/synthetic-unused", P256DH: keys.Public, Auth: base64.RawURLEncoding.EncodeToString(raw[:16]), ExpiresAt: signal.ExpiresAt}
	body, _ := json.Marshal(b)
	for _, tc := range []struct {
		path, auth, origin string
		code               int
	}{{"/api/push", "", "", 401}, {"/api/push?location=x", "Bearer " + token, "", 400}, {"/api/push", "Bearer " + token, "https://other.example.org", 403}} {
		if w := call("POST", tc.path, tc.auth, string(body), tc.origin); w.Code != tc.code {
			t.Fatal("push auth/query/origin boundary failed", w.Code)
		}
	}
	w = call("POST", "/api/push", "Bearer "+token, string(body), "")
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("opt-in failed", w.Code)
	}
	for _, secret := range []string{b.Endpoint, b.Auth, b.P256DH, token} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatal("private subscription echoed")
		}
	}
	current, e := backend.Status(ctx, hash)
	if e != nil || current.State != "gati" || current.Gathering != "" || current.ExpiresAt != signal.ExpiresAt {
		t.Fatal("opt-in changed willingness")
	}
	w = call("GET", "/api/push", "Bearer "+token, "", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":true`) || strings.Contains(w.Body.String(), b.Endpoint) {
		t.Fatal("private status contract failed")
	}
	// Removal stays available even if the operator disables transport after opt-in.
	engine.Push = nil
	w = call("DELETE", "/api/push", "Bearer "+token, "", "")
	if w.Code != 204 {
		t.Fatal("opt-out depended on active transport")
	}
	if _, e = backend.PushStatus(ctx, hash); e != store.ErrGone {
		t.Fatal("opt-out retained delivery binding")
	}
	engine.Push = service
	if stale := call("POST", "/api/push", "Bearer "+token, string(body), ""); stale.Code != 409 {
		t.Fatal("late pre-opt-out registration was accepted", stale.Code)
	}
	status := call("GET", "/api/push", "Bearer "+token, "", "")
	if !strings.Contains(status.Body.String(), `"revision":1`) || !strings.Contains(status.Body.String(), `"enabled":false`) {
		t.Fatal("opt-out fence missing from private metadata")
	}
	current, e = backend.Status(ctx, hash)
	if e != nil || current.State != "gati" {
		t.Fatal("opt-out cancelled willingness")
	}
}
