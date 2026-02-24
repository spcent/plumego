package scheduler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
)

// AdminHandler exposes minimal management endpoints over net/http.
type AdminHandler struct {
	scheduler *Scheduler
	prefix    string
}

// NewAdminHandler constructs an admin handler.
func NewAdminHandler(s *Scheduler) *AdminHandler {
	return &AdminHandler{scheduler: s, prefix: "/scheduler"}
}

// WithPrefix sets the handler path prefix (default: /scheduler).
func (h *AdminHandler) WithPrefix(prefix string) *AdminHandler {
	if prefix != "" {
		h.prefix = prefix
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.scheduler == nil {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "scheduler not configured", Category: contract.CategoryServer})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, h.prefix)
	if path == "" {
		path = "/"
	}

	switch {
	case r.Method == http.MethodGet && path == "/health":
		writeJSON(w, http.StatusOK, h.scheduler.Health())
		return
	case r.Method == http.MethodGet && path == "/stats":
		writeJSON(w, http.StatusOK, h.scheduler.Stats())
		return
	case r.Method == http.MethodGet && path == "/jobs":
		query := parseJobQuery(r)
		result := h.scheduler.QueryJobs(query)
		// Return the jobs slice for backward compatibility.
		// The total count is available in the X-Total-Count header.
		w.Header().Set("X-Total-Count", strconv.Itoa(result.Total))
		writeJSON(w, http.StatusOK, result.Jobs)
		return
	case strings.HasPrefix(path, "/jobs/"):
		h.handleJob(w, r, strings.TrimPrefix(path, "/jobs/"))
		return
	// Dead letter queue endpoints
	case r.Method == http.MethodGet && path == "/dlq":
		writeJSON(w, http.StatusOK, h.scheduler.ListDeadLetters())
		return
	case r.Method == http.MethodDelete && path == "/dlq":
		count := h.scheduler.ClearDeadLetters()
		writeJSON(w, http.StatusOK, map[string]int{"cleared": count})
		return
	case strings.HasPrefix(path, "/dlq/"):
		h.handleDLQEntry(w, r, strings.TrimPrefix(path, "/dlq/"))
		return
	// Bulk group operations: POST /scheduler/groups/{group}/{action}
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/groups/"):
		h.handleGroupAction(w, r, strings.TrimPrefix(path, "/groups/"))
		return
	// Bulk tag operations: POST /scheduler/tags/{tag}/{action}
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/tags/"):
		h.handleTagAction(w, r, strings.TrimPrefix(path, "/tags/"))
		return
	default:
		http.NotFound(w, r)
	}
}

func (h *AdminHandler) handleJob(w http.ResponseWriter, r *http.Request, suffix string) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	// Validate job ID length to prevent abuse via extremely long path segments.
	const maxJobIDLen = 256
	if len(parts[0]) > maxJobIDLen {
		contract.WriteError(w, r, contract.NewValidationError("job_id", "job ID too long"))
		return
	}
	id := JobID(parts[0])

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			contract.WriteError(w, r, contract.APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed", Category: contract.CategoryClient})
			return
		}
		status, ok := h.scheduler.Status(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, status)
		return
	}

	action := parts[1]
	if r.Method != http.MethodPost {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed", Category: contract.CategoryClient})
		return
	}
	switch action {
	case "pause":
		if !h.scheduler.Pause(id) {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
	case "resume":
		if !h.scheduler.Resume(id) {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
	case "cancel":
		if !h.scheduler.Cancel(id) {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "canceled"})
	case "trigger":
		if err := h.scheduler.TriggerNow(id); err != nil {
			if errors.Is(err, ErrJobNotFound) {
				http.NotFound(w, r)
				return
			}
			contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "TRIGGER_FAILED", Message: err.Error(), Category: contract.CategoryClient})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
	default:
		http.NotFound(w, r)
	}
}

// handleDLQEntry handles per-entry dead letter queue operations.
func (h *AdminHandler) handleDLQEntry(w http.ResponseWriter, r *http.Request, suffix string) {
	const maxJobIDLen = 256
	id := strings.Trim(suffix, "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	if len(id) > maxJobIDLen {
		contract.WriteError(w, r, contract.NewValidationError("job_id", "job ID too long"))
		return
	}
	jobID := JobID(id)

	switch r.Method {
	case http.MethodGet:
		entry, ok := h.scheduler.GetDeadLetter(jobID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, entry)
	case http.MethodDelete:
		if !h.scheduler.DeleteDeadLetter(jobID) {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		contract.WriteError(w, r, contract.APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed", Category: contract.CategoryClient})
	}
}

// handleGroupAction handles bulk operations on a job group.
// Path: /groups/{group}/{action}  Method: POST
// Supported actions: pause, resume, cancel
func (h *AdminHandler) handleGroupAction(w http.ResponseWriter, r *http.Request, suffix string) {
	parts := strings.SplitN(strings.Trim(suffix, "/"), "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	group, action := parts[0], parts[1]
	switch action {
	case "pause":
		n := h.scheduler.PauseByGroup(group)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	case "resume":
		n := h.scheduler.ResumeByGroup(group)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	case "cancel":
		n := h.scheduler.CancelByGroup(group)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	default:
		http.NotFound(w, r)
	}
}

// handleTagAction handles bulk operations on jobs sharing a tag.
// Path: /tags/{tag}/{action}  Method: POST
// Supported actions: pause, resume, cancel
func (h *AdminHandler) handleTagAction(w http.ResponseWriter, r *http.Request, suffix string) {
	parts := strings.SplitN(strings.Trim(suffix, "/"), "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	tag, action := parts[0], parts[1]
	switch action {
	case "pause":
		n := h.scheduler.PauseByTags(tag)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	case "resume":
		n := h.scheduler.ResumeByTags(tag)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	case "cancel":
		n := h.scheduler.CancelByTags(tag)
		writeJSON(w, http.StatusOK, map[string]int{"affected": n})
	default:
		http.NotFound(w, r)
	}
}

// parseJobQuery builds a JobQuery from URL query parameters.
// Supported parameters:
//
//	state=running&state=failed  — filter by job state (repeatable)
//	group=mygroup               — filter by group
//	tag=t1&tag=t2               — filter by tags (all must match)
//	kind=cron&kind=delay        — filter by kind (repeatable)
//	running=true                — filter by running flag
//	paused=false                — filter by paused flag
//	order_by=next_run           — sort field: id, next_run, last_run, group
//	asc=false                   — sort direction (default true)
//	limit=50                    — max results
//	offset=0                    — pagination offset
func parseJobQuery(r *http.Request) JobQuery {
	var q JobQuery
	if r == nil {
		return q
	}
	values := r.URL.Query()

	// States
	for _, v := range values["state"] {
		if s := parseJobState(v); s != "" {
			q.States = append(q.States, s)
		}
	}

	// Group
	q.Group = values.Get("group")

	// Tags (all must be present)
	q.Tags = values["tag"]

	// Kinds
	for _, k := range values["kind"] {
		k = strings.ToLower(k)
		if k == "cron" || k == "delay" {
			q.Kinds = append(q.Kinds, k)
		}
	}

	// Running / Paused boolean filters
	if v := values.Get("running"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			q.Running = &b
		}
	}
	if v := values.Get("paused"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			q.Paused = &b
		}
	}

	// Sorting
	q.OrderBy = values.Get("order_by")
	q.Ascending = true
	if v := values.Get("asc"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			q.Ascending = b
		}
	}

	// Pagination
	if v := values.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			q.Limit = n
		}
	}
	if v := values.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			q.Offset = n
		}
	}

	return q
}

func parseJobState(value string) JobState {
	switch JobState(strings.ToLower(value)) {
	case JobStateQueued,
		JobStateScheduled,
		JobStateRunning,
		JobStateFailed,
		JobStateRetrying,
		JobStateCanceled,
		JobStateCompleted:
		return JobState(strings.ToLower(value))
	default:
		return ""
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
