package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"

	"dbadmin/internal/domain/audit"
)

// AuditHandler exposes audit log events.
type AuditHandler struct {
	Store  *audit.Store
	Logger plumelog.StructuredLogger
}

// List returns recent audit events.
func (h AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, []audit.Event{}, nil))
		return
	}
	events, err := h.Store.List()
	if err != nil {
		h.Logger.Error("list audit events", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list audit events").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, events, map[string]any{"count": len(events)}))
}

// Export downloads audit events as JSON or NDJSON.
func (h AuditHandler) Export(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, []audit.Event{}, nil))
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "ndjson" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code("DBADMIN_AUDIT_EXPORT_FORMAT_INVALID").
			Message("audit export format must be json or ndjson").
			Detail("format", format).
			Build()))
		return
	}
	events, err := h.Store.List()
	if err != nil {
		h.Logger.Error("export audit events", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to export audit events").Build()))
		return
	}
	filename := "dbadmin-audit-events." + format
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if format == "ndjson" {
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		for _, event := range events {
			data, err := json.Marshal(event)
			if err != nil {
				h.Logger.Warn("encode audit event", plumelog.Fields{"error": err.Error(), "event_id": event.ID})
				continue
			}
			if _, err := fmt.Fprintf(w, "%s\n", data); err != nil {
				h.Logger.Warn("write audit export", plumelog.Fields{"error": err.Error()})
				return
			}
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.Logger.Warn("write audit export", plumelog.Fields{"error": err.Error()})
	}
}

// AuditMiddleware records non-GET protected requests without request bodies.
func AuditMiddleware(store *audit.Store, role string, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if store == nil || r.Method == http.MethodGet {
				return
			}
			user := ""
			if principal := authn.PrincipalFromContext(r.Context()); principal != nil {
				user = principal.Subject
			}
			id, _ := generateHistoryID()
			if err := store.Add(audit.Event{
				ID:           id,
				User:         user,
				Role:         role,
				Action:       r.Method + " " + r.URL.Path,
				Method:       r.Method,
				Path:         r.URL.Path,
				Status:       rec.status,
				RequestID:    contract.RequestIDFromContext(r.Context()),
				RemoteAddr:   clientAddr(r),
				DeniedReason: auditDeniedReason(role, r, rec.status),
				CreatedAt:    time.Now().UTC(),
			}); err != nil {
				logger.Warn("record audit event failed", plumelog.Fields{"error": err.Error()})
			}
		})
	}
}

func auditDeniedReason(role string, r *http.Request, status int) string {
	if status != http.StatusForbidden {
		return ""
	}
	if role == "readonly" && r.Method != http.MethodGet && r.URL.Path != "/api/auth/logout" {
		return "role_readonly"
	}
	return "forbidden"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// RoleMiddleware enforces the coarse app role.
func RoleMiddleware(role string, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if role == "readonly" && r.Method != http.MethodGet && r.URL.Path != "/api/auth/logout" {
				logger.Warn("rbac denied request", plumelog.Fields{"method": r.Method, "path": r.URL.Path})
				logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeForbidden).
					Code("DBADMIN_RBAC_DENIED").
					Message("role does not allow this operation").
					Detail("role", role).
					Detail("method", r.Method).
					Detail("path", r.URL.Path).
					Detail("reason", "role_readonly").
					Build()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
