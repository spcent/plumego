package log

import (
	"context"
	"testing"
)

func TestWithRequestIDRoundTrip(t *testing.T) {
	if id, ok := RequestIDFromContext(context.Background()); ok || id != "" {
		t.Fatalf("expected no request id on empty context, got %q (%v)", id, ok)
	}

	ctx := WithRequestID(context.Background(), "req-42")
	id, ok := RequestIDFromContext(ctx)
	if !ok || id != "req-42" {
		t.Fatalf("expected carrier to return req-42, got %q (%v)", id, ok)
	}
}

func TestRequestIDFromContextNilSafe(t *testing.T) {
	if id, ok := RequestIDFromContext(nil); ok || id != "" {
		t.Fatalf("expected nil-safe empty result, got %q (%v)", id, ok)
	}
}
