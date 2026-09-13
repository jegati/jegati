// Package notification owns optional Web Push. It never logs endpoints, keys,
// capabilities, payloads or participant state.
package notification

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"time"
)

func endpointURL(raw string, hosts []string) (*url.URL, error) {
	u, e := url.Parse(raw)
	if e != nil || len(raw) > 2048 || u.Scheme != "https" || u.User != nil || u.Fragment != "" || u.Opaque != "" || !slices.Contains(hosts, u.Hostname()) || (u.Port() != "" && u.Port() != "443") || strings.Contains(u.Host, "%") {
		return nil, errors.New("unsupported push endpoint")
	}
	return u, nil
}

var excludedPrefixes = func() []netip.Prefix {
	out := []netip.Prefix{}
	for _, raw := range []string{"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4", "2001::/23", "2001:db8::/32", "2002::/16", "3fff::/20"} {
		out = append(out, netip.MustParsePrefix(raw))
	}
	return out
}()

func publicAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range excludedPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

type guardedDialer struct {
	hosts  []string
	lookup func(context.Context, string, string) ([]netip.Addr, error)
	dial   func(context.Context, string, string) (net.Conn, error)
}

func (d guardedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, e := net.SplitHostPort(address)
	if e != nil || port != "443" || !slices.Contains(d.hosts, host) {
		return nil, errors.New("push connection blocked")
	}
	ips, e := d.lookup(ctx, "ip", host)
	if e != nil || len(ips) == 0 || len(ips) > 16 {
		return nil, errors.New("push resolution unavailable")
	}
	for _, ip := range ips {
		if !publicAddress(ip) {
			return nil, errors.New("push address blocked")
		}
	}
	// Dial the checked address directly: do not re-resolve after validating DNS.
	// net/http retains the original hostname for TLS/SNI certificate verification.
	return d.dial(ctx, "tcp", net.JoinHostPort(ips[0].Unmap().String(), "443"))
}
func protectedClient(hosts []string, timeout time.Duration, concurrency int) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	guard := guardedDialer{hosts: hosts, lookup: net.DefaultResolver.LookupNetIP, dial: dialer.DialContext}
	return &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("push redirects forbidden") }, Transport: &http.Transport{
		Proxy: nil, DialContext: guard.DialContext, TLSHandshakeTimeout: timeout, ResponseHeaderTimeout: timeout,
		MaxResponseHeaderBytes: 8192, MaxConnsPerHost: concurrency, MaxIdleConns: concurrency, MaxIdleConnsPerHost: concurrency, IdleConnTimeout: 30 * time.Second,
	}}
}
