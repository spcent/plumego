package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	"mini-saas-api/internal/domain/tenantspace"
	"mini-saas-api/internal/domain/user"
)

// MeHandler returns the authenticated caller's account and membership.
type MeHandler struct {
	Users  UserService
	Spaces WorkspaceService
	Logger plumelog.StructuredLogger
}

type meResponse struct {
	User       user.User              `json:"user"`
	Membership tenantspace.Membership `json:"membership"`
}

// Get serves GET /api/v1/me. RequireAuth guarantees a principal is present.
func (h MeHandler) Get(w http.ResponseWriter, r *http.Request) {
	p := authn.PrincipalFromContext(r.Context())
	if p == nil {
		writeUnauthorized(w, r, h.Logger, "auth.missing_principal", "authentication required")
		return
	}
	u, err := h.Users.ByID(r.Context(), p.Subject)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	m, err := h.Spaces.MembershipForUser(r.Context(), p.TenantID, p.Subject)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: u, Membership: m}, nil))
}
