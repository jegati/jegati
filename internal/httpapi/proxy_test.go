package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedProxyRejectsSpoofingAndCanonicalizesClients(t *testing.T) {
	for _, tc := range []struct {
		peer    string
		headers []string
		want    int
		address string
	}{
		{"10.22.0.2:456", []string{"198.51.100.7"}, 204, "198.51.100.7:0"},
		{"10.22.0.2:456", []string{"2001:db8::7"}, 204, "[2001:db8::7]:0"},
		{"10.22.0.2:456", []string{"::ffff:198.51.100.8"}, 204, "198.51.100.8:0"},
		{"10.22.0.3:456", []string{"198.51.100.7"}, 403, ""},
		{"10.22.0.2:456", nil, 403, ""},
		{"10.22.0.2:456", []string{"198.51.100.7", "198.51.100.8"}, 403, ""},
		{"10.22.0.2:456", []string{"198.51.100.7, 198.51.100.8"}, 403, ""},
		{"10.22.0.2:456", []string{"fe80::1%eth0"}, 403, ""},
		{"invalid", []string{"198.51.100.7"}, 403, ""},
	} {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			if r.RemoteAddr != tc.address || r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Gati-Client-IP") != "" {
				t.Fatal("untrusted metadata propagated")
			}
			w.WriteHeader(204)
		})
		h, err := TrustedProxy(next, "10.22.0.2/32")
		if err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest("GET", "http://gati.test/api/signal", nil)
		r.RemoteAddr = tc.peer
		r.Header.Set("X-Forwarded-For", "attacker-controlled")
		for _, v := range tc.headers {
			r.Header.Add("X-Gati-Client-IP", v)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want || called != (tc.want == 204) {
			t.Fatal(tc, w.Code)
		}
	}
	for _, value := range []string{"0.0.0.0/0", "::/0", "10.0.0.1/24", "invalid"} {
		if _, err := TrustedProxy(http.NotFoundHandler(), value); err == nil {
			t.Fatal("invalid trust config accepted")
		}
	}
}

func TestProxyHealthExceptionDoesNotOpenParticipantRoutes(t *testing.T) {
	h, err := TrustedProxy(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }), "10.22.0.2/32")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/healthz", "/healthz?x=1", "/api/signal"} {
		r := httptest.NewRequest("GET", path, nil)
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := 403
		if path == "/healthz" {
			want = 204
		}
		if w.Code != want {
			t.Fatal(path, w.Code)
		}
	}
}
