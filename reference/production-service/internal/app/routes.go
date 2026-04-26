package app

import (
	"net/http"
	"os"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/security/authn"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"github.com/spcent/plumego/x/tenant/resolve"
)

// RegisterRoutes wires all HTTP routes for the production reference.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/", http.HandlerFunc(a.root)); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", http.HandlerFunc(a.live)); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", http.HandlerFunc(a.ready)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/status", http.HandlerFunc(a.status)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/profile", a.protectedTenantAPIHandler(http.HandlerFunc(a.profile))); err != nil {
		return err
	}
	if err := a.Core.Get("/ops/metrics", a.protectedOpsHandler(http.HandlerFunc(a.metricStats))); err != nil {
		return err
	}
	return nil
}

type serviceResponse struct {
	Service   string   `json:"service"`
	Mode      string   `json:"mode"`
	Timestamp string   `json:"timestamp"`
	Features  []string `json:"features"`
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Check     string `json:"check"`
	Timestamp string `json:"timestamp"`
}

type statusResponse struct {
	Status     string                 `json:"status"`
	Service    string                 `json:"service"`
	Profile    string                 `json:"profile"`
	Timestamp  string                 `json:"timestamp"`
	Middleware []string               `json:"middleware"`
	Deployment statusDeploymentPolicy `json:"deployment"`
	Limits     statusLimits           `json:"limits"`
	Security   statusSecurityPolicy   `json:"security"`
	Storage    statusStoragePolicy    `json:"storage"`
	API        statusAPIPolicy        `json:"api"`
	Ops        statusOpsPolicy        `json:"ops"`
}

type statusDeploymentPolicy struct {
	Environment string `json:"environment"`
	Config      string `json:"config"`
	Secrets     string `json:"secrets"`
}

type statusLimits struct {
	BodyLimitBytes int64   `json:"body_limit_bytes"`
	RequestTimeout string  `json:"request_timeout"`
	RateLimit      float64 `json:"rate_limit"`
	RateBurst      int     `json:"rate_burst"`
}

type statusOpsPolicy struct {
	HealthRoutes string `json:"health_routes"`
	MetricsRoute string `json:"metrics_route"`
	OpsAuth      string `json:"ops_auth"`
	Devtools     string `json:"devtools"`
}

type statusSecurityPolicy struct {
	APIToken string `json:"api_token"`
	OpsToken string `json:"ops_token"`
}

type statusStoragePolicy struct {
	ProfileStore string `json:"profile_store"`
	Replacement  string `json:"replacement"`
}

type statusAPIPolicy struct {
	ProfileRoute string `json:"profile_route"`
	Auth         string `json:"auth"`
	Tenant       string `json:"tenant"`
	Storage      string `json:"storage"`
}

func (a *App) root(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, serviceResponse{
		Service:   a.Cfg.App.ServiceName,
		Mode:      "production-reference",
		Timestamp: utcNow(),
		Features: []string{
			"explicit_middleware",
			"security_headers",
			"abuse_guard",
			"request_metrics",
			"protected_tenant_api",
			"no_default_devtools",
		},
	}, nil)
}

func (a *App) live(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   a.Cfg.App.ServiceName,
		Check:     "liveness",
		Timestamp: utcNow(),
	}, nil)
}

func (a *App) ready(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   a.Cfg.App.ServiceName,
		Check:     "readiness",
		Timestamp: utcNow(),
	}, nil)
}

func (a *App) status(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, statusResponse{
		Status:    "healthy",
		Service:   a.Cfg.App.ServiceName,
		Profile:   "production-reference",
		Timestamp: utcNow(),
		Middleware: []string{
			"requestid",
			"recovery",
			"bodylimit",
			"timeout",
			"security_headers",
			"abuse_guard",
			"tracing",
			"httpmetrics",
			"accesslog",
		},
		Deployment: statusDeploymentPolicy{
			Environment: a.Cfg.App.Environment,
			Config:      "environment_and_flags",
			Secrets:     "environment_only_no_fallback_for_tokens",
		},
		Limits: statusLimits{
			BodyLimitBytes: a.Cfg.App.BodyLimitBytes,
			RequestTimeout: a.Cfg.App.RequestTimeout.String(),
			RateLimit:      a.Cfg.App.RateLimit,
			RateBurst:      a.Cfg.App.RateBurst,
		},
		Security: statusSecurityPolicy{
			APIToken: "APP_API_TOKEN required for /api/profile",
			OpsToken: "OPS_TOKEN required for /ops/metrics",
		},
		Storage: statusStoragePolicy{
			ProfileStore: "app_local_in_memory_reference",
			Replacement:  "replace profileStore behind App.Profiles in internal/app",
		},
		API: statusAPIPolicy{
			ProfileRoute: "/api/profile",
			Auth:         "bearer_token_required",
			Tenant:       "X-Tenant-ID required",
			Storage:      "see storage.profile_store",
		},
		Ops: statusOpsPolicy{
			HealthRoutes: "/healthz and /readyz are public by default",
			MetricsRoute: "/ops/metrics",
			OpsAuth:      "bearer_token_required",
			Devtools:     "not_mounted_by_default",
		},
	}, nil)
}

func (a *App) profile(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantcore.TenantIDFromContext(r.Context())
	profile, ok := a.Profiles.Get(tenantID)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("tenant profile not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, profile, nil)
}

func (a *App) metricStats(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, a.Metrics.GetStats(), nil)
}

func (a *App) protectedTenantAPIHandler(next http.Handler) http.Handler {
	return middleware.NewChain(
		auth.Authenticate(
			authn.StaticToken(a.Cfg.App.APIToken),
			auth.WithAuthRealm("production-api"),
		),
		resolve.Middleware(resolve.Options{}),
	).Build(next)
}

func (a *App) protectedOpsHandler(next http.Handler) http.Handler {
	return auth.Authenticate(
		authn.StaticToken(os.Getenv("OPS_TOKEN")),
		auth.WithAuthRealm("production-ops"),
	)(next)
}
