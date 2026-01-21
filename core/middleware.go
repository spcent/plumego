package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
)

// Use adds middleware to the application's middleware chain.
func (a *App) Use(middlewares ...middleware.Middleware) error {
	if a.started {
		return contract.WrapError(
			fmt.Errorf("cannot add middleware after app has started"),
			"use_middleware",
			"core",
			nil,
		)
	}

	if a.middlewareReg == nil {
		a.middlewareReg = middleware.NewRegistry()
	}

	a.middlewareReg.Use(middlewares...)
	return nil
}

func (a *App) applyGuardrails() {
	if a.guardsApplied {
		return
	}

	var guards []middleware.Middleware

	if a.config.Debug {
		cfg := middleware.DefaultDebugErrorConfig()
		cfg.NotFoundHint = devToolsRoutesPath
		guards = append(guards, middleware.DebugErrors(cfg))
	}

	if a.config.EnableSecurityHeaders {
		guards = append(guards, middleware.SecurityHeaders(a.config.SecurityHeadersPolicy))
	}

	if a.config.EnableAbuseGuard {
		cfg := middleware.DefaultAbuseGuardConfig()
		if a.config.AbuseGuardConfig != nil {
			cfg = *a.config.AbuseGuardConfig
		}
		guards = append(guards, middleware.AbuseGuard(cfg))
	}

	if a.config.MaxBodyBytes > 0 {
		guards = append(guards, middleware.BodyLimit(a.config.MaxBodyBytes, a.logger))
	}

	if a.config.MaxConcurrency > 0 {
		guards = append(guards, middleware.ConcurrencyLimit(
			a.config.MaxConcurrency,
			a.config.QueueDepth,
			a.config.QueueTimeout,
			a.logger))
	}

	if len(guards) > 0 {
		// Hardening middleware should execute before user-specified middleware.
		a.middlewareReg.Prepend(guards...)
	}

	a.guardsApplied = true
}

// buildHandler builds the combined handler with current middleware stack.
func (a *App) buildHandler() {
	chain := middleware.NewChain(a.middlewareReg.Middlewares()...)
	a.handler = chain.Apply(a.router)
}

func (a *App) enableLogging() error {
	if a.loggingEnabled {
		return nil
	}

	if err := a.Use(a.loggingMiddleware()); err != nil {
		return err
	}

	a.loggingEnabled = true
	return nil
}

func (a *App) loggingMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		var metricsCollector middleware.MetricsCollector
		if a.metricsCollector != nil {
			metricsCollector = &metricsAdapter{collector: a.metricsCollector}
		}
		return middleware.Logging(a.logger, metricsCollector, a.tracer)(next)
	}
}

// metricsAdapter adapts the unified MetricsCollector to the legacy interface
type metricsAdapter struct {
	collector metrics.MetricsCollector
}

func (m *metricsAdapter) Observe(ctx context.Context, metrics middleware.RequestMetrics) {
	if m.collector != nil {
		m.collector.ObserveHTTP(ctx, metrics.Method, metrics.Path, metrics.Status, metrics.Bytes, metrics.Duration)
	}
}

// EnableAuth enables the auth middleware.
func (a *App) EnableAuth() {
	a.Use(middleware.FromFuncMiddleware(middleware.Auth))
}

// EnableRateLimit enables the rate limiting middleware with the given configuration.
// rate: requests per second.
// capacity: maximum burst size.
func (a *App) EnableRateLimit(rate float64, capacity int) {
	a.Use(middleware.RateLimit(rate, capacity, time.Minute, 5*time.Minute))
}

func (a *App) enableRecovery() error {
	if a.recoveryEnabled {
		return nil
	}

	if err := a.Use(middleware.RecoveryMiddleware); err != nil {
		return err
	}

	a.recoveryEnabled = true
	return nil
}

func (a *App) enableCORS(opts *middleware.CORSOptions) error {
	if opts != nil {
		cfg := *opts
		a.corsOptions = &cfg
	} else {
		a.corsOptions = nil
	}

	if a.corsEnabled {
		return nil
	}

	if err := a.Use(a.corsMiddleware()); err != nil {
		return err
	}

	a.corsEnabled = true
	return nil
}

func (a *App) corsMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		if a.corsOptions == nil {
			return middleware.CORS(next)
		}
		return middleware.FromFuncMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
			return middleware.CORSWithOptions(*a.corsOptions, next)
		})(next)
	}
}
