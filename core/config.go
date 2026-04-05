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
	Addr            string        // Server address
	EnvFile         string        // Path to .env file
	TLS             TLSConfig     // TLS configuration
	Debug           bool          // Debug mode
	ShutdownTimeout time.Duration // Graceful shutdown timeout
	// HTTP server hardening
	ReadTimeout       time.Duration // Maximum duration for reading the entire request, including the body
	ReadHeaderTimeout time.Duration // Maximum duration for reading the request headers (slowloris protection)
	WriteTimeout      time.Duration // Maximum duration before timing out writes of the response
	IdleTimeout       time.Duration // Maximum time to wait for the next request when keep-alives are enabled
	MaxHeaderBytes    int           // Maximum size of request headers
	EnableHTTP2       bool          // Whether to keep HTTP/2 support enabled
	DrainInterval     time.Duration // How often to log in-flight connection counts while draining
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
	Debug             bool               `json:"debug"`
	EnvFile           string             `json:"env_file"`
	ShutdownTimeout   time.Duration      `json:"shutdown_timeout"`
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
