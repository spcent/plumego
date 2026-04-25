package devserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/x/devtools"
)

func TestGetAppSnapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_debug/config" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(struct {
			Data devtools.ConfigSnapshot `json:"data"`
		}{
			Data: devtools.ConfigSnapshot{
				Debug:   true,
				EnvFile: ".env.test",
				RuntimeSnapshot: devtools.RuntimeSnapshot{
					Addr:              ":8080",
					ReadTimeout:       30 * time.Second,
					ReadHeaderTimeout: 5 * time.Second,
					WriteTimeout:      30 * time.Second,
					IdleTimeout:       60 * time.Second,
					MaxHeaderBytes:    1048576,
					HTTP2Enabled:      true,
					DrainInterval:     500 * time.Millisecond,
					PreparationState:  "server_prepared",
					TLS: devtools.TLSSnapshot{
						Enabled:  true,
						CertFile: "cert.pem",
						KeyFile:  "key.pem",
					},
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
	if snapshot.PreparationState != "server_prepared" {
		t.Fatalf("unexpected preparation_state: %+v", snapshot)
	}
	if !snapshot.TLS.Enabled || snapshot.TLS.CertFile != "cert.pem" || snapshot.TLS.KeyFile != "key.pem" {
		t.Fatalf("unexpected tls snapshot: %+v", snapshot.TLS)
	}
}
