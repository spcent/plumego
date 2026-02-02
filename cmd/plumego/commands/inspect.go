package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

type InspectCmd struct{}

func (c *InspectCmd) Name() string {
	return "inspect"
}

func (c *InspectCmd) Short() string {
	return "Inspect running application"
}

func (c *InspectCmd) Long() string {
	return `Inspect a running plumego application via HTTP endpoints.

This command connects to a running application and fetches various
runtime information like health status, metrics, routes, and configuration.

Subcommands:
  health    - Check health endpoints
  metrics   - Fetch metrics
  routes    - List active routes (requires support in app)
  config    - View runtime config (requires support in app)
  info      - General application info

Examples:
  plumego inspect health --url http://localhost:8080
  plumego inspect metrics --url http://localhost:8080
  plumego inspect health --format json`
}

func (c *InspectCmd) Flags() []Flag {
	return []Flag{
		{Name: "url", Default: "http://localhost:8080", Usage: "Application URL"},
		{Name: "auth", Default: "", Usage: "Authentication token"},
		{Name: "timeout", Default: "10s", Usage: "Request timeout"},
	}
}

func (c *InspectCmd) Run(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)

	url := fs.String("url", "http://localhost:8080", "Application URL")
	auth := fs.String("auth", "", "Authentication token")
	timeoutStr := fs.String("timeout", "10s", "Request timeout")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Parse timeout
	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("invalid timeout: %v", err), 1)
	}

	// Get subcommand
	subcommand := "health"
	if fs.NArg() > 0 {
		subcommand = fs.Arg(0)
	}

	client := &http.Client{
		Timeout: timeout,
	}

	switch subcommand {
	case "health":
		return inspectHealth(client, *url, *auth)
	case "metrics":
		return inspectMetrics(client, *url, *auth)
	case "routes":
		return inspectRoutes(client, *url, *auth)
	case "config":
		return inspectConfig(client, *url, *auth)
	case "info":
		return inspectInfo(client, *url, *auth)
	default:
		return output.NewFormatter().Error(fmt.Sprintf("unknown subcommand: %s", subcommand), 1)
	}
}

func inspectHealth(client *http.Client, baseURL, auth string) error {
	// Try common health endpoints
	endpoints := []string{
		"/health",
		"/healthz",
		"/ready",
		"/livez",
		"/_health",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		url := strings.TrimSuffix(baseURL, "/") + endpoint

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		if auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Try to parse as JSON
		var healthData map[string]any
		if err := json.Unmarshal(body, &healthData); err == nil {
			healthData["endpoint"] = endpoint
			healthData["status_code"] = resp.StatusCode

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return output.NewFormatter().Success("Application is healthy", healthData)
			} else {
				return output.NewFormatter().Error("Application is unhealthy", 1, healthData)
			}
		}

		// Not JSON, return as text
		result := map[string]any{
			"endpoint":    endpoint,
			"status_code": resp.StatusCode,
			"body":        string(body),
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return output.NewFormatter().Success("Application is healthy", result)
		} else {
			return output.NewFormatter().Error("Application is unhealthy", 1, result)
		}
	}

	if lastErr != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to connect: %v", lastErr), 1)
	}

	return output.NewFormatter().Error("no health endpoints found", 1)
}

func inspectMetrics(client *http.Client, baseURL, auth string) error {
	// Try common metrics endpoints
	endpoints := []string{
		"/metrics",
		"/_metrics",
		"/debug/metrics",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		url := strings.TrimSuffix(baseURL, "/") + endpoint

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		if auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Try to parse as JSON
		var metricsData map[string]any
		if err := json.Unmarshal(body, &metricsData); err == nil {
			metricsData["endpoint"] = endpoint
			return output.NewFormatter().Success("Metrics retrieved", metricsData)
		}

		// Return as text (e.g., Prometheus format)
		result := map[string]any{
			"endpoint": endpoint,
			"format":   "text",
			"data":     string(body),
		}

		return output.NewFormatter().Success("Metrics retrieved", result)
	}

	if lastErr != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to fetch metrics: %v", lastErr), 1)
	}

	return output.NewFormatter().Error("no metrics endpoints found", 1)
}

func inspectRoutes(client *http.Client, baseURL, auth string) error {
	url := strings.TrimSuffix(baseURL, "/") + "/_routes"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to create request: %v", err), 1)
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return output.NewFormatter().Error(fmt.Sprintf("unexpected status: %d", resp.StatusCode), 1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to read response: %v", err), 1)
	}

	var routesData map[string]any
	if err := json.Unmarshal(body, &routesData); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to parse response: %v", err), 1)
	}

	return output.NewFormatter().Success("Routes retrieved", routesData)
}

func inspectConfig(client *http.Client, baseURL, auth string) error {
	url := strings.TrimSuffix(baseURL, "/") + "/_config"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to create request: %v", err), 1)
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return output.NewFormatter().Error(fmt.Sprintf("unexpected status: %d", resp.StatusCode), 1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to read response: %v", err), 1)
	}

	var configData map[string]any
	if err := json.Unmarshal(body, &configData); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to parse response: %v", err), 1)
	}

	return output.NewFormatter().Success("Configuration retrieved", configData)
}

func inspectInfo(client *http.Client, baseURL, auth string) error {
	url := strings.TrimSuffix(baseURL, "/") + "/_info"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to create request: %v", err), 1)
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return output.NewFormatter().Error(fmt.Sprintf("unexpected status: %d", resp.StatusCode), 1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to read response: %v", err), 1)
	}

	var infoData map[string]any
	if err := json.Unmarshal(body, &infoData); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to parse response: %v", err), 1)
	}

	return output.NewFormatter().Success("Application info retrieved", infoData)
}
