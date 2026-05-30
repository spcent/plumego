package diagnostics

import (
	"strings"

	"cloud-vault/internal/config"
)

// sensitiveKeys lists configuration keys that should be redacted in diagnostics.
// A key matches if it contains any of these substrings (case-insensitive).
var sensitiveKeys = []string{
	"secret_key",
	"api_key",
	"apikey",
	"password",
	"token",
	"session",
	"cookie",
	"authorization",
	"private_key",
}

const redactedValue = "<redacted>"

// RedactConfig returns a copy of the configuration with sensitive fields replaced
// by "<redacted>". The input is not modified.
func RedactConfig(cfg config.Config) RedactedConfig {
	return RedactedConfig{
		Server: redactMap(map[string]interface{}{
			"addr":        cfg.Core.Addr,
			"tls_enabled": cfg.Core.TLS.Enabled,
			"tls_cert":    redactIfSensitive("tls_cert_file", cfg.Core.TLS.CertFile),
			"tls_key":     redactIfSensitive("tls_key_file", cfg.Core.TLS.KeyFile),
		}),
		App: redactMap(map[string]interface{}{
			"config_file":        cfg.App.ConfigFile,
			"version":            cfg.App.Version,
			"max_upload_size_mb": cfg.App.MaxUploadSizeMB,
			"version_policy":     cfg.App.VersionPolicy,
		}),
		Database: redactMap(map[string]interface{}{
			"path": cfg.DB.Path,
		}),
		Storage: redactMap(map[string]interface{}{
			"provider":       cfg.Storage.Provider,
			"local_root":     cfg.Local.Root,
			"qiniu_access_key": redactAlways(cfg.Qiniu.AccessKey),
			"qiniu_secret_key": redactAlways(cfg.Qiniu.SecretKey),
			"qiniu_bucket":     cfg.Qiniu.Bucket,
			"qiniu_domain":     cfg.Qiniu.Domain,
			"qiniu_region":     cfg.Qiniu.Region,
			"qiniu_use_https":  cfg.Qiniu.UseHTTPS,
		}),
		Auth: redactMap(map[string]interface{}{
			"enabled":                   cfg.Auth.Enabled,
			"session_ttl_hours":         cfg.Auth.SessionTTLHours,
			"cookie_name":               redactIfSensitive("cookie_name", cfg.Auth.CookieName),
			"secure_cookie":             cfg.Auth.SecureCookie,
			"max_login_failures":        cfg.Auth.MaxLoginFailures,
			"login_failure_window_min":  cfg.Auth.LoginFailureWindowMinutes,
			"lockout_minutes":           cfg.Auth.LockoutMinutes,
			"password_min_length":       cfg.Auth.PasswordMinLength,
			"bootstrap_admin_enabled":   cfg.Auth.BootstrapAdminEnabled,
			"bootstrap_admin_username":  cfg.Auth.BootstrapAdminUsername,
			"bootstrap_admin_email":     cfg.Auth.BootstrapAdminEmail,
			"bootstrap_admin_password":  redactAlways(cfg.Auth.BootstrapAdminPassword),
		}),
		AI: redactMap(map[string]interface{}{
			"enabled":              cfg.AI.Enabled,
			"provider":             cfg.AI.Provider,
			"base_url":             cfg.AI.BaseURL,
			"api_key":              redactAlways(cfg.AI.APIKey),
			"model":                cfg.AI.Model,
			"max_context_tokens":   cfg.AI.MaxContextTokens,
			"max_retries":          cfg.AI.MaxRetries,
			"task_workers":         cfg.AI.TaskWorkers,
			"summary_enabled":      cfg.AI.SummaryEnabled,
			"qa_enabled":           cfg.AI.QAEnabled,
			"prompt_extract_enabled": cfg.AI.PromptExtractEnabled,
		}),
		Update: redactMap(map[string]interface{}{
			"enabled":              cfg.Update.Enabled,
			"check_on_startup":     cfg.Update.CheckOnStartup,
			"check_interval_min":   cfg.Update.CheckIntervalMin,
			"channel":              cfg.Update.Channel,
		}),
	}
}

// RedactLogLine redacts sensitive tokens from a single log line.
// It replaces patterns like `key=value`, `key="value"`, `Bearer xxx` with
// redacted placeholders.
func RedactLogLine(line string) string {
	result := line

	// Bearer tokens
	result = redactPattern(result, `Bearer\s+[A-Za-z0-9\-._~+/]+=*`, "Bearer <redacted>")
	result = redactPattern(result, `bearer\s+[A-Za-z0-9\-._~+/]+=*`, "bearer <redacted>")

	// Key=value patterns for sensitive keys
	for _, key := range sensitiveKeys {
		// key=value (until whitespace or end)
		result = redactKeyValue(result, key)
		// key="value" (quoted)
		result = redactKeyValueQuoted(result, key)
		// key: value (JSON-like)
		result = redactKeyColon(result, key)
	}

	return result
}

func redactAlways(value string) string {
	if value == "" {
		return ""
	}
	return redactedValue
}

func redactIfSensitive(key, value string) string {
	lk := strings.ToLower(key)
	for _, s := range sensitiveKeys {
		if strings.Contains(lk, s) {
			if value == "" {
				return ""
			}
			return redactedValue
		}
	}
	return value
}

func redactMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func redactPattern(line, pattern, replacement string) string {
	re, err := compileRegexp(pattern)
	if err != nil {
		return line
	}
	return re.ReplaceAllString(line, replacement)
}

func redactKeyValue(line, key string) string {
	// Match: key=value (value is non-whitespace)
	pattern := `(?i)` + regexpQuote(key) + `=[^\s"'` + "`" + `]+`
	re, err := compileRegexp(pattern)
	if err != nil {
		return line
	}
	return re.ReplaceAllStringFunc(line, func(match string) string {
		eqIdx := strings.Index(match, "=")
		if eqIdx < 0 {
			return match
		}
		return match[:eqIdx+1] + "<redacted>"
	})
}

func redactKeyValueQuoted(line, key string) string {
	// Match: key="..." or key='...'
	patterns := []string{
		`(?i)` + regexpQuote(key) + `="[^"]*"`,
		`(?i)` + regexpQuote(key) + `='[^']*'`,
	}
	for _, pattern := range patterns {
		re, err := compileRegexp(pattern)
		if err != nil {
			continue
		}
		line = re.ReplaceAllStringFunc(line, func(match string) string {
			eqIdx := strings.Index(match, "=")
			if eqIdx < 0 {
				return match
			}
			return match[:eqIdx+1] + `"<redacted>"`
		})
	}
	return line
}

func redactKeyColon(line, key string) string {
	// Match: key: value (until end or comma)
	pattern := `(?i)"?` + regexpQuote(key) + `"?\s*:\s*"[^"]*"`
	re, err := compileRegexp(pattern)
	if err != nil {
		return line
	}
	return re.ReplaceAllStringFunc(line, func(match string) string {
		colIdx := strings.Index(match, ":")
		if colIdx < 0 {
			return match
		}
		return match[:colIdx+1] + ` "<redacted>"`
	})
}
