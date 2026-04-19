package authn

import (
	"context"
	"testing"
)

func TestWithPrincipalRoundTrip(t *testing.T) {
	p := &Principal{
		Subject:  "user-1",
		TenantID: "tenant-a",
		Roles:    []string{"admin"},
		Scopes:   []string{"read", "write"},
	}

	ctx := WithPrincipal(t.Context(), p)
	got := PrincipalFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil principal from context")
	}
	if got.Subject != "user-1" {
		t.Errorf("Subject = %q, want user-1", got.Subject)
	}
	if got.TenantID != "tenant-a" {
		t.Errorf("TenantID = %q, want tenant-a", got.TenantID)
	}
}

func TestWithPrincipalNilContext(t *testing.T) {
	p := &Principal{Subject: "user-1"}

	ctx := WithPrincipal(nil, p)
	got := PrincipalFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil principal from nil parent context")
	}
	if got.Subject != "user-1" {
		t.Errorf("Subject = %q, want user-1", got.Subject)
	}
}

func TestPrincipalFromContextNil(t *testing.T) {
	got := PrincipalFromContext(t.Context())
	if got != nil {
		t.Fatalf("expected nil principal from empty context, got %+v", got)
	}
}

func TestPrincipalFromContextWrongType(t *testing.T) {
	ctx := context.WithValue(t.Context(), struct{}{}, "not-a-principal")
	got := PrincipalFromContext(ctx)
	if got != nil {
		t.Fatalf("expected nil for missing principal, got %+v", got)
	}
}

func TestWithPrincipalOverwrite(t *testing.T) {
	p1 := &Principal{Subject: "user-1"}
	p2 := &Principal{Subject: "user-2"}

	ctx := WithPrincipal(t.Context(), p1)
	ctx = WithPrincipal(ctx, p2)

	got := PrincipalFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil principal")
	}
	if got.Subject != "user-2" {
		t.Errorf("expected latest principal (user-2), got %q", got.Subject)
	}
}

func TestPrincipalFromContextNilPrincipalStored(t *testing.T) {
	ctx := WithPrincipal(t.Context(), nil)
	got := PrincipalFromContext(ctx)
	if got != nil {
		t.Fatalf("expected nil for explicitly stored nil principal, got %+v", got)
	}
}

func TestPrincipalFromContextNilContext(t *testing.T) {
	if got := PrincipalFromContext(nil); got != nil {
		t.Fatalf("expected nil for nil context, got %+v", got)
	}
}
