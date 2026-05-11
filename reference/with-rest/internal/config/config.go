// Package config loads the with-rest demo configuration.
package config

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/core"
)

// Config holds all application configuration.
type Config struct {
	Core core.AppConfig
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8084"
	return Config{Core: coreCfg}
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Defaults()
	if addr := os.Getenv("APP_ADDR"); addr != "" {
		cfg.Core.Addr = addr
	}
	return cfg, Validate(cfg)
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	return nil
}
