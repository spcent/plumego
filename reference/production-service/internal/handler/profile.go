package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"production-service/internal/domain/tenant"
)

// ProfileStore is the minimal persistence interface that ProfileHandler depends on.
// The concrete implementation (internal/app/profileStore) satisfies this interface;
// tests may pass a stub without constructing the full app.
type ProfileStore interface {
	Get(tenantID string) (tenant.Profile, bool)
}

// ProfileHandler serves GET /api/profile.
// It reads the tenant ID from request context (populated by x/tenant/resolve middleware)
// and returns the matching profile or a 404 TypeNotFound error.
type ProfileHandler struct {
	Profiles ProfileStore
}

const codeProfileNotFound = "profile.not_found"

// GetProfile handles GET /api/profile.
//
//	GET /api/profile (with X-Tenant-ID: tenant-a) → 200 tenant.Profile
//	GET /api/profile (unknown tenant)              → 404 TypeNotFound
func (h ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantcore.TenantIDFromContext(r.Context())
	profile, ok := h.Profiles.Get(tenantID)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeProfileNotFound).
			Detail("tenant_id", tenantID).
			Message("tenant profile not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, profile, nil)
}
