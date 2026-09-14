package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRequiresSuccessfulOriginResponse(t *testing.T) {
	for _, status := range []int{200, 302, 503} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if status == 302 {
				w.Header().Set("Location", "http://127.0.0.1:1/")
			}
			w.WriteHeader(status)
		}))
		if healthy(server.URL) != (status == 200) {
			t.Errorf("status %d accepted incorrectly", status)
		}
		server.Close()
		if healthy(server.URL) {
			t.Fatal("closed origin accepted")
		}
	}
}
