package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

func TestStrictSignalBody(t *testing.T) {
	good := `{"cell":"tirana-v1:1000:4:4","radius_km":3,"availability_minutes":30}`
	if _, e := readCreate(strings.NewReader(good)); e != nil {
		t.Fatal(e)
	}
	for _, body := range []string{good + "{}", `null`, `[]`, `{}`, strings.Replace(good, `"cell":`, `"lat":41.32,"cell":`, 1), strings.Replace(good, `"radius_km":3`, `"radius_km":3,"radius_km":5`, 1), strings.Replace(good, `"radius_km":3`, `"radius_km":null`, 1), strings.Replace(good, `"radius_km":3`, `"radius_km":"3"`, 1)} {
		if _, e := readCreate(strings.NewReader(body)); e == nil {
			t.Fatal("accepted malformed/private-location body")
		}
	}
}
func TestRealWillingnessAPI(t *testing.T) {
	if os.Getenv("GATI_INTEGRATION") != "1" {
		t.Skip("real store API test: use make test-store")
	}
	b, e := os.ReadFile(os.Getenv("GATI_TEST_PASSWORD_FILE"))
	if e != nil {
		t.Fatal("test credential unavailable")
	}
	backend, e := store.Connect(context.Background(), os.Getenv("GATI_TEST_ADDR"), "app", strings.TrimSpace(string(b)))
	if e != nil {
		t.Fatal(e)
	}
	defer backend.Client.Close()
	c, e := config.Load("../../config/gati.yaml", false)
	if e != nil {
		t.Fatal(e)
	}
	handler := Handler(c, backend, nil)
	token := make([]byte, 32)
	rand.Read(token)
	auth := "Bearer " + base64.RawURLEncoding.EncodeToString(token)
	request := func(method, path, body, origin, header string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://gati.test"+path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", header)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	body := `{"cell":"tirana-v1:1000:4:4","radius_km":3,"availability_minutes":30}`
	for _, header := range []string{"", "Bearer tiny", auth + "="} {
		if w := request("POST", "/api/signals", body, "", header); w.Code != 401 {
			t.Fatal("invalid capability accepted")
		}
	}
	if w := request("POST", "/api/signals", body, "https://other.test", auth); w.Code != 403 {
		t.Fatal("cross-origin request accepted")
	}
	for _, invalid := range []string{strings.Replace(body, `"cell":`, `"latitude":41.32,"cell":`, 1), strings.Replace(body, "30", "15", 1), strings.Replace(body, ":4:4", ":999:999", 1), strings.Replace(body, "radius_km\":3", "radius_km\":2", 1), strings.Repeat("x", 4096)} {
		if w := request("POST", "/api/signals", invalid, "", auth); w.Code != 400 {
			t.Fatalf("invalid body status: %d", w.Code)
		}
	}
	created := request("POST", "/api/signals", body, "", auth)
	if created.Code != 200 {
		t.Fatalf("create failed: %d", created.Code)
	}
	var first store.Signal
	if e = json.Unmarshal(created.Body.Bytes(), &first); e != nil {
		t.Fatal(e)
	}
	if first.ExpiresAt-first.CreatedAt != int64(30*time.Minute/time.Millisecond) {
		t.Fatal("incorrect server deadline")
	}
	if strings.Contains(created.Body.String(), store.Hash(token)) {
		t.Fatal("hash leaked in response")
	}
	again := request("POST", "/api/signals", body, "", auth)
	var second store.Signal
	json.Unmarshal(again.Body.Bytes(), &second)
	if first.ExpiresAt != second.ExpiresAt {
		t.Fatal("retry renewed willingness")
	}
	if w := request("GET", "/api/signal", "", "", auth); w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Set-Cookie") != "" {
		t.Fatal("private status contract failed")
	}
	if w := request("DELETE", "/api/signal", "", "", auth); w.Code != 204 {
		t.Fatal("cancellation failed")
	}
	if w := request("GET", "/api/signal", "", "", auth); w.Code != 410 {
		t.Fatal("cancelled signal readable")
	}
	if w := request("POST", "/api/signals", body, "", auth); w.Code != 410 {
		t.Fatal("cancelled capability replay")
	}
	if w := request("GET", "/api/signals", "", "", auth); w.Code != http.StatusNotFound {
		t.Fatal("public participant listing exists")
	}
}
