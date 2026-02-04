package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware/observability"
)

func TestOpenTelemetryTracer(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	ctx, span := tracer.Start(context.Background(), req)
	if ctx == nil || span == nil {
		t.Fatalf("tracer should return context and span")
	}

	// Add a small delay to ensure non-zero duration
	time.Sleep(1 * time.Millisecond)

	span.End(observability.RequestMetrics{
		Status:   http.StatusOK,
		Bytes:    10,
		TraceID:  "abc123",
		Duration: 100 * time.Millisecond,
	})

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
	if spans[0].Status != "OK" {
		t.Fatalf("expected OK status, got: %s", spans[0].Status)
	}
	if spans[0].Duration == 0 {
		t.Fatalf("expected non-zero duration")
	}
}

func TestOpenTelemetryTracerError(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodPost, "/api/error", nil)
	_, span := tracer.Start(context.Background(), req)
	if span == nil {
		t.Fatalf("tracer should return span")
	}

	span.End(observability.RequestMetrics{
		Status:   http.StatusInternalServerError,
		Bytes:    50,
		TraceID:  "error123",
		Duration: 200 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %d", len(spans))
	}
	if spans[0].Status != "ERROR" {
		t.Fatalf("expected ERROR status, got: %s", spans[0].Status)
	}
	if spans[0].StatusMessage == "" {
		t.Fatalf("expected non-empty status message for error")
	}
}

func TestOpenTelemetryTracerWithParent(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("X-Trace-ID", "parent-trace-id")
	_, span := tracer.Start(context.Background(), req)

	span.End(observability.RequestMetrics{
		Status:   http.StatusOK,
		Bytes:    10,
		TraceID:  "abc123",
		Duration: 100 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %d", len(spans))
	}
	if spans[0].TraceID != "parent-trace-id" {
		t.Fatalf("expected trace ID to be sourced from header, got: %s", spans[0].TraceID)
	}
	if spans[0].Attributes["parent.trace_id"] != "parent-trace-id" {
		t.Fatalf("expected parent trace attribute to be set, got: %s", spans[0].Attributes["parent.trace_id"])
	}
}

func TestOpenTelemetryTracerStats(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	// Add multiple spans
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, span := tracer.Start(context.Background(), req)

		// Add a small delay to ensure non-zero duration
		time.Sleep(1 * time.Millisecond)

		status := http.StatusOK
		if i == 3 {
			status = http.StatusInternalServerError
		}

		span.End(observability.RequestMetrics{
			Status:   status,
			Bytes:    100,
			TraceID:  "trace",
			Duration: time.Duration(100+i*10) * time.Millisecond,
		})
	}

	stats := tracer.GetStats()
	if stats.TotalSpans != 5 {
		t.Fatalf("expected 5 spans, got %d", stats.TotalSpans)
	}
	// Check that we have at least one error span (status 500)
	if stats.ErrorSpans < 1 {
		t.Fatalf("expected at least 1 error span, got %d", stats.ErrorSpans)
	}
	if stats.AverageDuration == 0 {
		t.Fatalf("expected non-zero average duration")
	}
}

func TestOpenTelemetryTracerClear(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, span := tracer.Start(context.Background(), req)
	span.End(observability.RequestMetrics{
		Status:  http.StatusOK,
		Bytes:   10,
		TraceID: "test",
	})

	if len(tracer.Spans()) != 1 {
		t.Fatalf("expected 1 span before clear")
	}

	tracer.Clear()
	if len(tracer.Spans()) != 0 {
		t.Fatalf("expected 0 spans after clear")
	}
}

func TestOpenTelemetryTracerConcurrency(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			_, span := tracer.Start(context.Background(), req)
			span.End(observability.RequestMetrics{
				Status:  http.StatusOK,
				Bytes:   10,
				TraceID: "concurrent",
			})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	spans := tracer.Spans()
	if len(spans) != 10 {
		t.Fatalf("expected 10 spans, got %d", len(spans))
	}
}

func TestOpenTelemetryTracerUsesContextTraceID(t *testing.T) {
	tracer := NewOpenTelemetryTracer("plumego-test")

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	ctx := context.WithValue(context.Background(), contract.TraceIDKey{}, "trace-ctx")
	ctx, span := tracer.Start(ctx, req)

	spanCtx, ok := span.(interface {
		TraceID() string
		SpanID() string
	})
	if !ok {
		t.Fatalf("expected span to expose context identifiers")
	}
	if spanCtx.TraceID() != "trace-ctx" {
		t.Fatalf("expected trace id from context, got %s", spanCtx.TraceID())
	}

	traceCtx := contract.TraceContextFromContext(ctx)
	if traceCtx == nil || string(traceCtx.TraceID) != "trace-ctx" {
		t.Fatalf("expected trace context to be set from context")
	}
	if traceCtx.SpanID == "" {
		t.Fatalf("expected span id in trace context")
	}
}

func TestSpanIDGeneration(t *testing.T) {
	spanID1 := generateSpanID()
	spanID2 := generateSpanID()

	if spanID1 == spanID2 {
		t.Fatalf("span IDs should be unique")
	}
	if len(spanID1) == 0 || len(spanID2) == 0 {
		t.Fatalf("span IDs should not be empty")
	}
}

func TestTraceIDGeneration(t *testing.T) {
	traceID1 := generateTraceID()
	traceID2 := generateTraceID()

	if traceID1 == traceID2 {
		t.Fatalf("trace IDs should be unique")
	}
	if len(traceID1) == 0 || len(traceID2) == 0 {
		t.Fatalf("trace IDs should not be empty")
	}
}
