// Package config loads production reference configuration.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
	applyFlags(&cfg)
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
	cfg.App.ProfileStorePath = envString("APP_PROFILE_STORE_PATH", cfg.App.ProfileStorePath)
	cfg.App.BodyLimitBytes = envInt64("APP_BODY_LIMIT_BYTES", cfg.App.BodyLimitBytes)
	cfg.App.RequestTimeout = envDuration("APP_REQUEST_TIMEOUT", cfg.App.RequestTimeout)
	cfg.App.RateLimit = envFloat("APP_RATE_LIMIT", cfg.App.RateLimit)
	cfg.App.RateBurst = envInt("APP_RATE_BURST", cfg.App.RateBurst)
}

func applyFlags(cfg *Config) {
	flag.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	flag.StringVar(&cfg.App.Environment, "env", cfg.App.Environment, "deployment environment")
	flag.StringVar(&cfg.App.ServiceName, "service-name", cfg.App.ServiceName, "service name")
	flag.StringVar(&cfg.App.APIToken, "api-token", cfg.App.APIToken, "bearer token for protected API routes")
	flag.StringVar(&cfg.App.ProfileStorePath, "profile-store-path", cfg.App.ProfileStorePath, "optional JSON profile store path")
	flag.Int64Var(&cfg.App.BodyLimitBytes, "body-limit-bytes", cfg.App.BodyLimitBytes, "request body limit")
	flag.DurationVar(&cfg.App.RequestTimeout, "request-timeout", cfg.App.RequestTimeout, "request timeout")
	flag.Float64Var(&cfg.App.RateLimit, "rate-limit", cfg.App.RateLimit, "requests per second per client")
	flag.IntVar(&cfg.App.RateBurst, "rate-burst", cfg.App.RateBurst, "rate-limit burst capacity")
	flag.Parse()
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
