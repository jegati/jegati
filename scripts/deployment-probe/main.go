// Local deployment-test connector. Never copied into a production image.
package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	check := flag.String("check", "", "one anonymous GET instead of serving")
	want := flag.Int("want", 403, "expected check status")
	flag.Parse()
	if *check != "" {
		client := http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{}}
		r, err := client.Get(*check)
		if err != nil {
			os.Exit(2)
		}
		r.Body.Close()
		if r.StatusCode != *want {
			os.Exit(3)
		}
		return
	}
	target, _ := url.Parse("http://web:8080")
	proxy := &httputil.ReverseProxy{Rewrite: func(p *httputil.ProxyRequest) {
		p.SetURL(target)
		p.Out.Host = p.In.Host
		ip, _, _ := net.SplitHostPort(p.In.RemoteAddr)
		// Only this disposable, loopback-published lab accepts synthetic networks.
		// It is not an origin-authentication mechanism or a production connector.
		if values, ok := p.In.Header["X-Gati-Lab-Ip"]; ok && len(values) == 1 {
			ip = values[0]
		}
		for _, h := range []string{"CF-Connecting-IP", "CF-Ray", "X-Gati-Client-IP", "X-Forwarded-For", "Forwarded", "X-Gati-Lab-IP"} {
			p.Out.Header.Del(h)
		}
		p.Out.Header.Set("CF-Connecting-IP", ip)
	}, ErrorLog: log.New(io.Discard, "", 0)}
	server := http.Server{Addr: ":8090", Handler: proxy, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.New(io.Discard, "", 0)}
	if server.ListenAndServe() != nil {
		os.Exit(1)
	}
}
