// Package config loads the with-events scenario configuration.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds scenario application configuration.
type Config struct {
	Addr             string
	LogLevel         string
	WebhookTargetURL string
}

// Defaults returns safe local development settings.
func Defaults() Config {
	return Config{
		Addr:     ":8083",
		LogLevel: "info",
	}
}

// Load reads configuration from environment variables and command flags.
func Load() (Config, error) {
	cfg := Defaults()

	if value := strings.TrimSpace(os.Getenv("APP_ADDR")); value != "" {
		cfg.Addr = value
	}
	if value := strings.TrimSpace(os.Getenv("APP_LOG_LEVEL")); value != "" {
		cfg.LogLevel = value
	}
	if value := strings.TrimSpace(os.Getenv("WEBHOOK_TARGET_URL")); value != "" {
		cfg.WebhookTargetURL = value
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level")
	flag.StringVar(&cfg.WebhookTargetURL, "webhook-target-url", cfg.WebhookTargetURL, "optional webhook target URL")
	flag.Parse()

	return cfg, Validate(cfg)
}

// Validate returns an error when cfg cannot start the app.
func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		return fmt.Errorf("addr is required")
	}
	if strings.TrimSpace(cfg.LogLevel) == "" {
		return fmt.Errorf("log level is required")
	}
	return nil
}
