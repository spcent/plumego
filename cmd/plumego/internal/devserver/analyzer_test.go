package devserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAppSnapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_debug/config" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"addr":                ":8080",
				"debug":               true,
				"env_file":            ".env.test",
				"shutdown_timeout":    int64((5 * time.Second).Nanoseconds()),
				"read_timeout":        int64((30 * time.Second).Nanoseconds()),
				"read_header_timeout": int64((5 * time.Second).Nanoseconds()),
				"write_timeout":       int64((30 * time.Second).Nanoseconds()),
				"idle_timeout":        int64((60 * time.Second).Nanoseconds()),
				"max_header_bytes":    1048576,
				"http2_enabled":       true,
				"drain_interval":      int64((500 * time.Millisecond).Nanoseconds()),
				"started":             true,
				"config_frozen":       true,
				"server_prepared":     true,
				"tls": map[string]any{
					"enabled":   true,
					"cert_file": "cert.pem",
					"key_file":  "key.pem",
				},
			},
		})
	}))
	defer srv.Close()

	analyzer := NewAnalyzer(srv.URL)
	snapshot, err := analyzer.GetAppSnapshot()
	if err != nil {
		t.Fatalf("GetAppSnapshot returned error: %v", err)
	}

	if snapshot.Addr != ":8080" {
		t.Fatalf("addr = %q, want %q", snapshot.Addr, ":8080")
	}
	if snapshot.EnvFile != ".env.test" {
		t.Fatalf("env_file = %q, want %q", snapshot.EnvFile, ".env.test")
	}
	if !snapshot.Started || !snapshot.ConfigFrozen || !snapshot.ServerPrepared {
		t.Fatalf("unexpected lifecycle flags: %+v", snapshot)
	}
	if !snapshot.TLS.Enabled || snapshot.TLS.CertFile != "cert.pem" || snapshot.TLS.KeyFile != "key.pem" {
		t.Fatalf("unexpected tls snapshot: %+v", snapshot.TLS)
	}
}
