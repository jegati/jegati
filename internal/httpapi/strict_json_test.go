package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type requestParserCase struct {
	name   string
	fields []string
	values []string
	parse  func(io.Reader) (any, error)
}

func strictRequestCases() []requestParserCase {
	return []requestParserCase{
		{"willingness", []string{"cell", "radius_km", "availability_minutes"}, []string{`"tirana-v1:100:55:55"`, `3`, `30`}, func(r io.Reader) (any, error) { return readCreate(r) }},
		{"join-and-preview", []string{"cell", "radius_km", "availability_minutes", "gathering_id"}, []string{`"tirana-v1:100:55:55"`, `3`, `30`, `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`}, func(r io.Reader) (any, error) {
			id, v, err := readJoin(r)
			return struct {
				ID    string
				Value createRequest
			}{id, v}, err
		}},
		{"push", []string{"revision", "binding", "endpoint", "p256dh", "auth", "expires_at"}, []string{`1`, `"synthetic-binding"`, `"https://push.invalid/synthetic"`, `"synthetic-key"`, `"synthetic-auth"`, `1801000`}, func(r io.Reader) (any, error) { return readPush(r) }},
		{"arrival", []string{"cell"}, []string{`"tirana-v1:100:55:55"`}, func(r io.Reader) (any, error) {
			var request struct {
				Cell string `json:"cell"`
			}
			err := readStrictObject(r, &request, "cell")
			return request, err
		}},
		{"going-and-decline", []string{"gathering_id"}, []string{`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`}, func(r io.Reader) (any, error) {
			var request struct {
				GatheringID string `json:"gathering_id"`
			}
			err := readStrictObject(r, &request, "gathering_id")
			return request, err
		}},
	}
}

func TestStrictRequestMatrix(t *testing.T) {
	for _, c := range strictRequestCases() {
		t.Run(c.name, func(t *testing.T) {
			pairs := make([]string, len(c.fields))
			for i, k := range c.fields {
				pairs[i] = fmt.Sprintf("%q:%s", k, c.values[i])
			}
			good := "{" + strings.Join(pairs, ",") + "}"
			want, err := c.parse(strings.NewReader(good))
			if err != nil {
				t.Fatal(err)
			}
			valid := []string{good, " \n" + good + "\t", strings.Replace(good, fmt.Sprintf("%q", c.fields[0]), fmt.Sprintf(`"\u%04x%s"`, c.fields[0][0], c.fields[0][1:]), 1)}
			reversed := append([]string(nil), pairs...)
			for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
				reversed[i], reversed[j] = reversed[j], reversed[i]
			}
			valid = append(valid, "{"+strings.Join(reversed, ",")+"}")
			for _, body := range valid {
				got, err := c.parse(strings.NewReader(body))
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatal("valid field ordering/escaping changed")
				}
			}
			invalid := []string{"", "null", "[]", "{}", good + "{}", good + "null", good + "x", good[:len(good)-1], good[:len(good)-1] + `,"latitude":41.32}`, strings.Replace(good, pairs[0], fmt.Sprintf("%q:%s", strings.ToUpper(c.fields[0]), c.values[0]), 1)}
			for i, k := range c.fields {
				invalid = append(invalid, good[:len(good)-1]+","+pairs[i]+"}")
				invalid = append(invalid, good[:len(good)-1]+fmt.Sprintf(`,"\u%04x%s":%s}`, k[0], k[1:], c.values[i]))
				missing := append([]string(nil), pairs[:i]...)
				missing = append(missing, pairs[i+1:]...)
				invalid = append(invalid, "{"+strings.Join(missing, ",")+"}")
				for _, replacement := range []string{"null", "true", "[]", "{}"} {
					invalid = append(invalid, strings.Replace(good, pairs[i], fmt.Sprintf("%q:%s", k, replacement), 1))
				}
			}
			for i, body := range invalid {
				if _, err := c.parse(strings.NewReader(body)); err == nil {
					t.Fatalf("invalid case %d accepted", i)
				}
			}
			limited := http.MaxBytesReader(httptest.NewRecorder(), io.NopCloser(strings.NewReader(good)), int64(len(good)-1))
			if _, err := c.parse(limited); err == nil {
				t.Fatal("body limit bypassed")
			}
		})
	}
}

func FuzzStrictRequestRoundTrip(f *testing.F) {
	for i, c := range strictRequestCases() {
		fields := map[string]json.RawMessage{}
		for j, k := range c.fields {
			fields[k] = json.RawMessage(c.values[j])
		}
		body, _ := json.Marshal(fields)
		f.Add(uint8(i), string(body))
	}
	f.Fuzz(func(t *testing.T, kind uint8, body string) {
		if len(body) > 4096 {
			t.Skip()
		}
		cases := strictRequestCases()
		c := cases[int(kind)%len(cases)]
		got, err := c.parse(strings.NewReader(body))
		if err != nil {
			return
		}
		// Normalizing field order/whitespace must preserve every accepted value.
		var fields map[string]json.RawMessage
		if json.Unmarshal([]byte(body), &fields) != nil || len(fields) != len(c.fields) {
			t.Fatal("accepted non-object or wrong field set")
		}
		for _, key := range c.fields {
			if _, ok := fields[key]; !ok {
				t.Fatal("accepted missing field")
			}
		}
		normalized, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		again, err := c.parse(strings.NewReader(string(normalized)))
		if err != nil || !reflect.DeepEqual(got, again) {
			t.Fatal("accepted request changed on normalization")
		}
	})
}
