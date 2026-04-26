package tenant_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/security/authn"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"github.com/spcent/plumego/x/tenant/policy"
	"github.com/spcent/plumego/x/tenant/quota"
	"github.com/spcent/plumego/x/tenant/ratelimit"
	"github.com/spcent/plumego/x/tenant/resolve"
	tenanttransport "github.com/spcent/plumego/x/tenant/transport"
)

func Example_saasAPIChain() {
	cfg := tenantcore.NewInMemoryConfigManager()
	cfg.SetTenantConfig(tenantcore.Config{
		TenantID: "tenant-a",
		Policy: tenantcore.PolicyConfig{
			AllowedModels: []string{"gpt-4o"},
		},
		Quota: tenantcore.QuotaConfig{
			Limits: []tenantcore.QuotaLimit{
				{Window: tenantcore.QuotaWindowMinute, Requests: 100},
			},
		},
	})

	rateLimits := tenantcore.NewInMemoryRateLimitManager()
	rateLimits.SetRateLimit("tenant-a", tenantcore.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
	})

	handler := middleware.NewChain(
		resolve.Middleware(resolve.Options{}),
		policy.Middleware(policy.Options{Evaluator: tenantcore.NewConfigPolicyEvaluator(cfg)}),
		quota.Middleware(quota.Options{Manager: tenantcore.NewFixedWindowQuotaManager(cfg)}),
		ratelimit.Middleware(ratelimit.Options{Limiter: tenantcore.NewTokenBucketRateLimiter(rateLimits)}),
	).Build(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	allowed := httptest.NewRecorder()
	allowedReq := httptest.NewRequest(http.MethodGet, "/models", nil)
	allowedReq = allowedReq.WithContext(authn.WithPrincipal(
		allowedReq.Context(),
		&authn.Principal{TenantID: "tenant-a"},
	))
	allowedReq.Header.Set("X-Model", "gpt-4o")
	handler.ServeHTTP(allowed, allowedReq)
	fmt.Println("allowed:", allowed.Code)

	denied := httptest.NewRecorder()
	deniedReq := httptest.NewRequest(http.MethodGet, "/models", nil)
	deniedReq.Header.Set("X-Model", "gpt-4o")
	handler.ServeHTTP(denied, deniedReq)

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.NewDecoder(denied.Body).Decode(&body)
	fmt.Println("missing tenant:", denied.Code, body.Error.Code == tenanttransport.CodeRequired)

	// Output:
	// allowed: 200
	// missing tenant: 401 true
}
