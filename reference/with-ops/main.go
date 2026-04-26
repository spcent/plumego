// Example: non-canonical
//
// This demo mounts protected x/ops routes and stable request observability
// middleware without mounting x/devtools.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/ops"
)

func main() {
	addr := envString("APP_ADDR", ":8087")
	token := envString("OPS_TOKEN", "local-admin-token")
	logger := plumelog.NewLogger()
	collector := metrics.NewBaseMetricsCollector()
	r := router.NewRouter()

	if err := r.AddRoute(http.MethodGet, "/", http.HandlerFunc(root)); err != nil {
		log.Fatalf("register root route: %v", err)
	}
	if err := r.AddRoute(http.MethodGet, "/metrics", http.HandlerFunc(metricStats(collector))); err != nil {
		log.Fatalf("register metrics route: %v", err)
	}

	opsHandler := ops.New(ops.Options{
		Enabled: true,
		Auth: ops.AuthConfig{
			Token: token,
		},
		Hooks: ops.Hooks{
			QueueStats: func(ctx context.Context, queue string) (ops.QueueStats, error) {
				return ops.QueueStats{Queue: queue, Queued: 3, UpdatedAt: time.Now().UTC()}, nil
			},
		},
		Logger: logger,
	})
	if err := opsHandler.RegisterRoutes(r); err != nil {
		log.Fatalf("register ops routes: %v", err)
	}

	handler := middleware.NewChain(
		requestid.Middleware(),
		recovery.Recovery(logger),
		httpmetrics.Middleware(collector),
		accesslog.Middleware(logger, nil, nil),
	).Build(r)

	log.Printf("Starting with-ops demo on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"service":  "with-ops",
		"ops":      "/ops",
		"metrics":  "/metrics",
		"devtools": "not_mounted",
	}, nil)
}

func metricStats(collector *metrics.BaseMetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, collector.GetStats(), nil)
	}
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
