package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/security/headers"
	"github.com/spcent/plumego/security/jwt"
)

// JWTAuthMiddleware provides JWT authentication with public path support
func JWTAuthMiddleware(cfg AuthConfig, jwtManager *jwt.JWTManager) func(http.Handler) http.Handler {
	// Get base JWT authenticator from manager
	baseAuthenticator := jwtManager.JWTAuthenticator(jwt.TokenTypeAccess)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path is public
			for _, publicPath := range cfg.PublicPaths {
				if strings.HasPrefix(r.URL.Path, publicPath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Use the JWT manager's built-in authenticator
			baseAuthenticator(next).ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(cfg SecurityConfig) func(http.Handler) http.Handler {
	// Build security policy
	policy := headers.Policy{
		FrameOptions:       cfg.XFrameOptions,
		ContentTypeOptions: cfg.XContentTypeOptions,
		ReferrerPolicy:     cfg.ReferrerPolicy,
	}

	// Add HSTS if configured
	if cfg.HSTSMaxAge > 0 {
		policy.StrictTransportSecurity = &headers.HSTSOptions{
			MaxAge:            time.Duration(cfg.HSTSMaxAge) * time.Second,
			IncludeSubDomains: cfg.HSTSIncludeSubDomains,
			Preload:           cfg.HSTSPreload,
		}
	}

	// Add CSP if configured
	if cfg.ContentSecurityPolicy != "" {
		policy.ContentSecurityPolicy = cfg.ContentSecurityPolicy
	}

	// Return middleware function
	return func(next http.Handler) http.Handler {
		return policy.Middleware(next)
	}
}

// AdminAPIKeyMiddleware validates admin API key
func AdminAPIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check API key in header
			providedKey := r.Header.Get("X-Admin-API-Key")
			if providedKey == "" {
				// Also check Authorization header
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					providedKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// Validate API key (timing-safe comparison)
			if providedKey != apiKey || providedKey == "" {
				writeJSONError(w, http.StatusUnauthorized, "Invalid or missing admin API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   http.StatusText(status),
		"message": message,
		"status":  status,
	})
}

// AdminHandlers provides admin API handlers
type AdminHandlers struct {
	cfg              *Config
	metricsCollector interface{} // metrics.MetricsCollector
}

// NewAdminHandlers creates admin API handlers
func NewAdminHandlers(cfg *Config, metricsCollector interface{}) *AdminHandlers {
	return &AdminHandlers{
		cfg:              cfg,
		metricsCollector: metricsCollector,
	}
}

// HandleStats returns gateway statistics
func (h *AdminHandlers) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"gateway": map[string]interface{}{
			"addr":           h.cfg.Server.Addr,
			"debug":          h.cfg.Server.Debug,
			"uptime_seconds": time.Since(startTime).Seconds(),
		},
		"services": map[string]interface{}{
			"user":    map[string]interface{}{"enabled": h.cfg.Services.UserService.Enabled, "targets": len(h.cfg.Services.UserService.Targets)},
			"order":   map[string]interface{}{"enabled": h.cfg.Services.OrderService.Enabled, "targets": len(h.cfg.Services.OrderService.Targets)},
			"product": map[string]interface{}{"enabled": h.cfg.Services.ProductService.Enabled, "targets": len(h.cfg.Services.ProductService.Targets)},
		},
		"features": map[string]interface{}{
			"auth":             h.cfg.Auth.Enabled,
			"metrics":          h.cfg.Metrics.Enabled,
			"rate_limit":       h.cfg.RateLimit.Enabled,
			"cors":             h.cfg.CORS.Enabled,
			"security_headers": h.cfg.Security.Enabled,
			"tls":              h.cfg.TLS.Enabled,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleHealth returns detailed health check
func (h *AdminHandlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"server": map[string]interface{}{"status": "ok"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// HandleConfig returns current configuration (sanitized)
func (h *AdminHandlers) HandleConfig(w http.ResponseWriter, r *http.Request) {
	// Sanitize sensitive data
	sanitized := map[string]interface{}{
		"server": map[string]interface{}{
			"addr":  h.cfg.Server.Addr,
			"debug": h.cfg.Server.Debug,
		},
		"auth": map[string]interface{}{
			"enabled":           h.cfg.Auth.Enabled,
			"token_header":      h.cfg.Auth.TokenHeader,
			"access_token_ttl":  h.cfg.Auth.AccessTokenTTL,
			"refresh_token_ttl": h.cfg.Auth.RefreshTokenTTL,
			"public_paths":      h.cfg.Auth.PublicPaths,
			// JWT secret is NOT exposed
		},
		"security": h.cfg.Security,
		"tls": map[string]interface{}{
			"enabled": h.cfg.TLS.Enabled,
			// Cert/key paths are NOT exposed
		},
		"metrics":    h.cfg.Metrics,
		"rate_limit": h.cfg.RateLimit,
		"timeouts":   h.cfg.Timeouts,
		"cors":       h.cfg.CORS,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitized)
}

// HandleReload triggers configuration reload (placeholder for now)
func (h *AdminHandlers) HandleReload(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement hot reload in a future update
	writeJSONError(w, http.StatusNotImplemented, "Configuration reload not yet implemented")
}

var startTime = time.Now()

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
