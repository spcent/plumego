package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/alerting"
	"guardus/internal/alerting/provider/custom"
	"guardus/internal/alerting/provider/discord"
	"guardus/internal/alerting/provider/email"
	"guardus/internal/alerting/provider/slack"
	"guardus/internal/alerting/provider/telegram"
	"guardus/internal/domain/connectivity"
	"guardus/internal/domain/maintenance"
	"guardus/internal/storage"
)

const envPrefix = "GUARDUS_"

const (
	envMetrics     = envPrefix + "METRICS"
	envConcurrency = envPrefix + "CONCURRENCY"

	envWebAddress        = envPrefix + "WEB_ADDRESS"
	envWebPort           = envPrefix + "WEB_PORT"
	envWebReadBufferSize = envPrefix + "WEB_READ_BUFFER_SIZE"
	envWebTLSCertFile    = envPrefix + "WEB_TLS_CERT_FILE"
	envWebTLSKeyFile     = envPrefix + "WEB_TLS_KEY_FILE"

	envStorageType       = envPrefix + "STORAGE_TYPE"
	envStoragePath       = envPrefix + "STORAGE_PATH"
	envStorageCaching    = envPrefix + "STORAGE_CACHING"
	envStorageMaxResults = envPrefix + "STORAGE_MAX_RESULTS"
	envStorageMaxEvents  = envPrefix + "STORAGE_MAX_EVENTS"

	envBootstrapFile = envPrefix + "BOOTSTRAP_FILE"

	envSecurityBasicUsername       = envPrefix + "SECURITY_BASIC_USERNAME"
	envSecurityBasicPasswordBcrypt = envPrefix + "SECURITY_BASIC_PASSWORD_BCRYPT_BASE64"
	envSecurityBasicPasswordPBKDF2 = envPrefix + "SECURITY_BASIC_PASSWORD_PBKDF2"

	envUITitle               = envPrefix + "UI_TITLE"
	envUIDescription         = envPrefix + "UI_DESCRIPTION"
	envUIHeader              = envPrefix + "UI_HEADER"
	envUIDashboardHeading    = envPrefix + "UI_DASHBOARD_HEADING"
	envUIDashboardSubheading = envPrefix + "UI_DASHBOARD_SUBHEADING"
	envUILogo                = envPrefix + "UI_LOGO"
	envUILink                = envPrefix + "UI_LINK"
	envUIDarkMode            = envPrefix + "UI_DARK_MODE"
	envUIDefaultSortBy       = envPrefix + "UI_DEFAULT_SORT_BY"
	envUIDefaultFilterBy     = envPrefix + "UI_DEFAULT_FILTER_BY"
	envUILoginSubtitle       = envPrefix + "UI_LOGIN_SUBTITLE"
	envUIFavicon             = envPrefix + "UI_FAVICON"
	envUIFavicon16           = envPrefix + "UI_FAVICON_16"
	envUIFavicon32           = envPrefix + "UI_FAVICON_32"

	envAlertingSlack    = envPrefix + "ALERTING_SLACK"
	envAlertingDiscord  = envPrefix + "ALERTING_DISCORD"
	envAlertingTelegram = envPrefix + "ALERTING_TELEGRAM"
	envAlertingEmail    = envPrefix + "ALERTING_EMAIL"
	envAlertingCustom   = envPrefix + "ALERTING_CUSTOM"

	envConnectivityTarget   = envPrefix + "CONNECTIVITY_TARGET"
	envConnectivityInterval = envPrefix + "CONNECTIVITY_INTERVAL"

	envMaintenanceEnabled  = envPrefix + "MAINTENANCE_ENABLED"
	envMaintenanceStart    = envPrefix + "MAINTENANCE_START"
	envMaintenanceDuration = envPrefix + "MAINTENANCE_DURATION"
	envMaintenanceTimezone = envPrefix + "MAINTENANCE_TIMEZONE"
	envMaintenanceEvery    = envPrefix + "MAINTENANCE_EVERY"
)

// LoadFromEnv builds a Config from process environment variables and runs the
// validation chain on every section. Endpoints are NOT populated here — they
// come from the storage layer (see storage.Store.ListEndpointConfigs) and any
// optional bootstrap JSON file referenced by GUARDUS_BOOTSTRAP_FILE. logger
// must not be nil.
func LoadFromEnv(logger plumelog.StructuredLogger) (*Config, error) {
	cfg := &Config{
		Metrics:     envBool(envMetrics, false),
		Concurrency: envInt(envConcurrency, DefaultConcurrency),
		Version:     "dev",
	}

	web, err := webFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.Web = web

	store, err := storageFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.Storage = store

	cfg.Security = securityFromEnv()
	cfg.UI = uiFromEnv()

	maint, err := maintenanceFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.Maintenance = maint

	conn, err := connectivityFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.Connectivity = conn

	alertCfg, err := alertingFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.Alerting = alertCfg

	// Endpoints are empty until the storage layer (or a future admin API)
	// populates them; alerting/endpoint validation is run after that step.
	if err := ValidateSecurityConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateWebConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateUIConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateMaintenanceConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateStorageConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateConnectivityConfig(cfg); err != nil {
		return nil, err
	}
	ValidateAndSetConcurrencyDefaults(cfg)
	if cfg.UI != nil && cfg.Storage != nil {
		cfg.UI.MaximumNumberOfResults = cfg.Storage.MaximumNumberOfResults
	}
	cfg.applyToCore()
	return cfg, nil
}

// FinalizeWithEndpoints runs the endpoint-aware validation chain after the
// caller has populated cfg.Endpoints / cfg.ExternalEndpoints from the store.
// Separated from LoadFromEnv because endpoints come from a different source
// (DB) than deployment config (env).
func FinalizeWithEndpoints(cfg *Config, logger plumelog.StructuredLogger) error {
	ValidateAlertingConfig(cfg.Alerting, cfg.Endpoints, cfg.ExternalEndpoints, logger)
	if err := ValidateEndpointsConfig(cfg, logger); err != nil {
		return err
	}
	if err := ValidateUniqueKeys(cfg); err != nil {
		return err
	}
	return nil
}

func webFromEnv() (*WebConfig, error) {
	w := defaultWebConfig()
	w.Address = envString(envWebAddress, w.Address)
	w.Port = envInt(envWebPort, w.Port)
	w.ReadBufferSize = envInt(envWebReadBufferSize, w.ReadBufferSize)
	cert := os.Getenv(envWebTLSCertFile)
	key := os.Getenv(envWebTLSKeyFile)
	if cert != "" || key != "" {
		w.TLS = &TLSConfig{CertificateFile: cert, PrivateKeyFile: key}
	}
	return w, nil
}

func storageFromEnv() (*storage.Config, error) {
	s := &storage.Config{
		Type:                   storage.Type(envString(envStorageType, string(storage.TypeMemory))),
		Path:                   os.Getenv(envStoragePath),
		Caching:                envBool(envStorageCaching, false),
		MaximumNumberOfResults: envInt(envStorageMaxResults, 0),
		MaximumNumberOfEvents:  envInt(envStorageMaxEvents, 0),
	}
	return s, nil
}

func securityFromEnv() *SecurityConfig {
	user := os.Getenv(envSecurityBasicUsername)
	bcrypt := os.Getenv(envSecurityBasicPasswordBcrypt)
	pbkdf2 := os.Getenv(envSecurityBasicPasswordPBKDF2)
	if user == "" && bcrypt == "" && pbkdf2 == "" {
		return nil
	}
	return &SecurityConfig{
		Basic: &BasicConfig{
			Username:                        user,
			PasswordBcryptHashBase64Encoded: bcrypt,
			PasswordPBKDF2Hash:              pbkdf2,
		},
	}
}

func uiFromEnv() *UIConfig {
	u := defaultUIConfig()
	u.Title = envString(envUITitle, u.Title)
	u.Description = envString(envUIDescription, u.Description)
	u.Header = envString(envUIHeader, u.Header)
	u.DashboardHeading = envString(envUIDashboardHeading, u.DashboardHeading)
	u.DashboardSubheading = envString(envUIDashboardSubheading, u.DashboardSubheading)
	u.Logo = envString(envUILogo, u.Logo)
	u.Link = envString(envUILink, u.Link)
	u.DefaultSortBy = envString(envUIDefaultSortBy, u.DefaultSortBy)
	u.DefaultFilterBy = envString(envUIDefaultFilterBy, u.DefaultFilterBy)
	u.LoginSubtitle = envString(envUILoginSubtitle, u.LoginSubtitle)
	if v, ok := envBoolPtr(envUIDarkMode); ok {
		u.DarkMode = v
	}
	u.Favicon.Default = envString(envUIFavicon, u.Favicon.Default)
	u.Favicon.Size16x16 = envString(envUIFavicon16, u.Favicon.Size16x16)
	u.Favicon.Size32x32 = envString(envUIFavicon32, u.Favicon.Size32x32)
	return u
}

func maintenanceFromEnv() (*maintenance.Config, error) {
	enabledRaw := os.Getenv(envMaintenanceEnabled)
	start := os.Getenv(envMaintenanceStart)
	duration := os.Getenv(envMaintenanceDuration)
	tz := os.Getenv(envMaintenanceTimezone)
	every := os.Getenv(envMaintenanceEvery)
	if enabledRaw == "" && start == "" && duration == "" && tz == "" && every == "" {
		return maintenance.GetDefaultConfig(), nil
	}
	c := &maintenance.Config{
		Start:    start,
		Timezone: tz,
	}
	if enabledRaw != "" {
		b, err := strconv.ParseBool(enabledRaw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envMaintenanceEnabled, err)
		}
		c.Enabled = &b
	}
	if duration != "" {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envMaintenanceDuration, err)
		}
		c.Duration = d
	}
	if every != "" {
		for _, p := range strings.Split(every, ",") {
			if v := strings.TrimSpace(p); v != "" {
				c.Every = append(c.Every, v)
			}
		}
	}
	return c, nil
}

func connectivityFromEnv() (*connectivity.Config, error) {
	target := os.Getenv(envConnectivityTarget)
	intervalRaw := os.Getenv(envConnectivityInterval)
	if target == "" && intervalRaw == "" {
		return nil, nil
	}
	checker := &connectivity.Checker{Target: target}
	if intervalRaw != "" {
		d, err := time.ParseDuration(intervalRaw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envConnectivityInterval, err)
		}
		checker.Interval = d
	}
	return &connectivity.Config{Checker: checker}, nil
}

func alertingFromEnv() (*alerting.Config, error) {
	cfg := &alerting.Config{}
	any := false
	if raw := os.Getenv(envAlertingSlack); raw != "" {
		p := &slack.AlertProvider{}
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("%s: %w", envAlertingSlack, err)
		}
		cfg.Slack = p
		any = true
	}
	if raw := os.Getenv(envAlertingDiscord); raw != "" {
		p := &discord.AlertProvider{}
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("%s: %w", envAlertingDiscord, err)
		}
		cfg.Discord = p
		any = true
	}
	if raw := os.Getenv(envAlertingTelegram); raw != "" {
		p := &telegram.AlertProvider{}
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("%s: %w", envAlertingTelegram, err)
		}
		cfg.Telegram = p
		any = true
	}
	if raw := os.Getenv(envAlertingEmail); raw != "" {
		p := &email.AlertProvider{}
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("%s: %w", envAlertingEmail, err)
		}
		cfg.Email = p
		any = true
	}
	if raw := os.Getenv(envAlertingCustom); raw != "" {
		p := &custom.AlertProvider{}
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("%s: %w", envAlertingCustom, err)
		}
		cfg.Custom = p
		any = true
	}
	if !any {
		return nil, nil
	}
	return cfg, nil
}

// BootstrapFilePath returns the optional bootstrap JSON path or "".
func BootstrapFilePath() string { return os.Getenv(envBootstrapFile) }

func envString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// envBoolPtr returns (*bool, true) when the variable is set to a parseable
// bool, otherwise (nil, false). Used for tri-state UIConfig fields where
// "unset" must remain distinct from "false".
func envBoolPtr(key string) (*bool, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return nil, false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil, false
	}
	return &b, true
}

// errMissing is exported only so tests can identify a structured failure.
var errMissing = errors.New("required env var missing")
