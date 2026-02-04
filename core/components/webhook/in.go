package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core/internal/contractio"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	webhookin "github.com/spcent/plumego/net/webhookin"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/utils/jsonx"
)

type WebhookInComponent struct {
	cfg        WebhookInConfig
	pub        pubsub.PubSub
	logger     log.StructuredLogger
	deduper    *webhookin.Deduper
	routesOnce sync.Once
}

func NewWebhookInComponent(cfg WebhookInConfig, fallbackPub pubsub.PubSub, logger log.StructuredLogger) *WebhookInComponent {
	if logger == nil {
		logger = log.NewGLogger()
	}
	pub := cfg.Pub
	if pub == nil {
		pub = fallbackPub
	}

	return &WebhookInComponent{cfg: cfg, pub: pub, logger: logger}
}

func (c *WebhookInComponent) RegisterRoutes(r *router.Router) {
	if !c.cfg.Enabled || c.pub == nil {
		return
	}

	c.routesOnce.Do(func() {
		gitHubPath := strings.TrimSpace(c.cfg.GitHubPath)
		if gitHubPath == "" {
			gitHubPath = "/webhooks/github"
		}
		r.PostCtx(gitHubPath, func(ctx *contract.Ctx) { c.webhookInGitHub(ctx) })

		stripePath := strings.TrimSpace(c.cfg.StripePath)
		if stripePath == "" {
			stripePath = "/webhooks/stripe"
		}
		r.PostCtx(stripePath, func(ctx *contract.Ctx) { c.webhookInStripe(ctx) })
	})
}

func (c *WebhookInComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *WebhookInComponent) Start(_ context.Context) error { return nil }

func (c *WebhookInComponent) Stop(_ context.Context) error { return nil }

func (c *WebhookInComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.cfg.Enabled}}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "webhook_in", status
}

func (c *WebhookInComponent) Dependencies() []reflect.Type { return nil }

func (c *WebhookInComponent) webhookInGitHub(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.GitHubSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET"))
	}
	if secret == "" {
		contractio.WriteContractError(ctx, http.StatusInternalServerError, "missing_secret", "GITHUB_WEBHOOK_SECRET is not configured")
		return
	}

	maxBody := c.cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	raw, err := webhookin.VerifyGitHub(ctx.R, secret, maxBody)
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusUnauthorized, "invalid_signature", "invalid GitHub signature")
		return
	}

	event := strings.TrimSpace(ctx.Headers.Get("X-GitHub-Event"))
	delivery := strings.TrimSpace(ctx.Headers.Get("X-GitHub-Delivery"))
	if event == "" {
		event = "unknown"
	}
	if delivery == "" {
		delivery = "unknown"
	}

	d := c.ensureWebhookInDeduper()
	if delivery != "unknown" && d.SeenBefore("github:"+delivery) {
		contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
			"ok":          true,
			"provider":    "github",
			"event_type":  event,
			"delivery_id": delivery,
			"deduped":     true,
		})
		return
	}

	topicPrefix := strings.TrimSpace(c.cfg.TopicPrefixGitHub)
	if topicPrefix == "" {
		topicPrefix = "in.github."
	}
	topic := topicPrefix + sanitizeTopicSuffix(event)

	msg := pubsub.Message{
		ID:    delivery,
		Topic: topic,
		Type:  event,
		Time:  time.Now(),
		Data:  json.RawMessage(raw),
		Meta: map[string]string{
			"source":      "github",
			"trace_id":    ctx.TraceID,
			"delivery_id": delivery,
			"event_type":  event,
			"client_ip":   ctx.ClientIP,
		},
	}

	err = c.pub.Publish(topic, msg)
	if err != nil {
		c.logger.Error("Failed to publish GitHub event", log.Fields{"error": err, "topic": topic, "event_id": delivery})
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"ok":          true,
		"provider":    "github",
		"topic":       topic,
		"event_type":  event,
		"delivery_id": delivery,
		"deduped":     false,
		"body_bytes":  len(raw),
	})
}

func (c *WebhookInComponent) webhookInStripe(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.StripeSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	}
	if secret == "" {
		contractio.WriteContractError(ctx, http.StatusInternalServerError, "missing_secret", "STRIPE_WEBHOOK_SECRET is not configured")
		return
	}

	maxBody := c.cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	tol := c.cfg.StripeTolerance
	if tol <= 0 {
		tol = 5 * time.Minute
	}

	raw, err := webhookin.VerifyStripe(ctx.R, secret, webhookin.StripeVerifyOptions{MaxBody: maxBody, Tolerance: tol})
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusUnauthorized, "invalid_signature", "invalid Stripe signature")
		return
	}

	evtID := jsonx.FieldString(raw, "id")
	evtType := jsonx.FieldString(raw, "type")
	if evtType == "" {
		evtType = "unknown"
	}
	if evtID == "" {
		evtID = "unknown"
	}

	d := c.ensureWebhookInDeduper()
	if evtID != "unknown" && d.SeenBefore("stripe:"+evtID) {
		contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
			"ok":         true,
			"provider":   "stripe",
			"event_type": evtType,
			"event_id":   evtID,
			"deduped":    true,
		})
		return
	}

	topicPrefix := strings.TrimSpace(c.cfg.TopicPrefixStripe)
	if topicPrefix == "" {
		topicPrefix = "in.stripe."
	}
	topic := topicPrefix + sanitizeTopicSuffix(evtType)

	msg := pubsub.Message{
		ID:    evtID,
		Topic: topic,
		Type:  evtType,
		Time:  time.Now(),
		Data:  json.RawMessage(raw),
		Meta: map[string]string{
			"source":      "stripe",
			"trace_id":    ctx.TraceID,
			"delivery_id": evtID,
			"event_type":  evtType,
			"client_ip":   ctx.ClientIP,
		},
	}

	err = c.pub.Publish(topic, msg)
	if err != nil {
		c.logger.Error("Failed to publish Stripe event", log.Fields{"error": err, "topic": topic, "event_id": evtID})
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"ok":         true,
		"provider":   "stripe",
		"topic":      topic,
		"event_type": evtType,
		"event_id":   evtID,
		"deduped":    false,
		"body_bytes": len(raw),
	})
}

func (c *WebhookInComponent) ensureWebhookInDeduper() *webhookin.Deduper {
	if c.cfg.Deduper != nil {
		return c.cfg.Deduper
	}
	if c.deduper != nil {
		return c.deduper
	}
	ttl := c.cfg.DedupTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	c.deduper = webhookin.NewDeduper(ttl)
	return c.deduper
}

// sanitizeTopicSuffix keeps topic safe and consistent.
// Allow [a-zA-Z0-9._-], replace others with '_'.
func sanitizeTopicSuffix(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '.' || c == '_' || c == '-'
		if ok {
			b.WriteByte(c)
		} else {
			b.WriteByte('_')
		}
	}
	return strings.ToLower(b.String())
}
