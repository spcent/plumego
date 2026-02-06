package protocol

import (
	"context"
	"io"
)

// GRPCAdapter is a specialized adapter contract for gRPC gateway
//
// Users implement this interface using google.golang.org/grpc library.
// This contract defines what plumego expects from a gRPC adapter
// without importing the gRPC library.
//
// Example implementation:
//
//	type MyGRPCAdapter struct {
//		conn *grpc.ClientConn  // User imports grpc library
//	}
//
//	func (a *MyGRPCAdapter) Name() string {
//		return "grpc"
//	}
//
//	func (a *MyGRPCAdapter) Handles(req *HTTPRequest) bool {
//		contentType := req.Headers["Content-Type"]
//		return len(contentType) > 0 && contentType[0] == "application/grpc"
//	}
//
//	func (a *MyGRPCAdapter) Service() string {
//		return "MyService"
//	}
//
//	func (a *MyGRPCAdapter) Methods() []string {
//		return []string{"GetUser", "ListUsers", "CreateUser"}
//	}
type GRPCAdapter interface {
	ProtocolAdapter

	// Service returns the gRPC service name
	// Example: "user.v1.UserService"
	Service() string

	// Methods returns supported RPC method names
	// Example: ["GetUser", "CreateUser", "UpdateUser"]
	Methods() []string
}

// GRPCMetadata represents gRPC-specific metadata
//
// Users can embed this in their Request/Response implementations
// to carry gRPC-specific information
type GRPCMetadata struct {
	// Service is the full service name (e.g., "user.v1.UserService")
	Service string

	// Method is the RPC method name (e.g., "GetUser")
	Method string

	// Authority is the :authority pseudo-header
	Authority string

	// Timeout is the grpc-timeout header value
	Timeout string

	// ContentType is the content-type (e.g., "application/grpc+proto")
	ContentType string

	// Encoding is the grpc-encoding header (e.g., "gzip")
	Encoding string

	// UserAgent is the grpc-user-agent
	UserAgent string

	// Custom holds additional custom metadata
	Custom map[string]string
}

// GRPCRequest is a helper type users can embed in their request implementations
type GRPCRequest struct {
	MethodName string
	Meta       *GRPCMetadata
	BodyReader io.Reader
	HeaderMap  map[string][]string
}

// Method returns the gRPC method name
func (r *GRPCRequest) Method() string {
	return r.MethodName
}

// Headers returns gRPC headers
func (r *GRPCRequest) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

// Body returns the request body
func (r *GRPCRequest) Body() io.Reader {
	return r.BodyReader
}

// Metadata returns gRPC metadata as generic map
func (r *GRPCRequest) Metadata() map[string]any {
	m := make(map[string]any)
	if r.Meta != nil {
		m["service"] = r.Meta.Service
		m["method"] = r.Meta.Method
		m["authority"] = r.Meta.Authority
		m["timeout"] = r.Meta.Timeout
		m["content_type"] = r.Meta.ContentType
		m["encoding"] = r.Meta.Encoding
		m["user_agent"] = r.Meta.UserAgent
		if r.Meta.Custom != nil {
			for k, v := range r.Meta.Custom {
				m[k] = v
			}
		}
	}
	return m
}

// GRPCResponse is a helper type users can embed in their response implementations
type GRPCResponse struct {
	Status     int
	Meta       *GRPCMetadata
	BodyReader io.Reader
	HeaderMap  map[string][]string
}

// StatusCode returns HTTP status code
func (r *GRPCResponse) StatusCode() int {
	if r.Status == 0 {
		return 200
	}
	return r.Status
}

// Headers returns response headers
func (r *GRPCResponse) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

// Body returns response body
func (r *GRPCResponse) Body() io.Reader {
	return r.BodyReader
}

// Metadata returns gRPC metadata
func (r *GRPCResponse) Metadata() map[string]any {
	m := make(map[string]any)
	if r.Meta != nil {
		m["service"] = r.Meta.Service
		m["method"] = r.Meta.Method
		m["content_type"] = r.Meta.ContentType
		m["encoding"] = r.Meta.Encoding
	}
	return m
}

// GRPCErrorCode maps gRPC error codes to HTTP status codes
// Users can use this in their implementations
func GRPCErrorCode(code int) int {
	// Standard gRPC to HTTP mapping
	switch code {
	case 0: // OK
		return 200
	case 1: // CANCELLED
		return 499
	case 2: // UNKNOWN
		return 500
	case 3: // INVALID_ARGUMENT
		return 400
	case 4: // DEADLINE_EXCEEDED
		return 504
	case 5: // NOT_FOUND
		return 404
	case 6: // ALREADY_EXISTS
		return 409
	case 7: // PERMISSION_DENIED
		return 403
	case 8: // RESOURCE_EXHAUSTED
		return 429
	case 9: // FAILED_PRECONDITION
		return 400
	case 10: // ABORTED
		return 409
	case 11: // OUT_OF_RANGE
		return 400
	case 12: // UNIMPLEMENTED
		return 501
	case 13: // INTERNAL
		return 500
	case 14: // UNAVAILABLE
		return 503
	case 15: // DATA_LOSS
		return 500
	case 16: // UNAUTHENTICATED
		return 401
	default:
		return 500
	}
}

// GRPCInvoker is an interface users can implement for custom gRPC invocation
//
// This provides a hook for users to customize how gRPC calls are made
type GRPCInvoker interface {
	// Invoke executes a gRPC call
	Invoke(ctx context.Context, method string, req Request) (Response, error)
}
