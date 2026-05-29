package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidationError aggregates one or more field-level configuration problems.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return "config validation failed: " + strings.Join(e.Errors, "; ")
}

// add appends a message if cond is true.
func (e *ValidationError) add(cond bool, format string, args ...any) {
	if cond {
		e.Errors = append(e.Errors, fmt.Sprintf(format, args...))
	}
}

// ValidateConfig checks cfg for completeness and correctness according to V0.8 rules.
// It returns nil when cfg is acceptable. It may create missing directories as a side effect
// (database parent dir, local storage root) and returns an error if they are unwritable.
// Secrets are never included in error messages.
func ValidateConfig(cfg Config) error {
	v := &ValidationError{}

	validateServer(cfg, v)
	validateDatabase(cfg, v)
	validateStorage(cfg, v)
	validateAuth(cfg, v)
	validateAI(cfg, v)

	if len(v.Errors) > 0 {
		return v
	}
	return nil
}

func validateServer(cfg Config, v *ValidationError) {
	addr := strings.TrimSpace(cfg.Core.Addr)
	v.add(addr == "", "server.addr is required")
	if addr != "" {
		// Accept ":port" or "host:port".
		colon := strings.LastIndex(addr, ":")
		if colon >= 0 && colon < len(addr)-1 {
			portStr := addr[colon+1:]
			if port, err := strconv.Atoi(portStr); err == nil {
				v.add(port < 1 || port > 65535, "server.addr port %d out of range 1..65535", port)
			} else {
				v.add(true, "server.addr has non-numeric port %q", portStr)
			}
		}
	}
}

func validateDatabase(cfg Config, v *ValidationError) {
	path := strings.TrimSpace(cfg.DB.Path)
	v.add(path == "", "database.path is required")
	if path == "" {
		return
	}
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		v.add(true, "database.path parent directory not writable: %v", err)
		return
	}
	// Probe writability: try creating then removing a temp file.
	probe := filepath.Join(dir, ".cloud-vault-write-probe")
	f, err := os.Create(probe)
	if err != nil {
		v.add(true, "database.path parent directory not writable: %v", err)
		return
	}
	f.Close()
	_ = os.Remove(probe)
}

func validateStorage(cfg Config, v *ValidationError) {
	switch cfg.Storage.Provider {
	case "local":
		root := strings.TrimSpace(cfg.Local.Root)
		v.add(root == "", "local.root is required when storage.provider=local")
		if root != "" {
			if err := os.MkdirAll(root, 0o755); err != nil {
				v.add(true, "local.root not writable: %v", err)
				return
			}
			probe := filepath.Join(root, ".cloud-vault-write-probe")
			f, err := os.Create(probe)
			if err != nil {
				v.add(true, "local.root not writable: %v", err)
				return
			}
			f.Close()
			_ = os.Remove(probe)
		}
	case "qiniu":
		v.add(strings.TrimSpace(cfg.Qiniu.AccessKey) == "", "qiniu.access_key is required when storage.provider=qiniu")
		v.add(strings.TrimSpace(cfg.Qiniu.SecretKey) == "", "qiniu.secret_key is required when storage.provider=qiniu")
		v.add(strings.TrimSpace(cfg.Qiniu.Bucket) == "", "qiniu.bucket is required when storage.provider=qiniu")
		v.add(strings.TrimSpace(cfg.Qiniu.Domain) == "", "qiniu.domain is required when storage.provider=qiniu")
		region := strings.TrimSpace(cfg.Qiniu.Region)
		validRegions := map[string]bool{"z0": true, "z1": true, "z2": true, "na0": true, "as0": true}
		v.add(region == "" || !validRegions[region], "qiniu.region %q is invalid (expected one of z0, z1, z2, na0, as0)", region)
	default:
		v.add(true, "storage.provider %q is invalid (expected local or qiniu)", cfg.Storage.Provider)
	}
}

func validateAuth(cfg Config, v *ValidationError) {
	if !cfg.Auth.Enabled {
		return
	}
	v.add(strings.TrimSpace(cfg.Auth.CookieName) == "", "auth.cookie_name is required when auth.enabled=true")
	v.add(cfg.Auth.SessionTTLHours <= 0, "auth.session_ttl_hours must be > 0")
	v.add(cfg.Auth.PasswordMinLength < 8, "auth.password_min_length must be >= 8")
	v.add(cfg.Auth.MaxLoginFailures <= 0, "auth.max_login_failures must be > 0")
	v.add(cfg.Auth.LockoutMinutes <= 0, "auth.lockout_minutes must be > 0")
}

func validateAI(cfg Config, v *ValidationError) {
	if !cfg.AI.Enabled {
		return
	}
	switch cfg.AI.Provider {
	case "openai_compatible":
		v.add(strings.TrimSpace(cfg.AI.BaseURL) == "", "ai.base_url is required when ai.provider=openai_compatible")
		v.add(strings.TrimSpace(cfg.AI.APIKey) == "", "ai.api_key is required when ai.provider=openai_compatible")
		v.add(strings.TrimSpace(cfg.AI.Model) == "", "ai.model is required when ai.provider=openai_compatible")
	case "local_mock":
		// No external credentials required.
	case "":
		v.add(true, "ai.provider is required when ai.enabled=true")
	default:
		v.add(true, "ai.provider %q is invalid (expected local_mock or openai_compatible)", cfg.AI.Provider)
	}
}

// Warnings returns non-fatal advisories about cfg (e.g. secure_cookie=false in auth-enabled mode).
func Warnings(cfg Config) []string {
	var w []string
	if cfg.Auth.Enabled && !cfg.Auth.SecureCookie {
		w = append(w, "secure_cookie=false is only recommended for local development; set secure_cookie=true behind HTTPS")
	}
	return w
}
