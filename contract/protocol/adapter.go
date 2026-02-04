// Package protocol provides interface contracts for protocol adapters
//
// This package defines interfaces for extending plumego to support different
// protocols (gRPC, GraphQL, WebSocket, etc.) WITHOUT adding dependencies to
// the plumego core.
//
// Users implement these interfaces using their chosen libraries:
//   - gRPC: google.golang.org/grpc
//   - GraphQL: github.com/graphql-go/graphql
//   - Custom protocols: User-defined implementations
//
// This contract-based approach maintains zero dependencies in plumego
// while providing full extensibility.
//
// Example usage:
//
//	// User's application code
//	type MyGRPCAdapter struct {
//		conn *grpc.ClientConn
//	}
//
//	func (a *MyGRPCAdapter) Handles(r *http.Request) bool {
//		return r.Header.Get("Content-Type") == "application/grpc"
//	}
//
//	func (a *MyGRPCAdapter) Execute(ctx context.Context, req protocol.Request) (protocol.Response, error) {
//		// User's gRPC implementation using grpc library
//	}
//
//	// Register adapter
//	registry := protocol.NewRegistry()
//	registry.Register(myAdapter)
//	app.Use(protomw.Middleware(registry))
package protocol

import (
	"context"
	"io"
)

// ProtocolAdapter converts HTTP requests to/from other protocols
//
// Users implement this interface to add support for:
//   - gRPC (using google.golang.org/grpc)
//   - GraphQL (using github.com/graphql-go/graphql or similar)
//   - WebSocket (using gorilla/websocket or nhooyr.io/websocket)
//   - Thrift, SOAP, custom protocols
//
// The adapter is responsible for:
//  1. Detecting if it can handle a request (Handles)
//  2. Transforming HTTP request to protocol request (Transform)
//  3. Executing the protocol request (Execute)
//  4. Encoding the protocol response as HTTP response (Encode)
type ProtocolAdapter interface {
	// Name returns the protocol name (e.g., "grpc", "graphql", "websocket")
	Name() string

	// Handles checks if this adapter can handle the request
	// Examples:
	//   - gRPC: Content-Type == "application/grpc"
	//   - GraphQL: Path == "/graphql"
	//   - Custom: Custom header or path pattern
	Handles(req *HTTPRequest) bool

	// Transform converts HTTP request to protocol-specific request
	// This is where the adapter parses HTTP into its native format
	Transform(ctx context.Context, req *HTTPRequest) (Request, error)

	// Execute sends the protocol request to backend
	// This is where the actual protocol call happens (gRPC call, GraphQL query, etc.)
	Execute(ctx context.Context, req Request) (Response, error)

	// Encode writes the protocol response as HTTP response
	// This converts the protocol response back to HTTP format
	Encode(ctx context.Context, resp Response, writer ResponseWriter) error
}

// HTTPRequest represents an incoming HTTP request
// This is a simplified interface to avoid importing net/http in user code
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string][]string
	Body    io.Reader

	// Metadata holds protocol-specific metadata
	Metadata map[string]interface{}
}

// Request represents a protocol-specific request
// Users populate this with their protocol's request format
type Request interface {
	// Method returns the operation name (e.g., gRPC method, GraphQL operation)
	Method() string

	// Headers returns protocol-specific headers/metadata
	Headers() map[string][]string

	// Body returns the request body reader
	Body() io.Reader

	// Metadata returns protocol-specific metadata
	// Examples:
	//   - gRPC: service name, authority, timeout
	//   - GraphQL: operation name, variables
	Metadata() map[string]interface{}
}

// Response represents a protocol-specific response
type Response interface {
	// StatusCode returns HTTP-compatible status code
	StatusCode() int

	// Headers returns protocol-specific headers/metadata
	Headers() map[string][]string

	// Body returns the response body reader
	Body() io.Reader

	// Metadata returns protocol-specific metadata
	Metadata() map[string]interface{}
}

// ResponseWriter provides a way to write HTTP responses
// This abstraction allows adapters to work without importing net/http
type ResponseWriter interface {
	// Header returns the header map
	Header() map[string][]string

	// Write writes data to the response body
	Write([]byte) (int, error)

	// WriteHeader sends an HTTP response header with status code
	WriteHeader(statusCode int)
}

// Registry manages protocol adapters
type Registry struct {
	adapters []ProtocolAdapter
}

// NewRegistry creates a new protocol adapter registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make([]ProtocolAdapter, 0),
	}
}

// Register adds a protocol adapter to the registry
func (r *Registry) Register(adapter ProtocolAdapter) {
	r.adapters = append(r.adapters, adapter)
}

// Find returns the first adapter that can handle the request
// Returns nil if no adapter is found
func (r *Registry) Find(req *HTTPRequest) ProtocolAdapter {
	for _, adapter := range r.adapters {
		if adapter.Handles(req) {
			return adapter
		}
	}
	return nil
}

// Adapters returns all registered adapters
func (r *Registry) Adapters() []ProtocolAdapter {
	return r.adapters
}

// Count returns the number of registered adapters
func (r *Registry) Count() int {
	return len(r.adapters)
}
