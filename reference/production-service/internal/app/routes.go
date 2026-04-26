package app

import (
	"net/http"

	"github.com/spcent/plumego/contract"
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
	if err := a.Core.Get("/ops/metrics", http.HandlerFunc(a.metricStats)); err != nil {
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
	Status     string          `json:"status"`
	Service    string          `json:"service"`
	Profile    string          `json:"profile"`
	Timestamp  string          `json:"timestamp"`
	Middleware []string        `json:"middleware"`
	Limits     statusLimits    `json:"limits"`
	Ops        statusOpsPolicy `json:"ops"`
}

type statusLimits struct {
	BodyLimitBytes int64   `json:"body_limit_bytes"`
	RequestTimeout string  `json:"request_timeout"`
	RateLimit      float64 `json:"rate_limit"`
	RateBurst      int     `json:"rate_burst"`
}

type statusOpsPolicy struct {
	MetricsRoute string `json:"metrics_route"`
	Devtools     string `json:"devtools"`
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
		Limits: statusLimits{
			BodyLimitBytes: a.Cfg.App.BodyLimitBytes,
			RequestTimeout: a.Cfg.App.RequestTimeout.String(),
			RateLimit:      a.Cfg.App.RateLimit,
			RateBurst:      a.Cfg.App.RateBurst,
		},
		Ops: statusOpsPolicy{
			MetricsRoute: "/ops/metrics",
			Devtools:     "not_mounted_by_default",
		},
	}, nil)
}

func (a *App) metricStats(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, a.Metrics.GetStats(), nil)
}
