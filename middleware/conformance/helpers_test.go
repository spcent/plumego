package conformance_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/security/authn"
)

type callCountingHandler struct {
	calls int
}

type canonicalErrorEnvelope struct {
	Error struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Category string `json:"category"`
	} `json:"error"`
}

func (h *callCountingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.calls++
	w.Header().Set("X-Conformance-Handler", "called")
	w.WriteHeader(http.StatusNoContent)
}

func assertCanonicalErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	if rec.Code < 400 {
		t.Fatalf("expected error status, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}

	var payload canonicalErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}
	if payload.Error.Message == "" {
		t.Fatalf("expected error.message field in payload %#v", payload)
	}
	if payload.Error.Category == "" {
		t.Fatalf("expected error.category field in payload %#v", payload)
	}
	if payload.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, payload.Error.Code)
	}
}

func execute(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func newConformanceAccessLog(t *testing.T) middleware.Middleware {
	t.Helper()
	mw, err := accesslog.Middleware(accesslog.Config{
		Logger: log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard}),
	})
	if err != nil {
		t.Fatalf("accesslog middleware: %v", err)
	}
	return mw
}

func newConformanceRecovery(t *testing.T) middleware.Middleware {
	t.Helper()
	mw, err := recovery.Middleware(recovery.Config{
		Logger: log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard}),
	})
	if err != nil {
		t.Fatalf("recovery middleware: %v", err)
	}
	return mw
}

func newConformanceAuth(t *testing.T, authenticator authn.Authenticator, opts ...auth.AuthOption) middleware.Middleware {
	t.Helper()
	mw, err := auth.Authenticate(authenticator, opts...)
	if err != nil {
		t.Fatalf("auth middleware: %v", err)
	}
	return mw
}
