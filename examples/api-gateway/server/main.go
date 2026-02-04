package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/middleware/proxy"
	"github.com/spcent/plumego/net/discovery"
)

func main() {
	// Create service discovery
	// In this example, we use static configuration
	// For production, you would use Consul, Kubernetes, etc.
	sd := discovery.NewStatic(map[string][]string{
		"user-service": {
			"http://localhost:8081",
			"http://localhost:8082",
		},
		"order-service": {
			"http://localhost:9001",
			"http://localhost:9002",
		},
		"product-service": {
			"http://localhost:7001",
		},
	})

	// Create application
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithRecovery(),
		core.WithLogging(),
	)

	// Configure CORS for browser access
	app.Use("/api/*", corsMiddleware())

	// Proxy to user service
	app.Use("/api/v1/users/*", proxy.New(proxy.Config{
		Targets: []string{
			"http://localhost:8081",
			"http://localhost:8082",
		},
		LoadBalancer: proxy.NewRoundRobinBalancer(),
		PathRewrite:  proxy.StripPrefix("/api/v1"),
		ModifyRequest: proxy.ChainRequestModifiers(
			proxy.AddHeader("X-Gateway", "plumego"),
			proxy.AddHeader("X-Gateway-Version", "1.0"),
		),
		ModifyResponse: proxy.AddResponseHeader("X-Served-By", "API-Gateway"),
	}))

	// Proxy to order service with service discovery
	app.Use("/api/v1/orders/*", proxy.New(proxy.Config{
		ServiceName:  "order-service",
		Discovery:    sd,
		LoadBalancer: proxy.NewWeightedRoundRobinBalancer(),
		PathRewrite:  proxy.StripPrefix("/api/v1"),
		HealthCheck: &proxy.HealthCheckConfig{
			Interval:       10 * time.Second,
			Timeout:        5 * time.Second,
			Path:           "/health",
			ExpectedStatus: http.StatusOK,
			OnHealthChange: func(backend *proxy.Backend, healthy bool) {
				status := "healthy"
				if !healthy {
					status = "unhealthy"
				}
				log.Printf("Backend %s is now %s", backend.URL, status)
			},
		},
	}))

	// Proxy to product service with custom error handling
	app.Use("/api/v1/products/*", proxy.New(proxy.Config{
		Targets: []string{
			"http://localhost:7001",
		},
		PathRewrite: proxy.StripPrefix("/api/v1"),
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

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down gateway...")

		// Close service discovery
		if err := sd.Close(); err != nil {
			log.Printf("Error closing service discovery: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.Shutdown(ctx); err != nil {
			log.Printf("Gateway shutdown error: %v", err)
		}
	}()

	// Start server
	log.Println("Starting API Gateway on :8080")
	log.Println("Visit http://localhost:8080/status for info")

	if err := app.Boot(); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
