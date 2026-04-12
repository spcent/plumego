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
