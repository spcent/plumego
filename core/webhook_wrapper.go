package core

import (
	webhook "github.com/spcent/plumego/core/components/webhook"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/pubsub"
)

type Rule = webhook.Rule
type WebhookBridge = webhook.WebhookBridge

// ConfigureWebhookIn mounts inbound webhook receivers for GitHub and Stripe.
// It remains for backward compatibility but now mounts a component into the lifecycle.
func (a *App) ConfigureWebhookIn() {
	if err := a.ensureMutable("configure_webhook_in", "configure webhook in"); err != nil {
		a.logError("ConfigureWebhookIn failed", err, nil)
		return
	}

	cfg := a.configSnapshot()

	a.mu.RLock()
	pub := a.pub
	logger := a.logger
	a.mu.RUnlock()

	comp := newWebhookInComponent(cfg.WebhookIn, pub, logger)
	comp.RegisterRoutes(a.ensureRouter())
	comp.RegisterMiddleware(a.ensureMiddlewareRegistry())

	a.mu.Lock()
	a.components = append(a.components, comp)
	a.mu.Unlock()
}

// ConfigureWebhookOut mounts outbound webhook management APIs.
// It remains for backward compatibility but now mounts a component into the lifecycle.
func (a *App) ConfigureWebhookOut() {
	if err := a.ensureMutable("configure_webhook_out", "configure webhook out"); err != nil {
		a.logError("ConfigureWebhookOut failed", err, nil)
		return
	}

	cfg := a.configSnapshot()

	comp := newWebhookOutComponent(cfg.WebhookOut)
	comp.RegisterRoutes(a.ensureRouter())
	comp.RegisterMiddleware(a.ensureMiddlewareRegistry())

	a.mu.Lock()
	a.components = append(a.components, comp)
	a.mu.Unlock()
}

func newWebhookInComponent(cfg WebhookInConfig, fallbackPub pubsub.PubSub, logger log.StructuredLogger) Component {
	if logger == nil {
		logger = log.NewGLogger()
	}
	return webhook.NewWebhookInComponent(cfg, fallbackPub, logger)
}

func newWebhookOutComponent(cfg WebhookOutConfig) Component {
	return webhook.NewWebhookOutComponent(cfg)
}
