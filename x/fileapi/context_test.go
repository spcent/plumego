package fileapi

import (
	"testing"
)

func TestWithUserID(t *testing.T) {
	ctx := WithUserID(t.Context(), "user-123")

	if got := UserIDFromContext(ctx); got != "user-123" {
		t.Fatalf("UserIDFromContext() = %q, want %q", got, "user-123")
	}
}

func TestWithUserID_NilContext(t *testing.T) {
	ctx := WithUserID(nil, "user-123")

	if ctx == nil {
		t.Fatal("WithUserID(nil, ...) returned nil context")
	}
	if got := UserIDFromContext(ctx); got != "user-123" {
		t.Fatalf("UserIDFromContext() = %q, want %q", got, "user-123")
	}
}

func TestUserIDFromContext_Missing(t *testing.T) {
	if got := UserIDFromContext(t.Context()); got != "" {
		t.Fatalf("UserIDFromContext() = %q, want empty string", got)
	}
}
