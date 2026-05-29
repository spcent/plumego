// Package auth provides authentication and session management.
package auth

import (
	"context"
	"errors"
	"time"
)

// Standard errors
var (
	ErrUserNotFound       = errors.New("auth: user not found")
	ErrSessionNotFound    = errors.New("auth: session not found")
	ErrSessionExpired     = errors.New("auth: session expired")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrUserAlreadyExists  = errors.New("auth: user already exists")
	ErrRateLimited        = errors.New("auth: rate limited, too many attempts")
	ErrUnauthorized       = errors.New("auth: unauthorized")
	ErrForbidden          = errors.New("auth: forbidden")
	ErrInvalidInput       = errors.New("auth: invalid input")
)

// User represents an authenticated user.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Role         string    `json:"role"`
	Locale       string    `json:"locale"`
	Theme        string    `json:"theme"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents an active user session.
type Session struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	TokenHash    string     `json:"-"` // Never expose in JSON
	UserAgent    string     `json:"user_agent"`
	IPAddress    string     `json:"ip_address"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SecurityEvent represents a security-related event.
type SecurityEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	EventType string    `json:"event_type"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginAttempt represents a login attempt record.
type LoginAttempt struct {
	ID         string    `json:"id"`
	Identifier string    `json:"identifier"`
	IPAddress  string    `json:"ip_address"`
	Success    bool      `json:"success"`
	CreatedAt  time.Time `json:"created_at"`
}

// Config holds authentication configuration.
type Config struct {
	Enabled                   bool
	SessionTTLHours           int
	CookieName                string
	SecureCookie              bool
	MaxLoginFailures          int
	LoginFailureWindowMinutes int
	LockoutMinutes            int
	PasswordMinLength         int
	BootstrapAdminEnabled     bool
	BootstrapAdminUsername    string
	BootstrapAdminEmail       string
	BootstrapAdminPassword    string
}

// contextKey is a private type for context keys.
type contextKey string

const (
	userContextKey    contextKey = "auth_user"
	sessionContextKey contextKey = "auth_session"
)

// UserFromContext retrieves the authenticated user from context.
func UserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// WithUser stores the authenticated user in context.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// SessionFromContext retrieves the current session from context.
func SessionFromContext(ctx context.Context) *Session {
	session, ok := ctx.Value(sessionContextKey).(*Session)
	if !ok {
		return nil
	}
	return session
}

// WithSession stores the current session in context.
func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}
