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
				contract.WriteContractError(ctx, http.StatusInternalServerError, "missing_pubsub", "pubsub is not configured")
				return
			}

			type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

			if ps, ok := pub.(snapshoter); ok {
				contract.WriteContractResponse(ctx, http.StatusOK, ps.Snapshot())
				return
			}

			contract.WriteContractError(ctx, http.StatusNotImplemented, "not_supported", "pubsub snapshot not supported by this implementation")
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
