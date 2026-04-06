package contract

import (
	"testing"
)

func TestRequestIDContextRoundTrip(t *testing.T) {
	ctx := WithRequestID(t.Context(), "req-123")
	if got := RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("expected request id %q, got %q", "req-123", got)
	}
}

func TestRequestIDFromContextNilSafe(t *testing.T) {
	if got := RequestIDFromContext(t.Context()); got != "" {
		t.Fatalf("expected empty request id for nil context, got %q", got)
	}
	if got := RequestIDFromContext(t.Context()); got != "" {
		t.Fatalf("expected empty request id for empty context, got %q", got)
	}
}
