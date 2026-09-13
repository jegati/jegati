package notification

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"
)

func TestEndpointAndAddressBoundaries(t *testing.T) {
	hosts := []string{"push.example.org"}
	for _, raw := range []string{"http://push.example.org/path", "https://push.example.org.evil.test/path", "https://user:secret@push.example.org/path", "https://push.example.org:444/path", "https://127.0.0.1/path", "https://push.example.org/path#fragment", "https://PUSH.EXAMPLE.ORG/path"} {
		if _, e := endpointURL(raw, hosts); e == nil {
			t.Fatal("endpoint policy accepted forbidden form")
		}
	}
	if _, e := endpointURL("https://push.example.org:443/opaque", hosts); e != nil {
		t.Fatal(e)
	}
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "100.64.0.1", "192.0.2.1", "198.18.1.1", "224.0.0.1", "::1", "::ffff:127.0.0.1", "fc00::1", "fe80::1", "2001:db8::1", "2002:7f00:1::1", "64:ff9b::7f00:1", "3fff::1"} {
		if publicAddress(netip.MustParseAddr(raw)) {
			t.Fatal("non-public address accepted", raw)
		}
	}
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2001:4860:4860::8888"} {
		if !publicAddress(netip.MustParseAddr(raw)) {
			t.Fatal("public address rejected", raw)
		}
	}
}
func TestDNSRebindingMixedAnswersAndNoRedirectProxy(t *testing.T) {
	calls := 0
	addresses := []netip.Addr{netip.MustParseAddr("8.8.8.8")}
	d := guardedDialer{hosts: []string{"push.example.org"}, lookup: func(context.Context, string, string) ([]netip.Addr, error) { return addresses, nil }, dial: func(_ context.Context, _ string, address string) (net.Conn, error) {
		calls++
		if address != "8.8.8.8:443" {
			t.Fatal("dial re-resolved hostname")
		}
		return nil, errors.New("synthetic stop")
	}}
	_, _ = d.DialContext(context.Background(), "tcp", "push.example.org:443")
	if calls != 1 {
		t.Fatal("checked public IP not dialed")
	}
	addresses = append(addresses, netip.MustParseAddr("10.0.0.1"))
	_, _ = d.DialContext(context.Background(), "tcp", "push.example.org:443")
	_, _ = d.DialContext(context.Background(), "tcp", "other.example.org:443")
	if calls != 1 {
		t.Fatal("mixed DNS or unlisted host reached network")
	}
	client := protectedClient(d.hosts, time.Second, 2)
	transport := client.Transport.(*http.Transport)
	if transport.Proxy != nil || transport.DialContext == nil || transport.MaxConnsPerHost != 2 || client.Timeout != time.Second || client.CheckRedirect(&http.Request{}, nil) == nil {
		t.Fatal("transport protection absent")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestRedirectDoesNotForwardAuthorization(t *testing.T) {
	client := protectedClient([]string{"push.example.org"}, time.Second, 1)
	calls := 0
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 307, Header: http.Header{"Location": []string{"https://other.example.org/collect"}}, Body: http.NoBody, Request: r}, nil
	})
	request, _ := http.NewRequest("POST", "https://push.example.org/synthetic", nil)
	request.Header.Set("Authorization", "synthetic-only")
	response, e := client.Do(request)
	if response != nil {
		response.Body.Close()
	}
	if e == nil || calls != 1 {
		t.Fatal("redirect forwarded authenticated request")
	}
}
