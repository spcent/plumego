package document

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"cloud-vault/internal/audit"
	"cloud-vault/internal/auth"
)

// Handler handles HTTP requests for the document resource.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
	audit  *audit.Logger
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service, logger plumelog.StructuredLogger, auditLogger *audit.Logger) *Handler {
	return &Handler{svc: svc, logger: logger, audit: auditLogger}
}

// logAudit records a document mutation, using the authenticated user from
// the request context as the actor. It never blocks or fails the request.
func (h *Handler) logAudit(r *http.Request, action, docID string, detail map[string]any) {
	if h.audit == nil {
		return
	}
	var actorID string
	if u := auth.UserFromContext(r.Context()); u != nil {
		actorID = u.ID
	}
	h.audit.Log(r.Context(), actorID, audit.ClientIP(r), action, audit.ResourceDocument, docID, detail)
}

// List handles GET /api/v1/documents
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var isFavorite *bool
	if fav := q.Get("is_favorite"); fav != "" {
		b := fav == "1" || strings.EqualFold(fav, "true")
		isFavorite = &b
	}

	result, err := h.svc.List(r.Context(), ListQuery{
		Q:            q.Get("q"),
		TagID:        q.Get("tag_id"),
		Status:       q.Get("status"),
		SourceType:   q.Get("source_type"),
		ImportJobID:  q.Get("import_job_id"),
		IsFavorite:   isFavorite,
		ReviewStatus: q.Get("review_status"),
		SortBy:       q.Get("sort_by"),
		Order:        q.Get("order"),
		Limit:        parseIntParam(r, "limit", 50, 100),
		Offset:       parseIntParam(r, "offset", 0, -1),
		AfterID:      q.Get("after_id"),
	})
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// Create handles POST /api/v1/documents
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionCreate, result.ID, map[string]any{"title": req.Title})
	h.writeOK(w, r, http.StatusCreated, result)
}

// GetByID handles GET /api/v1/documents/:id
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	detail, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, detail)
}

// Update handles PUT /api/v1/documents/:id
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req UpdateRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionUpdate, id, nil)
	h.writeOK(w, r, http.StatusOK, result)
}

// Delete handles DELETE /api/v1/documents/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.logAudit(r, audit.ActionDelete, id, nil)
	w.WriteHeader(http.StatusNoContent)
}

// UpdateFavorite handles PUT /api/v1/documents/:id/favorite
func (h *Handler) UpdateFavorite(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req FavoriteRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.UpdateFavorite(r.Context(), id, req.IsFavorite); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateStatus handles PUT /api/v1/documents/:id/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req StatusRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateReviewStatus handles PUT /api/v1/documents/:id/review-status
func (h *Handler) UpdateReviewStatus(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req ReviewStatusRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.UpdateReviewStatus(r.Context(), id, req.ReviewStatus); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BatchUpdateStatus handles PUT /api/v1/documents/batch-status
func (h *Handler) BatchUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req BatchStatusRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if len(req.IDs) == 0 {
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("ids must not be empty").
			Build()))
		return
	}
	if err := h.svc.BatchUpdateStatus(r.Context(), req.IDs, req.Status); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetVersions handles GET /api/v1/documents/:id/versions
func (h *Handler) GetVersions(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	result, err := h.svc.GetVersions(r.Context(), id)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// CreateSnapshot handles POST /api/v1/documents/:id/versions
func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var req CreateSnapshotRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.CreateSnapshot(r.Context(), id, req.Note)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusCreated, result)
}

// GetVersion handles GET /api/v1/documents/:id/versions/:version
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	versionStr := router.Param(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("version must be an integer").
			Build()))
		return
	}

	result, err := h.svc.GetVersion(r.Context(), id, version)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// RestoreVersion handles POST /api/v1/documents/:id/versions/:version/restore
func (h *Handler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	versionStr := router.Param(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("version must be an integer").
			Build()))
		return
	}

	result, err := h.svc.RestoreVersion(r.Context(), id, version)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
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
			Message("document not found").
			Build()))
	case errors.Is(err, ErrVersionConflict):
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeConflict).
			Message("document has been updated by another session").
			Build()))
	default:
		h.logger.Error("document handler error", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("internal server error").
			Build()))
	}
}

func parseIntParam(r *http.Request, name string, defaultVal, maxVal int) int {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return defaultVal
	}
	if maxVal > 0 && n > maxVal {
		return maxVal
	}
	return n
}

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err != nil && logger != nil {
		logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
