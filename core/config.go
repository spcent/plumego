package core

import (
	"fmt"
	"strings"
	"time"
)

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
	HTTP2Enabled      bool          // Whether to keep prepared-server TLS HTTP/2 support enabled
	DrainInterval     time.Duration // How often to log open HTTP connection counts while draining
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

func validateServerConfig(cfg AppConfig) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if cfg.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}
	if cfg.ReadHeaderTimeout < 0 {
		return fmt.Errorf("read header timeout cannot be negative")
	}
	if cfg.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}
	if cfg.IdleTimeout < 0 {
		return fmt.Errorf("idle timeout cannot be negative")
	}
	if cfg.MaxHeaderBytes < 0 {
		return fmt.Errorf("max header bytes cannot be negative")
	}
	return nil
}

// PreparationState describes the app's canonical kernel preparation phase.
type PreparationState string

const (
	PreparationStateMutable         PreparationState = "mutable"
	PreparationStateHandlerPrepared PreparationState = "handler_prepared"
	PreparationStateServerPrepared  PreparationState = "server_prepared"
)
