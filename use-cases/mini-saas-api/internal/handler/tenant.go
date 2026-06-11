package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	"mini-saas-api/internal/domain/tenantspace"
)

// ProjectUsage reports project count and plan limit; project.Service satisfies it.
type ProjectUsage interface {
	Usage(ctx context.Context, tenantID string) (count, limit int, err error)
}

// TenantHandler serves workspace details and updates.
type TenantHandler struct {
	Spaces   WorkspaceService
	Projects ProjectUsage
	Audit    AuditRecorder
	Logger   plumelog.StructuredLogger
}

type tenantResponse struct {
	Tenant tenantspace.Tenant `json:"tenant"`
	Usage  usage              `json:"usage"`
}

type usage struct {
	Projects     int `json:"projects"`
	ProjectLimit int `json:"project_limit"` // 0 = unlimited
}

// Get serves GET /api/v1/tenant (any member).
func (h TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	t, err := h.Spaces.Get(r.Context(), p.TenantID)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	count, limit, err := h.Projects.Usage(r.Context(), p.TenantID)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, tenantResponse{
		Tenant: t,
		Usage:  usage{Projects: count, ProjectLimit: limit},
	}, nil))
}

type tenantUpdateRequest struct {
	Name string `json:"name"`
	Plan string `json:"plan"`
}

// Update serves PATCH /api/v1/tenant (admin or owner; enforced in routes.go).
func (h TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	var req tenantUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	t, err := h.Spaces.Update(r.Context(), p.TenantID, req.Name, req.Plan)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, p.TenantID, p.Subject, "tenant.updated", "tenant", t.ID, "")
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, t, nil))
}
