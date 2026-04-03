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

func (h *Handler) RegisterRoutes(r *router.Router) {
	if !h.cfg.Enabled {
		return
	}

	h.routesOnce.Do(func() {
		pub := h.cfg.Pub
		if pub == nil {
			pub = h.defaultPub
		}
		path := strings.TrimSpace(h.cfg.Path)
		if path == "" {
			path = "/_debug/pubsub"
		}

		r.Get(path, contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
			if pub == nil {
				_ = contract.WriteError(ctx.W, ctx.R, contract.APIError{Status: http.StatusInternalServerError, Code: "missing_pubsub", Message: "pubsub is not configured", Category: contract.CategoryForStatus(http.StatusInternalServerError)})
				return
			}

			type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

			if ps, ok := pub.(snapshoter); ok {
				_ = ctx.Response(http.StatusOK, ps.Snapshot(), nil)
				return
			}

			_ = contract.WriteError(ctx.W, ctx.R, contract.APIError{Status: http.StatusNotImplemented, Code: "not_supported", Message: "pubsub snapshot not supported by this implementation", Category: contract.CategoryForStatus(http.StatusNotImplemented)})
		}, r.Logger()))
	})
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
