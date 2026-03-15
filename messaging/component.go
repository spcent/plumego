package messaging

import (
	"context"
	"reflect"

	"github.com/spcent/plumego/contract"
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
	// Send
	r.Post(c.prefix+"/send", contract.AdaptCtxHandler(c.svc.HandleSend, r.Logger()))
	r.Post(c.prefix+"/batch", contract.AdaptCtxHandler(c.svc.HandleBatchSend, r.Logger()))
	// Query
	r.Get(c.prefix+"/stats", contract.AdaptCtxHandler(c.svc.HandleStats, r.Logger()))
	r.Get(c.prefix+"/receipts", contract.AdaptCtxHandler(c.svc.HandleListReceipts, r.Logger()))
	r.Get(c.prefix+"/:id/receipt", contract.AdaptCtxHandler(c.svc.HandleGetReceipt, r.Logger()))
	// Operations
	r.Get(c.prefix+"/channels", contract.AdaptCtxHandler(c.svc.HandleChannelHealth, r.Logger()))
}

func (c *Component) RegisterMiddleware(_ *middleware.Chain) {}

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
	// Check channel health.
	for _, ch := range c.svc.Monitor().Status() {
		if ch.State == ChannelUnhealthy {
			return "messaging", health.HealthStatus{
				Status:  health.StatusDegraded,
				Message: string(ch.Channel) + " channel unhealthy: " + ch.Error,
			}
		}
	}
	return "messaging", health.HealthStatus{Status: health.StatusHealthy}
}

func (c *Component) Dependencies() []reflect.Type {
	return nil
}
