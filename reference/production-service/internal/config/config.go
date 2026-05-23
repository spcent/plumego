// Package config loads production reference configuration.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/core"
)

// Config holds all application configuration.
type Config struct {
	Core core.AppConfig
	App  AppConfig
}

// AppConfig holds app-local production profile settings.
type AppConfig struct {
	Environment      string
	ServiceName      string
	APIToken         string
	OpsToken         string
	ProfileStorePath string
	BodyLimitBytes   int64
	RequestTimeout   time.Duration
	RateLimit        float64
	RateBurst        int
}

// Defaults returns conservative local defaults for the production reference.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			Environment:    "local",
			ServiceName:    "plumego-production-reference",
			BodyLimitBytes: 1 << 20,
			RequestTimeout: 5 * time.Second,
			RateLimit:      20,
			RateBurst:      40,
		},
	}
}

// Load reads environment variables and command-line flags.
func Load() (Config, error) {
	cfg := Defaults()
	applyEnv(&cfg)
	if err := applyFlags(&cfg, os.Args); err != nil {
		return cfg, err
	}
	return cfg, Validate(cfg)
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	if cfg.App.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if cfg.App.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if cfg.App.BodyLimitBytes <= 0 {
		return fmt.Errorf("body limit bytes must be positive")
	}
	if cfg.App.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	if cfg.App.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}
	if cfg.App.RateBurst <= 0 {
		return fmt.Errorf("rate burst must be positive")
	}
	return nil
}

func applyEnv(cfg *Config) {
	cfg.Core.Addr = envString("APP_ADDR", cfg.Core.Addr)
	cfg.App.Environment = envString("APP_ENV", cfg.App.Environment)
	cfg.App.ServiceName = envString("APP_SERVICE_NAME", cfg.App.ServiceName)
	cfg.App.APIToken = envString("APP_API_TOKEN", cfg.App.APIToken)
	cfg.App.OpsToken = envString("OPS_TOKEN", cfg.App.OpsToken)
	cfg.App.ProfileStorePath = envString("APP_PROFILE_STORE_PATH", cfg.App.ProfileStorePath)
	cfg.App.BodyLimitBytes = envInt64("APP_BODY_LIMIT_BYTES", cfg.App.BodyLimitBytes)
	cfg.App.RequestTimeout = envDuration("APP_REQUEST_TIMEOUT", cfg.App.RequestTimeout)
	cfg.App.RateLimit = envFloat("APP_RATE_LIMIT", cfg.App.RateLimit)
	cfg.App.RateBurst = envInt("APP_RATE_BURST", cfg.App.RateBurst)
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("production-service", flag.ContinueOnError)
	fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	fs.StringVar(&cfg.App.Environment, "env", cfg.App.Environment, "deployment environment")
	fs.StringVar(&cfg.App.ServiceName, "service-name", cfg.App.ServiceName, "service name")
	fs.StringVar(&cfg.App.APIToken, "api-token", cfg.App.APIToken, "bearer token for protected API routes")
	fs.StringVar(&cfg.App.OpsToken, "ops-token", cfg.App.OpsToken, "bearer token for ops routes")
	fs.StringVar(&cfg.App.ProfileStorePath, "profile-store-path", cfg.App.ProfileStorePath, "optional JSON profile store path")
	fs.Int64Var(&cfg.App.BodyLimitBytes, "body-limit-bytes", cfg.App.BodyLimitBytes, "request body limit")
	fs.DurationVar(&cfg.App.RequestTimeout, "request-timeout", cfg.App.RequestTimeout, "request timeout")
	fs.Float64Var(&cfg.App.RateLimit, "rate-limit", cfg.App.RateLimit, "requests per second per client")
	fs.IntVar(&cfg.App.RateBurst, "rate-burst", cfg.App.RateBurst, "rate-limit burst capacity")
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
		case "addr", "env", "service-name", "api-token", "ops-token", "profile-store-path",
			"body-limit-bytes", "request-timeout", "rate-limit", "rate-burst":
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

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil || value == 0 {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value == 0 {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil || value == 0 {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value == 0 {
		return fallback
	}
	return value
}
