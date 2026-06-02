package config

import (
	"bufio"
	"flag"
	"fmt"
	"net"
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

// AppConfig holds dbadmin-specific configuration.
type AppConfig struct {
	EnvFile      string
	ServiceName  string
	MaxBodyBytes int64
	// MaxUploadBytes is the maximum size of a SQLite file upload.
	MaxUploadBytes int64
	// DataDir is where KV stores (sessions, connections, history) are persisted.
	DataDir string
	// AdminUser and AdminPassword are the login credentials for the web UI.
	AdminUser     string
	AdminPassword string
	// AdminRole is the single-user role: admin or readonly.
	AdminRole string
	// AllowDefaultPassword permits the demo-only admin/admin password.
	AllowDefaultPassword bool
	// SessionTTL controls how long web UI sessions remain valid.
	SessionTTL time.Duration
	// EncryptionKey is a 32-byte hex-encoded AES-GCM key for encrypting connection passwords.
	EncryptionKey string
	Version       string
	// QueryTimeoutSeconds is the maximum execution time for SQL queries (default: 30).
	QueryTimeoutSeconds int
	// QueryCancelEnabled enables query cancellation support (default: true).
	QueryCancelEnabled bool
	// RedisCommandTimeoutSeconds is the maximum execution time for Redis commands (default: 30).
	RedisCommandTimeoutSeconds int
	// MongoQueryTimeoutSeconds is the maximum execution time for MongoDB operations (default: 30).
	MongoQueryTimeoutSeconds int
	// ESQueryTimeoutSeconds is the maximum execution time for Elasticsearch operations (default: 30).
	ESQueryTimeoutSeconds int
	// ResourceListTimeoutSeconds is the maximum execution time for resource-tree listing (default: 30).
	ResourceListTimeoutSeconds int
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = "127.0.0.1:8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			EnvFile:                    ".env",
			ServiceName:                "dbadmin",
			MaxBodyBytes:               512 << 20, // 512 MiB (accommodates large SQLite uploads)
			MaxUploadBytes:             512 << 20, // 512 MiB per SQLite file
			DataDir:                    "./data",
			AdminUser:                  "admin",
			AdminPassword:              "",
			AdminRole:                  "admin",
			SessionTTL:                 24 * time.Hour,
			Version:                    "dev",
			QueryTimeoutSeconds:        30,
			QueryCancelEnabled:         true,
			RedisCommandTimeoutSeconds: 30,
			MongoQueryTimeoutSeconds:   30,
			ESQueryTimeoutSeconds:      30,
			ResourceListTimeoutSeconds: 30,
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
	if cfg.App.AdminUser == "" {
		return fmt.Errorf("DBADMIN_USER is required")
	}
	if cfg.App.AdminPassword == "" {
		return fmt.Errorf("DBADMIN_PASSWORD is required")
	}
	if isPublicAddr(cfg.Core.Addr) && cfg.App.AdminPassword == "admin" {
		return fmt.Errorf("refusing default password on non-loopback listen address")
	}
	if !cfg.App.AllowDefaultPassword && cfg.App.AdminUser == "admin" && cfg.App.AdminPassword == "admin" {
		return fmt.Errorf("refusing default admin/admin credentials; set DBADMIN_PASSWORD or DBADMIN_ALLOW_DEFAULT_PASSWORD=true for disposable loopback demos")
	}
	if cfg.App.AdminRole != "admin" && cfg.App.AdminRole != "readonly" {
		return fmt.Errorf("DBADMIN_ROLE must be admin or readonly")
	}
	if cfg.App.SessionTTL <= 0 {
		return fmt.Errorf("DBADMIN_SESSION_TTL must be positive")
	}
	return nil
}

func applyEnv(cfg *Config, lookupEnv func(string) (string, bool)) {
	if lookupEnv == nil {
		return
	}
	str := func(key string, dest *string) {
		if val, ok := lookupEnv(key); ok {
			if v := strings.TrimSpace(val); v != "" {
				*dest = v
			}
		}
	}
	int64f := func(key string, dest *int64) {
		if val, ok := lookupEnv(key); ok {
			if n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64); err == nil {
				*dest = n
			}
		}
	}
	intf := func(key string, dest *int) {
		if val, ok := lookupEnv(key); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				*dest = n
			}
		}
	}
	boolf := func(key string, dest *bool) {
		if val, ok := lookupEnv(key); ok {
			if b, err := strconv.ParseBool(strings.TrimSpace(val)); err == nil {
				*dest = b
			}
		}
	}
	durationf := func(key string, dest *time.Duration) {
		if val, ok := lookupEnv(key); ok {
			if d, err := time.ParseDuration(strings.TrimSpace(val)); err == nil && d > 0 {
				*dest = d
			}
		}
	}
	str("APP_ADDR", &cfg.Core.Addr)
	str("APP_ENV_FILE", &cfg.App.EnvFile)
	str("APP_SERVICE_NAME", &cfg.App.ServiceName)
	int64f("APP_MAX_BODY_BYTES", &cfg.App.MaxBodyBytes)
	int64f("APP_MAX_UPLOAD_BYTES", &cfg.App.MaxUploadBytes)
	boolf("APP_TLS_ENABLED", &cfg.Core.TLS.Enabled)
	str("APP_TLS_CERT_FILE", &cfg.Core.TLS.CertFile)
	str("APP_TLS_KEY_FILE", &cfg.Core.TLS.KeyFile)
	str("DBADMIN_DATA_DIR", &cfg.App.DataDir)
	str("DBADMIN_USER", &cfg.App.AdminUser)
	str("DBADMIN_PASSWORD", &cfg.App.AdminPassword)
	str("DBADMIN_ROLE", &cfg.App.AdminRole)
	boolf("DBADMIN_ALLOW_DEFAULT_PASSWORD", &cfg.App.AllowDefaultPassword)
	durationf("DBADMIN_SESSION_TTL", &cfg.App.SessionTTL)
	str("DBADMIN_ENCRYPTION_KEY", &cfg.App.EncryptionKey)
	intf("DBADMIN_QUERY_TIMEOUT_SECONDS", &cfg.App.QueryTimeoutSeconds)
	boolf("DBADMIN_QUERY_CANCEL_ENABLED", &cfg.App.QueryCancelEnabled)
	intf("DBADMIN_REDIS_COMMAND_TIMEOUT_SECONDS", &cfg.App.RedisCommandTimeoutSeconds)
	intf("DBADMIN_MONGO_QUERY_TIMEOUT_SECONDS", &cfg.App.MongoQueryTimeoutSeconds)
	intf("DBADMIN_ES_QUERY_TIMEOUT_SECONDS", &cfg.App.ESQueryTimeoutSeconds)
	intf("DBADMIN_RESOURCE_LIST_TIMEOUT_SECONDS", &cfg.App.ResourceListTimeoutSeconds)
}

func isPublicAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return true
	}
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return host != "localhost"
	}
	return !ip.IsLoopback()
}

func applyEnvMap(cfg *Config, values map[string]string) {
	if len(values) == 0 {
		return
	}
	applyEnv(cfg, func(key string) (string, bool) {
		v, ok := values[key]
		return v, ok
	})
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("dbadmin", flag.ContinueOnError)
	fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	fs.StringVar(&cfg.App.EnvFile, "env-file", cfg.App.EnvFile, "path to .env file")
	fs.StringVar(&cfg.App.DataDir, "data-dir", cfg.App.DataDir, "data directory for KV stores")
	if len(args) == 0 {
		return fs.Parse(nil)
	}
	known := make(map[string]bool)
	fs.VisitAll(func(f *flag.Flag) { known[f.Name] = true })
	return fs.Parse(filterFlagArgs(args[1:], known))
}

func filterFlagArgs(args []string, known map[string]bool) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, hasValue := flagName(arg)
		if !known[name] {
			continue
		}
		out = append(out, arg)
		if !hasValue && i+1 < len(args) {
			i++
			out = append(out, args[i])
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
