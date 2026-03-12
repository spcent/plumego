// Package handler contains the HTTP handlers for the reference application.
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// APIHandler handles the core JSON API endpoints.
type APIHandler struct{}

// Hello responds with service metadata and available endpoints.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"message":   "hello from plumego reference",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"features": []string{
			"WebSocket",
			"Documentation",
			"Webhook",
			"Metrics",
			"Health Check",
			"Middleware",
			"Pub/Sub",
		},
		"endpoints": map[string]string{
			"docs":      "/docs",
			"webhooks":  "/webhooks",
			"metrics":   "/metrics",
			"health":    "/health/ready",
			"websocket": "/ws",
			"api":       "/api",
		},
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Status responds with a summary of system health and component state.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().Add(-time.Hour) // simulate 1 hour uptime

	resp := map[string]any{
		"status":  "healthy",
		"service": "plumego-reference",
		"version": "1.0.0",
		"system": map[string]any{
			"uptime":     time.Since(startTime).String(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"go_version": "1.24+",
		},
		"components": map[string]any{
			"websocket":   "enabled",
			"webhook_in":  "enabled",
			"webhook_out": "enabled",
			"metrics":     "enabled",
			"docs":        "enabled",
			"pubsub":      "enabled",
		},
		"endpoints": map[string]string{
			"root":      "/",
			"docs":      "/docs",
			"api":       "/api",
			"metrics":   "/metrics",
			"health":    "/health/ready",
			"websocket": "/ws",
		},
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Test supports format and delay query parameters for integration testing.
func (h APIHandler) Test(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	delay := r.URL.Query().Get("delay")

	if delay != "" {
		if d, err := time.ParseDuration(delay); err == nil {
			const maxDelay = 2 * time.Second
			if d > maxDelay {
				d = maxDelay
			}
			select {
			case <-time.After(d):
			case <-r.Context().Done():
				return
			}
		}
	}

	switch format {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<response><timestamp>%s</timestamp><format>xml</format><status>success</status></response>`,
			time.Now().Format(time.RFC3339))
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		fmt.Fprintf(w, "timestamp,format,status\n%s,csv,success\n", time.Now().Format(time.RFC3339))
	case "plain":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Plain text response at %s", time.Now().Format(time.RFC3339))
	default:
		resp := map[string]any{
			"format":       "json",
			"timestamp":    time.Now().Format(time.RFC3339),
			"status":       "success",
			"query_params": r.URL.Query().Encode(),
		}
		if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
