package security

// Sessions is a placeholder for the OIDC session store described in the
// roadmap.
//
// v1 ships only Basic auth, so a session store is not wired into routes.
// The stub exists so that v1.1's OIDC implementation can extend it without
// rewriting callers.
type Sessions struct{}

// NewSessions returns a no-op session store.
func NewSessions() *Sessions { return &Sessions{} }
