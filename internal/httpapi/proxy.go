package httpapi

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// TrustedProxy is an explicit deployment boundary, disabled for direct local use.
// Caddy must overwrite X-Gati-Client-IP from its authenticated ingress chain.
// A configured boundary requires every request to arrive through a listed peer.
func TrustedProxy(next http.Handler, ranges string) (http.Handler, error) {
	if ranges == "" {
		return next, nil
	}
	parts := strings.Split(ranges, ",")
	if len(parts) > 16 {
		return nil, errors.New("too many trusted proxy ranges")
	}
	prefixes := []netip.Prefix{}
	for _, value := range parts {
		p, err := netip.ParsePrefix(value)
		if err != nil || p.Bits() == 0 || p != p.Masked() {
			return nil, errors.New("invalid trusted proxy range")
		}
		prefixes = append(prefixes, p)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		peer, e := netip.ParseAddr(host)
		// Owner-only container health probes carry no participant state.
		if err == nil && e == nil && peer.IsLoopback() && r.Method == "GET" && r.URL.Path == "/healthz" && r.URL.RawQuery == "" {
			next.ServeHTTP(w, r)
			return
		}
		trusted := false
		if err == nil && e == nil {
			for _, p := range prefixes {
				if p.Contains(peer.Unmap()) {
					trusted = true
					break
				}
			}
		}
		values := r.Header.Values("X-Gati-Client-IP")
		if !trusted || len(values) != 1 {
			writeError(w, http.StatusForbidden)
			return
		}
		client, err := netip.ParseAddr(values[0])
		if err != nil || client.Zone() != "" || client.IsUnspecified() || client.IsMulticast() {
			writeError(w, http.StatusForbidden)
			return
		}
		copy := r.Clone(r.Context())
		copy.RemoteAddr = net.JoinHostPort(client.Unmap().String(), "0")
		for _, key := range []string{"X-Gati-Client-IP", "Forwarded", "X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP", "CF-Ray"} {
			copy.Header.Del(key)
		}
		next.ServeHTTP(w, copy)
	}), nil
}
