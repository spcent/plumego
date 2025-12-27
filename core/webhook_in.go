package core

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
	webhookin "github.com/spcent/plumego/net/webhookin"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/utils/jsonx"
)

// ConfigureWebhookIn mounts inbound webhook receivers for GitHub and Stripe.
func (a *App) ConfigureWebhookIn() {
	cfg := a.config.WebhookIn
	if !cfg.Enabled {
		return
	}

	pub := cfg.Pub
	if pub == nil {
		pub = a.pub
	}
	if pub == nil {
		return
	}

	router := a.Router()

	gitHubPath := strings.TrimSpace(cfg.GitHubPath)
	if gitHubPath == "" {
		gitHubPath = "/webhooks/github"
	}
	router.PostCtx(gitHubPath, func(ctx *contract.Ctx) { a.webhookInGitHub(ctx, pub, cfg) })

	stripePath := strings.TrimSpace(cfg.StripePath)
	if stripePath == "" {
		stripePath = "/webhooks/stripe"
	}
	router.PostCtx(stripePath, func(ctx *contract.Ctx) { a.webhookInStripe(ctx, pub, cfg) })
}

func (a *App) webhookInGitHub(ctx *contract.Ctx, pub pubsub.PubSub, cfg WebhookInConfig) {
	secret := strings.TrimSpace(cfg.GitHubSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET"))
	}
	if secret == "" {
		ctx.ErrorJSON(http.StatusInternalServerError, "missing_secret", "GITHUB_WEBHOOK_SECRET is not configured", nil)
		return
	}

	maxBody := cfg.MaxBodyBytes
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

	d := a.ensureWebhookInDeduper(cfg)
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

	topicPrefix := strings.TrimSpace(cfg.TopicPrefixGitHub)
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

	err = pub.Publish(topic, msg)
	if err != nil {
		a.Logger().Error("Failed to publish GitHub event",
			log.Fields{"error": err, "topic": topic, "event_id": delivery})
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

func (a *App) webhookInStripe(ctx *contract.Ctx, pub pubsub.PubSub, cfg WebhookInConfig) {
	secret := strings.TrimSpace(cfg.StripeSecret)
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	}
	if secret == "" {
		ctx.ErrorJSON(http.StatusInternalServerError, "missing_secret", "STRIPE_WEBHOOK_SECRET is not configured", nil)
		return
	}

	maxBody := cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	tol := cfg.StripeTolerance
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

	d := a.ensureWebhookInDeduper(cfg)
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

	topicPrefix := strings.TrimSpace(cfg.TopicPrefixStripe)
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

	err = pub.Publish(topic, msg)
	if err != nil {
		a.Logger().Error("Failed to publish Stripe event",
			log.Fields{"error": err, "topic": topic, "event_id": evtID})
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

func (a *App) ensureWebhookInDeduper(cfg WebhookInConfig) *webhookin.Deduper {
	if cfg.Deduper != nil {
		return cfg.Deduper
	}
	if a.webhookInDeduper != nil {
		return a.webhookInDeduper
	}
	ttl := cfg.DedupTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	a.webhookInDeduper = webhookin.NewDeduper(ttl)
	return a.webhookInDeduper
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
