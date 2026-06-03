package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	svc    *Service
	config Config
	logger plumelog.StructuredLogger
}

// NewHandler creates a new authentication handler.
func NewHandler(svc *Service, config Config, logger plumelog.StructuredLogger) *Handler {
	return &Handler{
		svc:    svc,
		config: config,
		logger: logger,
	}
}

// Setup handles POST /api/v1/auth/setup
func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, r, http.StatusBadRequest, "Username and password are required")
		return
	}

	user, err := h.svc.Setup(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if err.Error() == "setup already completed: users already exist" {
			writeError(w, r, http.StatusConflict, err.Error())
			return
		}
		if isSetupValidationError(err) {
			writeError(w, r, http.StatusBadRequest, setupValidationMessage(err, h.config.PasswordMinLength))
			return
		}
		h.logger.Error("setup failed", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Failed to create admin user")
		return
	}

	// Auto-login after setup
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getIPAddress(r)
	_, _, token, err := h.svc.Login(r.Context(), user.Username, req.Password, userAgent, ipAddress)
	if err != nil {
		h.logger.Error("auto-login after setup failed", plumelog.Fields{"error": err.Error()})
	}

	if token != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     h.config.CookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.config.SecureCookie,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   h.config.SessionTTLHours * 3600,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user": user,
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, r, http.StatusBadRequest, "Username and password are required")
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := getIPAddress(r)

	user, session, token, err := h.svc.Login(r.Context(), req.Username, req.Password, userAgent, ipAddress)
	if err != nil {
		if err == ErrRateLimited {
			writeError(w, r, http.StatusTooManyRequests, "Too many login attempts, please try again later")
			return
		}
		if err == ErrInvalidCredentials {
			writeError(w, r, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		h.logger.Error("login error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"session": session,
	})
}

// GetStatus handles GET /api/v1/auth/status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userCount, err := h.svc.repo.GetUserCount(r.Context())
	if err != nil {
		h.logger.Error("failed to get user count", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Failed to check auth status")
		return
	}

	status := struct {
		Initialized   bool  `json:"initialized"`
		Authenticated bool  `json:"authenticated"`
		User          *User `json:"user,omitempty"`
	}{
		Initialized:   userCount > 0,
		Authenticated: UserFromContext(r.Context()) != nil,
		User:          UserFromContext(r.Context()),
	}

	writeJSON(w, http.StatusOK, status)
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session := SessionFromContext(r.Context())
	if session == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if err := h.svc.Logout(r.Context(), session.ID); err != nil {
		h.logger.Error("logout error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetCurrentUser handles GET /api/v1/auth/me
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateProfile handles PUT /api/v1/auth/me
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Locale      string `json:"locale"`
		Theme       string `json:"theme"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.svc.UpdateProfile(r.Context(), user.ID, req.DisplayName, req.Email, req.Locale, req.Theme); err != nil {
		h.logger.Error("update profile error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Fetch updated user
	updatedUser, err := h.svc.repo.GetUserByID(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("fetch updated user error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, updatedUser)
}

// ChangePassword handles POST /api/v1/auth/change-password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, r, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	if err := h.svc.ChangePassword(r.Context(), user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == ErrInvalidCredentials {
			writeError(w, r, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		h.logger.Error("change password error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Clear session cookie since all sessions are revoked
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// ListSessions handles GET /api/v1/auth/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	sessions, err := h.svc.ListSessions(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("list sessions error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
	})
}

// RevokeSession handles DELETE /api/v1/auth/sessions/:id
func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if err := h.svc.RevokeSession(r.Context(), user.ID, sessionID); err != nil {
		if err == ErrUnauthorized {
			writeError(w, r, http.StatusForbidden, "Cannot revoke this session")
			return
		}
		h.logger.Error("revoke session error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RevokeAllSessions handles POST /api/v1/auth/sessions/revoke-all
func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if err := h.svc.RevokeAllSessions(r.Context(), user.ID); err != nil {
		h.logger.Error("revoke all sessions error", plumelog.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		return
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	var errType contract.ErrorType
	switch status {
	case http.StatusBadRequest:
		errType = contract.TypeBadRequest
	case http.StatusUnauthorized:
		errType = contract.TypeUnauthorized
	case http.StatusForbidden:
		errType = contract.TypeForbidden
	case http.StatusNotFound:
		errType = contract.TypeNotFound
	case http.StatusConflict:
		errType = contract.TypeConflict
	case http.StatusTooManyRequests:
		errType = contract.TypeRateLimited
	case http.StatusInternalServerError:
		errType = contract.TypeInternal
	default:
		errType = contract.TypeInternal
	}

	apiErr := contract.NewErrorBuilder().
		Type(errType).
		Message(message).
		Build()

	contract.WriteError(w, r, apiErr)
}

func isSetupValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "password does not meet strength requirements") ||
		strings.Contains(msg, "user already exists")
}

func setupValidationMessage(err error, minLength int) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "password does not meet strength requirements"):
		return "Password must be at least " + strconv.Itoa(minLength) + " characters and include uppercase, lowercase, and a number"
	case strings.Contains(msg, "user already exists"):
		return "Username or email already exists"
	default:
		return "Invalid setup input"
	}
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
