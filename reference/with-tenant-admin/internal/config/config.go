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

	if err := applyFlags(&cfg, os.Args); err != nil {
		return cfg, err
	}

	return cfg, Validate(cfg)
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("with-tenant-admin", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	fs.StringVar(&cfg.AdminToken, "admin-token", cfg.AdminToken, "static admin token")
	fs.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level")
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
		case "addr", "admin-token", "log-level":
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
