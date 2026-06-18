package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"

	"dbadmin/internal/config"
	dbauthn "dbadmin/internal/domain/authn"
	"dbadmin/internal/domain/mfa"
	"dbadmin/internal/domain/session"
)

// AuthHandler handles login, logout, and current-user endpoints.
type AuthHandler struct {
	// Legacy single-user fields (kept for backward compat, used when Users is nil)
	AdminUser     string
	AdminPassword string
	AdminRole     string
	// Users, when non-empty, takes precedence over AdminUser/AdminPassword/AdminRole.
	Users        []config.UserConfig
	SessionTTL   time.Duration
	Sessions     *session.Store
	MFA          *mfa.Store
	LoginLimiter *LoginLimiter
	Logger       plumelog.StructuredLogger
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	User string `json:"user"`
	Role string `json:"role"`
}

// loginResponse is returned by Login. When MFA is required for the
// authenticated user, Status is "mfa_required" and ChallengeToken must be
// presented to MFAVerify along with a TOTP code to obtain a session; no
// session cookie is set in that case. Otherwise Status is "ok" and the
// session cookie has already been set.
type loginResponse struct {
	Status         string `json:"status"`
	User           string `json:"user,omitempty"`
	Role           string `json:"role,omitempty"`
	ChallengeToken string `json:"challenge_token,omitempty"`
}

// mfaVerifyRequest carries the MFA challenge token and TOTP code submitted
// to complete a login that required a second factor.
type mfaVerifyRequest struct {
	ChallengeToken string `json:"challenge_token"`
	Code           string `json:"code"`
}

// Login validates credentials and, when the user does not have MFA enabled,
// issues a session cookie. When MFA is enabled for the user, no session is
// created; instead a short-lived challenge token is returned and the caller
// must complete login via MFAVerify.
func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	remote := clientAddr(r)
	if h.LoginLimiter != nil && !h.LoginLimiter.Allow(remote) {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRateLimited).
			Message("too many failed login attempts; try again later").
			Build()))
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	user, ok := h.findUser(req.Username, req.Password)
	if !ok {
		if h.LoginLimiter != nil {
			h.LoginLimiter.RecordFailure(remote)
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("invalid credentials").
			Build()))
		return
	}
	if h.LoginLimiter != nil {
		h.LoginLimiter.Reset(remote)
	}

	if h.MFA != nil && h.MFA.IsEnabled(user.Username) {
		token, err := h.MFA.CreateChallenge(user.Username)
		if err != nil {
			h.Logger.Error("create mfa challenge", plumelog.Fields{"error": err.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).
				Message("failed to start mfa challenge").
				Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, loginResponse{
			Status:         "mfa_required",
			ChallengeToken: token,
		}, nil))
		return
	}

	if err := h.issueSession(w, r, user.Username); err != nil {
		h.Logger.Error("create session", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to create session").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, loginResponse{
		Status: "ok",
		User:   user.Username,
		Role:   h.roleForUser(user.Username),
	}, nil))
}

// MFAVerify completes a login that required a second factor: it validates
// the TOTP code against the user named by the challenge token and, on
// success, issues a session cookie. The challenge token is single-use and is
// consumed (deleted) on the first attempt regardless of outcome, so a
// rejected code requires the caller to log in again. Fails closed: any
// lookup or validation error rejects the request without issuing a session.
func (h AuthHandler) MFAVerify(w http.ResponseWriter, r *http.Request) {
	remote := clientAddr(r)
	if h.LoginLimiter != nil && !h.LoginLimiter.Allow(remote) {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRateLimited).
			Message("too many failed login attempts; try again later").
			Build()))
		return
	}
	if h.MFA == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("mfa is not configured").
			Build()))
		return
	}
	var req mfaVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	username, err := h.MFA.ConsumeChallenge(req.ChallengeToken)
	if err != nil {
		if h.LoginLimiter != nil {
			h.LoginLimiter.RecordFailure(remote)
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeExpiredToken).
			Message("mfa challenge expired or already used; please log in again").
			Build()))
		return
	}
	if !h.MFA.VerifyCode(username, req.Code) {
		if h.LoginLimiter != nil {
			h.LoginLimiter.RecordFailure(remote)
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("invalid mfa code").
			Build()))
		return
	}
	if h.LoginLimiter != nil {
		h.LoginLimiter.Reset(remote)
	}
	if err := h.issueSession(w, r, username); err != nil {
		h.Logger.Error("create session", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to create session").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, loginResponse{
		Status: "ok",
		User:   username,
		Role:   h.roleForUser(username),
	}, nil))
}

// issueSession creates a session for username and sets the session cookie.
func (h AuthHandler) issueSession(w http.ResponseWriter, r *http.Request, username string) error {
	token, err := h.Sessions.Create(username)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     dbauthn.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.sessionTTL().Seconds()),
	})
	return nil
}

// findUser looks up a user by username and password. When h.Users is set it
// takes precedence; otherwise it falls back to the legacy single-user fields.
func (h AuthHandler) findUser(username, password string) (config.UserConfig, bool) {
	if len(h.Users) > 0 {
		var match config.UserConfig
		found := false
		for _, u := range h.Users {
			if secureEqual(username, u.Username) {
				match = u
				found = true
			}
		}
		if !found {
			// Still perform a comparison to keep timing consistent.
			secureEqual(password, "")
			return config.UserConfig{}, false
		}
		if !secureEqual(password, match.Password) {
			return config.UserConfig{}, false
		}
		return match, true
	}
	if !secureEqual(username, h.AdminUser) || !secureEqual(password, h.AdminPassword) {
		return config.UserConfig{}, false
	}
	return config.UserConfig{Username: h.AdminUser, Password: h.AdminPassword, Role: h.AdminRole}, true
}

func (h AuthHandler) sessionTTL() time.Duration {
	if h.SessionTTL <= 0 {
		return 24 * time.Hour
	}
	return h.SessionTTL
}

func (h AuthHandler) role() string {
	if h.AdminRole == "" {
		return "admin"
	}
	return h.AdminRole
}

// roleForUser resolves the role for username, checking h.Users first and
// falling back to the legacy single-user AdminRole.
func (h AuthHandler) roleForUser(username string) string {
	for _, u := range h.Users {
		if u.Username == username {
			return u.Role
		}
	}
	if h.AdminRole != "" {
		return h.AdminRole
	}
	return "admin"
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
		Secure:   secureCookie(r),
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
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: principal.Subject, Role: h.roleForUser(principal.Subject)}, nil))
}

// RevokeAllSessions deletes all sessions belonging to the current user and
// clears their session cookie.
func (h AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	count, err := h.Sessions.DeleteAllByUser(principal.Subject)
	if err != nil {
		h.Logger.Error("revoke sessions", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to revoke sessions").
			Build()))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     dbauthn.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie(r),
		MaxAge:   -1,
	})
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"status": "ok", "revoked": count}, nil))
}

// mfaEnrollResponse returns a newly generated TOTP secret and otpauth URI.
// MFA is not enabled until the secret is confirmed via MFAConfirm.
type mfaEnrollResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
}

// mfaConfirmRequest carries the 6-digit TOTP code used to confirm enrollment.
type mfaConfirmRequest struct {
	Code string `json:"code"`
}

// mfaDisableRequest carries the current password required to disable MFA.
type mfaDisableRequest struct {
	Password string `json:"password"`
}

// mfaStatusResponse reports whether the current user has MFA enabled.
type mfaStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// MFAStatus reports whether the current user has MFA enabled.
func (h AuthHandler) MFAStatus(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil || h.MFA == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, mfaStatusResponse{
		Enabled: h.MFA.IsEnabled(principal.Subject),
	}, nil))
}

// MFAEnroll generates a new TOTP secret for the current user and returns it
// along with an otpauth:// URI for use with an external authenticator app.
// MFA is not enabled by this call — the caller must confirm possession of
// the secret via MFAConfirm before it takes effect at login.
func (h AuthHandler) MFAEnroll(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil || h.MFA == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	secret, uri, err := h.MFA.StartEnrollment(principal.Subject)
	if err != nil {
		h.Logger.Error("start mfa enrollment", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to start mfa enrollment").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, mfaEnrollResponse{
		Secret:     secret,
		OTPAuthURI: uri,
	}, nil))
}

// MFAConfirm validates a TOTP code against the pending enrollment for the
// current user and, on success, enables MFA for that user. Fails closed: any
// error rejects the request and leaves MFA disabled.
func (h AuthHandler) MFAConfirm(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil || h.MFA == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	var req mfaConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	if err := h.MFA.ConfirmEnrollment(principal.Subject, req.Code); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("invalid mfa code").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, mfaStatusResponse{Enabled: true}, nil))
}

// MFADisable disables MFA for the current user after re-verifying their
// current password. Fails closed: any error leaves the existing MFA state
// unchanged.
func (h AuthHandler) MFADisable(w http.ResponseWriter, r *http.Request) {
	principal := authn.PrincipalFromContext(r.Context())
	if principal == nil || h.MFA == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("not authenticated").
			Build()))
		return
	}
	var req mfaDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid request body").
			Build()))
		return
	}
	if _, ok := h.findUser(principal.Subject, req.Password); !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnauthorized).
			Message("invalid password").
			Build()))
		return
	}
	if err := h.MFA.Disable(principal.Subject); err != nil {
		h.Logger.Error("disable mfa", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to disable mfa").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, mfaStatusResponse{Enabled: false}, nil))
}

// ListUsers returns the configured users (username and role only).
func (h AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	type userInfo struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	users := make([]userInfo, 0, len(h.Users))
	for _, u := range h.Users {
		users = append(users, userInfo{Username: u.Username, Role: u.Role})
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, users, nil))
}

func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func secureCookie(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	proto := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")))
	return proto == "https"
}

func clientAddr(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

// LoginLimiter tracks failed login attempts per client address.
type LoginLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	attempts map[string]loginAttempt
}

type loginAttempt struct {
	count      int
	firstSeen  time.Time
	blockUntil time.Time
}

func NewLoginLimiter(max int, window time.Duration) *LoginLimiter {
	if max <= 0 {
		max = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	return &LoginLimiter{max: max, window: window, attempts: make(map[string]loginAttempt)}
}

func (l *LoginLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.attempts[key]
	if !ok {
		return true
	}
	if !a.blockUntil.IsZero() && now.Before(a.blockUntil) {
		return false
	}
	if now.Sub(a.firstSeen) > l.window {
		delete(l.attempts, key)
	}
	return true
}

func (l *LoginLimiter) RecordFailure(key string) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	a := l.attempts[key]
	if a.firstSeen.IsZero() || now.Sub(a.firstSeen) > l.window {
		a = loginAttempt{firstSeen: now}
	}
	a.count++
	if a.count >= l.max {
		a.blockUntil = now.Add(l.window)
	}
	l.attempts[key] = a
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
