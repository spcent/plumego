package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/internal/stringsx"
)

type Outbound struct {
	cfg        WebhookOutConfig
	routesOnce sync.Once
}

func NewOutbound(cfg WebhookOutConfig) *Outbound {
	return &Outbound{cfg: cfg}
}

func (c *Outbound) RegisterRoutes(r routeRegistrar) error {
	if !c.cfg.Enabled || c.cfg.Service == nil {
		return nil
	}

	var regErr error
	c.routesOnce.Do(func() {
		base := strings.TrimSpace(c.cfg.BasePath)
		if base == "" {
			base = "/webhooks"
		}

		svc := c.cfg.Service

		if regErr = r.AddRoute(http.MethodPost, base+"/targets", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookCreateTarget(ctx, svc) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodGet, base+"/targets", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookListTargets(ctx, svc) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodGet, base+"/targets/:id", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookGetTarget(ctx, svc) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodPatch, base+"/targets/:id", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookPatchTarget(ctx, svc) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodPost, base+"/targets/:id/enable", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookSetTargetEnabled(ctx, svc, true) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodPost, base+"/targets/:id/disable", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookSetTargetEnabled(ctx, svc, false) })); regErr != nil {
			return
		}

		if regErr = r.AddRoute(http.MethodPost, base+"/events/:event", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
			webhookTriggerEvent(ctx, svc, c.cfg.TriggerToken, c.cfg.AllowEmptyToken)
		})); regErr != nil {
			return
		}

		if regErr = r.AddRoute(http.MethodGet, base+"/deliveries", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookListDeliveries(ctx, svc, c.cfg.DefaultPageLimit) })); regErr != nil {
			return
		}
		if regErr = r.AddRoute(http.MethodGet, base+"/deliveries/:id", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookGetDelivery(ctx, svc) })); regErr != nil {
			return
		}
		regErr = r.AddRoute(http.MethodPost, base+"/deliveries/:id/replay", contract.AdaptCtxHandler(func(ctx *contract.Ctx) { webhookReplayDelivery(ctx, svc) }))
	})

	return regErr
}

func (c *Outbound) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.cfg.Enabled}}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
	}

	return "webhook_out", status
}

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

func webhookCreateTarget(ctx *contract.Ctx, svc *Service) {
	var req Target
	if err := ctx.BindJSON(&req, nil); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errInvalidJSON())
		return
	}

	t, err := svc.CreateTarget(ctx.R.Context(), req)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest(err.Error()))
		return
	}

	_ = ctx.Response(http.StatusCreated, targetToDTO(t), nil)
}

func webhookListTargets(ctx *contract.Ctx, svc *Service) {
	q := ctx.Query

	var enabled *bool
	if v := strings.TrimSpace(q.Get("enabled")); v != "" {
		b := (v == "1" || strings.EqualFold(v, "true"))
		enabled = &b
	}
	event := strings.TrimSpace(q.Get("event"))

	items, err := svc.ListTargets(ctx.R.Context(), TargetFilter{
		Enabled: enabled,
		Event:   event,
	})
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errStoreError(err.Error()))
		return
	}

	out := make([]targetDTO, 0, len(items))
	for _, t := range items {
		out = append(out, targetToDTO(t))
	}

	_ = ctx.Response(http.StatusOK, map[string]any{"items": out}, nil)
}

func webhookGetTarget(ctx *contract.Ctx, svc *Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("id is required"))
		return
	}

	t, ok := svc.GetTarget(ctx.R.Context(), id)
	if !ok {
		_ = contract.WriteError(ctx.W, ctx.R, errNotFound("target not found"))
		return
	}

	_ = ctx.Response(http.StatusOK, targetToDTO(t), nil)
}

func webhookPatchTarget(ctx *contract.Ctx, svc *Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("id is required"))
		return
	}

	var req TargetPatch
	if err := ctx.BindJSON(&req, nil); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errInvalidJSON())
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, req)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest(err.Error()))
		return
	}

	_ = ctx.Response(http.StatusOK, targetToDTO(t), nil)
}

func webhookSetTargetEnabled(ctx *contract.Ctx, svc *Service, enable bool) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("id is required"))
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, TargetPatch{Enabled: &enable})
	if err != nil {
		if err == ErrTargetNotFound {
			_ = contract.WriteError(ctx.W, ctx.R, errNotFound("target not found"))
			return
		}
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest(err.Error()))
		return
	}

	_ = ctx.Response(http.StatusOK, targetToDTO(t), nil)
}

func webhookTriggerEvent(ctx *contract.Ctx, svc *Service, token string, allowEmpty bool) {
	event, ok := ctx.Param("event")
	if !ok || strings.TrimSpace(event) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("event is required"))
		return
	}

	provided := strings.TrimSpace(ctx.Query.Get("token"))
	if provided == "" {
		provided = strings.TrimSpace(ctx.RequestHeaders().Get("X-Trigger-Token"))
	}

	if token == "" && !allowEmpty {
		_ = contract.WriteError(ctx.W, ctx.R, errForbidden("triggering is disabled"))
		return
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
		_ = contract.WriteError(ctx.W, ctx.R, errUnauthorized("invalid trigger token"))
		return
	}

	var payload struct {
		Data map[string]any `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := ctx.BindJSON(&payload, nil); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errInvalidJSON())
		return
	}

	enqueued, err := svc.TriggerEvent(ctx.R.Context(), Event{Type: event, Data: payload.Data, Meta: payload.Meta})
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest(err.Error()))
		return
	}

	_ = ctx.Response(http.StatusAccepted, map[string]any{
		"enqueued": enqueued,
		"event":    event,
	}, nil)
}

func webhookListDeliveries(ctx *contract.Ctx, svc *Service, defaultLimit int) {
	q := ctx.Query

	limit := defaultLimit
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := DeliveryFilter{Limit: limit, Cursor: strings.TrimSpace(q.Get("cursor"))}

	if v := strings.TrimSpace(q.Get("target_id")); v != "" {
		filter.TargetID = &v
	}
	if v := strings.TrimSpace(q.Get("event")); v != "" {
		filter.Event = &v
	}
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		status := DeliveryStatus(v)
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
		_ = contract.WriteError(ctx.W, ctx.R, errStoreError(err.Error()))
		return
	}

	type deliveryDTO struct {
		ID           string         `json:"id"`
		TargetID     string         `json:"target_id"`
		EventID      string         `json:"event_id"`
		EventType    string         `json:"event_type"`
		Status       DeliveryStatus `json:"status"`
		Attempt      int            `json:"attempt"`
		NextAt       *time.Time     `json:"next_at,omitempty"`
		LastHTTP     int            `json:"last_http_status,omitempty"`
		LastError    string         `json:"last_error,omitempty"`
		LastDuration int            `json:"last_duration_ms,omitempty"`
		LastResp     string         `json:"last_resp_snippet,omitempty"`
		CreatedAt    time.Time      `json:"created_at"`
		UpdatedAt    time.Time      `json:"updated_at"`
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
	_ = ctx.Response(http.StatusOK, resp, nil)
}

func webhookGetDelivery(ctx *contract.Ctx, svc *Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("id is required"))
		return
	}

	d, ok := svc.GetDelivery(ctx.R.Context(), id)
	if !ok {
		_ = contract.WriteError(ctx.W, ctx.R, errNotFound("delivery not found"))
		return
	}

	payload := json.RawMessage(d.PayloadJSON)

	_ = ctx.Response(http.StatusOK, map[string]any{
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
	}, nil)
}

func webhookReplayDelivery(ctx *contract.Ctx, svc *Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest("id is required"))
		return
	}

	d, err := svc.ReplayDelivery(ctx.R.Context(), id)
	if errors.Is(err, ErrTargetNotFound) {
		_ = contract.WriteError(ctx.W, ctx.R, errNotFound("delivery not found"))
		return
	}
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, errBadRequest(err.Error()))
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{"ok": true, "delivery_id": d.ID}, nil)
}

// --- package-local error helpers (use ErrorBuilder to guarantee fully populated APIError) ---

func errBadRequest(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusBadRequest).Code("bad_request").Message(msg).Build()
}

func errInvalidJSON() contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusBadRequest).Code("invalid_json").Message("invalid JSON payload").Build()
}

func errNotFound(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusNotFound).Code("not_found").Message(msg).Build()
}

func errStoreError(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusInternalServerError).Code("store_error").Message(msg).Build()
}

func errForbidden(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusForbidden).Code("forbidden").Message(msg).Build()
}

func errUnauthorized(msg string) contract.APIError {
	return contract.NewErrorBuilder().Status(http.StatusUnauthorized).Code("unauthorized").Message(msg).Build()
}

func targetToDTO(t Target) targetDTO {
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
