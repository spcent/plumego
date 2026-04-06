package webhook

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/internal/jsonx"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/x/pubsub"
)

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler) error
}

type Inbound struct {
	cfg        WebhookInConfig
	pub        pubsub.Broker
	logger     log.StructuredLogger
	deduper    *Deduper
	routesOnce sync.Once
}

func NewInbound(cfg WebhookInConfig, fallbackPub pubsub.Broker, logger log.StructuredLogger) *Inbound {
	if logger == nil {
		logger = log.NewNoOpLogger()
	}
	pub := cfg.Pub
	if pub == nil {
		pub = fallbackPub
	}

	return &Inbound{cfg: cfg, pub: pub, logger: logger}
}

func (c *Inbound) RegisterRoutes(r routeRegistrar) error {
	if !c.cfg.Enabled || c.pub == nil {
		return nil
	}

	var regErr error
	c.routesOnce.Do(func() {
		gitHubPath := strings.TrimSpace(c.cfg.GitHubPath)
		if gitHubPath == "" {
			gitHubPath = "/webhooks/github"
		}
		regErr = r.AddRoute(http.MethodPost, gitHubPath, contract.AdaptCtxHandler(func(ctx *contract.Ctx) { c.webhookInGitHub(ctx) }))
		if regErr != nil {
			return
		}

		stripePath := strings.TrimSpace(c.cfg.StripePath)
		if stripePath == "" {
			stripePath = "/webhooks/stripe"
		}
		regErr = r.AddRoute(http.MethodPost, stripePath, contract.AdaptCtxHandler(func(ctx *contract.Ctx) { c.webhookInStripe(ctx) }))
	})

	return regErr
}

func (c *Inbound) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.cfg.Enabled}}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "webhook_in", status
}

func (c *Inbound) webhookInGitHub(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.GitHubSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET"))
	}
	if secret == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errInMissingSecret("GITHUB_WEBHOOK_SECRET is not configured"))
		return
	}

	maxBody := c.cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	raw, err := VerifyGitHub(ctx.R, secret, maxBody)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errInInvalidSignature("invalid GitHub signature"))
		return
	}

	event := strings.TrimSpace(ctx.RequestHeaders().Get("X-GitHub-Event"))
	delivery := strings.TrimSpace(ctx.RequestHeaders().Get("X-GitHub-Delivery"))
	if event == "" {
		event = "unknown"
	}
	if delivery == "" {
		delivery = "unknown"
	}

	d := c.ensureWebhookInDeduper()
	if delivery != "unknown" && d.SeenBefore("github:"+delivery) {
		_ = ctx.Response(http.StatusOK, map[string]any{
			"ok":          true,
			"provider":    "github",
			"event_type":  event,
			"delivery_id": delivery,
			"deduped":     true,
		}, nil)
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
			"trace_id":    ctx.TraceID(),
			"delivery_id": delivery,
			"event_type":  event,
			"client_ip":   ctx.ClientIP,
		},
	}

	if err = c.pub.Publish(topic, msg); err != nil {
		c.logger.Error("Failed to publish GitHub event", log.Fields{"error": err, "topic": topic, "event_id": delivery})
		_ = contract.WriteError(ctx.W, ctx.R, errInPublishFailed())
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{
		"ok":          true,
		"provider":    "github",
		"topic":       topic,
		"event_type":  event,
		"delivery_id": delivery,
		"deduped":     false,
		"body_bytes":  len(raw),
	}, nil)
}

func (c *Inbound) webhookInStripe(ctx *contract.Ctx) {
	secret := strings.TrimSpace(c.cfg.StripeSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	}
	if secret == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errInMissingSecret("STRIPE_WEBHOOK_SECRET is not configured"))
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

	raw, err := VerifyStripe(ctx.R, secret, StripeVerifyOptions{MaxBody: maxBody, Tolerance: tol})
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errInInvalidSignature("invalid Stripe signature"))
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
		_ = ctx.Response(http.StatusOK, map[string]any{
			"ok":         true,
			"provider":   "stripe",
			"event_type": evtType,
			"event_id":   evtID,
			"deduped":    true,
		}, nil)
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
			"trace_id":    ctx.TraceID(),
			"delivery_id": evtID,
			"event_type":  evtType,
			"client_ip":   ctx.ClientIP,
		},
	}

	if err = c.pub.Publish(topic, msg); err != nil {
		c.logger.Error("Failed to publish Stripe event", log.Fields{"error": err, "topic": topic, "event_id": evtID})
		_ = contract.WriteError(ctx.W, ctx.R, errInPublishFailed())
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{
		"ok":         true,
		"provider":   "stripe",
		"topic":      topic,
		"event_type": evtType,
		"event_id":   evtID,
		"deduped":    false,
		"body_bytes": len(raw),
	}, nil)
}

func (c *Inbound) ensureWebhookInDeduper() *Deduper {
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
	c.deduper = NewDeduper(ttl)
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

// --- package-local error helpers for inbound webhook handlers ---

func errInMissingSecret(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusInternalServerError).Code("missing_secret").Message(msg).Build()
}

func errInInvalidSignature(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusUnauthorized).Code("invalid_signature").Message(msg).Build()
}

func errInPublishFailed() contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusInternalServerError).Code("publish_failed").Message("failed to forward event to internal bus").Build()
}
