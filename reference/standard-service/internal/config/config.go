// Package config loads and validates the reference application configuration
// from environment variables and command-line flags.
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
	Core core.AppConfig
	App  AppConfig
}

// AppConfig holds app-local, non-kernel configuration.
// Add application-specific fields here; keep framework/kernel config in Config.Core.
type AppConfig struct {
	EnvFile      string
	ServiceName  string // APP_SERVICE_NAME; used as the service identity in health responses.
	MaxBodyBytes int64  // APP_MAX_BODY_BYTES; maximum request body size. 0 disables the limit.
	WriteKey     string // APP_WRITE_KEY; when non-empty, POST/PUT/DELETE /api/v1/items require X-Write-Key header. Empty disables the guard.
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			EnvFile:      ".env",
			ServiceName:  "plumego-reference",
			MaxBodyBytes: 1 << 20, // 1 MiB
		},
	}
}

// Load reads configuration from environment variables and flags.
func Load() (Config, error) {
	return load(os.Args, os.LookupEnv)
}

func load(args []string, lookupEnv func(string) (string, bool)) (Config, error) {
	cfg := Defaults()
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	cfg.App.EnvFile = resolveEnvFile(args, lookupEnv, cfg.App.EnvFile)
	fileEnv, err := readEnvFile(cfg.App.EnvFile)
	if err != nil {
		return cfg, err
	}

	applyEnvMap(&cfg, fileEnv)
	applyEnv(&cfg, lookupEnv)
	if err := applyFlags(&cfg, args); err != nil {
		return cfg, err
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

func applyEnv(cfg *Config, lookupEnv func(string) (string, bool)) {
	cfg.Core.Addr = envString(lookupEnv, "APP_ADDR", cfg.Core.Addr)
	cfg.App.EnvFile = envString(lookupEnv, "APP_ENV_FILE", cfg.App.EnvFile)
	cfg.App.ServiceName = envString(lookupEnv, "APP_SERVICE_NAME", cfg.App.ServiceName)
	cfg.App.MaxBodyBytes = envInt64(lookupEnv, "APP_MAX_BODY_BYTES", cfg.App.MaxBodyBytes)
	cfg.App.WriteKey = envString(lookupEnv, "APP_WRITE_KEY", cfg.App.WriteKey)
}

func applyEnvMap(cfg *Config, values map[string]string) {
	if values == nil {
		return
	}
	if value := strings.TrimSpace(values["APP_ADDR"]); value != "" {
		cfg.Core.Addr = value
	}
	if value := strings.TrimSpace(values["APP_ENV_FILE"]); value != "" {
		cfg.App.EnvFile = value
	}
	if value := strings.TrimSpace(values["APP_SERVICE_NAME"]); value != "" {
		cfg.App.ServiceName = value
	}
	if value := strings.TrimSpace(values["APP_MAX_BODY_BYTES"]); value != "" {
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			cfg.App.MaxBodyBytes = n
		}
	}
	if value := strings.TrimSpace(values["APP_WRITE_KEY"]); value != "" {
		cfg.App.WriteKey = value
	}
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("standard-service", flag.ContinueOnError)
	fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	fs.StringVar(&cfg.App.EnvFile, "env-file", cfg.App.EnvFile, "path to .env file")
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
		case "addr", "env-file":
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

func resolveEnvFile(args []string, lookupEnv func(string) (string, bool), defaultPath string) string {
	if envPath, ok := lookupEnv("APP_ENV_FILE"); ok && strings.TrimSpace(envPath) != "" {
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

func readEnvFile(path string) (map[string]string, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		values[key] = value
	}
	return values, scanner.Err()
}

func envString(lookupEnv func(string) (string, bool), key, fallback string) string {
	if value, ok := lookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func envInt64(lookupEnv func(string) (string, bool), key string, fallback int64) int64 {
	if value, ok := lookupEnv(key); ok && value != "" {
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			return n
		}
	}
	return fallback
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
