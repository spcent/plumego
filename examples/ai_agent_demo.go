package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/router"
)

// DemoApp demonstrates AI Agent friendly features
type DemoApp struct {
	container *core.DIContainer
	dashboard *metrics.Dashboard
}

// NewDemoApp creates a new demo application
func NewDemoApp() *DemoApp {
	return &DemoApp{
		container: core.NewDIContainer(),
		dashboard: metrics.NewDashboard(metrics.DashboardConfig{
			UpdateInterval:   1 * time.Second,
			MaxHistoryPoints: 100,
			EnableAlerts:     true,
		}),
	}
}

// SetupDependencyInjection demonstrates enhanced DI features
func (d *DemoApp) SetupDependencyInjection() {
	fmt.Println("=== Enhanced Dependency Injection Demo ===")

	// Register with detailed documentation
	d.container.RegisterFactory(
		reflect.TypeOf((*DatabaseService)(nil)),
		func(c *core.DIContainer) interface{} {
			return NewDatabaseService("postgres://localhost:5432")
		},
		core.Singleton,
	)

	d.container.RegisterFactory(
		reflect.TypeOf((*CacheService)(nil)),
		func(c *core.DIContainer) interface{} {
			return NewCacheService(c)
		},
		core.Singleton,
	)

	// Resolve dependencies
	dbService, err := d.container.Resolve(reflect.TypeOf((*DatabaseService)(nil)))
	if err != nil {
		fmt.Printf("Error resolving database: %v\n", err)
		return
	}

	cacheService, err := d.container.Resolve(reflect.TypeOf((*CacheService)(nil)))
	if err != nil {
		fmt.Printf("Error resolving cache: %v\n", err)
		return
	}

	fmt.Printf("✅ Database Service: %s\n", dbService.(*DatabaseService).GetInfo())
	fmt.Printf("✅ Cache Service: %s\n", cacheService.(*CacheService).GetInfo())

	// Show registrations
	registrations := d.container.GetRegistrations()
	fmt.Printf("✅ Total Registrations: %d\n", len(registrations))
	fmt.Println()
}

// SetupConfiguration demonstrates configuration management
func (d *DemoApp) SetupConfiguration() {
	fmt.Println("=== Configuration Management Demo ===")

	// Demonstrate configuration with a simple map
	cfg := map[string]interface{}{
		"database_url": "postgres://user:pass@localhost:5432/mydb",
		"port":         8080,
		"timeout":      "30s",
		"enable_cors":  true,
	}

	// Show configuration values
	fmt.Printf("✅ Database URL: %v\n", cfg["database_url"])
	fmt.Printf("✅ Port: %v\n", cfg["port"])
	fmt.Printf("✅ Timeout: %v\n", cfg["timeout"])
	fmt.Printf("✅ Enable CORS: %v\n", cfg["enable_cors"])
	fmt.Println()
}

// SetupRouting demonstrates enhanced routing with documentation
func (d *DemoApp) SetupRouting() {
	fmt.Println("=== Enhanced Routing Demo ===")

	enhancedRouter := router.NewEnhancedRouterIntegration()

	// Add routes
	enhancedRouter.AddRoute("GET", "/users/:id", nil)
	enhancedRouter.AddRoute("POST", "/users", nil)
	enhancedRouter.AddRoute("GET", "/health", nil)

	// Generate documentation
	docs, err := enhancedRouter.GenerateAllDocumentation()
	if err != nil {
		fmt.Printf("Error generating docs: %v\n", err)
	} else {
		fmt.Printf("✅ Router Documentation:\n%s\n", docs)
	}

	// Generate Mermaid diagram
	diagram := enhancedRouter.GenerateMermaidDiagram()
	fmt.Printf("✅ Mermaid Diagram:\n%s\n", diagram)
	fmt.Println()
}

// SetupMetrics demonstrates performance monitoring
func (d *DemoApp) SetupMetrics() {
	fmt.Println("=== Performance Monitoring Demo ===")

	// Register metrics
	d.dashboard.RegisterMetric("http_request_duration", "HTTP request duration", "seconds", "performance", nil, metrics.AggregationAvg)
	d.dashboard.RegisterMetric("memory_usage", "Memory usage", "percent", "resource", nil, metrics.AggregationMax)
	d.dashboard.RegisterMetric("error_rate", "Error rate", "percent", "reliability", nil, metrics.AggregationAvg)

	// Add sample data
	d.dashboard.AddMetricValue("http_request_duration", 0.15, nil)
	d.dashboard.AddMetricValue("http_request_duration", 0.23, nil)
	d.dashboard.AddMetricValue("memory_usage", 65.5, nil)
	d.dashboard.AddMetricValue("error_rate", 2.3, nil)

	// Record events
	d.dashboard.RecordEvent("startup", "Application started", map[string]string{"version": "1.0.0"})
	d.dashboard.RecordEvent("request", "HTTP request processed", map[string]string{"method": "GET", "path": "/users/123"})

	// Get stats
	stats, err := d.dashboard.GetMetricStats("http_request_duration")
	if err == nil {
		fmt.Printf("✅ HTTP Request Stats: avg=%.3fs, max=%.3fs, min=%.3fs, count=%d\n",
			stats.Average, stats.Max, stats.Min, stats.Count)
	}

	// Generate report
	report := d.dashboard.GenerateReport()
	fmt.Printf("✅ Performance Report:\n%s\n", report)

	// Generate JSON
	jsonReport, _ := d.dashboard.GenerateJSON()
	fmt.Printf("✅ JSON Report (first 200 chars): %s...\n", jsonReport[:min(200, len(jsonReport))])

	// Generate Mermaid
	mermaid := d.dashboard.GenerateMermaid()
	fmt.Printf("✅ Mermaid Visualization:\n%s\n", mermaid)
	fmt.Println()
}

// SetupErrorHandling demonstrates structured error handling
func (d *DemoApp) SetupErrorHandling() {
	fmt.Println("=== Structured Error Handling Demo ===")

	// Create structured errors
	dbError := contract.NewStructuredError(
		"DB_CONN_ERR",
		"Failed to connect to database",
		nil,
	).WithField("host", "localhost").WithField("port", 5432)

	authError := contract.NewStructuredError(
		"AUTH_TOKEN_INVALID",
		"Invalid authentication token",
		nil,
	).WithField("token_prefix", "Bearer eyJ...")

	// Demonstrate error handling
	fmt.Printf("✅ Database Error: %v\n", dbError)
	fmt.Printf("✅ Auth Error: %v\n", authError)

	// Demonstrate error wrapping
	wrappedErr := contract.WrapError(dbError, "Failed to initialize application", "INIT_ERR", nil)
	fmt.Printf("✅ Wrapped Error: %v\n", wrappedErr)
	fmt.Println()
}

// SetupComponentManagement demonstrates component management
func (d *DemoApp) SetupComponentManagement() {
	fmt.Println("=== Component Management Demo ===")

	// Create enhanced component manager
	manager := core.NewEnhancedComponentManager()

	// Register components
	manager.RegisterComponent(
		"database",
		reflect.TypeOf((*DatabaseService)(nil)),
		"core",
		"Database Service",
		[]string{},
		core.LifecycleSingleton,
		[]string{"storage"},
		map[string]interface{}{"driver": "postgres"},
	)

	manager.RegisterComponent(
		"cache",
		reflect.TypeOf((*CacheService)(nil)),
		"core",
		"Cache Service",
		[]string{},
		core.LifecycleSingleton,
		[]string{"performance"},
		map[string]interface{}{"backend": "redis"},
	)

	manager.RegisterComponent(
		"webserver",
		"webserver_component",
		"network",
		"Web Server",
		[]string{"database", "cache"},
		core.LifecycleSingleton,
		[]string{"http"},
		map[string]interface{}{"port": 8080},
	)

	// Get component graph
	graph := manager.GetDependencyGraph()
	fmt.Printf("✅ Component Graph Nodes: %d\n", len(graph.Nodes))
	fmt.Printf("✅ Component Graph Edges: %d\n", len(graph.Edges))

	// Generate component report
	report := manager.GenerateComponentReport()
	fmt.Printf("✅ Component Report:\n%s\n", report)

	// Validate dependencies
	errors := manager.ValidateDependencies()
	if len(errors) == 0 {
		fmt.Println("✅ All dependencies are valid")
	} else {
		fmt.Printf("❌ Validation errors: %v\n", errors)
	}
	fmt.Println()
}

// DemoModels defines demo model types
type DatabaseService struct {
	connectionString string
}

func NewDatabaseService(connStr string) *DatabaseService {
	return &DatabaseService{connectionString: connStr}
}

func (db *DatabaseService) GetInfo() string {
	return fmt.Sprintf("Database connected to: %s", db.connectionString)
}

type CacheService struct {
	container *core.DIContainer
}

func NewCacheService(container *core.DIContainer) *CacheService {
	return &CacheService{container: container}
}

func (c *CacheService) GetInfo() string {
	return "Redis Cache Service ready"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("🚀 Plumego AI Agent Friendly Features Demo")
	fmt.Println("==========================================\n")

	demo := NewDemoApp()

	// Run all demos
	demo.SetupDependencyInjection()
	demo.SetupConfiguration()
	demo.SetupRouting()
	demo.SetupMetrics()
	demo.SetupErrorHandling()
	demo.SetupComponentManagement()

	fmt.Println("✅ All demos completed successfully!")
	fmt.Println("\nThese features make Plumego highly AI Agent friendly by providing:")
	fmt.Println("1. Comprehensive documentation generation")
	fmt.Println("2. Structured error handling with error codes")
	fmt.Println("3. Dependency visualization and management")
	fmt.Println("4. Configuration validation and schema generation")
	fmt.Println("5. Performance monitoring and reporting")
	fmt.Println("6. Route documentation and OpenAPI support")
}
