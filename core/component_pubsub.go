package core

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

type pubSubDebugComponent struct {
	BaseComponent
	cfg        PubSubConfig
	defaultPub pubsub.PubSub
	routesOnce sync.Once
}

func newPubSubDebugComponent(cfg PubSubConfig, fallbackPub pubsub.PubSub) Component {
	return &pubSubDebugComponent{cfg: cfg, defaultPub: fallbackPub}
}

func (c *pubSubDebugComponent) RegisterRoutes(r *router.Router) {
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

		r.GetCtx(path, func(ctx *contract.Ctx) {
			if pub == nil {
				writeContractError(ctx, http.StatusInternalServerError, "missing_pubsub", "pubsub is not configured")
				return
			}

			type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

			if ps, ok := pub.(snapshoter); ok {
				writeContractResponse(ctx, http.StatusOK, ps.Snapshot())
				return
			}

			writeContractError(ctx, http.StatusNotImplemented, "not_supported", "pubsub snapshot not supported by this implementation")
		})
	})
}

func (c *pubSubDebugComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *pubSubDebugComponent) Start(_ context.Context) error { return nil }

func (c *pubSubDebugComponent) Stop(_ context.Context) error { return nil }

func (c *pubSubDebugComponent) Health() (string, health.HealthStatus) {
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
