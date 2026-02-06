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

// maxInspectResponseBytes limits inspect response bodies to 10 MiB
// to prevent OOM from malicious or misconfigured servers.
const maxInspectResponseBytes = 10 << 20

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

func (c *InspectCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	url := fs.String("url", "http://localhost:8080", "Application URL")
	auth := fs.String("auth", "", "Authentication token")
	timeoutStr := fs.String("timeout", "10s", "Request timeout")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid timeout: %v", err), 1)
	}

	subcommand := "health"
	if fs.NArg() > 0 {
		subcommand = fs.Arg(0)
	}

	client := &http.Client{
		Timeout: timeout,
	}

	switch subcommand {
	case "health":
		return inspectHealth(ctx.Out, client, *url, *auth)
	case "metrics":
		return inspectMetrics(ctx.Out, client, *url, *auth)
	case "routes":
		return fetchSingleEndpoint(ctx.Out, client, *url, *auth, "/_routes", "Routes retrieved")
	case "config":
		return fetchSingleEndpoint(ctx.Out, client, *url, *auth, "/_config", "Configuration retrieved")
	case "info":
		return fetchSingleEndpoint(ctx.Out, client, *url, *auth, "/_info", "Application info retrieved")
	default:
		return ctx.Out.Error(fmt.Sprintf("unknown subcommand: %s", subcommand), 1)
	}
}

// doInspectRequest performs an HTTP GET with optional auth and a bounded body read.
// The response body is always closed before returning.
func doInspectRequest(client *http.Client, url, auth string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxInspectResponseBytes))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

// fetchSingleEndpoint fetches one endpoint and returns the parsed JSON result.
func fetchSingleEndpoint(out *output.Formatter, client *http.Client, baseURL, auth, path, successMsg string) error {
	url := strings.TrimSuffix(baseURL, "/") + path

	body, statusCode, err := doInspectRequest(client, url, auth)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}

	if statusCode != http.StatusOK {
		return out.Error(fmt.Sprintf("unexpected status: %d", statusCode), 1)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return out.Error(fmt.Sprintf("failed to parse response: %v", err), 1)
	}

	return out.Success(successMsg, data)
}

// probeEndpoints tries a list of endpoints and returns the first that responds.
func probeEndpoints(client *http.Client, baseURL, auth string, endpoints []string) (body []byte, statusCode int, endpoint string, err error) {
	var lastErr error
	for _, ep := range endpoints {
		url := strings.TrimSuffix(baseURL, "/") + ep

		b, code, reqErr := doInspectRequest(client, url, auth)
		if reqErr != nil {
			lastErr = reqErr
			continue
		}

		return b, code, ep, nil
	}

	if lastErr != nil {
		return nil, 0, "", lastErr
	}
	return nil, 0, "", fmt.Errorf("no endpoints responded")
}

func inspectHealth(out *output.Formatter, client *http.Client, baseURL, auth string) error {
	endpoints := []string{
		"/health",
		"/healthz",
		"/ready",
		"/livez",
		"/_health",
	}

	body, statusCode, endpoint, err := probeEndpoints(client, baseURL, auth, endpoints)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}

	var healthData map[string]any
	if jsonErr := json.Unmarshal(body, &healthData); jsonErr == nil {
		healthData["endpoint"] = endpoint
		healthData["status_code"] = statusCode

		if statusCode >= 200 && statusCode < 300 {
			return out.Success("Application is healthy", healthData)
		}
		return out.Error("Application is unhealthy", 1, healthData)
	}

	result := map[string]any{
		"endpoint":    endpoint,
		"status_code": statusCode,
		"body":        string(body),
	}

	if statusCode >= 200 && statusCode < 300 {
		return out.Success("Application is healthy", result)
	}
	return out.Error("Application is unhealthy", 1, result)
}

func inspectMetrics(out *output.Formatter, client *http.Client, baseURL, auth string) error {
	endpoints := []string{
		"/metrics",
		"/_metrics",
		"/debug/metrics",
	}

	body, statusCode, endpoint, err := probeEndpoints(client, baseURL, auth, endpoints)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to fetch metrics: %v", err), 1)
	}

	if statusCode != http.StatusOK {
		return out.Error("no metrics endpoints found", 1)
	}

	var metricsData map[string]any
	if jsonErr := json.Unmarshal(body, &metricsData); jsonErr == nil {
		metricsData["endpoint"] = endpoint
		return out.Success("Metrics retrieved", metricsData)
	}

	result := map[string]any{
		"endpoint": endpoint,
		"format":   "text",
		"data":     string(body),
	}

	return out.Success("Metrics retrieved", result)
}
