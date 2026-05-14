package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spcent/plumego/core"
)

const DefaultShutdownTimeout = 10 * time.Second

type ServerConfig struct {
	Core            core.AppConfig
	ShutdownTimeout time.Duration
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Core:            core.DefaultConfig(),
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

func LoadServerConfig(lookup func(string) (string, bool)) (ServerConfig, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	cfg := DefaultServerConfig()
	if value, ok := lookup("WORKERFLEET_HTTP_ADDR"); ok {
		addr := strings.TrimSpace(value)
		if addr == "" {
			return ServerConfig{}, fmt.Errorf("WORKERFLEET_HTTP_ADDR must not be empty")
		}
		cfg.Core.Addr = addr
	}
	if value, ok := lookup("WORKERFLEET_SHUTDOWN_TIMEOUT"); ok {
		timeout, err := time.ParseDuration(strings.TrimSpace(value))
		if err != nil {
			return ServerConfig{}, fmt.Errorf("parse WORKERFLEET_SHUTDOWN_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return ServerConfig{}, fmt.Errorf("WORKERFLEET_SHUTDOWN_TIMEOUT must be greater than zero")
		}
		cfg.ShutdownTimeout = timeout
	}
	return cfg, nil
}
