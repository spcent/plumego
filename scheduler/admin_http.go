package scheduler

import (
	"encoding/json"
	"net/http"
	"strings"
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
		http.Error(w, "scheduler not configured", http.StatusServiceUnavailable)
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
		writeJSON(w, http.StatusOK, h.scheduler.List())
		return
	case strings.HasPrefix(path, "/jobs/"):
		h.handleJob(w, r, strings.TrimPrefix(path, "/jobs/"))
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
	id := JobID(parts[0])

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
