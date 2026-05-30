package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
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
	AI       AIConfig
	Auth     AuthConfig
	Desktop  DesktopConfig
	Update   UpdateConfig
}

// AppConfig holds app-local configuration.
type AppConfig struct {
	ConfigFile      string
	Version         string
	MaxUploadSizeMB int64
	VersionPolicy   string
}

// DesktopConfig holds desktop application configuration.
type DesktopConfig struct {
	AppName                string
	DataDir                string
	CloseToTray            bool
	CloseToTraySet         bool // tracks if CloseToTray was explicitly set
	NativeNotifications    bool
	NativeNotificationsSet bool
	LaunchAtLogin          bool
}

// UpdateConfig holds V1.0 auto-update configuration.
type UpdateConfig struct {
	Enabled          bool
	CheckOnStartup   bool
	Channel          string // "stable" | "beta" | "dev"
	CheckIntervalMin int
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

	MaxComparePerBucket int
	SimilarityBatchSize int

	AutoArchiveDuplicates   bool
	AutoApplyTagSuggestions bool

	PromptCandidateDetection bool
}

// AIConfig holds V0.5 AI configuration.
type AIConfig struct {
	Enabled              bool
	Provider             string // "local_mock" | "openai_compatible"
	BaseURL              string // OpenAI-compatible endpoint, e.g. https://api.openai.com/v1
	APIKey               string
	Model                string
	MaxContextTokens     int
	MaxRetries           int
	TaskWorkers          int
	SummaryEnabled       bool
	QAEnabled            bool
	PromptExtractEnabled bool
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

// AuthConfig holds V0.7 authentication configuration.
type AuthConfig struct {
	Enabled                   bool
	SessionTTLHours           int
	CookieName                string
	SecureCookie              bool
	MaxLoginFailures          int
	LoginFailureWindowMinutes int
	LockoutMinutes            int
	PasswordMinLength         int
	BootstrapAdminEnabled     bool
	BootstrapAdminUsername    string
	BootstrapAdminEmail       string
	BootstrapAdminPassword    string
}

// tomlConfig mirrors Config for TOML deserialization with struct tags.
type tomlConfig struct {
	Server   tomlServer   `toml:"server"`
	App      tomlApp      `toml:"app"`
	Database tomlDB       `toml:"database"`
	Storage  tomlStorage  `toml:"storage"`
	Import   tomlImport   `toml:"import"`
	Search   tomlSearch   `toml:"search"`
	Organize tomlOrganize `toml:"organize"`
	AI       tomlAI       `toml:"ai"`
	Auth     tomlAuth     `toml:"auth"`
	Desktop  tomlDesktop  `toml:"desktop"`
	Update   tomlUpdate   `toml:"update"`
}

type tomlServer struct {
	Addr string `toml:"addr"`
	TLS  struct {
		Enabled  bool   `toml:"enabled"`
		CertFile string `toml:"cert_file"`
		KeyFile  string `toml:"key_file"`
	} `toml:"tls"`
}

type tomlApp struct {
	MaxUploadSizeMB int64  `toml:"max_upload_size_mb"`
	VersionPolicy   string `toml:"version_policy"`
}

type tomlDB struct {
	Path string `toml:"path"`
}

type tomlStorage struct {
	Provider string `toml:"provider"`
	Local    struct {
		Root string `toml:"root"`
	} `toml:"local"`
	Qiniu struct {
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		Bucket    string `toml:"bucket"`
		Domain    string `toml:"domain"`
		Region    string `toml:"region"`
		UseHTTPS  bool   `toml:"use_https"`
	} `toml:"qiniu"`
}

type tomlImport struct {
	MaxFileSizeMB int64 `toml:"max_file_size_mb"`
}

type tomlSearch struct {
	Enabled              bool   `toml:"enabled"`
	IndexOnSave          bool   `toml:"index_on_save"`
	IndexOnImport        string `toml:"index_on_import"`
	IndexBatchSize       int    `toml:"index_batch_size"`
	IndexIntervalSeconds int    `toml:"index_interval_seconds"`
	MaxContentSizeMB     int64  `toml:"max_content_size_mb"`
	SnippetTokens        int    `toml:"snippet_tokens"`
	HistoryLimit         int    `toml:"history_limit"`
}

type tomlOrganize struct {
	DuplicateDetection  bool    `toml:"duplicate_detection"`
	SimilarityDetection bool    `toml:"similarity_detection"`
	TagSuggestion       bool    `toml:"tag_suggestion"`
	TopicBuild          bool    `toml:"topic_build"`
	NearDupThreshold    float64 `toml:"near_dup_threshold"`
	RelatedThreshold    float64 `toml:"related_threshold"`
	MaxComparePerBucket int     `toml:"max_compare_per_bucket"`
	SimilarityBatchSize int     `toml:"similarity_batch_size"`
	AutoArchiveDups     bool    `toml:"auto_archive_duplicates"`
	AutoApplyTags       bool    `toml:"auto_apply_tag_suggestions"`
	PromptCandidate     bool    `toml:"prompt_candidate_detection"`
}

type tomlAI struct {
	Enabled          bool   `toml:"enabled"`
	Provider         string `toml:"provider"`
	BaseURL          string `toml:"base_url"`
	APIKey           string `toml:"api_key"`
	Model            string `toml:"model"`
	MaxContextTokens int    `toml:"max_context_tokens"`
	MaxRetries       int    `toml:"max_retries"`
	TaskWorkers      int    `toml:"task_workers"`
	SummaryEnabled   bool   `toml:"summary_enabled"`
	QAEnabled        bool   `toml:"qa_enabled"`
	PromptExtract    bool   `toml:"prompt_extract_enabled"`
}

type tomlAuth struct {
	Enabled                   bool   `toml:"enabled"`
	SessionTTLHours           int    `toml:"session_ttl_hours"`
	CookieName                string `toml:"cookie_name"`
	SecureCookie              bool   `toml:"secure_cookie"`
	MaxLoginFailures          int    `toml:"max_login_failures"`
	LoginFailureWindowMinutes int    `toml:"login_failure_window_minutes"`
	LockoutMinutes            int    `toml:"lockout_minutes"`
	PasswordMinLength         int    `toml:"password_min_length"`
	BootstrapAdmin            struct {
		Enabled  bool   `toml:"enabled"`
		Username string `toml:"username"`
		Email    string `toml:"email"`
		Password string `toml:"password"`
	} `toml:"bootstrap_admin"`
}

type tomlDesktop struct {
	Enabled             bool   `toml:"enabled"`
	AppName             string `toml:"app_name"`
	DataDir             string `toml:"data_dir"`
	StartHidden         bool   `toml:"start_hidden"`
	CloseToTray         bool   `toml:"close_to_tray"`
	ShowTrayIcon        bool   `toml:"show_tray_icon"`
	LaunchAtLogin       bool   `toml:"launch_at_login"`
	TrustedLocalMode    bool   `toml:"trusted_local_mode"`
	NativeNotifications bool   `toml:"native_notifications"`
	Import              struct {
		DefaultRecursive       bool `toml:"default_recursive"`
		DefaultAutoTagFromPath bool `toml:"default_auto_tag_from_path"`
		ShowScanPreview        bool `toml:"show_scan_preview"`
	} `toml:"import"`
}

type tomlUpdate struct {
	Enabled          bool   `toml:"enabled"`
	CheckOnStartup   bool   `toml:"check_on_startup"`
	CheckIntervalMin int    `toml:"check_interval_minutes"`
	Channel          string `toml:"channel"`
}

// Defaults returns safe configuration values for local development.
func Defaults() Config {
	coreCfg := core.DefaultConfig()
	coreCfg.Addr = ":8080"
	return Config{
		Core: coreCfg,
		App: AppConfig{
			ConfigFile:      "config.toml",
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
		AI: AIConfig{
			Enabled:              false,
			Provider:             "local_mock",
			Model:                "gpt-4o-mini",
			MaxContextTokens:     8000,
			MaxRetries:           2,
			TaskWorkers:          2,
			SummaryEnabled:       true,
			QAEnabled:            true,
			PromptExtractEnabled: true,
		},
		Auth: AuthConfig{
			Enabled:                   false,
			SessionTTLHours:           720,
			CookieName:                "cv_session",
			SecureCookie:              false,
			MaxLoginFailures:          5,
			LoginFailureWindowMinutes: 15,
			LockoutMinutes:            15,
			PasswordMinLength:         12,
			BootstrapAdminEnabled:     false,
		},
		Desktop: DesktopConfig{
			AppName:             "Markdown Cloud Vault",
			CloseToTray:         true,
			NativeNotifications: true,
		},
		Update: UpdateConfig{
			Enabled:          true,
			CheckOnStartup:   true,
			CheckIntervalMin: 1440, // 24 hours
			Channel:          "stable",
		},
	}
}

// Load reads configuration from the TOML config file and environment variable overrides.
func Load() (Config, error) {
	return load(os.Args, os.LookupEnv)
}

func load(args []string, lookupEnv func(string) (string, bool)) (Config, error) {
	cfg := Defaults()

	// Resolve config file path from flag or env before parsing.
	configFile := resolveConfigFile(args, lookupEnv, cfg.App.ConfigFile)

	// Read TOML file if it exists.
	if err := readTOMLFile(configFile, &cfg); err != nil {
		return cfg, err
	}

	// Layer 3: .env file overrides TOML.
	// Layer 4: real env vars override .env. Combined via combineLookup.
	dotenvPath := ".env"
	if lookupEnv != nil {
		if v, ok := lookupEnv("DOT_ENV_FILE"); ok && v != "" {
			dotenvPath = v
		}
	}
	dotenv := loadDotEnvFile(dotenvPath)
	combined := combineLookup(dotenv, lookupEnv)
	applyEnvOverrides(&cfg, combined)

	// CLI flags override everything.
	if err := applyFlags(&cfg, args); err != nil {
		return cfg, err
	}

	// Full V0.8 validation (includes side effects like creating data dirs).
	if err := ValidateConfig(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func readTOMLFile(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // missing file is fine, use defaults
		}
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	var tc tomlConfig
	if err := toml.Unmarshal(data, &tc); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	applyTOML(cfg, &tc)
	return nil
}

func applyTOML(cfg *Config, tc *tomlConfig) {
	// Server
	if tc.Server.Addr != "" {
		cfg.Core.Addr = tc.Server.Addr
	}
	cfg.Core.TLS.Enabled = tc.Server.TLS.Enabled
	if tc.Server.TLS.CertFile != "" {
		cfg.Core.TLS.CertFile = tc.Server.TLS.CertFile
	}
	if tc.Server.TLS.KeyFile != "" {
		cfg.Core.TLS.KeyFile = tc.Server.TLS.KeyFile
	}

	// App
	if tc.App.MaxUploadSizeMB > 0 {
		cfg.App.MaxUploadSizeMB = tc.App.MaxUploadSizeMB
	}
	if tc.App.VersionPolicy != "" {
		cfg.App.VersionPolicy = tc.App.VersionPolicy
	}

	// Database
	if tc.Database.Path != "" {
		cfg.DB.Path = tc.Database.Path
	}

	// Storage
	if tc.Storage.Provider != "" {
		cfg.Storage.Provider = tc.Storage.Provider
	}
	if tc.Storage.Local.Root != "" {
		cfg.Local.Root = tc.Storage.Local.Root
	}
	if tc.Storage.Qiniu.AccessKey != "" {
		cfg.Qiniu.AccessKey = tc.Storage.Qiniu.AccessKey
	}
	if tc.Storage.Qiniu.SecretKey != "" {
		cfg.Qiniu.SecretKey = tc.Storage.Qiniu.SecretKey
	}
	if tc.Storage.Qiniu.Bucket != "" {
		cfg.Qiniu.Bucket = tc.Storage.Qiniu.Bucket
	}
	if tc.Storage.Qiniu.Domain != "" {
		cfg.Qiniu.Domain = tc.Storage.Qiniu.Domain
	}
	if tc.Storage.Qiniu.Region != "" {
		cfg.Qiniu.Region = tc.Storage.Qiniu.Region
	}
	cfg.Qiniu.UseHTTPS = tc.Storage.Qiniu.UseHTTPS

	// Import
	if tc.Import.MaxFileSizeMB > 0 {
		cfg.Import.MaxFileSizeMB = tc.Import.MaxFileSizeMB
	}

	// Search
	cfg.Search.Enabled = tc.Search.Enabled
	cfg.Search.IndexOnSave = tc.Search.IndexOnSave
	if tc.Search.IndexOnImport != "" {
		cfg.Search.IndexOnImport = tc.Search.IndexOnImport
	}
	if tc.Search.IndexBatchSize > 0 {
		cfg.Search.IndexBatchSize = tc.Search.IndexBatchSize
	}
	if tc.Search.IndexIntervalSeconds > 0 {
		cfg.Search.IndexIntervalSeconds = tc.Search.IndexIntervalSeconds
	}
	if tc.Search.MaxContentSizeMB > 0 {
		cfg.Search.MaxContentSizeMB = tc.Search.MaxContentSizeMB
	}
	if tc.Search.SnippetTokens > 0 {
		cfg.Search.SnippetTokens = tc.Search.SnippetTokens
	}
	if tc.Search.HistoryLimit > 0 {
		cfg.Search.HistoryLimit = tc.Search.HistoryLimit
	}

	// Organize
	cfg.Organize.DuplicateDetectionEnabled = tc.Organize.DuplicateDetection
	cfg.Organize.SimilarityDetectionEnabled = tc.Organize.SimilarityDetection
	cfg.Organize.TagSuggestionEnabled = tc.Organize.TagSuggestion
	cfg.Organize.TopicBuildEnabled = tc.Organize.TopicBuild
	if tc.Organize.NearDupThreshold > 0 {
		cfg.Organize.NearDuplicateThreshold = tc.Organize.NearDupThreshold
	}
	if tc.Organize.RelatedThreshold > 0 {
		cfg.Organize.RelatedThreshold = tc.Organize.RelatedThreshold
	}
	if tc.Organize.MaxComparePerBucket > 0 {
		cfg.Organize.MaxComparePerBucket = tc.Organize.MaxComparePerBucket
	}
	if tc.Organize.SimilarityBatchSize > 0 {
		cfg.Organize.SimilarityBatchSize = tc.Organize.SimilarityBatchSize
	}
	cfg.Organize.AutoArchiveDuplicates = tc.Organize.AutoArchiveDups
	cfg.Organize.AutoApplyTagSuggestions = tc.Organize.AutoApplyTags
	cfg.Organize.PromptCandidateDetection = tc.Organize.PromptCandidate

	// AI
	cfg.AI.Enabled = tc.AI.Enabled
	if tc.AI.Provider != "" {
		cfg.AI.Provider = tc.AI.Provider
	}
	if tc.AI.BaseURL != "" {
		cfg.AI.BaseURL = tc.AI.BaseURL
	}
	if tc.AI.APIKey != "" {
		cfg.AI.APIKey = tc.AI.APIKey
	}
	if tc.AI.Model != "" {
		cfg.AI.Model = tc.AI.Model
	}
	if tc.AI.MaxContextTokens > 0 {
		cfg.AI.MaxContextTokens = tc.AI.MaxContextTokens
	}
	if tc.AI.MaxRetries > 0 {
		cfg.AI.MaxRetries = tc.AI.MaxRetries
	}
	if tc.AI.TaskWorkers > 0 {
		cfg.AI.TaskWorkers = tc.AI.TaskWorkers
	}
	cfg.AI.SummaryEnabled = tc.AI.SummaryEnabled
	cfg.AI.QAEnabled = tc.AI.QAEnabled
	cfg.AI.PromptExtractEnabled = tc.AI.PromptExtract

	// Auth
	cfg.Auth.Enabled = tc.Auth.Enabled
	if tc.Auth.SessionTTLHours > 0 {
		cfg.Auth.SessionTTLHours = tc.Auth.SessionTTLHours
	}
	if tc.Auth.CookieName != "" {
		cfg.Auth.CookieName = tc.Auth.CookieName
	}
	cfg.Auth.SecureCookie = tc.Auth.SecureCookie
	if tc.Auth.MaxLoginFailures > 0 {
		cfg.Auth.MaxLoginFailures = tc.Auth.MaxLoginFailures
	}
	if tc.Auth.LoginFailureWindowMinutes > 0 {
		cfg.Auth.LoginFailureWindowMinutes = tc.Auth.LoginFailureWindowMinutes
	}
	if tc.Auth.LockoutMinutes > 0 {
		cfg.Auth.LockoutMinutes = tc.Auth.LockoutMinutes
	}
	if tc.Auth.PasswordMinLength > 0 {
		cfg.Auth.PasswordMinLength = tc.Auth.PasswordMinLength
	}
	cfg.Auth.BootstrapAdminEnabled = tc.Auth.BootstrapAdmin.Enabled
	if tc.Auth.BootstrapAdmin.Username != "" {
		cfg.Auth.BootstrapAdminUsername = tc.Auth.BootstrapAdmin.Username
	}
	if tc.Auth.BootstrapAdmin.Email != "" {
		cfg.Auth.BootstrapAdminEmail = tc.Auth.BootstrapAdmin.Email
	}
	if tc.Auth.BootstrapAdmin.Password != "" {
		cfg.Auth.BootstrapAdminPassword = tc.Auth.BootstrapAdmin.Password
	}

	// Desktop
	if tc.Desktop.AppName != "" {
		cfg.Desktop.AppName = tc.Desktop.AppName
	}
	if tc.Desktop.DataDir != "" {
		cfg.Desktop.DataDir = tc.Desktop.DataDir
	}
	cfg.Desktop.CloseToTray = tc.Desktop.CloseToTray
	cfg.Desktop.CloseToTraySet = true
	cfg.Desktop.NativeNotifications = tc.Desktop.NativeNotifications
	cfg.Desktop.NativeNotificationsSet = true
	cfg.Desktop.LaunchAtLogin = tc.Desktop.LaunchAtLogin

	// Update
	cfg.Update.Enabled = tc.Update.Enabled
	cfg.Update.CheckOnStartup = tc.Update.CheckOnStartup
	if tc.Update.CheckIntervalMin > 0 {
		cfg.Update.CheckIntervalMin = tc.Update.CheckIntervalMin
	}
	if tc.Update.Channel != "" {
		cfg.Update.Channel = tc.Update.Channel
	}
}

// Validate returns an error if cfg is unusable.
func Validate(cfg Config) error {
	if cfg.Core.Addr == "" {
		return fmt.Errorf("server.addr is required")
	}
	if cfg.Storage.Provider != "local" && cfg.Storage.Provider != "qiniu" {
		return fmt.Errorf("storage.provider must be 'local' or 'qiniu'")
	}
	if cfg.Storage.Provider == "qiniu" {
		if cfg.Qiniu.AccessKey == "" {
			return fmt.Errorf("storage.qiniu.access_key is required when using qiniu storage")
		}
		if cfg.Qiniu.SecretKey == "" {
			return fmt.Errorf("storage.qiniu.secret_key is required when using qiniu storage")
		}
		if cfg.Qiniu.Bucket == "" {
			return fmt.Errorf("storage.qiniu.bucket is required when using qiniu storage")
		}
		if cfg.Qiniu.Domain == "" {
			return fmt.Errorf("storage.qiniu.domain is required when using qiniu storage")
		}
	}
	if cfg.Auth.Enabled {
		if cfg.Auth.SessionTTLHours <= 0 {
			return fmt.Errorf("auth.session_ttl_hours must be greater than 0")
		}
		if cfg.Auth.PasswordMinLength < 8 {
			return fmt.Errorf("auth.password_min_length must be at least 8")
		}
	}
	return nil
}

// applyEnvOverrides allows environment variables to override TOML values.
// This is useful for secrets (API keys, passwords) and deployment-specific settings.
func applyEnvOverrides(cfg *Config, lookupEnv func(string) (string, bool)) {
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
	intf := func(key string, dest *int) {
		if val, ok := lookupEnv(key); ok {
			if n, err := parseInt(strings.TrimSpace(val)); err == nil {
				*dest = n
			}
		}
	}
	int64f := func(key string, dest *int64) {
		if val, ok := lookupEnv(key); ok {
			if n, err := parseInt64(strings.TrimSpace(val)); err == nil {
				*dest = n
			}
		}
	}
	boolf := func(key string, dest *bool) {
		if val, ok := lookupEnv(key); ok {
			if b, err := parseBool(strings.TrimSpace(val)); err == nil {
				*dest = b
			}
		}
	}

	str("APP_ADDR", &cfg.Core.Addr)
	str("APP_CONFIG_FILE", &cfg.App.ConfigFile)
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

	boolf("AI_ENABLED", &cfg.AI.Enabled)
	str("AI_PROVIDER", &cfg.AI.Provider)
	str("AI_BASE_URL", &cfg.AI.BaseURL)
	str("AI_API_KEY", &cfg.AI.APIKey)
	str("AI_MODEL", &cfg.AI.Model)
	intf("AI_MAX_CONTEXT_TOKENS", &cfg.AI.MaxContextTokens)
	intf("AI_MAX_RETRIES", &cfg.AI.MaxRetries)
	intf("AI_TASK_WORKERS", &cfg.AI.TaskWorkers)
	boolf("AI_SUMMARY_ENABLED", &cfg.AI.SummaryEnabled)
	boolf("AI_QA_ENABLED", &cfg.AI.QAEnabled)
	boolf("AI_PROMPT_EXTRACT_ENABLED", &cfg.AI.PromptExtractEnabled)

	boolf("AUTH_ENABLED", &cfg.Auth.Enabled)
	intf("AUTH_SESSION_TTL_HOURS", &cfg.Auth.SessionTTLHours)
	str("AUTH_COOKIE_NAME", &cfg.Auth.CookieName)
	boolf("AUTH_SECURE_COOKIE", &cfg.Auth.SecureCookie)
	intf("AUTH_MAX_LOGIN_FAILURES", &cfg.Auth.MaxLoginFailures)
	intf("AUTH_LOGIN_FAILURE_WINDOW_MINUTES", &cfg.Auth.LoginFailureWindowMinutes)
	intf("AUTH_LOCKOUT_MINUTES", &cfg.Auth.LockoutMinutes)
	intf("AUTH_PASSWORD_MIN_LENGTH", &cfg.Auth.PasswordMinLength)
	boolf("AUTH_BOOTSTRAP_ADMIN_ENABLED", &cfg.Auth.BootstrapAdminEnabled)
	str("AUTH_BOOTSTRAP_ADMIN_USERNAME", &cfg.Auth.BootstrapAdminUsername)
	str("AUTH_BOOTSTRAP_ADMIN_EMAIL", &cfg.Auth.BootstrapAdminEmail)
	str("AUTH_BOOTSTRAP_ADMIN_PASSWORD", &cfg.Auth.BootstrapAdminPassword)

	// Desktop environment variables
	str("DESKTOP_APP_NAME", &cfg.Desktop.AppName)
	str("DESKTOP_DATA_DIR", &cfg.Desktop.DataDir)
	boolf("DESKTOP_CLOSE_TO_TRAY", &cfg.Desktop.CloseToTray)
	boolf("DESKTOP_NATIVE_NOTIFICATIONS", &cfg.Desktop.NativeNotifications)
	boolf("DESKTOP_LAUNCH_AT_LOGIN", &cfg.Desktop.LaunchAtLogin)

	// Update environment variables
	boolf("UPDATE_ENABLED", &cfg.Update.Enabled)
	boolf("UPDATE_CHECK_ON_STARTUP", &cfg.Update.CheckOnStartup)
	intf("UPDATE_CHECK_INTERVAL_MIN", &cfg.Update.CheckIntervalMin)
	str("UPDATE_CHANNEL", &cfg.Update.Channel)
}

func applyFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("cloud-vault", flag.ContinueOnError)
	fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
	fs.StringVar(&cfg.App.ConfigFile, "config", cfg.App.ConfigFile, "path to config.toml file")
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

func resolveConfigFile(args []string, lookupEnv func(string) (string, bool), defaultPath string) string {
	if lookupEnv != nil {
		if envPath, ok := lookupEnv("APP_CONFIG_FILE"); ok && strings.TrimSpace(envPath) != "" {
			defaultPath = envPath
		}
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--config" || arg == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		}
	}
	return defaultPath
}

// parseInt is a simple strconv-free int parser.
func parseInt(s string) (int, error) {
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid int: %q", s)
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// parseInt64 is a simple strconv-free int64 parser.
func parseInt64(s string) (int64, error) {
	var n int64
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid int64: %q", s)
		}
		n = n*10 + int64(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// parseBool is a simple strconv-free bool parser.
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool: %q", s)
	}
}
