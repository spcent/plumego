package authn

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestWithPrincipalRoundTrip(t *testing.T) {
	p := &Principal{
		Subject:  "user-1",
		TenantID: "tenant-a",
		Roles:    []string{"admin"},
		Scopes:   []string{"read", "write"},
	}

	ctx := WithPrincipal(context.Background(), p)
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

func TestPrincipalFromContextNil(t *testing.T) {
	got := PrincipalFromContext(context.Background())
	if got != nil {
		t.Fatalf("expected nil principal from empty context, got %+v", got)
	}
}

func TestPrincipalFromContextWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "not-a-principal")
	got := PrincipalFromContext(ctx)
	if got != nil {
		t.Fatalf("expected nil for missing principal, got %+v", got)
	}
}

func TestPrincipalFromRequest(t *testing.T) {
	p := &Principal{Subject: "req-user"}

	req := httptest.NewRequest("GET", "/", nil)
	req = RequestWithPrincipal(req, p)

	got := PrincipalFromRequest(req)
	if got == nil {
		t.Fatal("expected non-nil principal from request")
	}
	if got.Subject != "req-user" {
		t.Errorf("Subject = %q, want req-user", got.Subject)
	}
}

func TestPrincipalFromRequestNil(t *testing.T) {
	got := PrincipalFromRequest(nil)
	if got != nil {
		t.Fatalf("expected nil for nil request, got %+v", got)
	}
}

func TestPrincipalFromRequestNoAttached(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	got := PrincipalFromRequest(req)
	if got != nil {
		t.Fatalf("expected nil for request without principal, got %+v", got)
	}
}

func TestRequestWithPrincipalNil(t *testing.T) {
	got := RequestWithPrincipal(nil, &Principal{Subject: "u"})
	if got != nil {
		t.Fatalf("expected nil for nil request, got non-nil")
	}
}

func TestWithPrincipalOverwrite(t *testing.T) {
	p1 := &Principal{Subject: "user-1"}
	p2 := &Principal{Subject: "user-2"}

	ctx := WithPrincipal(context.Background(), p1)
	ctx = WithPrincipal(ctx, p2)

	got := PrincipalFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil principal")
	}
	if got.Subject != "user-2" {
		t.Errorf("expected latest principal (user-2), got %q", got.Subject)
	}
}

func TestRequestWithPrincipalNewRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/data", nil)
	original := req

	p := &Principal{Subject: "svc-account", Roles: []string{"service"}}
	newReq := RequestWithPrincipal(req, p)

	if PrincipalFromRequest(original) != nil {
		t.Fatal("original request should not have principal attached")
	}

	got := PrincipalFromRequest(newReq)
	if got == nil {
		t.Fatal("new request should have principal attached")
	}
	if got.Subject != "svc-account" {
		t.Errorf("Subject = %q, want svc-account", got.Subject)
	}
}

func TestPrincipalFromContextNilPrincipalStored(t *testing.T) {
	ctx := WithPrincipal(context.Background(), nil)
	got := PrincipalFromContext(ctx)
	if got != nil {
		t.Fatalf("expected nil for explicitly stored nil principal, got %+v", got)
	}
}
