// Package config loads the with-gateway demo application configuration.
package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spcent/plumego/core"
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
	cfg.Core.Addr = envString("APP_ADDR", cfg.Core.Addr)
	cfg.App.EnvFile = envString("APP_ENV_FILE", cfg.App.EnvFile)
	cfg.App.Debug = envBool("APP_DEBUG", cfg.App.Debug)
	cfg.GatewayBackend = envString("GATEWAY_BACKEND", cfg.GatewayBackend)
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
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func parseEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	idx := strings.IndexByte(line, '=')
	if idx < 0 {
		return "", "", false
	}

	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}

	value = strings.TrimSpace(line[idx+1:])
	if len(value) >= 2 {
		q := value[0]
		if (q == '"' || q == '\'') && value[len(value)-1] == q {
			value = value[1 : len(value)-1]
			value = strings.ReplaceAll(value, string([]byte{'\\', q}), string(q))
		}
	}

	return key, value, true
}
