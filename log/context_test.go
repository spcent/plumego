package log

import (
	"context"
	"testing"
)

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "test-trace-123"

	ctx = WithTraceID(ctx, traceID)
	extracted := TraceIDFromContext(ctx)

	if extracted != traceID {
		t.Errorf("expected trace ID %q, got %q", traceID, extracted)
	}
}

func TestTraceIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	extracted := TraceIDFromContext(ctx)

	if extracted != "" {
		t.Errorf("expected empty trace ID, got %q", extracted)
	}
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()
	traceID := "chain-trace-123"

	// Chain context operations
	ctx = WithTraceID(ctx, traceID)

	// Verify the trace id is preserved
	extractedTraceID := TraceIDFromContext(ctx)

	if extractedTraceID != traceID {
		t.Errorf("expected trace ID %q, got %q", traceID, extractedTraceID)
	}
}

func TestNilContextSafety(t *testing.T) {
	ctx := WithTraceID(nil, "trace-nil")
	if got := TraceIDFromContext(ctx); got != "trace-nil" {
		t.Fatalf("expected trace id from nil context wrapper, got %q", got)
	}

	if got := TraceIDFromContext(nil); got != "" {
		t.Fatalf("expected empty trace id for nil context, got %q", got)
	}
}
