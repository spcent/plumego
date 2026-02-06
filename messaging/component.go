package messaging

import (
	"context"
	"reflect"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// Component wraps Service as a core.Component for plug-and-play
// registration via core.WithComponent.
type Component struct {
	svc    *Service
	prefix string
}

// NewComponent creates a Component that registers routes under prefix.
// A typical prefix is "/api/v1/messages".
func NewComponent(svc *Service, prefix string) *Component {
	if prefix == "" {
		prefix = "/api/v1/messages"
	}
	return &Component{svc: svc, prefix: prefix}
}

func (c *Component) RegisterRoutes(r *router.Router) {
	r.PostCtx(c.prefix+"/send", c.svc.HandleSend)
	r.PostCtx(c.prefix+"/batch", c.svc.HandleBatchSend)
	r.GetCtx(c.prefix+"/stats", c.svc.HandleStats)
}

func (c *Component) RegisterMiddleware(_ *middleware.Registry) {}

func (c *Component) Start(ctx context.Context) error {
	return c.svc.Start(ctx)
}

func (c *Component) Stop(ctx context.Context) error {
	return c.svc.Stop(ctx)
}

func (c *Component) Health() (string, health.HealthStatus) {
	stats, err := c.svc.Stats(context.Background())
	if err != nil {
		return "messaging", health.HealthStatus{Status: health.StatusDegraded, Message: err.Error()}
	}
	if stats.Dead > 100 {
		return "messaging", health.HealthStatus{Status: health.StatusDegraded, Message: "high dead-letter count"}
	}
	return "messaging", health.HealthStatus{Status: health.StatusHealthy}
}

func (c *Component) Dependencies() []reflect.Type {
	return nil
}
