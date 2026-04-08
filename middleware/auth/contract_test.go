package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/security/authn"
)

type staticAuthenticator struct {
	principal *authn.Principal
	err       error
}

func (s staticAuthenticator) Authenticate(_ *http.Request) (*authn.Principal, error) {
	return s.principal, s.err
}

type contextMarkerKey struct{}

type enrichingAuthenticator struct {
	principal *authn.Principal
}

func (e enrichingAuthenticator) Authenticate(_ *http.Request) (*authn.Principal, error) {
	return nil, nil
}

func (e enrichingAuthenticator) AuthenticateRequest(r *http.Request) (*authn.Principal, *http.Request, error) {
	ctx := context.WithValue(r.Context(), contextMarkerKey{}, "enriched")
	return e.principal, r.WithContext(ctx), nil
}

type staticAuthorizer struct {
	err error
}

func (s staticAuthorizer) Authorize(_ *authn.Principal, _, _ string) error {
	return s.err
}

func TestAuthenticateMiddlewareSuccess(t *testing.T) {
	principal := &authn.Principal{Subject: "user-1"}
	authenticator := staticAuthenticator{principal: principal}

	handler := Authenticate(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := authn.PrincipalFromRequest(r)
		if got == nil || got.Subject != "user-1" {
			t.Fatalf("expected principal in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthenticateMiddlewareInvalidToken(t *testing.T) {
	authenticator := staticAuthenticator{err: authn.ErrInvalidToken}

	handler := Authenticate(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var payload contract.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}
	if payload.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected code UNAUTHORIZED, got %s", payload.Error.Code)
	}
}

func TestAuthenticateMiddlewareUsesEnrichedRequestContext(t *testing.T) {
	authenticator := enrichingAuthenticator{
		principal: &authn.Principal{Subject: "user-2"},
	}

	handler := Authenticate(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, _ := r.Context().Value(contextMarkerKey{}).(string); got != "enriched" {
			t.Fatalf("expected enriched request context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthorizeAllowsRequest(t *testing.T) {
	principal := &authn.Principal{Subject: "user-authz"}
	authorizer := staticAuthorizer{}

	handler := Authorize(authorizer, "read", "widgets")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(authn.WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthorizeRejectsForbiddenRequest(t *testing.T) {
	principal := &authn.Principal{Subject: "user-authz"}
	authorizer := staticAuthorizer{err: authn.ErrUnauthorized}

	handler := Authorize(authorizer, "write", "widgets")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(authn.WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
