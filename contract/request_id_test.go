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

func TestRequestIDTrimsWhitespace(t *testing.T) {
	ctx := WithRequestID(t.Context(), "  req-123  ")
	if got := RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("expected trimmed request id %q, got %q", "req-123", got)
	}
}

func TestRequestIDRejectsControlCharacters(t *testing.T) {
	tests := []string{
		"req\n123",
		"req\t123",
		"req\x00123",
		"req\x7f123",
	}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			ctx := WithRequestID(t.Context(), id)
			if got := RequestIDFromContext(ctx); got != "" {
				t.Fatalf("expected unsafe request id to be ignored, got %q", got)
			}
		})
	}
}

func TestRequestIDAllowsVisiblePunctuation(t *testing.T) {
	id := "req-123_ABC.:/+="
	ctx := WithRequestID(t.Context(), id)
	if got := RequestIDFromContext(ctx); got != id {
		t.Fatalf("expected request id %q, got %q", id, got)
	}
}
