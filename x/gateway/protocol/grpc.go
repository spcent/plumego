package protocol

import (
	"context"
	"io"
)

// GRPCAdapter is a specialized adapter contract for gRPC gateway use cases.
type GRPCAdapter interface {
	ProtocolAdapter
	Service() string
	Methods() []string
}

// GRPCMetadata represents gRPC-specific metadata.
type GRPCMetadata struct {
	Service     string
	Method      string
	Authority   string
	Timeout     string
	ContentType string
	Encoding    string
	UserAgent   string
	Custom      map[string]string
}

// GRPCRequest is a helper type users can embed in their request implementations.
type GRPCRequest struct {
	MethodName string
	Meta       *GRPCMetadata
	BodyReader io.Reader
	HeaderMap  map[string][]string
}

func (r *GRPCRequest) Method() string { return r.MethodName }

func (r *GRPCRequest) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

func (r *GRPCRequest) Body() io.Reader { return r.BodyReader }

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
		for k, v := range r.Meta.Custom {
			m[k] = v
		}
	}
	return m
}

// GRPCResponse is a helper type users can embed in their response implementations.
type GRPCResponse struct {
	Status     int
	Meta       *GRPCMetadata
	BodyReader io.Reader
	HeaderMap  map[string][]string
}

func (r *GRPCResponse) StatusCode() int {
	if r.Status == 0 {
		return 200
	}
	return r.Status
}

func (r *GRPCResponse) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

func (r *GRPCResponse) Body() io.Reader { return r.BodyReader }

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

// GRPCErrorCode maps gRPC error codes to HTTP status codes.
func GRPCErrorCode(code int) int {
	switch code {
	case 0:
		return 200
	case 1:
		return 499
	case 2:
		return 500
	case 3:
		return 400
	case 4:
		return 504
	case 5:
		return 404
	case 6:
		return 409
	case 7:
		return 403
	case 8:
		return 429
	case 9:
		return 400
	case 10:
		return 409
	case 11:
		return 400
	case 12:
		return 501
	case 13:
		return 500
	case 14:
		return 503
	case 15:
		return 500
	case 16:
		return 401
	default:
		return 500
	}
}

// GRPCInvoker is an interface users can implement for custom gRPC invocation.
type GRPCInvoker interface {
	Invoke(ctx context.Context, method string, req Request) (Response, error)
}
