package client

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetHTTPClient_DefaultsAndOverrides(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Timeout = 5 * time.Second
	c1 := GetHTTPClient(cfg)
	c2 := GetHTTPClient(cfg)
	if c1 != c2 {
		t.Errorf("expected GetHTTPClient to memoize the *http.Client per config")
	}
	if c1.Timeout != 5*time.Second {
		t.Errorf("expected client timeout 5s, got %s", c1.Timeout)
	}
}

func TestHTTPProbeAgainstHttptest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	cfg := GetDefaultConfig()
	c := GetHTTPClient(cfg)
	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("http client failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCanCreateNetworkConnection_TCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			_, _ = conn.Write([]byte("PONG"))
			_ = conn.Close()
		}
	}()
	cfg := GetDefaultConfig()
	cfg.Timeout = 1 * time.Second
	connected, body := CanCreateNetworkConnection("tcp", listener.Addr().String(), "PING", cfg)
	if !connected {
		t.Fatalf("expected connection")
	}
	if !strings.HasPrefix(string(body), "PONG") {
		t.Errorf("expected PONG, got %q", body)
	}
}

func TestCanCreateNetworkConnection_TCPRefused(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Timeout = 200 * time.Millisecond
	// :1 is reserved (port 1) — almost certain to refuse on most systems.
	connected, _ := CanCreateNetworkConnection("tcp", "127.0.0.1:1", "", cfg)
	if connected {
		t.Errorf("expected refused connection")
	}
}

func TestParseDNSResolver(t *testing.T) {
	c := &Config{DNSResolver: "tcp://1.1.1.1:53"}
	r, err := c.parseDNSResolver()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if r.Protocol != "tcp" || r.Host != "1.1.1.1" || r.Port != "53" {
		t.Errorf("unexpected parsed value: %+v", r)
	}

	if _, err := (&Config{DNSResolver: "garbage"}).parseDNSResolver(); err == nil {
		t.Errorf("expected error for malformed resolver")
	}
}
