package core

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/utils/stringsx"
)

// ConfigureWebhookOut mounts outbound webhook management APIs.
func (a *App) ConfigureWebhookOut() {
	cfg := a.config.WebhookOut
	if !cfg.Enabled || cfg.Service == nil {
		return
	}

	base := strings.TrimSpace(cfg.BasePath)
	if base == "" {
		base = "/webhooks"
	}

	svc := cfg.Service
	router := a.Router()

	router.PostCtx(base+"/targets", func(ctx *contract.Ctx) { a.webhookCreateTarget(ctx, svc) })
	router.GetCtx(base+"/targets", func(ctx *contract.Ctx) { a.webhookListTargets(ctx, svc) })
	router.GetCtx(base+"/targets/:id", func(ctx *contract.Ctx) { a.webhookGetTarget(ctx, svc) })
	router.PatchCtx(base+"/targets/:id", func(ctx *contract.Ctx) { a.webhookPatchTarget(ctx, svc) })
	router.PostCtx(base+"/targets/:id/enable", func(ctx *contract.Ctx) { a.webhookSetTargetEnabled(ctx, svc, true) })
	router.PostCtx(base+"/targets/:id/disable", func(ctx *contract.Ctx) { a.webhookSetTargetEnabled(ctx, svc, false) })

	router.PostCtx(base+"/events/:event", func(ctx *contract.Ctx) { a.webhookTriggerEvent(ctx, svc, cfg.TriggerToken, cfg.AllowEmptyToken) })

	router.GetCtx(base+"/deliveries", func(ctx *contract.Ctx) { a.webhookListDeliveries(ctx, svc, cfg.IncludeStats, cfg.DefaultPageLimit) })
	router.GetCtx(base+"/deliveries/:id", func(ctx *contract.Ctx) { a.webhookGetDelivery(ctx, svc) })
	router.PostCtx(base+"/deliveries/:id/replay", func(ctx *contract.Ctx) { a.webhookReplayDelivery(ctx, svc) })
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

func (a *App) webhookCreateTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	var req webhookout.Target
	if err := ctx.BindJSON(&req); err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "invalid_json", "invalid JSON payload", nil)
		return
	}

	t, err := svc.CreateTarget(ctx.R.Context(), req)
	if err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	ctx.JSON(http.StatusCreated, targetToDTO(t))
}

func (a *App) webhookListTargets(ctx *contract.Ctx, svc *webhookout.Service) {
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
		ctx.ErrorJSON(http.StatusInternalServerError, "store_error", err.Error(), nil)
		return
	}

	out := make([]targetDTO, 0, len(items))
	for _, t := range items {
		out = append(out, targetToDTO(t))
	}

	ctx.JSON(http.StatusOK, map[string]any{"items": out})
}

func (a *App) webhookGetTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "id is required", nil)
		return
	}

	t, ok := svc.GetTarget(ctx.R.Context(), id)
	if !ok {
		ctx.ErrorJSON(http.StatusNotFound, "not_found", "target not found", nil)
		return
	}

	ctx.JSON(http.StatusOK, targetToDTO(t))
}

func (a *App) webhookPatchTarget(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "id is required", nil)
		return
	}

	var req webhookout.TargetPatch
	if err := ctx.BindJSON(&req); err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "invalid_json", "invalid JSON payload", nil)
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, req)
	if err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	ctx.JSON(http.StatusOK, targetToDTO(t))
}

func (a *App) webhookSetTargetEnabled(ctx *contract.Ctx, svc *webhookout.Service, enable bool) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "id is required", nil)
		return
	}

	t, err := svc.UpdateTarget(ctx.R.Context(), id, webhookout.TargetPatch{Enabled: &enable})
	if err != nil {
		if err == webhookout.ErrNotFound {
			ctx.ErrorJSON(http.StatusNotFound, "not_found", "target not found", nil)
			return
		}
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	ctx.JSON(http.StatusOK, targetToDTO(t))
}

func (a *App) webhookTriggerEvent(ctx *contract.Ctx, svc *webhookout.Service, token string, allowEmpty bool) {
	event, ok := ctx.Param("event")
	if !ok || strings.TrimSpace(event) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "event is required", nil)
		return
	}

	provided := strings.TrimSpace(ctx.Query.Get("token"))
	if provided == "" {
		provided = strings.TrimSpace(ctx.Headers.Get("X-Trigger-Token"))
	}

	if token == "" && !allowEmpty {
		ctx.ErrorJSON(http.StatusForbidden, "forbidden", "triggering is disabled", nil)
		return
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
		ctx.ErrorJSON(http.StatusUnauthorized, "unauthorized", "invalid trigger token", nil)
		return
	}

	var payload struct {
		Data map[string]any `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "invalid_json", "invalid JSON payload", nil)
		return
	}

	enqueued, err := svc.TriggerEvent(ctx.R.Context(), webhookout.Event{Type: event, Data: payload.Data, Meta: payload.Meta})
	if err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{"enqueued": enqueued})
}

func (a *App) webhookListDeliveries(ctx *contract.Ctx, svc *webhookout.Service, includeStats bool, defaultLimit int) {
	q := ctx.Query

	limit := defaultLimit
	if limit <= 0 {
		limit = 25
	}

	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			limit = min(limit, p)
		}
	}

	status := strings.TrimSpace(q.Get("status"))
	targetID := strings.TrimSpace(q.Get("target_id"))
	event := strings.TrimSpace(q.Get("event"))

	filter := webhookout.DeliveryFilter{Limit: limit, Cursor: strings.TrimSpace(q.Get("cursor"))}
	if status != "" {
		st := webhookout.DeliveryStatus(status)
		filter.Status = &st
	}
	if targetID != "" {
		filter.TargetID = &targetID
	}
	if event != "" {
		filter.Event = &event
	}
	if t, ok := parseRFC3339(q.Get("from")); ok {
		filter.From = &t
	}
	if t, ok := parseRFC3339(q.Get("to")); ok {
		filter.To = &t
	}

	ds, err := svc.ListDeliveries(ctx.R.Context(), filter)
	if err != nil {
		ctx.ErrorJSON(http.StatusInternalServerError, "store_error", err.Error(), nil)
		return
	}

	for i := range ds {
		ds[i].PayloadJSON = nil
	}

	nextCursor := ""
	if len(ds) > 0 {
		nextCursor = ds[len(ds)-1].ID
	}

	resp := map[string]any{
		"items":       ds,
		"next_cursor": nextCursor,
	}

	includeStats = includeStats || strings.TrimSpace(q.Get("include_stats")) == "1"
	if includeStats {
		resp["stats"] = buildDeliveryStats(ds)
	}

	ctx.JSON(http.StatusOK, resp)
}

func (a *App) webhookGetDelivery(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "id is required", nil)
		return
	}

	d, ok := svc.GetDelivery(ctx.R.Context(), id)
	if !ok {
		ctx.ErrorJSON(http.StatusNotFound, "not_found", "delivery not found", nil)
		return
	}

	if !includePayload(ctx) {
		d.PayloadJSON = nil
	}
	ctx.JSON(http.StatusOK, d)
}

func (a *App) webhookReplayDelivery(ctx *contract.Ctx, svc *webhookout.Service) {
	id, ok := ctx.Param("id")
	if !ok || strings.TrimSpace(id) == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", "id is required", nil)
		return
	}

	d, err := svc.ReplayDelivery(ctx.R.Context(), id)
	if err != nil {
		if err == webhookout.ErrNotFound {
			ctx.ErrorJSON(http.StatusNotFound, "not_found", "delivery not found", nil)
			return
		}
		ctx.ErrorJSON(http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	if !includePayload(ctx) {
		d.PayloadJSON = nil
	}
	ctx.JSON(http.StatusOK, d)
}

func includePayload(ctx *contract.Ctx) bool {
	return strings.TrimSpace(ctx.Query.Get("include_payload")) == "1"
}

func parseRFC3339(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

func buildDeliveryStats(ds []webhookout.Delivery) map[string]any {
	by := map[string]int{}
	for _, d := range ds {
		by[string(d.Status)]++
	}
	return map[string]any{
		"total":     len(ds),
		"by_status": by,
	}
}

func targetToDTO(t webhookout.Target) targetDTO {
	return targetDTO{
		ID:            t.ID,
		Name:          t.Name,
		URL:           t.URL,
		Events:        t.Events,
		Enabled:       t.Enabled,
		Headers:       t.Headers,
		TimeoutMs:     t.TimeoutMs,
		MaxRetries:    t.MaxRetries,
		BackoffBaseMs: t.BackoffBaseMs,
		BackoffMaxMs:  t.BackoffMaxMs,
		RetryOn429:    t.RetryOn429,
		SecretMasked:  stringsx.MaskSecret(t.Secret),
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
