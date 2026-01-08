package core

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	webhookin "github.com/spcent/plumego/net/webhookin"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/utils/jsonx"
)

type webhookInComponent struct {
	BaseComponent
	cfg        WebhookInConfig
	pub        pubsub.PubSub
	logger     log.StructuredLogger
	deduper    *webhookin.Deduper
	routesOnce sync.Once
}

func newWebhookInComponent(cfg WebhookInConfig, fallbackPub pubsub.PubSub, logger log.StructuredLogger) Component {
	pub := cfg.Pub
	if pub == nil {
		pub = fallbackPub
	}

	return &webhookInComponent{cfg: cfg, pub: pub, logger: logger}
}

func (c *webhookInComponent) RegisterRoutes(r *router.Router) {
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

func (c *webhookInComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *webhookInComponent) Start(_ context.Context) error { return nil }

func (c *webhookInComponent) Stop(_ context.Context) error { return nil }

func (c *webhookInComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.cfg.Enabled}}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "webhook_in", status
}

func (c *webhookInComponent) webhookInGitHub(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.GitHubSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET"))
	}
	if secret == "" {
		ctx.ErrorJSON(http.StatusInternalServerError, "missing_secret", "GITHUB_WEBHOOK_SECRET is not configured", nil)
		return
	}

	maxBody := c.cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	raw, err := webhookin.VerifyGitHub(ctx.R, secret, maxBody)
	if err != nil {
		ctx.ErrorJSON(http.StatusUnauthorized, "invalid_signature", "invalid GitHub signature", nil)
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
		ctx.JSON(http.StatusOK, map[string]any{
			"ok":          true,
			"provider":    "github",
			"event_type":  event,
			"delivery_id": delivery,
			"deduped":     true,
			"trace_id":    ctx.TraceID,
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

	ctx.JSON(http.StatusOK, map[string]any{
		"ok":          true,
		"provider":    "github",
		"topic":       topic,
		"event_type":  event,
		"delivery_id": delivery,
		"deduped":     false,
		"trace_id":    ctx.TraceID,
		"body_bytes":  len(raw),
	})
}

func (c *webhookInComponent) webhookInStripe(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.StripeSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	}
	if secret == "" {
		ctx.ErrorJSON(http.StatusInternalServerError, "missing_secret", "STRIPE_WEBHOOK_SECRET is not configured", nil)
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
		ctx.ErrorJSON(http.StatusUnauthorized, "invalid_signature", "invalid Stripe signature", nil)
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
		ctx.JSON(http.StatusOK, map[string]any{
			"ok":         true,
			"provider":   "stripe",
			"event_type": evtType,
			"event_id":   evtID,
			"deduped":    true,
			"trace_id":   ctx.TraceID,
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

	ctx.JSON(http.StatusOK, map[string]any{
		"ok":         true,
		"provider":   "stripe",
		"topic":      topic,
		"event_type": evtType,
		"event_id":   evtID,
		"deduped":    false,
		"trace_id":   ctx.TraceID,
		"body_bytes": len(raw),
	})
}

func (c *webhookInComponent) ensureWebhookInDeduper() *webhookin.Deduper {
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

// ConfigureWebhookIn mounts inbound webhook receivers for GitHub and Stripe.
// It remains for backward compatibility but now mounts a component into the lifecycle.
func (a *App) ConfigureWebhookIn() {
	comp := newWebhookInComponent(a.config.WebhookIn, a.pub, a.logger)
	comp.RegisterRoutes(a.router)
	comp.RegisterMiddleware(a.middlewareReg)
	a.components = append(a.components, comp)
}
