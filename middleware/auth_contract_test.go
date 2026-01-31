package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

type staticAuthenticator struct {
	principal *contract.Principal
	err       error
}

func (s staticAuthenticator) Authenticate(_ *http.Request) (*contract.Principal, error) {
	return s.principal, s.err
}

type staticAuthorizer struct {
	err error
}

func (s staticAuthorizer) Authorize(_ *contract.Principal, _, _ string) error {
	return s.err
}

type stubSessionStore struct {
	session *contract.Session
	err     error
}

func (s *stubSessionStore) CreateSession(context.Context, *contract.Session) error {
	return nil
}

func (s *stubSessionStore) GetSession(context.Context, string) (*contract.Session, error) {
	return s.session, s.err
}

func (s *stubSessionStore) RotateRefresh(context.Context, string, string, string, time.Time) (*contract.Session, error) {
	return nil, nil
}

func (s *stubSessionStore) RevokeSession(context.Context, string, time.Time, string) error {
	return nil
}

func (s *stubSessionStore) ListSessionsBySubject(context.Context, string, string, int) ([]*contract.Session, error) {
	return nil, nil
}

type staticSessionValidator struct {
	err error
}

func (s staticSessionValidator) ValidateSession(context.Context, *contract.Session, time.Time) error {
	return s.err
}

func (s staticSessionValidator) ValidateRefresh(context.Context, *contract.Session, contract.RefreshClaims, time.Time) error {
	return nil
}

func TestAuthenticateMiddlewareSuccess(t *testing.T) {
	principal := &contract.Principal{Subject: "user-1"}
	authenticator := staticAuthenticator{principal: principal}

	handler := Authenticate(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := contract.PrincipalFromRequest(r)
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
	authenticator := staticAuthenticator{err: contract.ErrInvalidToken}

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

func TestAuthorizeMiddleware(t *testing.T) {
	principal := &contract.Principal{Subject: "user-authz"}
	authorizer := staticAuthorizer{}

	handler := Authorize(authorizer, "read", "widgets")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(contract.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthorizeMiddlewareForbidden(t *testing.T) {
	principal := &contract.Principal{Subject: "user-authz"}
	authorizer := staticAuthorizer{err: contract.ErrUnauthorized}

	handler := Authorize(authorizer, "write", "widgets")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(contract.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestSessionCheckMiddleware(t *testing.T) {
	principal := &contract.Principal{
		Subject: "user-1",
		Claims:  map[string]string{"session_id": "sess-1"},
	}
	store := &stubSessionStore{session: &contract.Session{SessionID: "sess-1"}}
	validator := staticSessionValidator{}

	handler := SessionCheck(store, validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(contract.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSessionCheckMissingSessionID(t *testing.T) {
	principal := &contract.Principal{Subject: "user-1"}
	store := &stubSessionStore{}
	validator := staticSessionValidator{}

	handler := SessionCheck(store, validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(contract.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSessionCheckRevokedSession(t *testing.T) {
	principal := &contract.Principal{
		Subject: "user-1",
		Claims:  map[string]string{"session_id": "sess-1"},
	}
	store := &stubSessionStore{session: &contract.Session{SessionID: "sess-1"}}
	validator := staticSessionValidator{err: contract.ErrSessionRevoked}

	handler := SessionCheck(store, validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(contract.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
