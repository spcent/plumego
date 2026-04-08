package conformance_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/limits"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
	"github.com/spcent/plumego/security/authn"
	tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
)

func TestMiddlewareTypeShape(t *testing.T) {
	var _ middleware.Middleware = fixtureNoOpMiddleware()
}

func TestMiddlewareNextCallAtMostOnce(t *testing.T) {
	recoveryMw := recovery.Recovery(log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard}))
	tests := []struct {
		name string
		mw   middleware.Middleware
		req  *http.Request
	}{
		{
			name: "noop fixture",
			mw:   fixtureNoOpMiddleware(),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "request id",
			mw:   requestid.Middleware(),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "tracing",
			mw:   mwtracing.Middleware(nil),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "http metrics",
			mw:   httpmetrics.Middleware(nil),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "access log",
			mw:   accesslog.Logging(log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard}), nil, nil),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "recovery",
			mw:   recoveryMw,
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "tenant resolver valid",
			mw: tenantresolve.Middleware(tenantresolve.Options{
				AllowMissing: true,
			}),
			req: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name: "auth valid token",
			mw:   auth.Authenticate(authn.StaticToken("secret")),
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "auth valid token" {
				tc.req.Header.Set("Authorization", "Bearer secret")
			}

			next := &callCountingHandler{}
			h := tc.mw(next)
			h.ServeHTTP(httptest.NewRecorder(), tc.req)

			if next.calls > 1 {
				t.Fatalf("middleware invoked next more than once: %d", next.calls)
			}
		})
	}
}

func TestMiddlewareNoSuccessResponseEmission(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "handler-success")
	})

	h := fixtureNoOpMiddleware()(next)
	rec := execute(h, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status from handler, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "handler-success" {
		t.Fatalf("expected body from handler, got %q", rec.Body.String())
	}
}

func TestMiddlewareOrderingDeterministic(t *testing.T) {
	order := make([]string, 0, 3)

	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first:before")
			next.ServeHTTP(w, r)
			order = append(order, "first:after")
		})
	}
	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second:before")
			next.ServeHTTP(w, r)
			order = append(order, "second:after")
		})
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusNoContent)
	})

	h := middleware.Apply(final, first, second)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"first:before", "second:before", "handler", "second:after", "first:after"}
	if len(order) != len(want) {
		t.Fatalf("order length mismatch: got %v want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected execution order: got %v want %v", order, want)
		}
	}
}

func TestMiddlewareErrorSchemaCanonical(t *testing.T) {
	recoveryLogger := log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard})
	tests := []struct {
		name         string
		expectedCode string
		handler      http.Handler
		request      *http.Request
		prepare      func(http.Handler, *http.Request)
	}{
		{
			name:         "auth unauthenticated",
			expectedCode: middleware.CodeAuthUnauthenticated,
			handler: auth.Authenticate(authn.StaticToken("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			handler: tenantresolve.Middleware(tenantresolve.Options{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			prepare: func(handler http.Handler, req *http.Request) {
				handler.ServeHTTP(httptest.NewRecorder(), req.Clone(req.Context()))
			},
		},
		{
			name:         "recovery internal",
			expectedCode: middleware.CodeInternalError,
			handler: recovery.Recovery(recoveryLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:         "fixture canonical error",
			expectedCode: middleware.CodeInternalError,
			handler: fixtureCanonicalErrorMiddleware(middleware.CodeInternalError)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			request: httptest.NewRequest(http.MethodGet, "/", nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(tc.handler, tc.request)
			}

			rec := execute(tc.handler, tc.request)
			assertCanonicalErrorEnvelope(t, rec, tc.expectedCode)
		})
	}
}
