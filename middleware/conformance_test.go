package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/security/authn"
	tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
)

func TestMiddlewareErrorConformance(t *testing.T) {
	recoveryLogger := log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard})
	tests := []struct {
		name         string
		expectedCode string
		handler      http.Handler
		request      *http.Request
	}{
		{
			name:         "auth unauthenticated",
			expectedCode: contract.CodeUnauthorized,
			handler: auth.Authenticate(authn.StaticToken("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "body too large",
			expectedCode: contract.CodeRequestBodyTooLarge,
			handler: bodylimit.BodyLimit(4, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.ReadAll(r.Body)
			})),
			request: httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolarge")),
		},
		{
			name:         "tenant required",
			expectedCode: middleware.CodeTenantRequired,
			handler: tenantresolve.Middleware(tenantresolve.Options{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "abuse guard rate limited",
			expectedCode: contract.CodeRateLimited,
			handler: ratelimit.AbuseGuard(ratelimit.AbuseGuardConfig{Rate: 1, Capacity: 1, KeyFunc: func(*http.Request) string { return "k" }})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "recovery internal",
			expectedCode: contract.CodeInternalError,
			handler: recovery.Recovery(recoveryLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if tc.name == "abuse guard rate limited" {
				tc.handler.ServeHTTP(httptest.NewRecorder(), tc.request.Clone(tc.request.Context()))
			}

			tc.handler.ServeHTTP(rec, tc.request)

			assertCanonicalEnvelope(t, rec, tc.expectedCode)
		})
	}
}

func assertCanonicalEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	if rec.Code < 400 {
		t.Fatalf("expected error status, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	errorObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected top-level error object, got %#v", payload)
	}

	for _, field := range []string{"code", "message", "category"} {
		if _, exists := errorObj[field]; !exists {
			t.Fatalf("expected field %q in error payload: %#v", field, errorObj)
		}
	}

	if code, _ := errorObj["code"].(string); code != expectedCode {
		t.Fatalf("expected code %q, got %q", expectedCode, code)
	}
}

func TestMiddlewareCodeRegistryStability(t *testing.T) {
	expected := map[string]string{
		"tenant_required": middleware.CodeTenantRequired,
		"tenant_invalid":  middleware.CodeTenantInvalidID,
		"tenant_policy":   middleware.CodeTenantPolicyDenied,
		"tenant_quota":    middleware.CodeTenantQuotaExceeded,
		"tenant_rate":     middleware.CodeTenantRateLimited,
		"server_busy":     middleware.CodeServerBusy,
		"queue_timeout":   middleware.CodeServerQueueTimeout,
		"upstream_failed": middleware.CodeUpstreamFailed,
	}

	// Keep this test explicit; values are part of wire-level contract.
	if expected["tenant_required"] != "tenant_required" ||
		expected["tenant_invalid"] != "tenant_invalid_id" ||
		expected["tenant_policy"] != "tenant_policy_denied" ||
		expected["tenant_quota"] != "tenant_quota_exceeded" ||
		expected["tenant_rate"] != "tenant_rate_limited" ||
		expected["server_busy"] != "server_busy" ||
		expected["queue_timeout"] != "server_queue_timeout" ||
		expected["upstream_failed"] != "upstream_failed" {
		t.Fatalf("middleware error code registry changed: %#v", expected)
	}

}
