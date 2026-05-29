package organize

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

// Handler exposes organize HTTP endpoints.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// --- Duplicates ---

// DetectDuplicates handles POST /api/v1/organize/detect-duplicates
func (h *Handler) DetectDuplicates(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.DetectDuplicates(r.Context())
	if err != nil {
		h.internalErr(w, r, "detect duplicates failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// ListDuplicates handles GET /api/v1/organize/duplicates
func (h *Handler) ListDuplicates(w http.ResponseWriter, r *http.Request) {
	groups, err := h.svc.ListDuplicates(r.Context())
	if err != nil {
		h.internalErr(w, r, "list duplicates failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{
		"groups": groups,
		"total":  len(groups),
	})
}

// ResolveDuplicates handles POST /api/v1/organize/duplicates/resolve
func (h *Handler) ResolveDuplicates(w http.ResponseWriter, r *http.Request) {
	var req ResolveDuplicatesRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.ResolveDuplicates(r.Context(), req); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Similarity ---

// DetectSimilarity handles POST /api/v1/organize/detect-similarity
func (h *Handler) DetectSimilarity(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.DetectSimilarity(r.Context())
	if err != nil {
		h.internalErr(w, r, "detect similarity failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// GetSimilarDocuments handles GET /api/v1/documents/:id/similar
func (h *Handler) GetSimilarDocuments(w http.ResponseWriter, r *http.Request) {
	docID := router.Param(r, "id")
	docs, err := h.svc.GetSimilarDocuments(r.Context(), docID)
	if err != nil {
		h.internalErr(w, r, "get similar documents failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{"items": docs})
}

// IgnoreSimilarity handles POST /api/v1/organize/similarity/:id/ignore
func (h *Handler) IgnoreSimilarity(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.IgnoreSimilarity(r.Context(), id); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ConfirmSimilarity handles POST /api/v1/organize/similarity/:id/confirm
func (h *Handler) ConfirmSimilarity(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.ConfirmSimilarity(r.Context(), id); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Tag suggestions ---

// SuggestTags handles POST /api/v1/organize/suggest-tags
func (h *Handler) SuggestTags(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.SuggestTags(r.Context())
	if err != nil {
		h.internalErr(w, r, "suggest tags failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// GetTagSuggestions handles GET /api/v1/documents/:id/tag-suggestions
func (h *Handler) GetTagSuggestions(w http.ResponseWriter, r *http.Request) {
	docID := router.Param(r, "id")
	suggestions, err := h.svc.GetTagSuggestions(r.Context(), docID)
	if err != nil {
		h.internalErr(w, r, "get tag suggestions failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{"items": suggestions})
}

// AcceptTagSuggestion handles POST /api/v1/tag-suggestions/:id/accept
func (h *Handler) AcceptTagSuggestion(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.AcceptTagSuggestion(r.Context(), id); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RejectTagSuggestion handles POST /api/v1/tag-suggestions/:id/reject
func (h *Handler) RejectTagSuggestion(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if err := h.svc.RejectTagSuggestion(r.Context(), id); err != nil {
		h.handleErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BatchAcceptTagSuggestions handles POST /api/v1/tag-suggestions/batch/accept
func (h *Handler) BatchAcceptTagSuggestions(w http.ResponseWriter, r *http.Request) {
	var req BatchAcceptTagSuggestionsRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	accepted, err := h.svc.BatchAcceptTagSuggestions(r.Context(), req.IDs)
	if err != nil {
		h.internalErr(w, r, "batch accept failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{"accepted": accepted})
}

// --- Topics ---

// BuildTopics handles POST /api/v1/organize/build-topics
func (h *Handler) BuildTopics(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.BuildTopics(r.Context())
	if err != nil {
		h.internalErr(w, r, "build topics failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// ListTopics handles GET /api/v1/topics
func (h *Handler) ListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.svc.ListTopics(r.Context())
	if err != nil {
		h.internalErr(w, r, "list topics failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{"items": topics, "total": len(topics)})
}

// GetTopic handles GET /api/v1/topics/:id
func (h *Handler) GetTopic(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	topic, err := h.svc.GetTopic(r.Context(), id)
	if err != nil {
		h.handleErr(w, r, err)
		return
	}

	limit := parseIntParam(r, "limit", 50, 100)
	offset := parseIntParam(r, "offset", 0, -1)
	docs, total, err := h.svc.GetTopicDocuments(r.Context(), id, limit, offset)
	if err != nil {
		h.internalErr(w, r, "get topic documents failed", err)
		return
	}

	h.ok(w, r, http.StatusOK, map[string]any{
		"topic":     topic,
		"documents": docs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// --- Quality + prompts ---

// ScoreQuality handles POST /api/v1/organize/score-quality
func (h *Handler) ScoreQuality(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.ScoreQuality(r.Context())
	if err != nil {
		h.internalErr(w, r, "score quality failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// DetectPromptCandidates handles POST /api/v1/organize/detect-prompt-candidates
func (h *Handler) DetectPromptCandidates(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.DetectPromptCandidates(r.Context())
	if err != nil {
		h.internalErr(w, r, "detect prompt candidates failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// RunAll handles POST /api/v1/organize/run-all
func (h *Handler) RunAll(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RunAll(r.Context()); err != nil {
		h.internalErr(w, r, "run all failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]string{"status": "done"})
}

// --- Review queue ---

// GetReviewQueue handles GET /api/v1/review/queue
func (h *Handler) GetReviewQueue(w http.ResponseWriter, r *http.Request) {
	q := ReviewQueueQuery{
		Type:   r.URL.Query().Get("type"),
		Limit:  parseIntParam(r, "limit", 50, 100),
		Offset: parseIntParam(r, "offset", 0, -1),
	}
	result, err := h.svc.GetReviewQueue(r.Context(), q)
	if err != nil {
		h.internalErr(w, r, "get review queue failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, result)
}

// --- Jobs ---

// ListJobs handles GET /api/v1/organize/jobs
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.svc.ListJobs(r.Context())
	if err != nil {
		h.internalErr(w, r, "list jobs failed", err)
		return
	}
	h.ok(w, r, http.StatusOK, map[string]any{"items": jobs})
}

// GetJob handles GET /api/v1/organize/jobs/:id
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	job, err := h.svc.GetJob(r.Context(), id)
	if err != nil {
		h.handleErr(w, r, err)
		return
	}
	h.ok(w, r, http.StatusOK, job)
}

// --- helpers ---

func (h *Handler) ok(w http.ResponseWriter, r *http.Request, status int, data any) {
	switch status {
	case http.StatusCreated:
		logWriteErr(h.logger, contract.WriteResponse(w, r, http.StatusCreated, data, nil))
	default:
		logWriteErr(h.logger, contract.WriteResponse(w, r, http.StatusOK, data, nil))
	}
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

func (h *Handler) handleErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		logWriteErr(h.logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("not found").
			Build()))
		return
	}
	h.internalErr(w, r, "internal error", err)
}

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err != nil && logger != nil {
		logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}

func parseIntParam(r *http.Request, key string, defaultVal, maxVal int) int {
	s := r.URL.Query().Get(key)
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
