package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"

	dbauthn "dbadmin/internal/domain/authn"
	"dbadmin/internal/domain/session"
)

// AuthHandler handles login, logout, and current-user endpoints.
type AuthHandler struct {
	AdminUser     string
	AdminPassword string
	Sessions      *session.Store
	Logger        plumelog.StructuredLogger
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	User string `json:"user"`
}

// Login validates credentials and issues a session cookie.
func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	if req.Username != h.AdminUser || req.Password != h.AdminPassword {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("invalid credentials").
			Build()))
		return
	}
	token, err := h.Sessions.Create(req.Username)
	if err != nil {
		h.Logger.Error("create session", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to create session").
			Build()))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     dbauthn.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: req.Username}, nil))
}

// Logout deletes the session and clears the cookie.
func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(dbauthn.CookieName())
	if err == nil && cookie.Value != "" {
		if delErr := h.Sessions.Delete(cookie.Value); delErr != nil {
			h.Logger.Warn("delete session", plumelog.Fields{"error": delErr.Error()})
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     dbauthn.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]string{"status": "ok"}, nil))
}

// Me returns the current user from the request context.
func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: principal.Subject}, nil))
}
