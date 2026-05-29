package collection

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

// Handler exposes collection HTTP endpoints.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// List handles GET /api/v1/collections
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.List(r.Context())
	if err != nil {
		h.internalErr(w, r, "list collections failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, result)
}

// Create handles POST /api/v1/collections
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCollectionRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	c, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	h.ok(w, r, http.StatusCreated, c)
}

// GetByID handles GET /api/v1/collections/:id
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	result, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleErr(w, r, err)
		return
	}
	h.ok(w, r, http.StatusOK, result)
}

// Update handles PUT /api/v1/collections/:id
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req UpdateCollectionRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	c, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.notFound(w, r)
			return
		}
		h.badRequest(w, r, err.Error())
		return
	}
	h.ok(w, r, http.StatusOK, c)
}

// Delete handles DELETE /api/v1/collections/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddDocument handles POST /api/v1/collections/:id/documents
func (h *Handler) AddDocument(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req AddDocumentRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.AddDocument(r.Context(), id, req); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveDocument handles DELETE /api/v1/collections/:id/documents/:document_id
func (h *Handler) RemoveDocument(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	docID := router.Param(r, "document_id")
	if err := h.svc.RemoveDocument(r.Context(), id, docID); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reorder handles PUT /api/v1/collections/:id/documents/reorder
func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req ReorderRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.Reorder(r.Context(), id, req); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateFromSearch handles POST /api/v1/collections/from-search
func (h *Handler) CreateFromSearch(w http.ResponseWriter, r *http.Request) {
	var req CreateFromSearchRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	c, err := h.svc.CreateFromSearch(r.Context(), req)
	if err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	h.ok(w, r, http.StatusCreated, c)
}

// --- helpers ---

func (h *Handler) ok(w http.ResponseWriter, r *http.Request, status int, data any) {
	logWriteErr(h.logger, contract.WriteResponse(w, r, status, data, nil))
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

func (h *Handler) internalErr(w http.ResponseWriter, r *http.Request, msg string, err error) {
	if h.logger != nil {
		h.logger.Error(msg, plumelog.Fields{"error": err.Error()})
	}
	logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).
		Message(msg).
		Build()))
}

func (h *Handler) badRequest(w http.ResponseWriter, r *http.Request, msg string) {
	logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeBadRequest).
		Message(msg).
		Build()))
}

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotFound).
		Message("collection not found").
		Build()))
}

func (h *Handler) handleErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		h.notFound(w, r)
		return
	}
	h.internalErr(w, r, "internal error", err)
}

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err != nil && logger != nil {
		logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
