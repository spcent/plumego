package core

import (
	"testing"

	"github.com/spcent/plumego/pubsub"
)

func TestConfigureWebhookIn(t *testing.T) {
	app := New()
	ps := pubsub.New()
	defer ps.Close()

	app.pub = ps
	app.config.WebhookIn = WebhookInConfig{
		Enabled:      true,
		GitHubSecret: "secret",
		StripeSecret: "secret",
	}

	app.ConfigureWebhookIn()

	// Verify component was added
	if len(app.components) == 0 {
		t.Error("expected component to be added")
	}
}

func TestConfigureWebhookOut(t *testing.T) {
	app := New()

	app.config.WebhookOut = WebhookOutConfig{
		Enabled: true,
	}

	app.ConfigureWebhookOut()

	// Verify component was added
	if len(app.components) == 0 {
		t.Error("expected component to be added")
	}
}
