package handler

import (
	"encoding/json"
	"net/http"

	"guardus/internal/config"
)

// Config returns a handler that exposes whether OIDC is enabled and whether
// the caller is authenticated.
//
// guardus v1 has no OIDC, so oidc is always false; the SPA renders the
// login flow accordingly. authenticated mirrors gatus's check: if security
// is configured, it reflects whether the request carries valid credentials,
// otherwise it defaults to true.
func Config(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"oidc":          false,
			"authenticated": true,
			"announcements": []any{},
		}
		if cfg != nil && cfg.Security != nil && cfg.Security.Basic != nil {
			// If basic auth is configured, the protected /api/v1 routes will
			// 401 by themselves; here we report authenticated=false so the
			// SPA can show the login form.
			_, _, ok := r.BasicAuth()
			response["authenticated"] = ok
		}
		body, err := json.Marshal(response)
		if err != nil {
			WriteErrorString(w, r, http.StatusInternalServerError, "marshal: "+err.Error())
			return
		}
		WriteRawJSON(w, http.StatusOK, body)
	})
}
