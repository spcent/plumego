package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Addr       string
	AdminToken string
	LogLevel   string
}

func Defaults() Config {
	return Config{
		Addr:       ":8086",
		AdminToken: "local-admin-token",
		LogLevel:   "info",
	}
}

func Load() (Config, error) {
	cfg := Defaults()

	if value := strings.TrimSpace(os.Getenv("APP_ADDR")); value != "" {
		cfg.Addr = value
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_TOKEN")); value != "" {
		cfg.AdminToken = value
	}
	if value := strings.TrimSpace(os.Getenv("APP_LOG_LEVEL")); value != "" {
		cfg.LogLevel = value
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	flag.StringVar(&cfg.AdminToken, "admin-token", cfg.AdminToken, "static admin token")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level")
	flag.Parse()

	return cfg, Validate(cfg)
}

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		return fmt.Errorf("addr is required")
	}
	if strings.TrimSpace(cfg.AdminToken) == "" {
		return fmt.Errorf("admin token is required")
	}
	if strings.TrimSpace(cfg.LogLevel) == "" {
		return fmt.Errorf("log level is required")
	}
	return nil
}
