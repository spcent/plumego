package webhook

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core/internal/contractio"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/utils/stringsx"
)

type WebhookOutComponent struct {
	cfg        WebhookOutConfig
	routesOnce sync.Once
}

func NewWebhookOutComponent(cfg WebhookOutConfig) *WebhookOutComponent {
	return &WebhookOutComponent{cfg: cfg}
}

func (c *WebhookOutComponent) RegisterRoutes(r *router.Router) {
	if !c.cfg.Enabled || c.cfg.Service == nil {
		return
	}

	c.routesOnce.Do(func() {
		base := strings.TrimSpace(c.cfg.BasePath)
		if base == "" {
			base = "/webhooks"
		}

		svc := c.cfg.Service

		r.PostCtx(base+"/targets", func(ctx *contract.Ctx) { webhookCreateTarget(ctx, svc) })
		r.GetCtx(base+"/targets", func(ctx *contract.Ctx) { webhookListTargets(ctx, svc) })
		r.GetCtx(base+"/targets/:id", func(ctx *contract.Ctx) { webhookGetTarget(ctx, svc) })
		r.PatchCtx(base+"/targets/:id", func(ctx *contract.Ctx) { webhookPatchTarget(ctx, svc) })
		r.PostCtx(base+"/targets/:id/enable", func(ctx *contract.Ctx) { webhookSetTargetEnabled(ctx, svc, true) })
		r.PostCtx(base+"/targets/:id/disable", func(ctx *contract.Ctx) { webhookSetTargetEnabled(ctx, svc, false) })

		r.PostCtx(base+"/events/:event", func(ctx *contract.Ctx) {
			webhookTriggerEvent(ctx, svc, c.cfg.TriggerToken, c.cfg.AllowEmptyToken)
		})

		r.GetCtx(base+"/deliveries", func(ctx *contract.Ctx) { webhookListDeliveries(ctx, svc, c.cfg.DefaultPageLimit) })
		r.GetCtx(base+"/deliveries/:id", func(ctx *contract.Ctx) { webhookGetDelivery(ctx, svc) })
		r.PostCtx(base+"/deliveries/:id/replay", func(ctx *contract.Ctx) { webhookReplayDelivery(ctx, svc) })
	})
}

func (c *WebhookOutComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *WebhookOutComponent) Start(_ context.Context) error { return nil }

func (c *WebhookOutComponent) Stop(_ context.Context) error { return nil }

func (c *WebhookOutComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.cfg.Enabled}}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "webhook_out", status
}

func (c *WebhookOutComponent) Dependencies() []reflect.Type { return nil }

type targetDTO struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Enabled bool     `json:"enabled"`

	Headers map[string]string `json:"headers,omitempty"`

	TimeoutMs     int   `json:"timeout_ms,omitempty"`
	MaxRetries    int   `json:"max_retries,omitempty"`
	BackoffBaseMs int   `json:"backoff_base_ms,omitempty"`
	BackoffMaxMs  int   `json:"backoff_max_ms,omitempty"`
	RetryOn429    *bool `json:"retry_on_429,omitempty"`

	SecretMasked string    `json:"secret_masked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func webhookCreateTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	var req webhookout.Target
	if err := ctx.BindJSON(&req); err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	t, err := svc.CreateTarget(ctx.R.Context(), req)
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusCreated, targetToDTO(t))
}

func webhookListTargets(ctx *contract.Ctx, svc *webhookout.Service) {
	q := ctx.Query

	var enabled *bool
	if v := strings.TrimSpace(q.Get("enabled")); v != "" {
		b := (v == "1" || strings.EqualFold(v, "true"))
		enabled = &b
	}
	event := strings.TrimSpace(q.Get("event"))

	items, err := svc.ListTargets(ctx.R.Context(), webhookout.TargetFilter{
		Enabled: enabled,
		Event:   event,
	})
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusInternalServerError, "store_error", err.Error())
		return
	}

	out := make([]targetDTO, 0, len(items))
	for _, t := range items {
		out = append(out, targetToDTO(t))
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{"items": out})
}

func webhookGetTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "id is required")
		return
	}

	t, ok := svc.GetTarget(ctx.R.Context(), id)
	if !ok {
		contractio.WriteContractError(ctx, http.StatusNotFound, "not_found", "target not found")
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, targetToDTO(t))
}

func webhookPatchTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "id is required")
		return
	}

	var req webhookout.TargetPatch
	if err := ctx.BindJSON(&req); err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, req)
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, targetToDTO(t))
}

func webhookSetTargetEnabled(ctx *contract.Ctx, svc *webhookout.Service, enable bool) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "id is required")
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, webhookout.TargetPatch{Enabled: &enable})
	if err != nil {
		if err == webhookout.ErrNotFound {
			contractio.WriteContractError(ctx, http.StatusNotFound, "not_found", "target not found")
			return
		}
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, targetToDTO(t))
}

func webhookTriggerEvent(ctx *contract.Ctx, svc *webhookout.Service, token string, allowEmpty bool) {
	event, ok := ctx.Param("event")
	if !ok || strings.TrimSpace(event) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "event is required")
		return
	}

	provided := strings.TrimSpace(ctx.Query.Get("token"))
	if provided == "" {
		provided = strings.TrimSpace(ctx.Headers.Get("X-Trigger-Token"))
	}

	if token == "" && !allowEmpty {
		contractio.WriteContractError(ctx, http.StatusForbidden, "forbidden", "triggering is disabled")
		return
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
		contractio.WriteContractError(ctx, http.StatusUnauthorized, "unauthorized", "invalid trigger token")
		return
	}

	var payload struct {
		Data map[string]any `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	enqueued, err := svc.TriggerEvent(ctx.R.Context(), webhookout.Event{Type: event, Data: payload.Data, Meta: payload.Meta})
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusAccepted, map[string]any{
		"enqueued": enqueued,
		"event":    event,
	})
}

func webhookListDeliveries(ctx *contract.Ctx, svc *webhookout.Service, defaultLimit int) {
	q := ctx.Query

	limit := defaultLimit
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := webhookout.DeliveryFilter{Limit: limit, Cursor: strings.TrimSpace(q.Get("cursor"))}

	if v := strings.TrimSpace(q.Get("target_id")); v != "" {
		filter.TargetID = &v
	}
	if v := strings.TrimSpace(q.Get("event")); v != "" {
		filter.Event = &v
	}
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		status := webhookout.DeliveryStatus(v)
		filter.Status = &status
	}
	if v := strings.TrimSpace(q.Get("from")); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &ts
		}
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &ts
		}
	}

	deliveries, err := svc.ListDeliveries(ctx.R.Context(), filter)
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusInternalServerError, "store_error", err.Error())
		return
	}

	type deliveryDTO struct {
		ID           string                    `json:"id"`
		TargetID     string                    `json:"target_id"`
		EventID      string                    `json:"event_id"`
		EventType    string                    `json:"event_type"`
		Status       webhookout.DeliveryStatus `json:"status"`
		Attempt      int                       `json:"attempt"`
		NextAt       *time.Time                `json:"next_at,omitempty"`
		LastHTTP     int                       `json:"last_http_status,omitempty"`
		LastError    string                    `json:"last_error,omitempty"`
		LastDuration int                       `json:"last_duration_ms,omitempty"`
		LastResp     string                    `json:"last_resp_snippet,omitempty"`
		CreatedAt    time.Time                 `json:"created_at"`
		UpdatedAt    time.Time                 `json:"updated_at"`
	}

	out := make([]deliveryDTO, 0, len(deliveries))
	for _, d := range deliveries {
		out = append(out, deliveryDTO{
			ID:           d.ID,
			TargetID:     d.TargetID,
			EventID:      d.EventID,
			EventType:    d.EventType,
			Status:       d.Status,
			Attempt:      d.Attempt,
			NextAt:       d.NextAt,
			LastHTTP:     d.LastHTTPStatus,
			LastError:    d.LastError,
			LastDuration: d.LastDurationMs,
			LastResp:     d.LastRespSnippet,
			CreatedAt:    d.CreatedAt,
			UpdatedAt:    d.UpdatedAt,
		})
	}

	resp := map[string]any{"items": out}
	contractio.WriteContractResponse(ctx, http.StatusOK, resp)
}

func webhookGetDelivery(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "id is required")
		return
	}

	d, ok := svc.GetDelivery(ctx.R.Context(), id)
	if !ok {
		contractio.WriteContractError(ctx, http.StatusNotFound, "not_found", "delivery not found")
		return
	}

	payload := json.RawMessage(d.PayloadJSON)

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"id":                d.ID,
		"target_id":         d.TargetID,
		"event_id":          d.EventID,
		"event_type":        d.EventType,
		"status":            d.Status,
		"attempt":           d.Attempt,
		"next_at":           d.NextAt,
		"last_http_status":  d.LastHTTPStatus,
		"last_error":        d.LastError,
		"last_duration_ms":  d.LastDurationMs,
		"last_resp_snippet": d.LastRespSnippet,
		"payload_json":      payload,
		"created_at":        d.CreatedAt,
		"updated_at":        d.UpdatedAt,
	})
}

func webhookReplayDelivery(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", "id is required")
		return
	}

	d, err := svc.ReplayDelivery(ctx.R.Context(), id)
	if errors.Is(err, webhookout.ErrNotFound) {
		contractio.WriteContractError(ctx, http.StatusNotFound, "not_found", "delivery not found")
		return
	}
	if err != nil {
		contractio.WriteContractError(ctx, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{"ok": true, "delivery_id": d.ID})
}

func targetToDTO(t webhookout.Target) targetDTO {
	dto := targetDTO{
		ID:            t.ID,
		Name:          t.Name,
		URL:           t.URL,
		Events:        t.Events,
		Enabled:       t.Enabled,
		Headers:       t.Headers,
		SecretMasked:  stringsx.MaskSecret(t.Secret),
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		TimeoutMs:     t.TimeoutMs,
		MaxRetries:    t.MaxRetries,
		BackoffBaseMs: t.BackoffBaseMs,
		BackoffMaxMs:  t.BackoffMaxMs,
	}

	if t.RetryOn429 != nil {
		dto.RetryOn429 = t.RetryOn429
	}

	return dto
}
