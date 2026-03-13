package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

func TestMiddlewareDenied(t *testing.T) {
	evaluator := tenantcore.NewConfigPolicyEvaluator(&tenantcore.InMemoryConfigManager{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = tenantcore.RequestWithTenantID(req, "t-1")
	req.Header.Set("X-Model", "gpt-4o")
	rec := httptest.NewRecorder()

	mw := Middleware(Options{Evaluator: evaluator})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}
