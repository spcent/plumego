package core

import (
	"reflect"

	"github.com/spcent/plumego/core/components/builtin"
	"github.com/spcent/plumego/core/components/pubsubdebug"
	"github.com/spcent/plumego/core/components/webhook"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/pubsub"
)

func (a *App) builtInComponents() []Component {
	a.mu.RLock()
	pubSubConfig := a.config.PubSub
	webhookOutConfig := a.config.WebhookOut
	webhookInConfig := a.config.WebhookIn
	debug := a.config.Debug
	pub := a.pub
	logger := a.logger
	a.mu.RUnlock()

	comps := builtin.BuiltInComponents(builtin.Hooks{
		Debug:            debug,
		PubSubConfig:     pubSubConfig,
		WebhookOutConfig: webhookOutConfig,
		WebhookInConfig:  webhookInConfig,
		PubSub:           pub,
		Logger:           logger,
		HasComponentType: a.hasComponentType,
		NewDevTools: func() builtin.Component {
			return newDevToolsComponent(a)
		},
		NewPubSubDebug: func(cfg pubsubdebug.PubSubConfig, fallback pubsub.PubSub) builtin.Component {
			return newPubSubDebugComponent(cfg, fallback)
		},
		NewWebhookOut: func(cfg webhook.WebhookOutConfig) builtin.Component {
			return newWebhookOutComponent(cfg)
		},
		NewWebhookIn: func(cfg webhook.WebhookInConfig, fallback pubsub.PubSub, logger log.StructuredLogger) builtin.Component {
			return newWebhookInComponent(cfg, fallback, logger)
		},
	})

	out := make([]Component, 0, len(comps))
	for _, c := range comps {
		out = append(out, c)
	}
	return out
}

func (a *App) hasComponentType(target any) bool {
	if target == nil {
		return false
	}

	typeOfTarget := reflect.TypeOf(target)
	if typeOfTarget == nil {
		return false
	}

	a.mu.RLock()
	comps := append([]Component{}, a.components...)
	a.mu.RUnlock()

	for _, c := range comps {
		if reflect.TypeOf(c) == typeOfTarget {
			return true
		}
	}

	return false
}
