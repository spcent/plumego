package pubsubdebug

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core/internal/contractio"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

type PubSubDebugComponent struct {
	cfg        PubSubConfig
	defaultPub pubsub.PubSub
	routesOnce sync.Once
}

func NewPubSubDebugComponent(cfg PubSubConfig, fallbackPub pubsub.PubSub) *PubSubDebugComponent {
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

		r.GetCtx(path, func(ctx *contract.Ctx) {
			if pub == nil {
				contractio.WriteContractError(ctx, http.StatusInternalServerError, "missing_pubsub", "pubsub is not configured")
				return
			}

			type snapshoter interface{ Snapshot() pubsub.MetricsSnapshot }

			if ps, ok := pub.(snapshoter); ok {
				contractio.WriteContractResponse(ctx, http.StatusOK, ps.Snapshot())
				return
			}

			contractio.WriteContractError(ctx, http.StatusNotImplemented, "not_supported", "pubsub snapshot not supported by this implementation")
		})
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

func (c *PubSubDebugComponent) Dependencies() []reflect.Type { return nil }
