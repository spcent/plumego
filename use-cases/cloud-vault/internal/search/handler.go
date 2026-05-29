package search

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler exposes search HTTP endpoints.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Search handles GET /api/v1/search.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var isFavorite *bool
	if fav := q.Get("is_favorite"); fav != "" {
		b := fav == "1" || strings.EqualFold(fav, "true")
		isFavorite = &b
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	query := SearchQuery{
		Q:            q.Get("q"),
		TagID:        q.Get("tag"),
		Status:       q.Get("status"),
		ReviewStatus: q.Get("review_status"),
		SourceType:   q.Get("source_type"),
		ImportJobID:  q.Get("import_job_id"),
		IsFavorite:   isFavorite,
		From:         q.Get("from"),
		To:           q.Get("to"),
		Sort:         q.Get("sort"),
		Order:        q.Get("order"),
		Limit:        limit,
		Offset:       offset,
	}

	result, err := h.svc.Search(r.Context(), query)
	if err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("search failed").
			Build()))
		return
	}
	h.logW(contract.WriteResponse(w, r, http.StatusOK, result, nil))
}

// GetIndexStatus handles GET /api/v1/search/index-status.
func (h *Handler) GetIndexStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.GetIndexStatus(r.Context())
	if err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to get index status").
			Build()))
		return
	}
	h.logW(contract.WriteResponse(w, r, http.StatusOK, status, nil))
}

// Reindex handles POST /api/v1/search/reindex.
func (h *Handler) Reindex(w http.ResponseWriter, r *http.Request) {
	var req ReindexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	if strings.TrimSpace(req.Scope) == "" {
		req.Scope = "all"
	}
	if err := h.svc.Reindex(r.Context(), req); err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(err.Error()).
			Build()))
		return
	}
	h.logW(contract.WriteResponse(w, r, http.StatusOK,
		map[string]string{"status": "reindex scheduled"}, nil))
}

// GetHistory handles GET /api/v1/search/history.
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetHistory(r.Context())
	if err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to get history").
			Build()))
		return
	}
	h.logW(contract.WriteResponse(w, r, http.StatusOK, resp, nil))
}

// ClearHistory handles DELETE /api/v1/search/history.
func (h *Handler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ClearHistory(r.Context()); err != nil {
		h.logW(contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to clear history").
			Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) logW(err error) {
	if err != nil && h.logger != nil {
		h.logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
