// Example: non-canonical
//
// This demo adds x/tenant resolution, policy, quota, and rate limiting to a
// small API while keeping route registration explicit.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"github.com/spcent/plumego/x/tenant/policy"
	"github.com/spcent/plumego/x/tenant/quota"
	"github.com/spcent/plumego/x/tenant/ratelimit"
	"github.com/spcent/plumego/x/tenant/resolve"
)

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = envString("APP_ADDR", ":8085")

	app := core.New(cfg, core.AppDependencies{Logger: plumelog.NewLogger()})
	if err := app.Use(requestid.Middleware(), recovery.Recovery(app.Logger())); err != nil {
		log.Fatalf("register middleware: %v", err)
	}

	tenantConfig := tenantcore.NewInMemoryConfigManager()
	tenantConfig.SetTenantConfig(tenantcore.Config{
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

	tenantChain := middleware.NewChain(
		resolve.Middleware(resolve.Options{DisablePrincipal: true}),
		policy.Middleware(policy.Options{Evaluator: tenantcore.NewConfigPolicyEvaluator(tenantConfig)}),
		quota.Middleware(quota.Options{Manager: tenantcore.NewFixedWindowQuotaManager(tenantConfig)}),
		ratelimit.Middleware(ratelimit.Options{Limiter: tenantcore.NewTokenBucketRateLimiter(rateLimits)}),
	)

	if err := app.Get("/api/models", tenantChain.Build(http.HandlerFunc(models))); err != nil {
		log.Fatalf("register tenant route: %v", err)
	}

	if err := serve(app, cfg); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func models(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantcore.TenantIDFromContext(r.Context())
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"tenant_id": tenantID,
		"models":    []string{"gpt-4o"},
	}, nil)
}

func serve(app *core.App, cfg core.AppConfig) error {
	if err := app.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := app.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	defer app.Shutdown(context.Background())

	log.Printf("Starting with-tenant demo on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
