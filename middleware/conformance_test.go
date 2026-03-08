package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/middleware/limits"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/recovery"
	tenantmw "github.com/spcent/plumego/middleware/tenant"
	"github.com/spcent/plumego/middleware/versioning"
)

func TestMiddlewareErrorConformance(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode string
		handler      http.Handler
		request      *http.Request
	}{
		{
			name:         "auth unauthenticated",
			expectedCode: middleware.CodeAuthUnauthenticated,
			handler: auth.NewSimpleAuthMiddleware("secret").Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "body too large",
			expectedCode: middleware.CodeRequestBodyTooLarge,
			handler: limits.BodyLimit(4, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.ReadAll(r.Body)
			})),
			request: httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolarge")),
		},
		{
			name:         "tenant required",
			expectedCode: middleware.CodeTenantRequired,
			handler: tenantmw.TenantResolver(tenantmw.TenantResolverOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "abuse guard rate limited",
			expectedCode: middleware.CodeRateLimited,
			handler: ratelimit.AbuseGuard(ratelimit.AbuseGuardConfig{Rate: 1, Capacity: 1, KeyFunc: func(*http.Request) string { return "k" }})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "recovery internal",
			expectedCode: middleware.CodeInternalError,
			handler: recovery.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "unsupported version",
			expectedCode: middleware.CodeUnsupportedVersion,
			handler: versioning.Middleware(versioning.Config{DefaultVersion: 1, SupportedVersions: []int{2}})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
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
		"auth":               middleware.CodeAuthUnauthenticated,
		"rate":               middleware.CodeRateLimited,
		"tenant_required":    middleware.CodeTenantRequired,
		"tenant_invalid":     middleware.CodeTenantInvalidID,
		"tenant_policy":      middleware.CodeTenantPolicyDenied,
		"tenant_quota":       middleware.CodeTenantQuotaExceeded,
		"tenant_rate":        middleware.CodeTenantRateLimited,
		"body_too_large":     middleware.CodeRequestBodyTooLarge,
		"server_busy":        middleware.CodeServerBusy,
		"queue_timeout":      middleware.CodeServerQueueTimeout,
		"request_timeout":    middleware.CodeRequestTimeout,
		"upstream_failed":    middleware.CodeUpstreamFailed,
		"unsupported_ver":    middleware.CodeUnsupportedVersion,
		"transform_failed":   middleware.CodeTransformFailed,
		"protocol_transform": middleware.CodeProtocolTransformFail,
		"protocol_exec":      middleware.CodeProtocolExecutionFail,
		"internal":           middleware.CodeInternalError,
	}

	// Keep this test explicit; values are part of wire-level contract.
	if expected["auth"] != "auth_unauthenticated" ||
		expected["rate"] != "rate_limited" ||
		expected["tenant_required"] != "tenant_required" ||
		expected["tenant_invalid"] != "tenant_invalid_id" ||
		expected["tenant_policy"] != "tenant_policy_denied" ||
		expected["tenant_quota"] != "tenant_quota_exceeded" ||
		expected["tenant_rate"] != "tenant_rate_limited" ||
		expected["body_too_large"] != "request_body_too_large" ||
		expected["server_busy"] != "server_busy" ||
		expected["queue_timeout"] != "server_queue_timeout" ||
		expected["request_timeout"] != "request_timeout" ||
		expected["upstream_failed"] != "upstream_failed" ||
		expected["unsupported_ver"] != "unsupported_version" ||
		expected["transform_failed"] != "transform_failed" ||
		expected["protocol_transform"] != "protocol_transform_failed" ||
		expected["protocol_exec"] != "protocol_execution_failed" ||
		expected["internal"] != "internal_error" {
		t.Fatalf("middleware error code registry changed: %#v", expected)
	}

}
