package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/core"
)

const (
	DefaultShutdownTimeout = 10 * time.Second
	DefaultHandlerTimeout  = 30 * time.Second
	DefaultMaxBodyBytes    = 2 * 1024 * 1024 // 2 MiB — generous for heartbeat JSON
)

type ServerConfig struct {
	Core            core.AppConfig
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
	HandlerTimeout  time.Duration
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Core:            core.DefaultConfig(),
		ShutdownTimeout: DefaultShutdownTimeout,
		MaxBodyBytes:    DefaultMaxBodyBytes,
		HandlerTimeout:  DefaultHandlerTimeout,
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
	if value, ok := lookup("WORKERFLEET_MAX_BODY_BYTES"); ok {
		bytes, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || bytes <= 0 {
			return ServerConfig{}, fmt.Errorf("parse WORKERFLEET_MAX_BODY_BYTES: must be a positive integer")
		}
		cfg.MaxBodyBytes = bytes
	}
	if value, ok := lookup("WORKERFLEET_HANDLER_TIMEOUT"); ok {
		timeout, err := time.ParseDuration(strings.TrimSpace(value))
		if err != nil {
			return ServerConfig{}, fmt.Errorf("parse WORKERFLEET_HANDLER_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return ServerConfig{}, fmt.Errorf("WORKERFLEET_HANDLER_TIMEOUT must be greater than zero")
		}
		cfg.HandlerTimeout = timeout
	}
	return cfg, nil
}
