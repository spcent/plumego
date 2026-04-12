package pubsubdebug

import (
	"net/http"
	"strings"
	"sync"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/pubsub"
)

type Handler struct {
	cfg        PubSubConfig
	defaultPub pubsub.Broker
	routesOnce sync.Once
}

func New(cfg PubSubConfig, fallbackPub pubsub.Broker) *Handler {
	return &Handler{cfg: cfg, defaultPub: fallbackPub}
}

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

func (h *Handler) RegisterRoutes(r routeRegistrar) error {
	if !h.cfg.Enabled {
		return nil
	}

	var regErr error
	h.routesOnce.Do(func() {
		pub := h.cfg.Pub
		if pub == nil {
			pub = h.defaultPub
		}
		path := strings.TrimSpace(h.cfg.Path)
		if path == "" {
			path = "/_debug/pubsub"
		}

		regErr = r.AddRoute(http.MethodGet, path, adaptCtx(func(ctx *contract.Ctx) {
			if pub == nil {
				_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
					Status(http.StatusInternalServerError).
					Category(contract.CategoryServer).
					Type(contract.TypeInternal).
					Code(contract.CodeInternalError).
					Message("pubsub is not configured").
					Build())
				return
			}

			type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

			if ps, ok := pub.(snapshoter); ok {
				_ = contract.WriteResponse(ctx.W, ctx.R, http.StatusOK, ps.Snapshot(), nil)
				return
			}

			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().Status(http.StatusNotImplemented).Code("not_supported").Message("pubsub snapshot not supported by this implementation").Category(contract.CategoryServer).Build())
		}))
	})
	return regErr
}

func (h *Handler) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{
		Status:  health.StatusHealthy,
		Details: map[string]any{"enabled": h.cfg.Enabled},
	}

	if !h.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "pubsub debug disabled"
	}

	return "pubsub_debug", status
}
