package tag

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"cloud-vault/internal/audit"
	"cloud-vault/internal/auth"
)

// Handler handles HTTP requests for the tag resource.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
	audit  *audit.Logger
}

func NewHandler(svc *Service, logger plumelog.StructuredLogger, auditLogger *audit.Logger) *Handler {
	return &Handler{svc: svc, logger: logger, audit: auditLogger}
}

// logAudit records a tag mutation, using the authenticated user from the
// request context as the actor. It never blocks or fails the request.
func (h *Handler) logAudit(r *http.Request, action, tagID string, detail map[string]any) {
	if h.audit == nil {
		return
	}
	var actorID string
	if u := auth.UserFromContext(r.Context()); u != nil {
		actorID = u.ID
	}
	h.audit.Log(r.Context(), actorID, audit.ClientIP(r), action, audit.ResourceTag, tagID, detail)
}

// ListTags handles GET /api/v1/tags
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListTags(r.Context())
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// CreateTag handles POST /api/v1/tags
func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req CreateTagRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.CreateTag(r.Context(), req)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionCreate, result.ID, map[string]any{"name": result.Name})
	h.writeOK(w, r, http.StatusCreated, result)
}

// UpdateTag handles PUT /api/v1/tags/:id
func (h *Handler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req UpdateTagRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.UpdateTag(r.Context(), id, req)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionUpdate, id, nil)
	h.writeOK(w, r, http.StatusOK, result)
}

// DeleteTag handles DELETE /api/v1/tags/:id
func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.DeleteTag(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionDelete, id, nil)
	w.WriteHeader(http.StatusNoContent)
}

// GetDocumentTags handles GET /api/v1/documents/:id/tags
func (h *Handler) GetDocumentTags(w http.ResponseWriter, r *http.Request) {
	docID := router.Param(r, "id")
	result, err := h.svc.GetDocumentTags(r.Context(), docID)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// SetDocumentTags handles PUT /api/v1/documents/:id/tags
func (h *Handler) SetDocumentTags(w http.ResponseWriter, r *http.Request) {
	docID := router.Param(r, "id")
	var req SetDocumentTagsRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.SetDocumentTags(r.Context(), docID, req); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveDocumentTag handles DELETE /api/v1/documents/:id/tags/:tag_id
func (h *Handler) RemoveDocumentTag(w http.ResponseWriter, r *http.Request) {
	docID := router.Param(r, "id")
	tagID := router.Param(r, "tag_id")
	if err := h.svc.RemoveDocumentTag(r.Context(), docID, tagID); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(contract.CodeInvalidJSON).
			Message("invalid JSON request body").
			Build()))
		return false
	}
	return true
}

func (h *Handler) writeOK(w http.ResponseWriter, r *http.Request, status int, data any) {
	switch status {
	case http.StatusCreated:
		logWriteErr(h.logger, contract.WriteResponse(w, r, http.StatusCreated, data, nil))
	default:
		logWriteErr(h.logger, contract.WriteResponse(w, r, http.StatusOK, data, nil))
	}
}

func (h *Handler) writeErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("tag not found").
			Build()))
	case errors.Is(err, ErrDuplicate):
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeConflict).
			Message("tag name already exists").
			Build()))
	default:
		h.logger.Error("tag handler error", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("internal server error").
			Build()))
	}
}

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err != nil && logger != nil {
		logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
