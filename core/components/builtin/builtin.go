package builtin

import (
	"context"
	"reflect"

	"github.com/spcent/plumego/core/components/devtools"
	"github.com/spcent/plumego/core/components/pubsubdebug"
	"github.com/spcent/plumego/core/components/webhook"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

// Component mirrors core.Component for selection purposes.
type Component interface {
	RegisterRoutes(r *router.Router)
	RegisterMiddleware(m *middleware.Registry)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() (name string, status health.HealthStatus)
	Dependencies() []reflect.Type
}

// Hooks provide the wiring points needed to build default components.
type Hooks struct {
	Debug            bool
	PubSubConfig     pubsubdebug.PubSubConfig
	WebhookOutConfig webhook.WebhookOutConfig
	WebhookInConfig  webhook.WebhookInConfig
	PubSub           pubsub.PubSub
	Logger           log.StructuredLogger

	HasComponentType func(target any) bool

	NewDevTools    func() Component
	NewPubSubDebug func(cfg pubsubdebug.PubSubConfig, fallback pubsub.PubSub) Component
	NewWebhookOut  func(cfg webhook.WebhookOutConfig) Component
	NewWebhookIn   func(cfg webhook.WebhookInConfig, fallback pubsub.PubSub, logger log.StructuredLogger) Component
}

// BuiltInComponents selects and builds the default component set.
func BuiltInComponents(h Hooks) []Component {
	var comps []Component

	if h.HasComponentType == nil {
		return comps
	}

	if h.Debug && !h.HasComponentType((*devtools.DevToolsComponent)(nil)) {
		if h.NewDevTools != nil {
			comps = append(comps, h.NewDevTools())
		}
	}

	if h.PubSubConfig.Enabled && !h.HasComponentType((*pubsubdebug.PubSubDebugComponent)(nil)) {
		if h.NewPubSubDebug != nil {
			comps = append(comps, h.NewPubSubDebug(h.PubSubConfig, h.PubSub))
		}
	}

	if h.WebhookOutConfig.Enabled && !h.HasComponentType((*webhook.WebhookOutComponent)(nil)) {
		if h.NewWebhookOut != nil {
			comps = append(comps, h.NewWebhookOut(h.WebhookOutConfig))
		}
	}

	if h.WebhookInConfig.Enabled && !h.HasComponentType((*webhook.WebhookInComponent)(nil)) {
		if h.NewWebhookIn != nil {
			comps = append(comps, h.NewWebhookIn(h.WebhookInConfig, h.PubSub, h.Logger))
		}
	}

	return comps
}
