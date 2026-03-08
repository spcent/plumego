package core

import (
	pubsubdebug "github.com/spcent/plumego/core/components/pubsubdebug"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

// ConfigurePubSub registers a Pub/Sub debug snapshot endpoint when enabled.
//
// Extension capability: not part of the minimal core HTTP runtime. For
// production pub/sub setup, use core.WithComponent with a PubSub component.
func (a *App) ConfigurePubSub() {
	pubsubdebug.Configure(pubsubdebug.Hooks{
		EnsureMutable: a.ensureMutable,
		LogError: func(msg string, err error) {
			a.logError(msg, err, nil)
		},
		ConfigSnapshot: func() pubsubdebug.PubSubConfig {
			return a.configSnapshot().PubSub
		},
		DefaultPubSub: func() pubsub.PubSub {
			a.mu.RLock()
			defer a.mu.RUnlock()
			return a.pub
		},
		EnsureRouter: func() *router.Router {
			return a.ensureRouter()
		},
	})
}

func newPubSubDebugComponent(cfg PubSubConfig, fallbackPub pubsub.PubSub) Component {
	return pubsubdebug.NewPubSubDebugComponent(cfg, fallbackPub)
}
