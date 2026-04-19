package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/pubsub"
)

func TestWebhookBridge(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	// Test with nil PubSub
	bridgeBad := &WebhookBridge{
		Out:   nil,
		Rules: []Rule{{InTopic: "test", OutEventType: "test"}},
	}
	ctx, cancel := context.WithCancel(t.Context())
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

func TestWebhookInboundBasic(t *testing.T) {
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

	handler := NewInbound(cfg, ps, nil)

	// Test Health
	name, health := handler.Health()
	if name != "webhook_in" {
		t.Errorf("expected name 'webhook_in', got %s", name)
	}
	if health.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", health.Status)
	}

	// Test disabled component
	disabledCfg := WebhookInConfig{Enabled: false}
	disabledHandler := NewInbound(disabledCfg, ps, nil)
	_, disabledHealth := disabledHandler.Health()
	if disabledHealth.Status != "degraded" {
		t.Errorf("expected degraded status for disabled component, got %s", disabledHealth.Status)
	}
}

func TestWebhookOutboundBasic(t *testing.T) {
	cfg := WebhookOutConfig{
		Enabled:  true,
		BasePath: "/webhooks",
	}

	handler := NewOutbound(cfg)

	// Test Health
	name, health := handler.Health()
	if name != "webhook_out" {
		t.Errorf("expected name 'webhook_out', got %s", name)
	}
	if health.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", health.Status)
	}

	// Test disabled component
	disabledCfg := WebhookOutConfig{Enabled: false}
	disabledHandler := NewOutbound(disabledCfg)
	_, disabledHealth := disabledHandler.Health()
	if disabledHealth.Status != "degraded" {
		t.Errorf("expected degraded status for disabled component, got %s", disabledHealth.Status)
	}
}

func TestInboundRegisterRoutesDuplicateReturnsError(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	handler := NewInbound(WebhookInConfig{
		Enabled:      true,
		Pub:          ps,
		GitHubSecret: "secret123",
		StripeSecret: "secret456",
	}, ps, nil)

	r := router.NewRouter()
	if err := handler.RegisterRoutes(r); err != nil {
		t.Fatalf("first RegisterRoutes failed: %v", err)
	}
	if err := handler.RegisterRoutes(r); err == nil {
		t.Fatal("expected duplicate RegisterRoutes to return an error")
	}
}

func TestOutboundRegisterRoutesDuplicateReturnsError(t *testing.T) {
	handler := NewOutbound(WebhookOutConfig{
		Enabled: true,
		Service: &Service{},
	})

	r := router.NewRouter()
	if err := handler.RegisterRoutes(r); err != nil {
		t.Fatalf("first RegisterRoutes failed: %v", err)
	}
	if err := handler.RegisterRoutes(r); err == nil {
		t.Fatal("expected duplicate RegisterRoutes to return an error")
	}
}

func TestInboundMissingSecretWritesCanonicalError(t *testing.T) {
	ps := pubsub.New()
	defer ps.Close()

	handler := NewInbound(WebhookInConfig{
		Enabled:    true,
		Pub:        ps,
		GitHubPath: "/webhooks/github",
	}, ps, nil)

	r := router.NewRouter()
	if err := handler.RegisterRoutes(r); err != nil {
		t.Fatalf("RegisterRoutes failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "missing_secret" {
		t.Fatalf("error code = %q, want missing_secret", body.Error.Code)
	}
}

func TestOutboundTriggerDisabledWritesCanonicalError(t *testing.T) {
	handler := NewOutbound(WebhookOutConfig{
		Enabled:         true,
		Service:         &Service{},
		AllowEmptyToken: false,
	})

	r := router.NewRouter()
	if err := handler.RegisterRoutes(r); err != nil {
		t.Fatalf("RegisterRoutes failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhooks/events/demo", httptest.NewRequest(http.MethodPost, "/", nil).Body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "FORBIDDEN" {
		t.Fatalf("error code = %q, want FORBIDDEN", body.Error.Code)
	}
}
