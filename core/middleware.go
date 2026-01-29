package core

import (
	"context"
	"net/http"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
)

// Use adds middleware to the application's middleware chain.
func (a *App) Use(middlewares ...middleware.Middleware) error {
	if err := a.ensureMutable("use_middleware", "add middleware"); err != nil {
		return err
	}

	reg := a.ensureMiddlewareRegistry()
	reg.Use(middlewares...)
	return nil
}

func (a *App) applyGuardrails() {
	a.mu.Lock()
	if a.guardsApplied {
		a.mu.Unlock()
		return
	}
	a.guardsApplied = true
	a.mu.Unlock()

	var guards []middleware.Middleware

	cfg := a.configSnapshot()
	a.mu.RLock()
	logger := a.logger
	requestIDEnabled := a.requestIDEnabled
	a.mu.RUnlock()

	if requestIDEnabled {
		guards = append(guards, middleware.RequestID())
	}

	if cfg.Debug {
		cfg := middleware.DefaultDebugErrorConfig()
		cfg.NotFoundHint = devToolsRoutesPath
		guards = append(guards, middleware.DebugErrors(cfg))
	}

	if cfg.EnableSecurityHeaders {
		guards = append(guards, middleware.SecurityHeaders(cfg.SecurityHeadersPolicy))
	}

	if cfg.EnableAbuseGuard {
		guardCfg := middleware.DefaultAbuseGuardConfig()
		if cfg.AbuseGuardConfig != nil {
			guardCfg = *cfg.AbuseGuardConfig
		}
		guards = append(guards, middleware.AbuseGuard(guardCfg))
	}

	if cfg.MaxBodyBytes > 0 {
		guards = append(guards, middleware.BodyLimit(cfg.MaxBodyBytes, logger))
	}

	if cfg.MaxConcurrency > 0 {
		guards = append(guards, middleware.ConcurrencyLimit(
			cfg.MaxConcurrency,
			cfg.QueueDepth,
			cfg.QueueTimeout,
			logger))
	}

	if len(guards) > 0 {
		// Hardening middleware should execute before user-specified middleware.
		reg := a.ensureMiddlewareRegistry()
		reg.Prepend(guards...)
	}
}

// buildHandler builds the combined handler with current middleware stack.
func (a *App) buildHandler() {
	reg := a.ensureMiddlewareRegistry()
	r := a.ensureRouter()
	chain := middleware.NewChain(reg.Middlewares()...)
	handler := chain.Apply(r)

	a.mu.Lock()
	a.handler = handler
	a.mu.Unlock()
}

func (a *App) enableLogging() error {
	a.mu.RLock()
	enabled := a.loggingEnabled
	a.mu.RUnlock()
	if enabled {
		return nil
	}

	if err := a.Use(a.loggingMiddleware()); err != nil {
		return err
	}

	a.mu.Lock()
	a.loggingEnabled = true
	a.mu.Unlock()
	return nil
}

func (a *App) enableRequestID() error {
	a.mu.RLock()
	enabled := a.requestIDEnabled
	a.mu.RUnlock()
	if enabled {
		return nil
	}
	if err := a.ensureMutable("enable_request_id", "enable request id"); err != nil {
		return err
	}

	a.mu.Lock()
	a.requestIDEnabled = true
	a.mu.Unlock()
	return nil
}

func (a *App) loggingMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		a.mu.RLock()
		logger := a.logger
		collector := a.metricsCollector
		tracer := a.tracer
		a.mu.RUnlock()

		var metricsCollector middleware.MetricsCollector
		if collector != nil {
			metricsCollector = &metricsAdapter{collector: collector}
		}
		return middleware.Logging(logger, metricsCollector, tracer)(next)
	}
}

// metricsAdapter adapts the unified MetricsCollector to the legacy interface
type metricsAdapter struct {
	collector metrics.MetricsCollector
}

func (m *metricsAdapter) Observe(ctx context.Context, metrics middleware.RequestMetrics) {
	if m.collector != nil {
		path := metrics.Path
		if metrics.Route != "" {
			path = metrics.Route
		}
		m.collector.ObserveHTTP(ctx, metrics.Method, path, metrics.Status, metrics.Bytes, metrics.Duration)
	}
}

// EnableAuth enables the auth middleware.
func (a *App) EnableAuth() {
	if err := a.Use(middleware.FromFuncMiddleware(middleware.Auth)); err != nil {
		a.logError("EnableAuth failed", err, nil)
	}
}

// EnableRateLimit enables the rate limiting middleware with the given configuration.
// maxConcurrent: maximum concurrent requests.
// queueDepth: maximum queue depth for waiting requests.
func (a *App) EnableRateLimit(maxConcurrent int64, queueDepth int64) {
	config := middleware.RateLimiterConfig{
		MaxConcurrent: maxConcurrent,
		QueueDepth:    queueDepth,
	}
	if err := a.Use(middleware.RateLimitMiddleware(config)); err != nil {
		a.logError("EnableRateLimit failed", err, log.Fields{
			"maxConcurrent": maxConcurrent,
			"queueDepth":    queueDepth,
		})
	}
}

func (a *App) enableRecovery() error {
	a.mu.RLock()
	enabled := a.recoveryEnabled
	a.mu.RUnlock()
	if enabled {
		return nil
	}

	if err := a.Use(middleware.RecoveryMiddleware); err != nil {
		return err
	}

	a.mu.Lock()
	a.recoveryEnabled = true
	a.mu.Unlock()
	return nil
}

func (a *App) enableCORS(opts *middleware.CORSOptions) error {
	if err := a.ensureMutable("enable_cors", "enable CORS"); err != nil {
		return err
	}

	var cfg *middleware.CORSOptions
	if opts != nil {
		copy := *opts
		cfg = &copy
	}

	a.mu.Lock()
	a.corsOptions = cfg
	enabled := a.corsEnabled
	a.mu.Unlock()
	if enabled {
		return nil
	}

	if err := a.Use(a.corsMiddleware()); err != nil {
		return err
	}

	a.mu.Lock()
	a.corsEnabled = true
	a.mu.Unlock()
	return nil
}

func (a *App) corsMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		a.mu.RLock()
		opts := a.corsOptions
		a.mu.RUnlock()

		if opts == nil {
			return middleware.CORS(next)
		}
		return middleware.FromFuncMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
			return middleware.CORSWithOptions(*opts, next)
		})(next)
	}
}
