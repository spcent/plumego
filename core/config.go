package core

import "time"

// TLSConfig defines TLS configuration.
type TLSConfig struct {
	Enabled  bool   // Whether to enable TLS
	CertFile string // Path to TLS certificate file
	KeyFile  string // Path to TLS private key file
}

// AppConfig defines application configuration.
type AppConfig struct {
	Addr string    // Server address
	TLS  TLSConfig // TLS configuration
	// HTTP server hardening
	ReadTimeout       time.Duration // Maximum duration for reading the entire request, including the body
	ReadHeaderTimeout time.Duration // Maximum duration for reading the request headers (slowloris protection)
	WriteTimeout      time.Duration // Maximum duration before timing out writes of the response
	IdleTimeout       time.Duration // Maximum time to wait for the next request when keep-alives are enabled
	MaxHeaderBytes    int           // Maximum size of request headers
	EnableHTTP2       bool          // Whether to keep HTTP/2 support enabled
	DrainInterval     time.Duration // How often to log in-flight connection counts while draining
}

// DefaultConfig returns the canonical baseline application configuration.
func DefaultConfig() AppConfig {
	return AppConfig{
		Addr:              ":8080",
		TLS:               TLSConfig{Enabled: false},
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
		EnableHTTP2:       true,
		DrainInterval:     500 * time.Millisecond,
	}
}

// RuntimeTLSSnapshot exposes the stable TLS subset used by first-party tooling.
type RuntimeTLSSnapshot struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// RuntimeSnapshot exposes the stable runtime/config introspection contract used
// by first-party tooling and debug surfaces.
type RuntimeSnapshot struct {
	Addr              string             `json:"addr"`
	ReadTimeout       time.Duration      `json:"read_timeout"`
	ReadHeaderTimeout time.Duration      `json:"read_header_timeout"`
	WriteTimeout      time.Duration      `json:"write_timeout"`
	IdleTimeout       time.Duration      `json:"idle_timeout"`
	MaxHeaderBytes    int                `json:"max_header_bytes"`
	HTTP2Enabled      bool               `json:"http2_enabled"`
	DrainInterval     time.Duration      `json:"drain_interval"`
	TLS               RuntimeTLSSnapshot `json:"tls"`
	Started           bool               `json:"started"`
	ConfigFrozen      bool               `json:"config_frozen"`
	ServerPrepared    bool               `json:"server_prepared"`
}

func projectRuntimeSnapshot(cfg AppConfig) RuntimeSnapshot {
	return RuntimeSnapshot{
		Addr:              cfg.Addr,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		HTTP2Enabled:      cfg.EnableHTTP2,
		DrainInterval:     cfg.DrainInterval,
		TLS: RuntimeTLSSnapshot{
			Enabled:  cfg.TLS.Enabled,
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
		},
	}
}
