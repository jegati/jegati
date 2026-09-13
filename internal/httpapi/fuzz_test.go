package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func FuzzRequestParsers(f *testing.F) {
	for _, s := range []string{`{}`, `{"cell":"tirana-v1:100:55:55","radius_km":0.1,"availability_minutes":30}`, `{"revision":0,"binding":"x","endpoint":"https://example.test","p256dh":"p","auth":"a","expires_at":123}`, `null`, `{"cell":null,"cell":"a"}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2048 {
			return
		}
		value, e := readCreate(bytes.NewReader(data))
		if e == nil {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			again, err := readCreate(bytes.NewReader(encoded))
			if err != nil || again != value {
				t.Fatal("create roundtrip")
			}
		}
		_, _ = readPush(bytes.NewReader(data))
		_, _, _ = readJoin(bytes.NewReader(data))
	})
}
func FuzzCapability(f *testing.F) {
	f.Add("Bearer " + base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	f.Add("Bearer bad")
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 2048 {
			return
		}
		r := httptest.NewRequest("GET", "/api/signal", nil)
		r.Header.Set("Authorization", input)
		hash, e := capability(r)
		if e == nil {
			if len(hash) != 64 {
				t.Fatal("unbounded capability hash")
			}
			r.Header.Add("Authorization", input)
			if _, e := capability(r); e == nil {
				t.Fatal("duplicate header accepted")
			}
		}
	})
}
