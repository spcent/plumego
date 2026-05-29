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
	Core     core.AppConfig
	App      AppConfig
	DB       DBConfig
	Storage  StorageConfig
	Local    LocalConfig
	Qiniu    QiniuConfig
	Import   ImportConfig
	Search   SearchConfig
	Organize OrganizeConfig
}

// AppConfig holds app-local configuration.
type AppConfig struct {
	EnvFile         string
	Version         string
	MaxUploadSizeMB int64
	VersionPolicy   string
}

// DBConfig holds database configuration.
type DBConfig struct {
	Path string
}

// StorageConfig holds storage provider configuration.
type StorageConfig struct {
	Provider string // "local" or "qiniu"
}

// LocalConfig holds local filesystem storage configuration.
type LocalConfig struct {
	Root string
}

// QiniuConfig holds Qiniu Kodo configuration.
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string
	UseHTTPS  bool
}

// ImportConfig holds importer configuration.
type ImportConfig struct {
	MaxFileSizeMB int64
}

// OrganizeConfig holds V0.4 organize configuration.
type OrganizeConfig struct {
	DuplicateDetectionEnabled  bool
	SimilarityDetectionEnabled bool
	TagSuggestionEnabled       bool
	TopicBuildEnabled          bool

	NearDuplicateThreshold float64
	RelatedThreshold       float64

	MaxComparePerBucket   int
	SimilarityBatchSize   int

	AutoArchiveDuplicates   bool
	AutoApplyTagSuggestions bool

	PromptCandidateDetection bool
}

// SearchConfig holds full-text search configuration.
type SearchConfig struct {
	Enabled              bool
	IndexOnSave          bool
	IndexOnImport        string // "inline" | "background" | "disabled"
	IndexBatchSize       int
	IndexIntervalSeconds int
	MaxContentSizeMB     int64
	SnippetTokens        int
	HistoryLimit         int
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			EnvFile:         ".env",
			Version:         "dev",
			MaxUploadSizeMB: 10,
			VersionPolicy:   "always",
		},
		DB: DBConfig{
			Path: "./data/app.db",
		},
		Storage: StorageConfig{
			Provider: "local",
		},
		Local: LocalConfig{
			Root: "./data/objects",
		},
		Qiniu: QiniuConfig{
			Region:   "z0",
			UseHTTPS: true,
		},
		Import: ImportConfig{
			MaxFileSizeMB: 50,
		},
		Search: SearchConfig{
			Enabled:              true,
			IndexOnSave:          true,
			IndexOnImport:        "background",
			IndexBatchSize:       100,
			IndexIntervalSeconds: 5,
			MaxContentSizeMB:     5,
			SnippetTokens:        20,
			HistoryLimit:         100,
		},
		Organize: OrganizeConfig{
			DuplicateDetectionEnabled:  true,
			SimilarityDetectionEnabled: true,
			TagSuggestionEnabled:       true,
			TopicBuildEnabled:          true,
			NearDuplicateThreshold:     0.85,
			RelatedThreshold:           0.70,
			MaxComparePerBucket:        1000,
			SimilarityBatchSize:        500,
			AutoArchiveDuplicates:      false,
			AutoApplyTagSuggestions:    false,
			PromptCandidateDetection:   true,
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
	if cfg.Storage.Provider != "local" && cfg.Storage.Provider != "qiniu" {
		return fmt.Errorf("STORAGE_PROVIDER must be 'local' or 'qiniu'")
	}
	if cfg.Storage.Provider == "qiniu" {
		if cfg.Qiniu.AccessKey == "" {
			return fmt.Errorf("QINIU_ACCESS_KEY is required when using qiniu storage")
		}
		if cfg.Qiniu.SecretKey == "" {
			return fmt.Errorf("QINIU_SECRET_KEY is required when using qiniu storage")
		}
		if cfg.Qiniu.Bucket == "" {
			return fmt.Errorf("QINIU_BUCKET is required when using qiniu storage")
		}
		if cfg.Qiniu.Domain == "" {
			return fmt.Errorf("QINIU_DOMAIN is required when using qiniu storage")
		}
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
	boolf := func(key string, dest *bool) {
		if val, ok := lookupEnv(key); ok {
			if b, err := strconv.ParseBool(strings.TrimSpace(val)); err == nil {
				*dest = b
			}
		}
	}

	str("APP_ADDR", &cfg.Core.Addr)
	str("APP_ENV_FILE", &cfg.App.EnvFile)
	int64f("APP_MAX_UPLOAD_SIZE_MB", &cfg.App.MaxUploadSizeMB)
	str("APP_VERSION_POLICY", &cfg.App.VersionPolicy)

	str("DB_PATH", &cfg.DB.Path)

	str("STORAGE_PROVIDER", &cfg.Storage.Provider)
	str("LOCAL_ROOT", &cfg.Local.Root)

	str("QINIU_ACCESS_KEY", &cfg.Qiniu.AccessKey)
	str("QINIU_SECRET_KEY", &cfg.Qiniu.SecretKey)
	str("QINIU_BUCKET", &cfg.Qiniu.Bucket)
	str("QINIU_DOMAIN", &cfg.Qiniu.Domain)
	str("QINIU_REGION", &cfg.Qiniu.Region)
	boolf("QINIU_USE_HTTPS", &cfg.Qiniu.UseHTTPS)

	boolf("APP_TLS_ENABLED", &cfg.Core.TLS.Enabled)
	str("APP_TLS_CERT_FILE", &cfg.Core.TLS.CertFile)
	str("APP_TLS_KEY_FILE", &cfg.Core.TLS.KeyFile)
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
	fs := flag.NewFlagSet("cloud-vault", flag.ContinueOnError)
	fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	fs.StringVar(&cfg.App.EnvFile, "env-file", cfg.App.EnvFile, "path to .env file")
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
