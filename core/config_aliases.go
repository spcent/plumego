package core

import (
	pubsubdebug "github.com/spcent/plumego/core/components/pubsubdebug"
	webhook "github.com/spcent/plumego/core/components/webhook"
)

// Aliases to keep AppConfig stable while components move under core/components.
type PubSubConfig = pubsubdebug.PubSubConfig
type WebhookOutConfig = webhook.WebhookOutConfig
type WebhookInConfig = webhook.WebhookInConfig
