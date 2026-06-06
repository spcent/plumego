package importer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

// Handler handles HTTP requests for import jobs.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
}

func NewHandler(svc *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// CreateJob handles POST /api/v1/imports
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	result, err := h.svc.CreateJob(r.Context(), req)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusCreated, result)
}

// ListSources handles GET /api/v1/imports/sources.
func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListSources()
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// ListJobs handles GET /api/v1/imports
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 20, 100)
	offset := parseIntParam(r, "offset", 0, -1)
	result, err := h.svc.ListJobs(r.Context(), limit, offset)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// GetJob handles GET /api/v1/imports/:id
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	result, err := h.svc.GetJob(r.Context(), id)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	h.writeOK(w, r, http.StatusOK, result)
}

// StartJob handles POST /api/v1/imports/:id/start
func (h *Handler) StartJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.StartJob(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PauseJob handles POST /api/v1/imports/:id/pause
func (h *Handler) PauseJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.PauseJob(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CancelJob handles POST /api/v1/imports/:id/cancel
func (h *Handler) CancelJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.CancelJob(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RetryJob handles POST /api/v1/imports/:id/retry
func (h *Handler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.RetryJob(r.Context(), id); err != nil {
		h.writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListItems handles GET /api/v1/imports/:id/items
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	statusFilter := r.URL.Query().Get("status")
	limit := parseIntParam(r, "limit", 50, 200)
	offset := parseIntParam(r, "offset", 0, -1)

	result, err := h.svc.ListItems(r.Context(), id, statusFilter, limit, offset)
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
			Message("import job not found").
			Build()))
	case errors.Is(err, ErrInvalidSource):
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(err.Error()).
			Build()))
	default:
		h.logger.Error("importer handler error", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message(err.Error()).
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
