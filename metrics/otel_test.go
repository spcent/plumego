package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/middleware"
)

func TestOpenTelemetryTracer(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	ctx, span := tracer.Start(context.Background(), req)
	if ctx == nil || span == nil {
		t.Fatalf("tracer should return context and span")
	}

	span.End(middleware.RequestMetrics{Status: http.StatusOK, Bytes: 10, TraceID: "abc123"})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %d", len(spans))
	}
	if spans[0].Name != "http.request" {
		t.Fatalf("unexpected span name: %s", spans[0].Name)
	}
	if spans[0].Attributes["plumego.trace_id"] != "abc123" {
		t.Fatalf("trace id not recorded: %#v", spans[0].Attributes)
	}
}
