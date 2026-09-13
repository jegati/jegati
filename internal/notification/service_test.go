package notification

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"github.com/jegati/jegati/internal/matching"
	"github.com/jegati/jegati/internal/store"
)

type fakeClient struct {
	mu    sync.Mutex
	calls int
	code  int
	body  []byte
	ttl   int
	auth  string
}

func (f *fakeClient) Do(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.body, _ = io.ReadAll(r.Body)
	f.ttl, _ = strconv.Atoi(r.Header.Get("TTL"))
	f.auth = r.Header.Get("Authorization")
	return &http.Response{StatusCode: f.code, Body: io.NopCloser(strings.NewReader("synthetic provider")), Header: http.Header{}}, nil
}
func setup(t *testing.T) (*Service, *fakeClient, Registration) {
	t.Helper()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	c.Notifications.PushEndpointHosts = []string{"push.example.org"}
	keys, e := GenerateKeys()
	if e != nil {
		t.Fatal(e)
	}
	client := &fakeClient{code: 201}
	s, e := newService(c, nil, keys, client)
	if e != nil {
		t.Fatal(e)
	}
	recipient, e := ecdh.P256().GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	auth := make([]byte, 16)
	rand.Read(auth)
	binding := make([]byte, 32)
	rand.Read(binding)
	r := Registration{Binding: hex.EncodeToString(binding), Endpoint: "https://push.example.org/synthetic-only", P256DH: base64.RawURLEncoding.EncodeToString(recipient.PublicKey().Bytes()), Auth: base64.RawURLEncoding.EncodeToString(auth), ExpiresAt: time.Now().Add(time.Minute).UnixMilli()}
	return s, client, r
}
func TestKeysAndRegistrationValidation(t *testing.T) {
	s, _, r := setup(t)
	if e := s.Validate(r); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*Registration){func(r *Registration) { r.Auth = "bad" }, func(r *Registration) { r.P256DH = base64.RawURLEncoding.EncodeToString(make([]byte, 65)) }, func(r *Registration) { r.Endpoint = "http://push.example.org/path" }, func(r *Registration) { r.Binding = "permanent-device-id" }, func(r *Registration) { r.ExpiresAt = 0 }} {
		copy := r
		mutate(&copy)
		if s.Validate(copy) == nil {
			t.Fatal("invalid registration accepted")
		}
	}
	keys := s.keys
	keys.Public = "invalid"
	if _, e := newService(s.config, nil, keys, s.client); e == nil {
		t.Fatal("mismatched VAPID key accepted")
	}
	s.config.Notifications.PushEnabled = true
	s.config.Profile = "simulation"
	if _, e := New(s.config, &store.Store{}, "absent"); e == nil {
		t.Fatal("simulation enabled real provider delivery")
	}
}
func TestEncryptedStoreAndFakeProviderDelivery(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("real temporary store: make test-store")
	}
	s, client, r := setup(t)
	secret, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential unavailable")
	}
	ctx := context.Background()
	backend, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(secret)))
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Client.Close()
	s.Store = backend
	grid, _ := geography.NewGrid(1000)
	cell := geography.Cell{X: 5, Y: 5}
	proposal := matching.Proposal{Intersection: geography.Intersection{ID: r.Binding, Point: grid.Center(cell)}}
	for n := 0; n < 3; n++ {
		raw := make([]byte, 32)
		rand.Read(raw)
		hash := store.Hash(raw)
		signal, e := backend.Create(ctx, hash, grid.ID(cell), 3, 30, 10*time.Minute, 1000)
		if e != nil {
			t.Fatal(e)
		}
		defer backend.Cancel(ctx, hash)
		proposal.Founders = append(proposal.Founders, matching.Signal{Hash: hash, Area: geography.ParticipantArea{Cell: cell, RadiusKM: 3}, CreatedAt: signal.CreatedAt, ExpiresAt: signal.ExpiresAt})
	}
	hash := proposal.Founders[0].Hash
	r.ExpiresAt = proposal.Founders[0].ExpiresAt
	if _, e = s.Register(ctx, hash, r); e != nil {
		t.Fatal(e)
	}
	stored, e := backend.Client.Get(ctx, "gati:push:"+hash).Result()
	if e != nil {
		t.Fatal(e)
	}
	for _, secret := range []string{r.Endpoint, r.Auth, r.P256DH} {
		if strings.Contains(stored, secret) {
			t.Fatal("plaintext transport material in store")
		}
	}
	var binding store.PushBinding
	json.Unmarshal([]byte(stored), &binding)
	job := store.PushJob{PushBinding: binding}
	plain, e := s.decrypt(hash, job)
	if e != nil || plain != r {
		t.Fatal("transport encryption round trip failed")
	}
	if _, e = s.decrypt("wrong-session", job); e == nil {
		t.Fatal("ciphertext movable to another session")
	}
	copy := job
	copy.Binding = strings.Repeat("0", 64)
	if _, e = s.decrypt(hash, copy); e == nil {
		t.Fatal("ciphertext movable to another binding")
	}
	if e = s.Step(ctx); e != nil {
		t.Fatal(e)
	}
	if client.calls != 0 {
		t.Fatal("subscription sent before gathering")
	}
	id := r.Binding[:32]
	if _, e = backend.Reserve(ctx, id, "synthetic", proposal, grid, 20, 1000, 1000); e != nil {
		t.Fatal(e)
	}
	time.Sleep(30 * time.Millisecond)
	if _, e = backend.Activate(ctx, id, "synthetic", 1000, 600000); e != nil {
		t.Fatal(e)
	}
	// Test-only due scheduling; no production test route or real endpoint is used.
	jobClaim, e := backend.ClaimPush(ctx, hash, "first", s.policy())
	if e != nil || jobClaim == nil {
		t.Fatal("activated push not claimable", e)
	}
	outcome, _ := s.send(ctx, hash, *jobClaim, 1000)
	if outcome != "sent" || client.calls != 1 || client.ttl < 1 || client.ttl > 300 || client.auth == "" {
		t.Fatal("Web Push request failed", outcome, client.calls, client.ttl)
	}
	if len(client.body) != 512 || bytes.Contains(client.body, []byte(r.Binding)) || bytes.Contains(client.body, []byte(hash)) || bytes.Contains(client.body, []byte("expires_at")) {
		t.Fatal("payload not encrypted/padded")
	}
	client.code = 410
	outcome, _ = s.send(ctx, hash, *jobClaim, 1000)
	if outcome != "gone" {
		t.Fatal("provider expiry not removed")
	}
	before := client.calls
	if e = backend.Cancel(ctx, hash); e != nil {
		t.Fatal(e)
	}
	outcome, _ = s.send(ctx, hash, *jobClaim, 1000)
	if outcome != "gone" || client.calls != before {
		t.Fatal("cancelled queued item reached provider")
	}
}
