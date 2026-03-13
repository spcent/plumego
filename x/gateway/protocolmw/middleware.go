package protocolmw

import (
	"bytes"
	"io"
	"net/http"

	"github.com/spcent/plumego/contract"
	mw "github.com/spcent/plumego/middleware"
	gatewayproto "github.com/spcent/plumego/x/gateway/protocol"
)

// Middleware creates protocol gateway middleware.
func Middleware(registry *gatewayproto.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpReq := &gatewayproto.HTTPRequest{
				Method:   r.Method,
				URL:      r.URL.String(),
				Headers:  convertHeaders(r.Header),
				Body:     r.Body,
				Metadata: make(map[string]any),
			}

			adapter := findAdapter(registry, httpReq)
			if adapter == nil {
				next.ServeHTTP(w, r)
				return
			}

			req, err := adapter.Transform(r.Context(), httpReq)
			if err != nil {
				mw.WriteTransportError(w, r, http.StatusBadRequest, mw.CodeProtocolTransformFail, "protocol transformation failed", contract.CategoryClient, map[string]any{"cause": err.Error()})
				return
			}

			resp, err := adapter.Execute(r.Context(), req)
			if err != nil {
				mw.WriteTransportError(w, r, http.StatusBadGateway, mw.CodeProtocolExecutionFail, "protocol execution failed", contract.CategoryServer, map[string]any{"cause": err.Error()})
				return
			}

			respWriter := &responseWriter{
				ResponseWriter: w,
				headers:        make(map[string][]string),
			}

			if err := adapter.Encode(r.Context(), resp, respWriter); err != nil {
				mw.WriteTransportError(w, r, http.StatusInternalServerError, mw.CodeInternalError, "protocol encoding failed", contract.CategoryServer, map[string]any{"cause": err.Error()})
				return
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	headers map[string][]string
	written bool
}

func (w *responseWriter) Header() map[string][]string { return w.headers }

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
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
		for key, values := range w.headers {
			for _, value := range values {
				w.ResponseWriter.Header().Add(key, value)
			}
		}
		w.ResponseWriter.WriteHeader(statusCode)
		w.written = true
	}
}

func findAdapter(registry *gatewayproto.Registry, req *gatewayproto.HTTPRequest) gatewayproto.ProtocolAdapter {
	if registry == nil {
		return nil
	}
	return registry.Find(req)
}

func convertHeaders(h http.Header) map[string][]string {
	headers := make(map[string][]string)
	for key, values := range h {
		headers[key] = values
	}
	return headers
}

// Config holds protocol middleware configuration.
type Config struct {
	Registry          *gatewayproto.Registry
	OnAdapterNotFound func(w http.ResponseWriter, r *http.Request)
	OnTransformError  func(w http.ResponseWriter, r *http.Request, err error)
	OnExecuteError    func(w http.ResponseWriter, r *http.Request, err error)
	OnEncodeError     func(w http.ResponseWriter, r *http.Request, err error)
}

// MiddlewareWithConfig creates protocol gateway middleware with configuration.
func MiddlewareWithConfig(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyReader := r.Body
			if bodyReader == nil {
				bodyReader = http.NoBody
			}
			body, err := io.ReadAll(bodyReader)
			if err != nil {
				if config.OnTransformError != nil {
					config.OnTransformError(w, r, err)
					return
				}
				mw.WriteTransportError(w, r, http.StatusBadRequest, mw.CodeProtocolTransformFail, "protocol request read failed", contract.CategoryClient, map[string]any{"cause": err.Error()})
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			httpReq := &gatewayproto.HTTPRequest{
				Method:   r.Method,
				URL:      r.URL.String(),
				Headers:  convertHeaders(r.Header),
				Body:     bytes.NewReader(body),
				Metadata: make(map[string]any),
			}

			adapter := findAdapter(config.Registry, httpReq)
			if adapter == nil {
				if config.OnAdapterNotFound != nil {
					config.OnAdapterNotFound(w, r)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			req, err := adapter.Transform(r.Context(), httpReq)
			if err != nil {
				if config.OnTransformError != nil {
					config.OnTransformError(w, r, err)
					return
				}
				mw.WriteTransportError(w, r, http.StatusBadRequest, mw.CodeProtocolTransformFail, "protocol transformation failed", contract.CategoryClient, map[string]any{"cause": err.Error()})
				return
			}

			resp, err := adapter.Execute(r.Context(), req)
			if err != nil {
				if config.OnExecuteError != nil {
					config.OnExecuteError(w, r, err)
					return
				}
				mw.WriteTransportError(w, r, http.StatusBadGateway, mw.CodeProtocolExecutionFail, "protocol execution failed", contract.CategoryServer, map[string]any{"cause": err.Error()})
				return
			}

			respWriter := &responseWriter{
				ResponseWriter: w,
				headers:        make(map[string][]string),
			}

			if err := adapter.Encode(r.Context(), resp, respWriter); err != nil {
				if config.OnEncodeError != nil {
					config.OnEncodeError(w, r, err)
					return
				}
				mw.WriteTransportError(w, r, http.StatusInternalServerError, mw.CodeInternalError, "protocol encoding failed", contract.CategoryServer, map[string]any{"cause": err.Error()})
				return
			}
		})
	}
}
