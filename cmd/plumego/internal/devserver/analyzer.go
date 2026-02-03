package devserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/metrics"
)

// RouteInfo represents information about a route
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler,omitempty"`
	Middleware  []string `json:"middleware,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Analyzer analyzes the running application
type Analyzer struct {
	appURL string
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(appURL string) *Analyzer {
	return &Analyzer{
		appURL: appURL,
	}
}

// GetRoutes attempts to fetch routes from the running application
func (a *Analyzer) GetRoutes() ([]RouteInfo, error) {
	// Try to fetch from plumego debug endpoint (JSON format)
	debugURL := a.appURL + "/_debug/routes.json"

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(debugURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes (app may not be running): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("debug endpoint returned status %d", resp.StatusCode)
	}

	// Parse response - plumego returns routes in data.routes format
	var response struct {
		Data struct {
			Routes []struct {
				Method string                 `json:"method"`
				Path   string                 `json:"path"`
				Meta   map[string]interface{} `json:"meta"`
			} `json:"routes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}

	// Convert to RouteInfo format
	var routes []RouteInfo
	for _, r := range response.Data.Routes {
		routes = append(routes, RouteInfo{
			Method: r.Method,
			Path:   r.Path,
		})
	}

	return routes, nil
}

// ProbeEndpoints probes common endpoints to discover routes
func (a *Analyzer) ProbeEndpoints() []RouteInfo {
	commonPaths := []string{
		"/",
		"/health",
		"/ping",
		"/api",
		"/api/status",
		"/api/health",
		"/_debug/routes",
		"/_debug/config",
		"/_debug/middleware",
	}

	var discovered []RouteInfo
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	for _, path := range commonPaths {
		// Try HEAD first (lightweight)
		req, _ := http.NewRequest("HEAD", a.appURL+path, nil)
		resp, err := client.Do(req)

		if err == nil && resp.StatusCode < 500 {
			discovered = append(discovered, RouteInfo{
				Method:      "HEAD",
				Path:        path,
				Description: fmt.Sprintf("Discovered (status: %d)", resp.StatusCode),
			})
			resp.Body.Close()
		}
	}

	return discovered
}

// GetAppInfo fetches application information
func (a *Analyzer) GetAppInfo() (map[string]interface{}, error) {
	configURL := a.appURL + "/_debug/config"

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(configURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// HealthCheck checks if the application is healthy
func (a *Analyzer) HealthCheck() (bool, map[string]interface{}, error) {
	healthURL := a.appURL + "/health"

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		return false, nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == http.StatusOK

	var details map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		// Health endpoint might not return JSON
		details = map[string]interface{}{
			"status": resp.Status,
		}
	}

	return healthy, details, nil
}

// DevMetricsSnapshot contains dev HTTP + DB metrics.
type DevMetricsSnapshot struct {
	HTTP metrics.DevHTTPSnapshot `json:"http"`
	DB   metrics.DevDBSnapshot   `json:"db"`
}

// GetDevMetrics fetches dev metrics from the application.
func (a *Analyzer) GetDevMetrics() (*DevMetricsSnapshot, error) {
	metricsURL := a.appURL + "/_debug/metrics"

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(metricsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
	}

	var payload struct {
		Enabled bool                    `json:"enabled"`
		HTTP    metrics.DevHTTPSnapshot `json:"http"`
		DB      metrics.DevDBSnapshot   `json:"db"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	if !payload.Enabled {
		return nil, fmt.Errorf("dev metrics disabled")
	}

	return &DevMetricsSnapshot{
		HTTP: payload.HTTP,
		DB:   payload.DB,
	}, nil
}
