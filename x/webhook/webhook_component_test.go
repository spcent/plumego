package webhook

import (
	"context"
	"testing"

	"github.com/spcent/plumego/pubsub"
)

func TestWebhookBridge(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	// Test with nil PubSub
	bridgeBad := &WebhookBridge{
		Out:   nil,
		Rules: []Rule{{InTopic: "test", OutEventType: "test"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := bridgeBad.Start(ctx)
	if err == nil {
		t.Error("expected error for nil PubSub")
	}

	// Test with nil Service
	bridgeBad2 := &WebhookBridge{
		Pub:   ps,
		Rules: []Rule{{InTopic: "test", OutEventType: "test"}},
	}
	_, err = bridgeBad2.Start(ctx)
	if err == nil {
		t.Error("expected error for nil Service")
	}

	// Test with no rules
	bridgeBad3 := &WebhookBridge{
		Pub: ps,
		Out: nil,
	}
	_, err = bridgeBad3.Start(ctx)
	if err == nil {
		t.Error("expected error for no rules")
	}
}

func TestWebhookBridgeStopAll(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	bridge := &WebhookBridge{
		Pub:   ps,
		Out:   nil,
		Rules: []Rule{},
	}

	// Test stopAll when not started
	bridge.stopAll()

	// Test stopAll when subs is nil
	bridge.subs = nil
	bridge.stopAll()

	// Test stopAll with empty subs
	bridge.subs = []pubsub.Subscription{}
	bridge.stopAll()
}

func TestWebhookBridgeFirstNonEmpty(t *testing.T) {
	// Test the firstNonEmpty helper function
	tests := []struct {
		a, b     string
		expected string
	}{
		{"", "", ""},
		{"a", "", "a"},
		{"", "b", "b"},
		{"a", "b", "a"},
	}

	for _, tt := range tests {
		result := firstNonEmpty(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("firstNonEmpty(%q, %q) = %q, want %q", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestWebhookInComponentBasic(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	cfg := WebhookInConfig{
		Enabled:      true,
		Pub:          ps,
		GitHubSecret: "secret123",
		StripeSecret: "secret456",
		GitHubPath:   "/webhooks/github",
		StripePath:   "/webhooks/stripe",
	}

	comp := NewWebhookInComponent(cfg, ps, nil)

	// Test Health
	name, health := comp.Health()
	if name != "webhook_in" {
		t.Errorf("expected name 'webhook_in', got %s", name)
	}
	if health.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", health.Status)
	}

	// Test disabled component
	disabledCfg := WebhookInConfig{Enabled: false}
	disabledComp := NewWebhookInComponent(disabledCfg, ps, nil)
	_, disabledHealth := disabledComp.Health()
	if disabledHealth.Status != "degraded" {
		t.Errorf("expected degraded status for disabled component, got %s", disabledHealth.Status)
	}

	// Test Start/Stop
	if err := comp.Start(context.Background()); err != nil {
		t.Errorf("unexpected error from Start: %v", err)
	}
	if err := comp.Stop(context.Background()); err != nil {
		t.Errorf("unexpected error from Stop: %v", err)
	}
}

func TestWebhookOutComponentBasic(t *testing.T) {
	cfg := WebhookOutConfig{
		Enabled:  true,
		BasePath: "/webhooks",
	}

	comp := NewWebhookOutComponent(cfg)

	// Test Health
	name, health := comp.Health()
	if name != "webhook_out" {
		t.Errorf("expected name 'webhook_out', got %s", name)
	}
	if health.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", health.Status)
	}

	// Test disabled component
	disabledCfg := WebhookOutConfig{Enabled: false}
	disabledComp := NewWebhookOutComponent(disabledCfg)
	_, disabledHealth := disabledComp.Health()
	if disabledHealth.Status != "degraded" {
		t.Errorf("expected degraded status for disabled component, got %s", disabledHealth.Status)
	}

	// Test Start/Stop
	if err := comp.Start(context.Background()); err != nil {
		t.Errorf("unexpected error from Start: %v", err)
	}
	if err := comp.Stop(context.Background()); err != nil {
		t.Errorf("unexpected error from Stop: %v", err)
	}
}
