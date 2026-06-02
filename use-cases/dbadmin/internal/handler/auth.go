package handler

import (
	"crypto/sha256"
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

	dbauthn "dbadmin/internal/domain/authn"
	"dbadmin/internal/domain/session"
)

// AuthHandler handles login, logout, and current-user endpoints.
type AuthHandler struct {
	AdminUser     string
	AdminPassword string
	AdminRole     string
	SessionTTL    time.Duration
	Sessions      *session.Store
	LoginLimiter  *LoginLimiter
	Logger        plumelog.StructuredLogger
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	User string `json:"user"`
	Role string `json:"role"`
}

// Login validates credentials and issues a session cookie.
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
	if !secureEqual(req.Username, h.AdminUser) || !secureEqual(req.Password, h.AdminPassword) {
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
		Secure:   secureCookie(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.sessionTTL().Seconds()),
	})
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: req.Username, Role: h.role()}, nil))
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
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, meResponse{User: principal.Subject, Role: h.role()}, nil))
}

func secureEqual(a, b string) bool {
	ha := sha256.Sum256([]byte(a))
	hb := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
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
