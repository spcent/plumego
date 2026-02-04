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
	Server    ServerConfig    `json:"server"`
	Services  ServicesConfig  `json:"services"`
	CORS      CORSConfig      `json:"cors"`
	Metrics   MetricsConfig   `json:"metrics"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Timeouts  TimeoutsConfig  `json:"timeouts"`
}

// ServerConfig holds server-level configuration
type ServerConfig struct {
	Addr  string `json:"addr"`
	Debug bool   `json:"debug"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled   bool   `json:"enabled"`
	Path      string `json:"path"`
	Namespace string `json:"namespace"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerSecond int  `json:"requests_per_second"`
	BurstSize         int  `json:"burst_size"`
}

// TimeoutsConfig holds timeout configuration
type TimeoutsConfig struct {
	Gateway time.Duration `json:"gateway"` // Gateway-level timeout
	Service time.Duration `json:"service"` // Default service timeout
}

// ServicesConfig holds all service configurations
type ServicesConfig struct {
	UserService    ServiceConfig `json:"user_service"`
	OrderService   ServiceConfig `json:"order_service"`
	ProductService ServiceConfig `json:"product_service"`
}

// ServiceConfig holds configuration for a single backend service
type ServiceConfig struct {
	Enabled      bool          `json:"enabled"`
	Targets      []string      `json:"targets"`
	PathPrefix   string        `json:"path_prefix"`
	LoadBalancer string        `json:"load_balancer"`
	HealthCheck  bool          `json:"health_check"`
	Timeout      time.Duration `json:"timeout"` // Service-specific timeout
	RetryCount   int           `json:"retry_count"`
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
		Metrics: MetricsConfig{
			Enabled:   getEnvBool("METRICS_ENABLED", true),
			Path:      getEnv("METRICS_PATH", "/metrics"),
			Namespace: getEnv("METRICS_NAMESPACE", "api_gateway"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: getEnvInt("RATE_LIMIT_RPS", 1000),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 2000),
		},
		Timeouts: TimeoutsConfig{
			Gateway: getEnvDuration("TIMEOUT_GATEWAY", 30*time.Second),
			Service: getEnvDuration("TIMEOUT_SERVICE", 10*time.Second),
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
				Timeout:      getEnvDuration("USER_SERVICE_TIMEOUT", 10*time.Second),
				RetryCount:   getEnvInt("USER_SERVICE_RETRY", 3),
			},
			OrderService: ServiceConfig{
				Enabled:      getEnvBool("ORDER_SERVICE_ENABLED", true),
				Targets:      getEnvSlice("ORDER_SERVICE_TARGETS", []string{"http://localhost:9001", "http://localhost:9002"}),
				PathPrefix:   getEnv("ORDER_SERVICE_PATH_PREFIX", "/api/v1/orders"),
				LoadBalancer: getEnv("ORDER_SERVICE_LB", "weighted-round-robin"),
				HealthCheck:  getEnvBool("ORDER_SERVICE_HEALTH_CHECK", true),
				Timeout:      getEnvDuration("ORDER_SERVICE_TIMEOUT", 15*time.Second),
				RetryCount:   getEnvInt("ORDER_SERVICE_RETRY", 2),
			},
			ProductService: ServiceConfig{
				Enabled:      getEnvBool("PRODUCT_SERVICE_ENABLED", true),
				Targets:      getEnvSlice("PRODUCT_SERVICE_TARGETS", []string{"http://localhost:7001"}),
				PathPrefix:   getEnv("PRODUCT_SERVICE_PATH_PREFIX", "/api/v1/products"),
				LoadBalancer: getEnv("PRODUCT_SERVICE_LB", "round-robin"),
				HealthCheck:  getEnvBool("PRODUCT_SERVICE_HEALTH_CHECK", false),
				Timeout:      getEnvDuration("PRODUCT_SERVICE_TIMEOUT", 10*time.Second),
				RetryCount:   getEnvInt("PRODUCT_SERVICE_RETRY", 3),
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

// GetTimeout returns the service-specific timeout or default
func (sc *ServiceConfig) GetTimeout(defaultTimeout time.Duration) time.Duration {
	if sc.Timeout > 0 {
		return sc.Timeout
	}
	return defaultTimeout
}

// GetRetryCount returns the service-specific retry count
func (sc *ServiceConfig) GetRetryCount() int {
	if sc.RetryCount > 0 {
		return sc.RetryCount
	}
	return 0 // No retries by default
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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return i
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Support both seconds and duration strings
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Try parsing as seconds
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Addr == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	// Metrics validation
	if c.Metrics.Enabled {
		if c.Metrics.Path == "" {
			return fmt.Errorf("metrics path cannot be empty when metrics enabled")
		}
		if c.Metrics.Namespace == "" {
			return fmt.Errorf("metrics namespace cannot be empty when metrics enabled")
		}
	}

	// Rate limit validation
	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerSecond <= 0 {
			return fmt.Errorf("rate limit requests per second must be positive")
		}
		if c.RateLimit.BurstSize <= 0 {
			return fmt.Errorf("rate limit burst size must be positive")
		}
		if c.RateLimit.BurstSize < c.RateLimit.RequestsPerSecond {
			return fmt.Errorf("rate limit burst size should be >= requests per second")
		}
	}

	// Timeout validation
	if c.Timeouts.Gateway <= 0 {
		return fmt.Errorf("gateway timeout must be positive")
	}
	if c.Timeouts.Service <= 0 {
		return fmt.Errorf("service timeout must be positive")
	}
	if c.Timeouts.Service > c.Timeouts.Gateway {
		return fmt.Errorf("service timeout should not exceed gateway timeout")
	}

	// Validate at least one service is enabled
	hasService := false
	services := []struct {
		name   string
		config ServiceConfig
	}{
		{"user", c.Services.UserService},
		{"order", c.Services.OrderService},
		{"product", c.Services.ProductService},
	}

	for _, svc := range services {
		if svc.config.Enabled {
			if len(svc.config.Targets) == 0 {
				return fmt.Errorf("%s service is enabled but has no targets", svc.name)
			}
			// Validate each target URL
			for _, target := range svc.config.Targets {
				if target == "" {
					return fmt.Errorf("%s service has empty target URL", svc.name)
				}
				if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
					return fmt.Errorf("%s service target must start with http:// or https://: %s", svc.name, target)
				}
			}
			// Validate timeout
			if svc.config.Timeout <= 0 {
				return fmt.Errorf("%s service timeout must be positive", svc.name)
			}
			if svc.config.Timeout > c.Timeouts.Gateway {
				return fmt.Errorf("%s service timeout exceeds gateway timeout", svc.name)
			}
			// Validate retry count
			if svc.config.RetryCount < 0 {
				return fmt.Errorf("%s service retry count cannot be negative", svc.name)
			}
			hasService = true
		}
	}

	if !hasService {
		return fmt.Errorf("at least one service must be enabled with valid targets")
	}

	return nil
}
