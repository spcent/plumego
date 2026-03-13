// Package protocol provides interface contracts for gateway protocol adapters.
package protocol

import (
	"context"
	"io"
)

// ProtocolAdapter converts HTTP requests to and from other protocols.
type ProtocolAdapter interface {
	Name() string
	Handles(req *HTTPRequest) bool
	Transform(ctx context.Context, req *HTTPRequest) (Request, error)
	Execute(ctx context.Context, req Request) (Response, error)
	Encode(ctx context.Context, resp Response, writer ResponseWriter) error
}

// HTTPRequest represents an incoming HTTP request without pulling net/http into adapter code.
type HTTPRequest struct {
	Method   string
	URL      string
	Headers  map[string][]string
	Body     io.Reader
	Metadata map[string]any
}

// Request represents a protocol-specific request.
type Request interface {
	Method() string
	Headers() map[string][]string
	Body() io.Reader
	Metadata() map[string]any
}

// Response represents a protocol-specific response.
type Response interface {
	StatusCode() int
	Headers() map[string][]string
	Body() io.Reader
	Metadata() map[string]any
}

// ResponseWriter provides a transport-agnostic way to write an HTTP response.
type ResponseWriter interface {
	Header() map[string][]string
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

// Registry manages protocol adapters.
type Registry struct {
	adapters []ProtocolAdapter
}

// NewRegistry creates a new protocol adapter registry.
func NewRegistry() *Registry {
	return &Registry{adapters: make([]ProtocolAdapter, 0)}
}

// Register adds a protocol adapter to the registry.
func (r *Registry) Register(adapter ProtocolAdapter) {
	r.adapters = append(r.adapters, adapter)
}

// Find returns the first adapter that can handle the request.
func (r *Registry) Find(req *HTTPRequest) ProtocolAdapter {
	for _, adapter := range r.adapters {
		if adapter.Handles(req) {
			return adapter
		}
	}
	return nil
}

// Adapters returns the registered adapters.
func (r *Registry) Adapters() []ProtocolAdapter {
	out := make([]ProtocolAdapter, len(r.adapters))
	copy(out, r.adapters)
	return out
}

// Count returns the number of registered adapters.
func (r *Registry) Count() int {
	return len(r.adapters)
}
