package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/middleware/proxy"
)

// Config holds the API gateway configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Services ServicesConfig `json:"services"`
	CORS     CORSConfig     `json:"cors"`
}

// ServerConfig holds server-level configuration
type ServerConfig struct {
	Addr  string `json:"addr"`
	Debug bool   `json:"debug"`
}

// ServicesConfig holds all service configurations
type ServicesConfig struct {
	UserService    ServiceConfig `json:"user_service"`
	OrderService   ServiceConfig `json:"order_service"`
	ProductService ServiceConfig `json:"product_service"`
}

// ServiceConfig holds configuration for a single backend service
type ServiceConfig struct {
	Enabled      bool     `json:"enabled"`
	Targets      []string `json:"targets"`
	PathPrefix   string   `json:"path_prefix"`
	LoadBalancer string   `json:"load_balancer"`
	HealthCheck  bool     `json:"health_check"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Addr:  getEnv("GATEWAY_ADDR", ":8080"),
			Debug: getEnvBool("GATEWAY_DEBUG", false),
		},
		CORS: CORSConfig{
			Enabled:        getEnvBool("CORS_ENABLED", true),
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
		Services: ServicesConfig{
			UserService: ServiceConfig{
				Enabled:      getEnvBool("USER_SERVICE_ENABLED", true),
				Targets:      getEnvSlice("USER_SERVICE_TARGETS", []string{"http://localhost:8081", "http://localhost:8082"}),
				PathPrefix:   getEnv("USER_SERVICE_PATH_PREFIX", "/api/v1/users"),
				LoadBalancer: getEnv("USER_SERVICE_LB", "round-robin"),
				HealthCheck:  getEnvBool("USER_SERVICE_HEALTH_CHECK", false),
			},
			OrderService: ServiceConfig{
				Enabled:      getEnvBool("ORDER_SERVICE_ENABLED", true),
				Targets:      getEnvSlice("ORDER_SERVICE_TARGETS", []string{"http://localhost:9001", "http://localhost:9002"}),
				PathPrefix:   getEnv("ORDER_SERVICE_PATH_PREFIX", "/api/v1/orders"),
				LoadBalancer: getEnv("ORDER_SERVICE_LB", "weighted-round-robin"),
				HealthCheck:  getEnvBool("ORDER_SERVICE_HEALTH_CHECK", true),
			},
			ProductService: ServiceConfig{
				Enabled:      getEnvBool("PRODUCT_SERVICE_ENABLED", true),
				Targets:      getEnvSlice("PRODUCT_SERVICE_TARGETS", []string{"http://localhost:7001"}),
				PathPrefix:   getEnv("PRODUCT_SERVICE_PATH_PREFIX", "/api/v1/products"),
				LoadBalancer: getEnv("PRODUCT_SERVICE_LB", "round-robin"),
				HealthCheck:  getEnvBool("PRODUCT_SERVICE_HEALTH_CHECK", false),
			},
		},
	}

	return cfg, nil
}

// GetLoadBalancer returns a load balancer instance based on config
func (sc *ServiceConfig) GetLoadBalancer() proxy.LoadBalancer {
	switch sc.LoadBalancer {
	case "weighted-round-robin":
		return proxy.NewWeightedRoundRobinBalancer()
	case "least-connections":
		return proxy.NewLeastConnectionsBalancer()
	case "ip-hash":
		return proxy.NewIPHashBalancer()
	default:
		return proxy.NewRoundRobinBalancer()
	}
}

// GetHealthCheckConfig returns health check configuration
func (sc *ServiceConfig) GetHealthCheckConfig() *proxy.HealthCheckConfig {
	if !sc.HealthCheck {
		return nil
	}

	return &proxy.HealthCheckConfig{
		Interval:       10 * time.Second,
		Timeout:        5 * time.Second,
		Path:           "/health",
		ExpectedStatus: 200,
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	// Validate at least one service is enabled
	hasService := false
	if c.Services.UserService.Enabled && len(c.Services.UserService.Targets) > 0 {
		hasService = true
	}
	if c.Services.OrderService.Enabled && len(c.Services.OrderService.Targets) > 0 {
		hasService = true
	}
	if c.Services.ProductService.Enabled && len(c.Services.ProductService.Targets) > 0 {
		hasService = true
	}

	if !hasService {
		return fmt.Errorf("at least one service must be enabled with valid targets")
	}

	return nil
}
