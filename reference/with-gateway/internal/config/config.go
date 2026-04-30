// Package config loads the with-gateway demo application configuration.
package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spcent/plumego/core"
	plumecfg "github.com/spcent/plumego/internal/config"
	plumelog "github.com/spcent/plumego/log"
)

// Config holds all application configuration.
type Config struct {
	Core           core.AppConfig
	App            AppConfig
	GatewayBackend string // backend URL to proxy to
}

// AppConfig holds app-local, non-kernel configuration.
type AppConfig struct {
	EnvFile string
	Debug   bool
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8083"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			EnvFile: ".env",
		},
		GatewayBackend: "http://localhost:9090",
	}
}

// Load reads configuration from environment variables and flags.
func Load() (Config, error) {
	cfg := Defaults()

	cfg.App.EnvFile = resolveEnvFile(os.Args, cfg.App.EnvFile)
	if err := loadEnvFile(cfg.App.EnvFile); err != nil {
		return cfg, err
	}

	if err := applyEnv(&cfg); err != nil {
		return cfg, err
	}
	applyFlags(&cfg)

	return cfg, Validate(cfg)
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	if cfg.GatewayBackend == "" {
		return fmt.Errorf("gateway backend URL is required (set GATEWAY_BACKEND)")
	}
	return nil
}

func applyEnv(cfg *Config) error {
	manager := plumecfg.NewManager(plumelog.NewLogger())
	if err := manager.AddSource(plumecfg.NewEnvSource("")); err != nil {
		return err
	}
	if err := manager.Load(context.Background()); err != nil {
		return err
	}
	cfg.Core.Addr = manager.GetString("app_addr", cfg.Core.Addr)
	cfg.App.EnvFile = manager.GetString("app_env_file", cfg.App.EnvFile)
	cfg.App.Debug = manager.GetBool("app_debug", cfg.App.Debug)
	cfg.GatewayBackend = manager.GetString("gateway_backend", cfg.GatewayBackend)
	return nil
}

func applyFlags(cfg *Config) {
	flag.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	flag.StringVar(&cfg.App.EnvFile, "env-file", cfg.App.EnvFile, "path to .env file")
	flag.BoolVar(&cfg.App.Debug, "debug", cfg.App.Debug, "enable debug mode")
	flag.StringVar(&cfg.GatewayBackend, "gateway-backend", cfg.GatewayBackend, "backend URL to proxy to")
	flag.Parse()
}

func resolveEnvFile(args []string, defaultPath string) string {
	if envPath := strings.TrimSpace(os.Getenv("APP_ENV_FILE")); envPath != "" {
		defaultPath = envPath
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--env-file" || arg == "-env-file" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--env-file=") {
			return strings.TrimPrefix(arg, "--env-file=")
		}
		if strings.HasPrefix(arg, "-env-file=") {
			return strings.TrimPrefix(arg, "-env-file=")
		}
	}
	return defaultPath
}

func loadEnvFile(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return plumecfg.LoadEnvFile(path, true)
}
