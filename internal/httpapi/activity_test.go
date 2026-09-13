package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/activity"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

func TestRealPublicActivityCacheAndNoQueryContract(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("real Valkey: make test-store")
	}
	secret, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential unavailable")
	}
	ctx := context.Background()
	s, e := store.Connect(ctx, os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(secret)))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Client.Close()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	c.Limits.MaxActiveSignals = 123455
	_, hash := c.Canonical()
	now, e := s.Now(ctx)
	if e != nil {
		t.Fatal(e)
	}
	epoch := now/(int64(c.PublicActivity.ReleaseSeconds)*1000) - int64(c.PublicActivity.DelayEpochs) - 1
	key := fmt.Sprintf("gati:activity:%s:%d", hash, epoch)
	s.Client.Del(ctx, key)
	defer s.Client.Del(ctx, key)
	handler := Handler(c, s, nil)
	read := func(path, auth, etag string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		req.Header.Set("If-None-Match", etag)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	if w := read("/api/activity/latest", "", ""); w.Code != 204 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("absence invented data")
	}
	r := activity.Release{Version: 1, ID: "synthetic-release", ConfigHash: hash, ObservedFrom: now - 600000, ObservedUntil: now - 600000, ReleaseAt: now - 1000, ExpiresAt: now + 60000, Areas: []activity.Area{{Cell: "tirana-v1:1000:4:4", Willing: 50}}, Gatherings: []activity.Event{}}
	data, _ := json.Marshal(r)
	if e = s.Client.Set(ctx, key, data, time.Minute).Err(); e != nil {
		t.Fatal(e)
	}
	w := read("/api/activity/latest", "", "")
	if w.Code != 200 || w.Body.String() != string(data) || w.Header().Get("Cache-Control") != "public, max-age=30, must-revalidate" {
		t.Fatal("public cache contract", w.Code)
	}
	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag")
	}
	if w = read("/api/activity/latest", "", etag); w.Code != 304 || w.Body.Len() != 0 {
		t.Fatal("conditional response")
	}
	if w = read("/api/activity/latest", "Bearer synthetic", ""); w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("authenticated shared cache")
	}
	if w = read("/api/activity/latest?cell=anything", "", ""); w.Code != 400 {
		t.Fatal("custom query allowed")
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := read("/api/activity/latest", "", "")
			if w.Code != 200 || w.Body.String() != string(data) {
				t.Error("concurrent read changed release")
			}
		}()
	}
	wg.Wait()
	r.ExpiresAt = now - 1
	data, _ = json.Marshal(r)
	s.Client.Set(ctx, key, data, time.Minute)
	if w = read("/api/activity/latest", "", etag); w.Code != 204 {
		t.Fatal("logical expiry bypassed by native TTL or conditional request")
	}
}
