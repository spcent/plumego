// Package session defines session lifecycle types and interfaces for x/tenant.
// Authentication primitives (Principal, Authenticator, context accessors)
// remain in security/authn.
package session

import (
	"context"
	"time"
)

// SessionStatus is the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive  SessionStatus = "active"
	SessionRevoked SessionStatus = "revoked"
	SessionExpired SessionStatus = "expired"
)

// Session represents a long-lived login session, typically backing refresh tokens.
//
// Storage model guidance:
//   - Primary key: SessionID
//   - Secondary indexes: Subject, TenantID, Status, ExpiresAt, LastSeenAt
type Session struct {
	// ---- Identity ----

	// SessionID is the server-side identifier of this session.
	SessionID string

	// Subject identifies the authenticated principal (usually user id).
	Subject string

	// TenantID is optional. For SaaS/multi-tenant systems, it is strongly recommended.
	TenantID string

	// ---- Security versions ----

	// TokenVersion is a monotonic number used for forced logout.
	TokenVersion int64

	// KeyID identifies which signing key/version was used when issuing tokens.
	KeyID string

	// ---- Refresh token tracking ----

	// RefreshID is the identifier of the current refresh token in rotation.
	RefreshID string

	// PreviousRefreshID records the previous refresh id for replay detection.
	PreviousRefreshID string

	// RefreshExpiresAt is when the refresh capability expires for this session.
	RefreshExpiresAt time.Time

	// ---- Lifecycle ----

	// Status is the current lifecycle state.
	Status SessionStatus

	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	RevokedAt  time.Time

	// RevokeReason is a short machine-friendly string.
	// Example: "logout" | "password_changed" | "admin_forced" | "token_reuse_detected"
	RevokeReason string

	// ---- Client metadata (optional, privacy-aware) ----

	ClientID      string
	DeviceID      string
	UserAgentHash string
	IPHash        string

	// ---- Extensibility ----

	// Attributes is a small string map for controlled extensions (< 20 keys, no secrets).
	Attributes map[string]string
}

// RefreshClaims holds the claims extracted from a refresh token.
type RefreshClaims struct {
	SessionID    string
	RefreshID    string
	Subject      string
	TenantID     string
	KeyID        string
	TokenVersion int64
	ExpiresAt    time.Time
}

// SessionStore persists session state.
type SessionStore interface {
	CreateSession(ctx context.Context, s *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// RotateRefresh atomically rotates the refresh token ID.
	RotateRefresh(ctx context.Context, sessionID, currentRefreshID, nextRefreshID string, now time.Time) (*Session, error)

	RevokeSession(ctx context.Context, sessionID string, now time.Time, reason string) error

	ListSessionsBySubject(ctx context.Context, tenantID, subject string, limit int) ([]*Session, error)
}

// SessionValidator evaluates session and refresh validity rules.
type SessionValidator interface {
	// ValidateSession applies session status, expiry, and version checks.
	ValidateSession(ctx context.Context, s *Session, now time.Time) error
	// ValidateRefresh applies refresh id and refresh expiry checks.
	ValidateRefresh(ctx context.Context, s *Session, claims RefreshClaims, now time.Time) error
}

// RefreshManager handles refresh token issuance and rotation semantics.
type RefreshManager interface {
	IssueRefreshClaims(ctx context.Context, s *Session, now time.Time) (RefreshClaims, error)
	RotateRefreshClaims(ctx context.Context, s *Session, currentRefreshID string, now time.Time) (RefreshClaims, error)
}

// SessionManager is a composite contract for session lifecycle management.
type SessionManager interface {
	SessionStore
	SessionValidator
	RefreshManager
}
