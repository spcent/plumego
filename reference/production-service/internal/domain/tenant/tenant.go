// Package tenant contains the canonical per-tenant domain model for the
// production reference application.
package tenant

// Profile is the per-tenant profile resource. It is the shared domain type
// used by both the persistence layer (internal/app/profiles.go) and the HTTP
// handler layer (internal/handler/profile.go), eliminating duplicate struct
// definitions and keeping the JSON contract in one place.
type Profile struct {
	TenantID string   `json:"tenant_id"`
	Name     string   `json:"name"`
	Plan     string   `json:"plan"`
	Features []string `json:"features"`
}
