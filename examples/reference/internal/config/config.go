// Package config loads and validates the reference application configuration
// from environment variables and command-line flags.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	plumecfg "github.com/spcent/plumego/config"
	"github.com/spcent/plumego/core"
)

// Config holds all application configuration.
type Config struct {
	Core core.AppConfig

	WebSocketSecret string
	GitHubSecret    string
	StripeSecret    string
	WebhookToken    string
	EnableDocs      bool
	EnableMetrics   bool
	EnableWebhooks  bool
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	return Config{
		Core: core.AppConfig{
			Addr:    ":8080",
			EnvFile: ".env",
			Debug:   false,
		},
		WebSocketSecret: "dev-secret",
		GitHubSecret:    "dev-github-secret",
		StripeSecret:    "whsec_dev",
		WebhookToken:    "dev-trigger",
		EnableDocs:      true,
		EnableMetrics:   true,
		EnableWebhooks:  true,
	}
}

// Load reads configuration from environment variables and flags.
func Load() (Config, error) {
	cfg := Defaults()

	cfg.Core.EnvFile = resolveEnvFile(os.Args, cfg.Core.EnvFile)
	if err := loadEnvFile(cfg.Core.EnvFile); err != nil {
		return cfg, err
	}

	applyEnv(&cfg)
	applyFlags(&cfg)

	return cfg, Validate(cfg)
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	return nil
}

func applyEnv(cfg *Config) {
	cfg.Core.Addr = plumecfg.GetString("APP_ADDR", cfg.Core.Addr)
	cfg.Core.EnvFile = plumecfg.GetString("APP_ENV_FILE", cfg.Core.EnvFile)
	cfg.Core.Debug = plumecfg.GetBool("APP_DEBUG", cfg.Core.Debug)

	cfg.WebSocketSecret = plumecfg.GetString("WS_SECRET", cfg.WebSocketSecret)
	cfg.GitHubSecret = plumecfg.GetString("GITHUB_WEBHOOK_SECRET", cfg.GitHubSecret)
	cfg.StripeSecret = plumecfg.GetString("STRIPE_WEBHOOK_SECRET", cfg.StripeSecret)
	cfg.WebhookToken = plumecfg.GetString("WEBHOOK_TRIGGER_TOKEN", cfg.WebhookToken)
	cfg.EnableDocs = plumecfg.GetBool("ENABLE_DOCS", cfg.EnableDocs)
	cfg.EnableMetrics = plumecfg.GetBool("ENABLE_METRICS", cfg.EnableMetrics)
	cfg.EnableWebhooks = plumecfg.GetBool("ENABLE_WEBHOOKS", cfg.EnableWebhooks)
}

func applyFlags(cfg *Config) {
	flag.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	flag.StringVar(&cfg.Core.EnvFile, "env-file", cfg.Core.EnvFile, "path to .env file")
	flag.BoolVar(&cfg.Core.Debug, "debug", cfg.Core.Debug, "enable debug mode")
	flag.BoolVar(&cfg.EnableDocs, "enable-docs", cfg.EnableDocs, "enable docs site")
	flag.BoolVar(&cfg.EnableMetrics, "enable-metrics", cfg.EnableMetrics, "enable metrics")
	flag.BoolVar(&cfg.EnableWebhooks, "enable-webhooks", cfg.EnableWebhooks, "enable webhooks")
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
	return plumecfg.LoadEnv(path, true)
}
