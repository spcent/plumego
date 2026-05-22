// Package handler contains HTTP handlers for the with-tenant reference app.
package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

// ModelsHandler returns the models available to the resolved tenant.
// The tenant identity is extracted from context — it was placed there by
// the resolve middleware that runs before this handler in the per-route chain.
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantcore.TenantIDFromContext(r.Context())
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"tenant_id": tenantID,
		"models":    []string{"gpt-4o"},
	}, nil)
}
