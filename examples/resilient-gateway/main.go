package main

import (
	"log"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
	cb "github.com/spcent/plumego/security/resilience/circuitbreaker"
	"github.com/spcent/plumego/x/gateway"
	"github.com/spcent/plumego/x/tenant/core"
	tenantquota "github.com/spcent/plumego/x/tenant/quota"
	tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
)

func main() {
	// Create tenant config manager
	tenantConfig := tenant.NewInMemoryConfigManager()
	tenantConfig.SetTenantConfig(tenant.Config{
		TenantID: "tenant-1",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 100,
		},
	})

	// Create quota manager
	quotaMgr := tenant.NewInMemoryQuotaManager(tenantConfig)

	// Create application
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithLogger(plog.NewGLogger()),
	)
	app.Use(requestid.Middleware())
	app.Use(mwtracing.Middleware(nil))
	app.Use(httpmetrics.Middleware(nil))
	app.Use(accesslog.Middleware(app.Logger()))
	app.Use(recovery.Recovery(app.Logger()))

	// Create /api route group for middleware
	apiGroup := app.Router().Group("/api")

	// Add tenant middleware to the API group using the canonical x/tenant path.
	// AllowMissing=true so requests without a tenant ID proceed unauthenticated.
	apiGroup.Use(tenantresolve.Middleware(tenantresolve.Options{
		HeaderName:   "X-Tenant-ID",
		AllowMissing: true,
	}))
	apiGroup.Use(tenantquota.Middleware(tenantquota.Options{
		Manager: quotaMgr,
	}))

	// Add circuit breaker for all API routes
	apiGroup.Use(cb.Middleware(cb.Config{
		Name:             "api",
		FailureThreshold: 0.5,
		Timeout:          10 * time.Second,
		OnStateChange: func(from, to cb.State) {
			log.Printf("Circuit breaker state changed: %s -> %s", from, to)
		},
	}))

	// Create /api/v1 group
	v1Group := apiGroup.Group("/v1")

	// Proxy with circuit breaker per backend - now using Any() since proxy is a handler
	v1Group.Any("/users/*", gateway.New(gateway.Config{
		Targets: []string{
			"http://localhost:8081",
			"http://localhost:8082",
		},
		LoadBalancer:          gateway.NewRoundRobinBalancer(),
		PathRewrite:           gateway.StripPrefix("/api/v1"),
		CircuitBreakerEnabled: true,
		CircuitBreakerConfig: &gateway.CircuitBreakerConfig{
			FailureThreshold: 0.5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
		},
	}))

	// Health endpoint
	app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	log.Println("Starting resilient gateway on :8080")
	log.Println("Features: Circuit Breaker + Multi-Tenant + Load Balancing")

	if err := app.Boot(); err != nil {
		log.Fatal(err)
	}
}
