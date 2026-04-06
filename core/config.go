package core

import "time"

// TLSConfig defines TLS configuration.
type TLSConfig struct {
	Enabled  bool   // Whether to enable TLS
	CertFile string // Path to TLS certificate file
	KeyFile  string // Path to TLS private key file
}

// RouterConfig defines owned router behavior policy.
type RouterConfig struct {
	MethodNotAllowed bool // Whether to return 405 with Allow header on method mismatch
}

// AppConfig defines application configuration.
type AppConfig struct {
	Addr   string       // Server address
	TLS    TLSConfig    // TLS configuration
	Router RouterConfig // Router behavior policy
	// HTTP server hardening
	ReadTimeout       time.Duration // Maximum duration for reading the entire request, including the body
	ReadHeaderTimeout time.Duration // Maximum duration for reading the request headers (slowloris protection)
	WriteTimeout      time.Duration // Maximum duration before timing out writes of the response
	IdleTimeout       time.Duration // Maximum time to wait for the next request when keep-alives are enabled
	MaxHeaderBytes    int           // Maximum size of request headers
	HTTP2Enabled      bool          // Whether to keep HTTP/2 support enabled
	DrainInterval     time.Duration // How often to log in-flight connection counts while draining
}

// DefaultConfig returns the canonical baseline application configuration.
func DefaultConfig() AppConfig {
	return AppConfig{
		Addr:              ":8080",
		TLS:               TLSConfig{Enabled: false},
		Router:            RouterConfig{MethodNotAllowed: false},
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
		HTTP2Enabled:      true,
		DrainInterval:     500 * time.Millisecond,
	}
}

// RuntimeTLSSnapshot exposes the stable TLS subset used by first-party tooling.
type RuntimeTLSSnapshot struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// PreparationState describes the app's canonical kernel preparation phase.
type PreparationState string

const (
	PreparationStateMutable         PreparationState = "mutable"
	PreparationStateHandlerPrepared PreparationState = "handler_prepared"
	PreparationStateServerPrepared  PreparationState = "server_prepared"
)

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
	PreparationState  PreparationState   `json:"preparation_state"`
}

// runtimeSnapshot projects the AppConfig into a RuntimeSnapshot for introspection.
func (cfg AppConfig) runtimeSnapshot(state PreparationState) RuntimeSnapshot {
	return RuntimeSnapshot{
		Addr:              cfg.Addr,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		HTTP2Enabled:      cfg.HTTP2Enabled,
		DrainInterval:     cfg.DrainInterval,
		TLS: RuntimeTLSSnapshot{
			Enabled:  cfg.TLS.Enabled,
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
		},
		PreparationState: state,
	}
}
