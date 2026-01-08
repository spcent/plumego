package core

import (
	"fmt"
	"time"

	"github.com/spcent/plumego/contract"
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

// EnableLogging enables the logging middleware.
func (a *App) EnableLogging() {
	_ = a.enableLogging()
}

func (a *App) enableLogging() error {
	if a.loggingEnabled {
		return nil
	}

	if err := a.Use(middleware.Logging(a.logger, a.metricsCollector, a.tracer)); err != nil {
		return err
	}

	a.loggingEnabled = true
	return nil
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

// EnableCORS enables the CORS middleware.
func (a *App) EnableCORS() {
	a.Use(middleware.CORS)
}

// EnableRecovery enables the recovery middleware.
func (a *App) EnableRecovery() {
	a.Use(middleware.RecoveryMiddleware)
}
