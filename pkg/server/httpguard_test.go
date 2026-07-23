package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestGuard(bindHost, extraHosts, allowedOrigins string) *hostOriginGuard {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return newHostOriginGuard(next, bindHost, extraHosts, allowedOrigins)
}

func doRequest(g *hostOriginGuard, host, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Host = host
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rr := httptest.NewRecorder()
	g.ServeHTTP(rr, req)
	return rr
}

func TestHostOriginGuard_DNSRebindingBlocked(t *testing.T) {
	g := newTestGuard("127.0.0.1", "", "")

	// The MSRC reproduction: attacker-controlled Host/Origin via DNS rebinding.
	rr := doRequest(g, "attacker.example:8084", "http://attacker.example:8084")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected rebinding request to be forbidden, got %d", rr.Code)
	}
}

func TestHostOriginGuard_AttackerHostBlocked(t *testing.T) {
	g := newTestGuard("127.0.0.1", "", "")
	// Host alone is enough to reject, even without an Origin header.
	rr := doRequest(g, "attacker.example:8084", "")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected attacker Host to be forbidden, got %d", rr.Code)
	}
}

func TestHostOriginGuard_LoopbackAllowed(t *testing.T) {
	g := newTestGuard("127.0.0.1", "", "")
	cases := []struct{ host, origin string }{
		{"127.0.0.1:8084", ""},
		{"localhost:8084", "http://localhost:8084"},
		{"[::1]:8084", ""},
		{"127.0.0.1:8084", "http://127.0.0.1:8084"},
	}
	for _, c := range cases {
		rr := doRequest(g, c.host, c.origin)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected loopback host=%q origin=%q to be allowed, got %d", c.host, c.origin, rr.Code)
		}
	}
}

func TestHostOriginGuard_ConfiguredHostAllowed(t *testing.T) {
	// Binding to a non-loopback address must still trust that host.
	g := newTestGuard("10.0.0.5", "", "")
	rr := doRequest(g, "10.0.0.5:8084", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected configured bind host to be allowed, got %d", rr.Code)
	}
}

func TestHostOriginGuard_ExtraAllowedHost(t *testing.T) {
	g := newTestGuard("127.0.0.1", "mcp.internal", "")
	rr := doRequest(g, "mcp.internal:8084", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected explicitly allowed host to pass, got %d", rr.Code)
	}
	rr = doRequest(g, "other.internal:8084", "")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected non-allowed host to be forbidden, got %d", rr.Code)
	}
}

func TestHostOriginGuard_AllowedOrigin(t *testing.T) {
	g := newTestGuard("127.0.0.1", "", "https://app.example")
	// Host must still be loopback; a trusted cross-origin caller is permitted.
	rr := doRequest(g, "127.0.0.1:8084", "https://app.example")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected trusted origin to pass, got %d", rr.Code)
	}
	// Trailing slash normalized.
	rr = doRequest(g, "127.0.0.1:8084", "https://app.example/")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected trusted origin with trailing slash to pass, got %d", rr.Code)
	}
	// A different origin is rejected even with a valid loopback Host.
	rr = doRequest(g, "127.0.0.1:8084", "https://evil.example")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected untrusted origin to be forbidden, got %d", rr.Code)
	}
}

func TestHostOriginGuard_WildcardDisablesChecks(t *testing.T) {
	g := newTestGuard("127.0.0.1", "*", "*")
	rr := doRequest(g, "attacker.example:8084", "http://attacker.example:8084")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected wildcard config to allow any host/origin, got %d", rr.Code)
	}
}

func TestHostOriginGuard_NoOriginHeaderAllowedForLoopback(t *testing.T) {
	// Non-browser clients (no Origin) on a valid Host must not be blocked.
	g := newTestGuard("127.0.0.1", "", "")
	rr := doRequest(g, "localhost:8084", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected request without Origin to be allowed, got %d", rr.Code)
	}
}

func TestHostOriginGuard_WildcardBindHostNotTrusted(t *testing.T) {
	// Binding to 0.0.0.0/:: must not add the wildcard literal to the allow-list;
	// a request whose Host is the wildcard literal is meaningless and rejected,
	// while loopback still works.
	for _, bind := range []string{"0.0.0.0", "::"} {
		g := newTestGuard(bind, "", "")
		if rr := doRequest(g, bind+":8084", ""); rr.Code != http.StatusForbidden {
			t.Fatalf("bind=%q: expected wildcard Host to be forbidden, got %d", bind, rr.Code)
		}
		if rr := doRequest(g, "127.0.0.1:8084", ""); rr.Code != http.StatusOK {
			t.Fatalf("bind=%q: expected loopback Host to still be allowed, got %d", bind, rr.Code)
		}
	}
}

func TestHostOriginGuard_MalformedAllowHostIgnored(t *testing.T) {
	// A malformed --allow-hosts entry like ":8080" normalizes to "" and must
	// not be added to the allow-list (which would let an empty Host through).
	g := newTestGuard("127.0.0.1", ":8080", "")
	if rr := doRequest(g, "", ""); rr.Code != http.StatusForbidden {
		t.Fatalf("expected empty Host to be forbidden, got %d", rr.Code)
	}
	if rr := doRequest(g, "attacker.example:8080", ""); rr.Code != http.StatusForbidden {
		t.Fatalf("expected attacker Host to be forbidden, got %d", rr.Code)
	}
}
