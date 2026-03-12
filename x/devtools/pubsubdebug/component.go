package pubsubdebug

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/pubsub"
)

type PubSubDebugComponent struct {
	cfg        PubSubConfig
	defaultPub pubsub.Broker
	routesOnce sync.Once
}

func NewPubSubDebugComponent(cfg PubSubConfig, fallbackPub pubsub.Broker) *PubSubDebugComponent {
	return &PubSubDebugComponent{cfg: cfg, defaultPub: fallbackPub}
}

func (c *PubSubDebugComponent) RegisterRoutes(r *router.Router) {
	if !c.cfg.Enabled {
		return
	}

	c.routesOnce.Do(func() {
		pub := c.cfg.Pub
		if pub == nil {
			pub = c.defaultPub
		}
		path := strings.TrimSpace(c.cfg.Path)
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

func (c *PubSubDebugComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *PubSubDebugComponent) Start(_ context.Context) error { return nil }

func (c *PubSubDebugComponent) Stop(_ context.Context) error { return nil }

func (c *PubSubDebugComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{
		Status:  health.StatusHealthy,
		Details: map[string]any{"enabled": c.cfg.Enabled},
	}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "pubsub_debug", status
}
