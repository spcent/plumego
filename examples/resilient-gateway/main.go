package main

import (
	"log"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/middleware/circuitbreaker"
	"github.com/spcent/plumego/middleware/proxy"
	"github.com/spcent/plumego/tenant"
)

func main() {
	// Create tenant config manager
	tenantConfig := tenant.NewInMemoryConfigManager()
	tenantConfig.CreateTenant("tenant-1", tenant.Config{
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
		core.WithRecovery(),
		core.WithLogging(),
	)

	// Add tenant middleware
	app.Use("/api/*", tenant.Middleware(tenant.MiddlewareConfig{
		HeaderName:   "X-Tenant-ID",
		AllowMissing: true,
		QuotaManager: quotaMgr,
	}))

	// Add circuit breaker for all API routes
	app.Use("/api/*", circuitbreaker.Middleware(circuitbreaker.Config{
		Name:             "api",
		FailureThreshold: 0.5,
		Timeout:          10 * time.Second,
		OnStateChange: func(from, to circuitbreaker.State) {
			log.Printf("Circuit breaker state changed: %s -> %s", from, to)
		},
	}))

	// Proxy with circuit breaker per backend
	app.Use("/api/v1/users/*", proxy.New(proxy.Config{
		Targets: []string{
			"http://localhost:8081",
			"http://localhost:8082",
		},
		LoadBalancer:          proxy.NewRoundRobinBalancer(),
		PathRewrite:           proxy.StripPrefix("/api/v1"),
		CircuitBreakerEnabled: true,
		CircuitBreakerConfig: &proxy.CircuitBreakerConfig{
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
