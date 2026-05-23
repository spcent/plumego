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

	if err := applyFlags(&cfg, os.Args); err != nil {
		return cfg, err
	}

	return cfg, Validate(cfg)
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("with-events", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	fs.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level")
	fs.StringVar(&cfg.WebhookTargetURL, "webhook-target-url", cfg.WebhookTargetURL, "optional webhook target URL")
	if len(args) == 0 {
		return fs.Parse(nil)
	}
	return fs.Parse(configFlagArgs(args[1:]))
}

func configFlagArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, hasValue := flagName(arg)
		switch name {
		case "addr", "log-level", "webhook-target-url":
			out = append(out, arg)
			if !hasValue && i+1 < len(args) {
				i++
				out = append(out, args[i])
			}
		}
	}
	return out
}

func flagName(arg string) (name string, hasValue bool) {
	if strings.HasPrefix(arg, "--") {
		arg = strings.TrimPrefix(arg, "--")
	} else if strings.HasPrefix(arg, "-") {
		arg = strings.TrimPrefix(arg, "-")
	} else {
		return "", false
	}
	if idx := strings.IndexByte(arg, '='); idx >= 0 {
		return arg[:idx], true
	}
	return arg, false
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
