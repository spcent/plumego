package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/metrics"
)

type spanContextSpan struct {
	traceID string
	spanID  string
	ended   bool
}

func (s *spanContextSpan) End(status, bytes int, traceID string) { s.ended = true }
func (s *spanContextSpan) TraceID() string                       { return s.traceID }
func (s *spanContextSpan) SpanID() string                        { return s.spanID }

type spanContextTracer struct {
	span *spanContextSpan
}

func (t *spanContextTracer) Start(ctx context.Context, r *http.Request) (context.Context, metrics.TraceSpan) {
	t.span = &spanContextSpan{traceID: "trace-ctx", spanID: "span-123"}
	return ctx, t.span
}

func TestMiddlewareSetsTraceHeadersAndSpanContext(t *testing.T) {
	tracer := &spanContextTracer{}
	handler := Middleware(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := contract.TraceContextFromContext(r.Context())
		if tc == nil || tc.SpanID != "span-123" {
			t.Fatalf("expected trace context with span id, got %+v", tc)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/trace", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(contract.RequestIDHeader) != "trace-ctx" {
		t.Fatalf("expected request id header to be set from tracer")
	}
	if rec.Header().Get("X-Span-ID") != "span-123" {
		t.Fatalf("expected span id header to be set")
	}
	if tracer.span == nil || !tracer.span.ended {
		t.Fatalf("expected tracer span to be ended")
	}
}
