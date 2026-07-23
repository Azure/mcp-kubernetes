package server

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// hostOriginGuard is an http.Handler middleware that protects the HTTP-based
// MCP transports (streamable-http and sse) against DNS-rebinding attacks.
//
// A browser-origin attacker can rebind an attacker-controlled hostname to the
// victim's loopback address while preserving the browser Origin. Without
// Host/Origin validation the browser can then complete an MCP session against
// the local server and drive Kubernetes operations through the server-held
// kubeconfig. The MCP specification requires servers that bind to localhost to
// validate the Host and Origin headers for exactly this reason.
//
// The guard enforces two independent checks before any request reaches MCP
// session handling:
//
//  1. Host allow-list. The request Host must be a loopback name, the configured
//     bind host, or an explicitly allowed host.
//  2. Origin allow-list. If an Origin header is present it must be explicitly
//     allowed. A browser always sends Origin on cross-origin requests, so a
//     rebinding page is rejected here.
type hostOriginGuard struct {
	next           http.Handler
	allowedHosts   map[string]struct{}
	allowedOrigin  map[string]struct{}
	allowAnyHost   bool
	allowAnyOrigin bool
}

// newHostOriginGuard builds the middleware. bindHost is the --host value the
// server listens on; extraHosts / allowedOrigins are the operator-configured
// comma-separated allow-lists. A "*" entry in either list disables that check.
func newHostOriginGuard(next http.Handler, bindHost, extraHosts, allowedOrigins string) *hostOriginGuard {
	g := &hostOriginGuard{
		next:          next,
		allowedHosts:  map[string]struct{}{},
		allowedOrigin: map[string]struct{}{},
	}

	// Always-trusted loopback host names.
	for _, h := range []string{"localhost", "127.0.0.1", "::1", "[::1]"} {
		g.allowedHosts[h] = struct{}{}
	}
	if h := normalizeHost(bindHost); h != "" {
		g.allowedHosts[h] = struct{}{}
	}
	for _, h := range splitList(extraHosts) {
		if h == "*" {
			g.allowAnyHost = true
			continue
		}
		g.allowedHosts[normalizeHost(h)] = struct{}{}
	}

	for _, o := range splitList(allowedOrigins) {
		if o == "*" {
			g.allowAnyOrigin = true
			continue
		}
		g.allowedOrigin[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
	}

	return g
}

func (g *hostOriginGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !g.hostAllowed(r.Host) {
		http.Error(w, "forbidden: Host header not allowed", http.StatusForbidden)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && !g.originAllowed(origin) {
		http.Error(w, "forbidden: Origin header not allowed", http.StatusForbidden)
		return
	}
	g.next.ServeHTTP(w, r)
}

func (g *hostOriginGuard) hostAllowed(host string) bool {
	if g.allowAnyHost {
		return true
	}
	_, ok := g.allowedHosts[normalizeHost(host)]
	return ok
}

func (g *hostOriginGuard) originAllowed(origin string) bool {
	if g.allowAnyOrigin {
		return true
	}
	normalized := strings.ToLower(strings.TrimRight(strings.TrimSpace(origin), "/"))
	if _, ok := g.allowedOrigin[normalized]; ok {
		return true
	}
	// Accept loopback origins so a legitimate local client on the same host is
	// not blocked, while a rebound attacker origin (a public hostname) is not.
	if u, err := url.Parse(normalized); err == nil {
		if host := u.Hostname(); isLoopbackHost(host) {
			return true
		}
	}
	return false
}

// normalizeHost lowercases the host and strips any port, so "attacker:8084"
// and "127.0.0.1:8084" compare against the bare host allow-list.
func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.Trim(host, "[]")
}

func isLoopbackHost(host string) bool {
	host = strings.Trim(strings.ToLower(host), "[]")
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
