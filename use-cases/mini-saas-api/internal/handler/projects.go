package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/security/authn"
	"github.com/spcent/plumego/x/rest"
	"mini-saas-api/internal/domain/project"
)

// ProjectService is the project dependency of the controller.
type ProjectService interface {
	List(ctx context.Context, tenantID string) ([]project.Project, error)
	Get(ctx context.Context, tenantID, id string) (project.Project, error)
	Create(ctx context.Context, tenantID, createdBy, name, description string) (project.Project, error)
	Update(ctx context.Context, tenantID, id, name, description string, status project.Status) (project.Project, error)
	Delete(ctx context.Context, tenantID, id string) error
}

// ProjectsController is an x/rest ResourceController over the project service.
// It embeds rest.BaseResourceController so unimplemented REST verbs (batch,
// etc.) answer with the canonical not-implemented error, and overrides the
// CRUD verbs this API exposes. Every service call passes the caller's tenant
// ID explicitly — the controller never touches another tenant's data.
type ProjectsController struct {
	*rest.BaseResourceController
	Service ProjectService
	Audit   AuditRecorder
	Logger  plumelog.StructuredLogger
}

// NewProjectsController constructs the controller.
func NewProjectsController(svc ProjectService, audit AuditRecorder, logger plumelog.StructuredLogger) *ProjectsController {
	return &ProjectsController{
		BaseResourceController: rest.NewBaseResourceController("projects"),
		Service:                svc,
		Audit:                  audit,
		Logger:                 logger,
	}
}

// Index serves GET /api/v1/projects.
func (c *ProjectsController) Index(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, c.Logger, "auth.missing_principal", "authentication required")
		return
	}
	projects, err := c.Service.List(r.Context(), p.TenantID)
	if err != nil {
		writeDomainError(w, r, c.Logger, err)
		return
	}
	logWriteErr(c.Logger, contract.WriteResponse(w, r, http.StatusOK, struct {
		Projects []any `json:"projects"`
		Total    int   `json:"total"`
	}{asAny(projects), len(projects)}, nil))
}

// Show serves GET /api/v1/projects/:id.
func (c *ProjectsController) Show(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, c.Logger, "auth.missing_principal", "authentication required")
		return
	}
	proj, err := c.Service.Get(r.Context(), p.TenantID, router.Param(r, "id"))
	if err != nil {
		writeDomainError(w, r, c.Logger, err)
		return
	}
	logWriteErr(c.Logger, contract.WriteResponse(w, r, http.StatusOK, proj, nil))
}

type projectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Create serves POST /api/v1/projects.
func (c *ProjectsController) Create(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, c.Logger, "auth.missing_principal", "authentication required")
		return
	}
	var req projectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, c.Logger)
		return
	}
	proj, err := c.Service.Create(r.Context(), p.TenantID, p.Subject, req.Name, req.Description)
	if err != nil {
		writeDomainError(w, r, c.Logger, err)
		return
	}
	recordAudit(r.Context(), c.Audit, c.Logger, p.TenantID, p.Subject, "project.created", "project", proj.ID, proj.Name)
	logWriteErr(c.Logger, contract.WriteResponse(w, r, http.StatusCreated, proj, nil))
}

type projectUpdateRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      project.Status `json:"status"`
}

// Update serves PUT /api/v1/projects/:id.
func (c *ProjectsController) Update(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, c.Logger, "auth.missing_principal", "authentication required")
		return
	}
	var req projectUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, c.Logger)
		return
	}
	id := router.Param(r, "id")
	proj, err := c.Service.Update(r.Context(), p.TenantID, id, req.Name, req.Description, req.Status)
	if err != nil {
		writeDomainError(w, r, c.Logger, err)
		return
	}
	recordAudit(r.Context(), c.Audit, c.Logger, p.TenantID, p.Subject, "project.updated", "project", proj.ID, "")
	logWriteErr(c.Logger, contract.WriteResponse(w, r, http.StatusOK, proj, nil))
}

// Delete serves DELETE /api/v1/projects/:id (admin+; enforced in routes.go).
func (c *ProjectsController) Delete(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, c.Logger, "auth.missing_principal", "authentication required")
		return
	}
	id := router.Param(r, "id")
	if err := c.Service.Delete(r.Context(), p.TenantID, id); err != nil {
		writeDomainError(w, r, c.Logger, err)
		return
	}
	recordAudit(r.Context(), c.Audit, c.Logger, p.TenantID, p.Subject, "project.deleted", "project", id, "")
	w.WriteHeader(http.StatusNoContent)
}
