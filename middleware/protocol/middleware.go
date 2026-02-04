// Package protocol provides middleware for protocol gateway functionality
//
// This middleware enables plumego to act as a gateway for different protocols
// (gRPC, GraphQL, etc.) without adding dependencies to the core.
//
// Users register protocol adapters that implement the ProtocolAdapter interface,
// and the middleware routes requests to the appropriate adapter.
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/contract/protocol"
//		protomw "github.com/spcent/plumego/middleware/protocol"
//	)
//
//	// User implements adapter using their chosen library
//	grpcAdapter := NewMyGRPCAdapter()
//	graphqlAdapter := NewMyGraphQLAdapter()
//
//	// Register adapters
//	registry := protocol.NewRegistry()
//	registry.Register(grpcAdapter)
//	registry.Register(graphqlAdapter)
//
//	// Use middleware
//	app.Use(protomw.Middleware(registry))
package protocol

import (
	"bytes"
	"io"
	"net/http"

	"github.com/spcent/plumego/contract/protocol"
)

// Middleware creates a protocol gateway middleware
func Middleware(registry *protocol.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Convert http.Request to HTTPRequest
			httpReq := &protocol.HTTPRequest{
				Method:   r.Method,
				URL:      r.URL.String(),
				Headers:  convertHeaders(r.Header),
				Body:     r.Body,
				Metadata: make(map[string]interface{}),
			}

			// Try to find adapter
			adapter := registry.Find(httpReq)
			if adapter == nil {
				// No adapter found, pass through to next handler
				next.ServeHTTP(w, r)
				return
			}

			// Transform request
			req, err := adapter.Transform(r.Context(), httpReq)
			if err != nil {
				http.Error(w, "Protocol transformation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Execute protocol request
			resp, err := adapter.Execute(r.Context(), req)
			if err != nil {
				http.Error(w, "Protocol execution failed: "+err.Error(), http.StatusBadGateway)
				return
			}

			// Create response writer wrapper
			respWriter := &responseWriter{
				ResponseWriter: w,
				headers:        make(map[string][]string),
			}

			// Encode response
			if err := adapter.Encode(r.Context(), resp, respWriter); err != nil {
				http.Error(w, "Protocol encoding failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to implement protocol.ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	headers map[string][]string
	written bool
}

func (w *responseWriter) Header() map[string][]string {
	return w.headers
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		// Copy headers to underlying writer before first write
		for key, values := range w.headers {
			for _, value := range values {
				w.ResponseWriter.Header().Add(key, value)
			}
		}
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.written {
		// Copy headers to underlying writer
		for key, values := range w.headers {
			for _, value := range values {
				w.ResponseWriter.Header().Add(key, value)
			}
		}
		w.ResponseWriter.WriteHeader(statusCode)
		w.written = true
	}
}

// convertHeaders converts http.Header to map[string][]string
func convertHeaders(h http.Header) map[string][]string {
	headers := make(map[string][]string)
	for key, values := range h {
		headers[key] = values
	}
	return headers
}

// Config holds protocol middleware configuration
type Config struct {
	// Registry holds registered protocol adapters
	Registry *protocol.Registry

	// OnAdapterNotFound is called when no adapter is found
	// If nil, request passes through to next handler
	OnAdapterNotFound func(w http.ResponseWriter, r *http.Request)

	// OnTransformError is called when transformation fails
	OnTransformError func(w http.ResponseWriter, r *http.Request, err error)

	// OnExecuteError is called when execution fails
	OnExecuteError func(w http.ResponseWriter, r *http.Request, err error)

	// OnEncodeError is called when encoding fails
	OnEncodeError func(w http.ResponseWriter, r *http.Request, err error)
}

// MiddlewareWithConfig creates a protocol gateway middleware with configuration
func MiddlewareWithConfig(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Convert http.Request to HTTPRequest
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body)) // Restore for potential passthrough

			httpReq := &protocol.HTTPRequest{
				Method:   r.Method,
				URL:      r.URL.String(),
				Headers:  convertHeaders(r.Header),
				Body:     bytes.NewReader(body),
				Metadata: make(map[string]interface{}),
			}

			// Try to find adapter
			adapter := config.Registry.Find(httpReq)
			if adapter == nil {
				if config.OnAdapterNotFound != nil {
					config.OnAdapterNotFound(w, r)
					return
				}
				// Pass through to next handler
				next.ServeHTTP(w, r)
				return
			}

			// Transform request
			req, err := adapter.Transform(r.Context(), httpReq)
			if err != nil {
				if config.OnTransformError != nil {
					config.OnTransformError(w, r, err)
					return
				}
				http.Error(w, "Protocol transformation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Execute protocol request
			resp, err := adapter.Execute(r.Context(), req)
			if err != nil {
				if config.OnExecuteError != nil {
					config.OnExecuteError(w, r, err)
					return
				}
				http.Error(w, "Protocol execution failed: "+err.Error(), http.StatusBadGateway)
				return
			}

			// Create response writer wrapper
			respWriter := &responseWriter{
				ResponseWriter: w,
				headers:        make(map[string][]string),
			}

			// Encode response
			if err := adapter.Encode(r.Context(), resp, respWriter); err != nil {
				if config.OnEncodeError != nil {
					config.OnEncodeError(w, r, err)
					return
				}
				http.Error(w, "Protocol encoding failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
