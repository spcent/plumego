package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	"mini-saas-api/internal/domain/audit"
)

// AuditReader lists audit entries; audit.Recorder satisfies it.
type AuditReader interface {
	List(ctx context.Context, tenantID string, limit int) ([]audit.Entry, error)
}

// AuditHandler serves the tenant audit log.
type AuditHandler struct {
	Audit  AuditReader
	Logger plumelog.StructuredLogger
}

const defaultAuditLimit = 50

// List serves GET /api/v1/tenant/audit?limit=N (admin+; enforced in routes.go).
// Entries are newest first.
func (h AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	limit := defaultAuditLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).
				Code("audit.invalid_limit").
				Message("limit must be a positive integer").
				Build()))
			return
		}
		limit = n
	}
	entries, err := h.Audit.List(r.Context(), p.TenantID, limit)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, struct {
		Entries []any `json:"entries"`
		Total   int   `json:"total"`
	}{asAny(entries), len(entries)}, nil))
}
