package ai

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

// Handler exposes AI endpoints.
type Handler struct {
	svc     *Service
	enabled bool
	logger  plumelog.StructuredLogger
}

func NewHandler(svc *Service, enabled bool, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, enabled: enabled, logger: logger}
}

func (h *Handler) checkEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !h.enabled {
		h.writeErr(w, r, contract.NewErrorBuilder().Type("AI_DISABLED").Message("AI features are disabled").Build())
		return false
	}
	return true
}

func (h *Handler) ok(w http.ResponseWriter, r *http.Request, data any) {
	if err := contract.WriteResponse(w, r, http.StatusOK, data, nil); err != nil && h.logger != nil {
		h.logger.Error("ai handler: write response", plumelog.Fields{"err": err.Error()})
	}
}

func (h *Handler) accepted(w http.ResponseWriter, r *http.Request, data any) {
	if err := contract.WriteResponse(w, r, http.StatusAccepted, data, nil); err != nil && h.logger != nil {
		h.logger.Error("ai handler: write response", plumelog.Fields{"err": err.Error()})
	}
}

func (h *Handler) writeErr(w http.ResponseWriter, r *http.Request, e contract.APIError) {
	if err := contract.WriteError(w, r, e); err != nil && h.logger != nil {
		h.logger.Error("ai handler: write error", plumelog.Fields{"err": err.Error()})
	}
}

func (h *Handler) badRequest(w http.ResponseWriter, r *http.Request, msg string) {
	h.writeErr(w, r, contract.NewErrorBuilder().Type("INVALID_REQUEST").Message(msg).Build())
}

func (h *Handler) internalErr(w http.ResponseWriter, r *http.Request, err error) {
	h.writeErr(w, r, contract.NewErrorBuilder().Type("INTERNAL").Message(err.Error()).Build())
}

// POST /api/v1/ai/tasks/summary
func (h *Handler) EnqueueSummary(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	var req struct {
		DocumentID string `json:"document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DocumentID == "" {
		h.badRequest(w, r, "document_id required")
		return
	}
	task, err := h.svc.EnqueueSummary(r.Context(), req.DocumentID)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.accepted(w, r, map[string]any{"task": task})
}

// POST /api/v1/ai/tasks/qa
func (h *Handler) EnqueueQA(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	var req struct {
		Question    string   `json:"question"`
		DocumentIDs []string `json:"document_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid body")
		return
	}
	if req.Question == "" {
		h.badRequest(w, r, "question required")
		return
	}
	if len(req.DocumentIDs) == 0 {
		h.badRequest(w, r, "document_ids required; select at least one document")
		return
	}
	task, err := h.svc.EnqueueQA(r.Context(), req.Question, req.DocumentIDs)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.accepted(w, r, map[string]any{"task": task})
}

// POST /api/v1/ai/tasks/prompt-extract
func (h *Handler) EnqueuePromptExtract(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	var req struct {
		DocumentID string `json:"document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DocumentID == "" {
		h.badRequest(w, r, "document_id required")
		return
	}
	task, err := h.svc.EnqueuePromptExtract(r.Context(), req.DocumentID)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.accepted(w, r, map[string]any{"task": task})
}

// GET /api/v1/ai/tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	status := r.URL.Query().Get("status")
	limit := parseIntParam(r, "limit", 20)
	offset := parseIntParam(r, "offset", 0)

	tasks, total, err := h.svc.ListTasks(r.Context(), status, limit, offset)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.ok(w, r, map[string]any{
		"items": tasks, "total": total, "limit": limit, "offset": offset,
	})
}

// GET /api/v1/ai/tasks/:id
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	id := router.Param(r, "id")
	task, err := h.svc.GetTask(r.Context(), id)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	if task == nil {
		h.writeErr(w, r, contract.NewErrorBuilder().Type("NOT_FOUND").Message("task not found").Build())
		return
	}
	h.ok(w, r, map[string]any{"task": task})
}

// POST /api/v1/ai/tasks/:id/cancel
func (h *Handler) CancelTask(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	id := router.Param(r, "id")
	if err := h.svc.CancelTask(r.Context(), id); err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.ok(w, r, map[string]any{"ok": true})
}

// GET /api/v1/ai/documents/:id/summary
func (h *Handler) GetDocumentSummary(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	id := router.Param(r, "id")
	sum, err := h.svc.GetSummary(r.Context(), id)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	if sum == nil {
		h.writeErr(w, r, contract.NewErrorBuilder().Type("NOT_FOUND").Message("no AI summary for this document").Build())
		return
	}
	h.ok(w, r, map[string]any{"summary": sum})
}

// GET /api/v1/ai/prompts
func (h *Handler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	scenario := r.URL.Query().Get("scenario")
	limit := parseIntParam(r, "limit", 20)
	offset := parseIntParam(r, "offset", 0)
	prompts, total, err := h.svc.ListPrompts(r.Context(), scenario, limit, offset)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.ok(w, r, map[string]any{
		"items": prompts, "total": total, "limit": limit, "offset": offset,
	})
}

// GET /api/v1/ai/prompts/:id
func (h *Handler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	id := router.Param(r, "id")
	p, err := h.svc.GetPrompt(r.Context(), id)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	if p == nil {
		h.writeErr(w, r, contract.NewErrorBuilder().Type("NOT_FOUND").Message("prompt not found").Build())
		return
	}
	h.ok(w, r, map[string]any{"prompt": p})
}

// DELETE /api/v1/ai/prompts/:id
func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w, r) {
		return
	}
	id := router.Param(r, "id")
	if err := h.svc.DeletePrompt(r.Context(), id); err != nil {
		h.internalErr(w, r, err)
		return
	}
	h.ok(w, r, map[string]any{"ok": true})
}

func parseIntParam(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return def
}
