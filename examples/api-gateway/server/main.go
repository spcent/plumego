package main

import (
	"context"
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/proxy"
	"github.com/spcent/plumego/net/discovery"
	"github.com/spcent/plumego/router"
)

func main() {
	// Load configuration from environment variables
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Starting API Gateway with configuration:")
	log.Printf("  Server: %s (Debug: %v)", cfg.Server.Addr, cfg.Server.Debug)
	log.Printf("  Metrics: %v (%s)", cfg.Metrics.Enabled, cfg.Metrics.Path)
	log.Printf("  Rate Limit: %v (%d req/s)", cfg.RateLimit.Enabled, cfg.RateLimit.RequestsPerSecond)
	log.Printf("  Timeouts: Gateway=%v, Service=%v", cfg.Timeouts.Gateway, cfg.Timeouts.Service)
	log.Printf("  CORS: %v", cfg.CORS.Enabled)
	log.Printf("  User Service: %v (%d targets, timeout=%v, retries=%d)",
		cfg.Services.UserService.Enabled, len(cfg.Services.UserService.Targets),
		cfg.Services.UserService.Timeout, cfg.Services.UserService.RetryCount)
	log.Printf("  Order Service: %v (%d targets, timeout=%v, retries=%d)",
		cfg.Services.OrderService.Enabled, len(cfg.Services.OrderService.Targets),
		cfg.Services.OrderService.Timeout, cfg.Services.OrderService.RetryCount)
	log.Printf("  Product Service: %v (%d targets, timeout=%v, retries=%d)",
		cfg.Services.ProductService.Enabled, len(cfg.Services.ProductService.Targets),
		cfg.Services.ProductService.Timeout, cfg.Services.ProductService.RetryCount)

	// Create Prometheus metrics collector
	var metricsCollector metrics.MetricsCollector
	if cfg.Metrics.Enabled {
		metricsCollector = metrics.NewPrometheusCollector(cfg.Metrics.Namespace)
		log.Printf("✓ Prometheus metrics enabled at %s", cfg.Metrics.Path)
	}

	// Create service discovery for services that need it
	sd := discovery.NewStatic(map[string][]string{
		"order-service": cfg.Services.OrderService.Targets,
	})

	// Create application options
	appOptions := []core.Option{
		core.WithAddr(cfg.Server.Addr),
		core.WithRecovery(),
		core.WithLogging(),
	}

	// Add debug option if enabled
	if cfg.Server.Debug {
		appOptions = append(appOptions, core.WithDebug())
	}

	app := core.New(appOptions...)

	// Get logger for access logging
	logger := plog.NewGLogger()

	// Configure middleware stack for /api routes
	apiGroup := app.Router().Group("/api")

	// Apply access logging middleware
	apiGroup.Use(AccessLogMiddleware(logger, metricsCollector))
	log.Printf("✓ Access logging enabled")

	// Apply rate limiting if enabled
	if cfg.RateLimit.Enabled {
		apiGroup.Use(RateLimitMiddleware(cfg.RateLimit))
		log.Printf("✓ Rate limiting enabled: %d req/s (burst: %d)",
			cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.BurstSize)
	}

	// Apply gateway timeout if configured
	if cfg.Timeouts.Gateway > 0 {
		apiGroup.Use(TimeoutMiddleware(cfg.Timeouts.Gateway))
		log.Printf("✓ Gateway timeout enabled: %v", cfg.Timeouts.Gateway)
	}

	// Apply CORS if enabled
	if cfg.CORS.Enabled {
		apiGroup.Use(corsMiddleware(cfg.CORS))
		log.Printf("✓ CORS enabled")
	}

	// Create /api/v1 group
	v1Group := apiGroup.Group("/v1")

	// Register enabled services
	registerServices(v1Group, cfg, sd, metricsCollector)

	// Metrics endpoint (Prometheus)
	if cfg.Metrics.Enabled {
		if promCollector, ok := metricsCollector.(*metrics.PrometheusCollector); ok {
			app.Handle(cfg.Metrics.Path, promCollector.Handler())
			log.Printf("✓ Metrics endpoint registered: %s", cfg.Metrics.Path)
		}
	}

	// Health check endpoint for the gateway itself
	app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "api-gateway"}`))
	})

	// Gateway status endpoint
	app.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
Plumego API Gateway
===================

Endpoints:
  - GET  /health                    Gateway health check
  - GET  /status                    This status page
  - *    /api/v1/users/*            User service (round-robin, 2 backends)
  - *    /api/v1/orders/*           Order service (weighted, service discovery)
  - *    /api/v1/products/*         Product service (single backend)

Load Balancing:
  - User service: Round-robin
  - Order service: Weighted round-robin
  - Product service: Single backend

Service Discovery:
  - Type: Static (in-memory)
  - Services: user-service, order-service, product-service

Features:
  - ✅ Reverse proxy
  - ✅ Load balancing
  - ✅ Health checking
  - ✅ Service discovery
  - ✅ Path rewriting
  - ✅ Request/response modification
  - ✅ Custom error handling
  - ✅ CORS support
`))
	})

	// Register shutdown hook to close service discovery
	app.OnShutdown(func(ctx context.Context) error {
		log.Println("Closing service discovery...")
		return sd.Close()
	})

	// Start server (includes built-in graceful shutdown)
	log.Printf("Starting API Gateway on %s", cfg.Server.Addr)
	log.Println("Visit http://localhost:8080/status for info")

	if err := app.Boot(); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

// registerServices registers all enabled services to the router group
func registerServices(v1Group *router.Router, cfg *Config, sd *discovery.Static, metricsCollector metrics.MetricsCollector) {
	// User Service
	if cfg.Services.UserService.Enabled {
		log.Printf("Registering User Service: %s -> %v (timeout=%v, retries=%d)",
			cfg.Services.UserService.PathPrefix,
			cfg.Services.UserService.Targets,
			cfg.Services.UserService.Timeout,
			cfg.Services.UserService.RetryCount)

		v1Group.Any("/users/*", proxy.New(proxy.Config{
			Targets:      cfg.Services.UserService.Targets,
			LoadBalancer: cfg.Services.UserService.GetLoadBalancer(),
			PathRewrite:  proxy.StripPrefix("/api/v1"),
			Timeout:      cfg.Services.UserService.GetTimeout(cfg.Timeouts.Service),
			RetryCount:   cfg.Services.UserService.GetRetryCount(),
			ModifyRequest: proxy.ChainRequestModifiers(
				proxy.AddHeader("X-Gateway", "plumego"),
				proxy.AddHeader("X-Gateway-Version", "1.0"),
			),
			ModifyResponse: proxy.AddResponseHeader("X-Served-By", "API-Gateway"),
			HealthCheck:    cfg.Services.UserService.GetHealthCheckConfig(),
		}))
	}

	// Order Service
	if cfg.Services.OrderService.Enabled {
		log.Printf("Registering Order Service: %s -> %v (timeout=%v, retries=%d, service discovery)",
			cfg.Services.OrderService.PathPrefix,
			cfg.Services.OrderService.Targets,
			cfg.Services.OrderService.Timeout,
			cfg.Services.OrderService.RetryCount)

		v1Group.Any("/orders/*", proxy.New(proxy.Config{
			ServiceName:  "order-service",
			Discovery:    sd,
			LoadBalancer: cfg.Services.OrderService.GetLoadBalancer(),
			PathRewrite:  proxy.StripPrefix("/api/v1"),
			Timeout:      cfg.Services.OrderService.GetTimeout(cfg.Timeouts.Service),
			RetryCount:   cfg.Services.OrderService.GetRetryCount(),
			HealthCheck:  cfg.Services.OrderService.GetHealthCheckConfig(),
		}))
	}

	// Product Service
	if cfg.Services.ProductService.Enabled {
		log.Printf("Registering Product Service: %s -> %v (timeout=%v, retries=%d)",
			cfg.Services.ProductService.PathPrefix,
			cfg.Services.ProductService.Targets,
			cfg.Services.ProductService.Timeout,
			cfg.Services.ProductService.RetryCount)

		v1Group.Any("/products/*", proxy.New(proxy.Config{
			Targets:      cfg.Services.ProductService.Targets,
			LoadBalancer: cfg.Services.ProductService.GetLoadBalancer(),
			PathRewrite:  proxy.StripPrefix("/api/v1"),
			Timeout:      cfg.Services.ProductService.GetTimeout(cfg.Timeouts.Service),
			RetryCount:   cfg.Services.ProductService.GetRetryCount(),
			HealthCheck:  cfg.Services.ProductService.GetHealthCheckConfig(),
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error: %v", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{
					"error": "Service temporarily unavailable",
					"message": "The product service is currently unavailable. Please try again later."
				}`))
			},
		}))
	}
}

// corsMiddleware adds CORS headers based on configuration
func corsMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Join allowed origins
			origins := "*"
			if len(cfg.AllowedOrigins) > 0 {
				origins = cfg.AllowedOrigins[0]
			}

			// Join allowed methods
			methods := "GET, POST, PUT, DELETE, OPTIONS"
			if len(cfg.AllowedMethods) > 0 {
				methods = joinStrings(cfg.AllowedMethods, ", ")
			}

			// Join allowed headers
			headers := "Content-Type, Authorization"
			if len(cfg.AllowedHeaders) > 0 {
				headers = joinStrings(cfg.AllowedHeaders, ", ")
			}

			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// joinStrings joins a slice of strings with a separator
func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}
