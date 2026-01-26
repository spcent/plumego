package contract

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// SessionStatus is the lifecycle state of a session.
// Keep it stable and small. Avoid adding many states.
type SessionStatus string

const (
	SessionActive  SessionStatus = "active"
	SessionRevoked SessionStatus = "revoked"
	SessionExpired SessionStatus = "expired"
)

var (
	// ErrUnauthenticated indicates authentication is required or failed.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrUnauthorized indicates the principal lacks permission.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidToken indicates a token is malformed or unverifiable.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken indicates a token is no longer valid due to expiry.
	ErrExpiredToken = errors.New("expired token")
	// ErrSessionRevoked indicates the session is revoked.
	ErrSessionRevoked = errors.New("session revoked")
	// ErrSessionExpired indicates the session is expired.
	ErrSessionExpired = errors.New("session expired")
	// ErrRefreshReused indicates refresh token reuse was detected.
	ErrRefreshReused = errors.New("refresh token reuse detected")
	// ErrTokenVersionMismatch indicates the token/session version is no longer valid.
	ErrTokenVersionMismatch = errors.New("token version mismatch")
)

// Session represents a long-lived login session, typically backing refresh tokens.
//
// Storage model guidance:
// - Primary key: SessionID
// - Secondary indexes: Subject, TenantID, Status, ExpiresAt, LastSeenAt
type Session struct {
	// ---- Identity ----

	// SessionID is the server-side identifier of this session.
	// It must be globally unique and stable for the session lifetime.
	// Example: "ses_01H..." / UUID / ULID.
	SessionID string

	// Subject identifies the authenticated principal (usually user id).
	// This is the stable link back to the account.
	Subject string

	// TenantID is optional. For SaaS/multi-tenant systems, it is strongly recommended.
	TenantID string

	// ---- Security versions ----

	// TokenVersion is a monotonic number used for forced logout.
	// Common pattern:
	// - Store a current token_version on the user/account record
	// - Tokens/sessions must match that version
	// - Increment token_version to force all sessions invalid
	TokenVersion int64

	// KeyID identifies which signing key/version was used when issuing tokens for this session.
	// This enables global key rotation and fast invalidation.
	// Optional but recommended when supporting key rotation.
	KeyID string

	// ---- Refresh token tracking ----

	// RefreshID is the identifier of the *current* refresh token in rotation.
	// This must change on every successful refresh if rotation is enabled.
	// It enables:
	// - single-use refresh token enforcement
	// - replay detection
	RefreshID string

	// PreviousRefreshID records the previous refresh id to support "grace window"
	// or immediate replay detection. Optional.
	PreviousRefreshID string

	// RefreshExpiresAt is when the refresh capability expires for this session.
	// After this time, refresh must fail even if session is "active".
	RefreshExpiresAt time.Time

	// ---- Lifecycle ----

	Status SessionStatus

	// CreatedAt is when the session was created (login time).
	CreatedAt time.Time

	// LastSeenAt is updated on successful refresh (and optionally on access usage).
	// This supports session management UI and anomaly detection.
	LastSeenAt time.Time

	// ExpiresAt is the absolute session expiration time.
	// Often equals RefreshExpiresAt, but can differ if you want stricter server-side session lifetime.
	ExpiresAt time.Time

	// RevokedAt is set when Status becomes revoked. Optional.
	RevokedAt time.Time

	// RevokeReason is a short machine-friendly string.
	// Example: "logout" | "password_changed" | "admin_forced" | "token_reuse_detected"
	RevokeReason string

	// ---- Client metadata (optional, privacy-aware) ----

	// ClientID identifies the app/client type if multiple clients exist.
	// Example: "web" | "ios" | "android" | "cli"
	ClientID string

	// DeviceID is optional. If present, it should be a stable, non-PII identifier.
	DeviceID string

	// UserAgentHash is a hash of user-agent string (avoid storing raw UA).
	UserAgentHash string

	// IPHash is a hash of IP (avoid storing raw IP if not required).
	IPHash string

	// ---- Extensibility ----

	// Attributes is a small string map for controlled extensions.
	// RULES:
	// - must remain small (recommend < 20 keys)
	// - string-only to discourage dumping complex objects
	// - do not store secrets
	Attributes map[string]string
}

type Principal struct {
	Subject  string            // user id / service id
	TenantID string            // optional, SaaS
	Roles    []string          // small list
	Scopes   []string          // optional
	Claims   map[string]string // minimal extra (string-only to avoid abuse)
}

type principalContextKey struct{}

var principalContextKeyVar principalContextKey

// ContextWithPrincipal attaches a principal to a context.
func ContextWithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalContextKeyVar, p)
}

// PrincipalFromContext extracts a principal from a context.
func PrincipalFromContext(ctx context.Context) *Principal {
	if v := ctx.Value(principalContextKeyVar); v != nil {
		if p, ok := v.(*Principal); ok {
			return p
		}
	}
	return nil
}

// PrincipalFromRequest extracts a principal from a request context.
func PrincipalFromRequest(r *http.Request) *Principal {
	if r == nil {
		return nil
	}
	return PrincipalFromContext(r.Context())
}

// RequestWithPrincipal returns a shallow copy of r with the principal attached.
func RequestWithPrincipal(r *http.Request, p *Principal) *http.Request {
	if r == nil {
		return nil
	}
	return r.WithContext(ContextWithPrincipal(r.Context(), p))
}

type RefreshClaims struct {
	SessionID    string
	RefreshID    string
	Subject      string
	TenantID     string
	KeyID        string
	TokenVersion int64
	ExpiresAt    time.Time
}

type Authenticator interface {
	Authenticate(r *http.Request) (*Principal, error)
}

type Authorizer interface {
	Authorize(p *Principal, action string, resource string) error
}

type SessionStore interface {
	CreateSession(ctx context.Context, s *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// Rotate refresh id; should be atomic.
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
	// IssueRefreshClaims issues refresh claims for a session.
	IssueRefreshClaims(ctx context.Context, s *Session, now time.Time) (RefreshClaims, error)
	// RotateRefreshClaims validates the current refresh id and returns new claims.
	RotateRefreshClaims(ctx context.Context, s *Session, currentRefreshID string, now time.Time) (RefreshClaims, error)
}

// SessionManager is a composite contract for session lifecycle management.
type SessionManager interface {
	SessionStore
	SessionValidator
	RefreshManager
}
