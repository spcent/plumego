package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/internal/stringsx"
)

type Outbound struct {
	cfg WebhookOutConfig
}

func NewOutbound(cfg WebhookOutConfig) *Outbound {
	return &Outbound{cfg: cfg}
}

func (c *Outbound) RegisterRoutes(r routeRegistrar) error {
	if !c.cfg.Enabled || c.cfg.Service == nil {
		return nil
	}

	base := strings.TrimSpace(c.cfg.BasePath)
	if base == "" {
		base = "/webhooks"
	}

	svc := c.cfg.Service

	if err := r.AddRoute(http.MethodPost, base+"/targets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookCreateTarget(w, r, svc) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, base+"/targets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookListTargets(w, r, svc) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, base+"/targets/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookGetTarget(w, r, svc) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPatch, base+"/targets/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookPatchTarget(w, r, svc) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, base+"/targets/:id/enable", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookSetTargetEnabled(w, r, svc, true) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, base+"/targets/:id/disable", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookSetTargetEnabled(w, r, svc, false) })); err != nil {
		return err
	}
	token, allowEmpty := c.cfg.TriggerToken, c.cfg.AllowEmptyToken
	if err := r.AddRoute(http.MethodPost, base+"/events/:event", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookTriggerEvent(w, r, svc, token, allowEmpty)
	})); err != nil {
		return err
	}
	pageLimit := c.cfg.DefaultPageLimit
	if err := r.AddRoute(http.MethodGet, base+"/deliveries", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookListDeliveries(w, r, svc, pageLimit) })); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, base+"/deliveries/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookGetDelivery(w, r, svc) })); err != nil {
		return err
	}
	return r.AddRoute(http.MethodPost, base+"/deliveries/:id/replay", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { webhookReplayDelivery(w, r, svc) }))
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

func webhookCreateTarget(w http.ResponseWriter, r *http.Request, svc *Service) {
	var req Target
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeInvalidJSON).Message("invalid JSON payload").Build())
		return
	}

	t, err := svc.CreateTarget(r.Context(), req)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeBadRequest).Message(err.Error()).Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusCreated, targetToDTO(t), nil)
}

func webhookListTargets(w http.ResponseWriter, r *http.Request, svc *Service) {
	q := r.URL.Query()

	var enabled *bool
	if v := strings.TrimSpace(q.Get("enabled")); v != "" {
		b := (v == "1" || strings.EqualFold(v, "true"))
		enabled = &b
	}
	event := strings.TrimSpace(q.Get("event"))

	items, err := svc.ListTargets(r.Context(), TargetFilter{
		Enabled: enabled,
		Event:   event,
	})
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeInternal).Code(contract.CodeInternalError).Message(err.Error()).Build())
		return
	}

	out := make([]targetDTO, 0, len(items))
	for _, t := range items {
		out = append(out, targetToDTO(t))
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{"items": out}, nil)
}

func webhookGetTarget(w http.ResponseWriter, r *http.Request, svc *Service) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if strings.TrimSpace(id) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("id is required").Build())
		return
	}

	t, ok := svc.GetTarget(r.Context(), id)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeNotFound).Code(contract.CodeResourceNotFound).Message("target not found").Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, targetToDTO(t), nil)
}

func webhookPatchTarget(w http.ResponseWriter, r *http.Request, svc *Service) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if strings.TrimSpace(id) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("id is required").Build())
		return
	}

	var req TargetPatch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeInvalidJSON).Message("invalid JSON payload").Build())
		return
	}

	t, err := svc.UpdateTarget(r.Context(), id, req)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeBadRequest).Message(err.Error()).Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, targetToDTO(t), nil)
}

func webhookSetTargetEnabled(w http.ResponseWriter, r *http.Request, svc *Service, enable bool) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if strings.TrimSpace(id) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("id is required").Build())
		return
	}

	t, err := svc.UpdateTarget(r.Context(), id, TargetPatch{Enabled: &enable})
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeNotFound).Code(contract.CodeResourceNotFound).Message("target not found").Build())
			return
		}
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeBadRequest).Message(err.Error()).Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, targetToDTO(t), nil)
}

func webhookTriggerEvent(w http.ResponseWriter, r *http.Request, svc *Service, token string, allowEmpty bool) {
	event := contract.RequestContextFromContext(r.Context()).Params["event"]
	if strings.TrimSpace(event) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("event is required").Build())
		return
	}

	provided := strings.TrimSpace(r.URL.Query().Get("token"))
	if provided == "" {
		provided = strings.TrimSpace(r.Header.Get("X-Trigger-Token"))
	}

	if token == "" && !allowEmpty {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeForbidden).Code(contract.CodeForbidden).Message("triggering is disabled").Build())
		return
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeUnauthorized).Code(contract.CodeUnauthorized).Message("invalid trigger token").Build())
		return
	}

	var payload struct {
		Data map[string]any `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeInvalidJSON).Message("invalid JSON payload").Build())
		return
	}

	enqueued, err := svc.TriggerEvent(r.Context(), Event{Type: event, Data: payload.Data, Meta: payload.Meta})
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeBadRequest).Message(err.Error()).Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusAccepted, map[string]any{
		"enqueued": enqueued,
		"event":    event,
	}, nil)
}

func webhookListDeliveries(w http.ResponseWriter, r *http.Request, svc *Service, defaultLimit int) {
	q := r.URL.Query()

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

	deliveries, err := svc.ListDeliveries(r.Context(), filter)
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeInternal).Code(contract.CodeInternalError).Message(err.Error()).Build())
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
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}

func webhookGetDelivery(w http.ResponseWriter, r *http.Request, svc *Service) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if strings.TrimSpace(id) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("id is required").Build())
		return
	}

	d, ok := svc.GetDelivery(r.Context(), id)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeNotFound).Code(contract.CodeResourceNotFound).Message("delivery not found").Build())
		return
	}

	payload := json.RawMessage(d.PayloadJSON)

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
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

func webhookReplayDelivery(w http.ResponseWriter, r *http.Request, svc *Service) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if strings.TrimSpace(id) == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeRequired).Code(contract.CodeBadRequest).Message("id is required").Build())
		return
	}

	d, err := svc.ReplayDelivery(r.Context(), id)
	if errors.Is(err, ErrTargetNotFound) {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeNotFound).Code(contract.CodeResourceNotFound).Message("delivery not found").Build())
		return
	}
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().Type(contract.TypeValidation).Code(contract.CodeBadRequest).Message(err.Error()).Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true, "delivery_id": d.ID}, nil)
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
