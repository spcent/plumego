// Package gateway adapts grpc-gateway runtime handlers to net/http.
package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HTTPTranscoder struct {
	mux    *runtime.ServeMux
	target string
}

func New(target string, opts ...runtime.ServeMuxOption) *HTTPTranscoder {
	return &HTTPTranscoder{
		mux:    runtime.NewServeMux(opts...),
		target: strings.TrimSpace(target),
	}
}

func (t *HTTPTranscoder) Register(ctx context.Context, handler runtime.HandlerFunc, pattern string) error {
	_ = ctx
	method, path := splitPattern(pattern)
	return t.mux.HandlePath(method, path, handler)
}

func (t *HTTPTranscoder) Handler() http.Handler {
	return t.mux
}

func (t *HTTPTranscoder) Target() string {
	return t.target
}

func splitPattern(pattern string) (string, string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return http.MethodGet, "/"
	}
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		return http.MethodGet, pattern
	}
	method = strings.TrimSpace(method)
	path = strings.TrimSpace(path)
	if method == "" {
		method = http.MethodGet
	}
	if path == "" {
		path = "/"
	}
	return method, path
}
