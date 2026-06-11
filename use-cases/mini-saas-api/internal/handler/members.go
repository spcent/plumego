package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/security/authn"
	"mini-saas-api/internal/domain/access"
	"mini-saas-api/internal/domain/audit"
)

// MembersHandler serves membership administration.
type MembersHandler struct {
	Users  UserService
	Spaces WorkspaceService
	Audit  AuditRecorder
	Logger plumelog.StructuredLogger
}

// List serves GET /api/v1/tenant/members (any member).
func (h MembersHandler) List(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	members, err := h.Spaces.Members(r.Context(), p.TenantID)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, struct {
		Members []any `json:"members"`
		Total   int   `json:"total"`
	}{asAny(members), len(members)}, nil))
}

type addMemberRequest struct {
	Email string      `json:"email"`
	Role  access.Role `json:"role"`
}

// Add serves POST /api/v1/tenant/members (admin+). The user must already have
// an account; there is no email delivery in this use-case.
func (h MembersHandler) Add(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	u, err := h.Users.ByEmail(r.Context(), req.Email)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	m, err := h.Spaces.AddMember(r.Context(), p.TenantID, u.ID, req.Role)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, p.TenantID, p.Subject, "member.added", "membership", m.ID, "user "+u.Email+" as "+string(m.Role))
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, m, nil))
}

type changeRoleRequest struct {
	Role access.Role `json:"role"`
}

// ChangeRole serves PATCH /api/v1/tenant/members/:id (admin+).
func (h MembersHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	var req changeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	id := router.Param(r, "id")
	m, err := h.Spaces.ChangeRole(r.Context(), p.TenantID, id, req.Role)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, p.TenantID, p.Subject, "member.role_changed", "membership", m.ID, "now "+string(m.Role))
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, m, nil))
}

// Remove serves DELETE /api/v1/tenant/members/:id (admin+).
func (h MembersHandler) Remove(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	id := router.Param(r, "id")
	if err := h.Spaces.RemoveMember(r.Context(), p.TenantID, id); err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, p.TenantID, p.Subject, "member.removed", "membership", id, "")
	w.WriteHeader(http.StatusNoContent)
}

// recordAudit appends an audit entry; failures are logged, never surfaced.
func recordAudit(ctx context.Context, rec AuditRecorder, logger plumelog.StructuredLogger, tenantID, actorID, action, resourceType, resourceID, detail string) {
	if rec == nil {
		return
	}
	if err := rec.Record(ctx, audit.Entry{
		TenantID:     tenantID,
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
	}); err != nil && logger != nil {
		logger.Warn("audit record failed", plumelog.Fields{"error": err.Error()})
	}
}

// asAny converts a typed slice for the wrapped list envelope.
func asAny[T any](in []T) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
