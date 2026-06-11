// Package session owns auth token lifecycle: JWT access tokens plus opaque,
// single-use refresh tokens with rotation and family invalidation.
//
// Refresh tokens are 256-bit random values handed to the client once and
// stored server-side only as SHA-256 hashes. Each rotation invalidates the
// presented token; reuse of an already-rotated token revokes the whole token
// family (fail closed against token theft).
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spcent/plumego/security/jwt"
	"mini-saas-api/internal/domain/access"
	"mini-saas-api/internal/domain/ident"
)

// Sentinel errors translated to HTTP status codes by the handler layer.
var (
	ErrInvalidToken = errors.New("session: invalid or expired refresh token")
	// ErrReused marks reuse of a rotated refresh token; the token family has
	// been revoked and the client must log in again.
	ErrReused = errors.New("session: refresh token reuse detected; session revoked")
)

// KV is the narrow key-value contract the refresh store needs.
// store/kv.KVStore is adapted to this interface in app wiring; tests use an
// in-memory fake.
type KV interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Get returns found=false for missing or expired keys.
	Get(ctx context.Context, key string) (value []byte, found bool, err error)
	Delete(ctx context.Context, key string) error
}

// Record is the server-side state bound to a refresh token.
type Record struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	FamilyID string `json:"family_id"`
}

// TokenSet is the client-facing token bundle.
type TokenSet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

const (
	keyToken   = "rt:"  // + token hash → Record JSON (active token)
	keyUsed    = "rtu:" // + token hash → family ID (rotated, must not be seen again)
	keyRevoked = "rtr:" // + family ID → "1" (family revoked after reuse)
)

// Store manages refresh tokens in a KV backend.
type Store struct {
	kv  KV
	ttl time.Duration
}

// NewStore returns a refresh-token store with the given token lifetime.
func NewStore(kv KV, ttl time.Duration) *Store {
	return &Store{kv: kv, ttl: ttl}
}

// Issue creates a new refresh token in a new token family.
func (s *Store) Issue(ctx context.Context, userID, tenantID string) (string, error) {
	family, err := ident.New()
	if err != nil {
		return "", err
	}
	return s.issue(ctx, Record{UserID: userID, TenantID: tenantID, FamilyID: family})
}

func (s *Store) issue(ctx context.Context, rec Record) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(rec)
	if err != nil {
		return "", err
	}
	if err := s.kv.Set(ctx, keyToken+hashToken(token), payload, s.ttl); err != nil {
		return "", err
	}
	return token, nil
}

// Rotate consumes a refresh token and returns its record plus a replacement
// token in the same family. Reuse of an already-rotated token revokes the
// family and returns ErrReused.
func (s *Store) Rotate(ctx context.Context, token string) (Record, string, error) {
	h := hashToken(token)

	raw, found, err := s.kv.Get(ctx, keyToken+h)
	if err != nil {
		return Record{}, "", err
	}
	if !found {
		// Not an active token. If it was rotated before, this is reuse:
		// revoke the whole family so the stolen successor stops working too.
		famRaw, usedBefore, err := s.kv.Get(ctx, keyUsed+h)
		if err != nil {
			return Record{}, "", err
		}
		if usedBefore {
			if err := s.kv.Set(ctx, keyRevoked+string(famRaw), []byte("1"), s.ttl); err != nil {
				return Record{}, "", err
			}
			return Record{}, "", ErrReused
		}
		return Record{}, "", ErrInvalidToken
	}

	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return Record{}, "", fmt.Errorf("session: decode record: %w", err)
	}

	// A revoked family rejects even its currently active token.
	if _, revoked, err := s.kv.Get(ctx, keyRevoked+rec.FamilyID); err != nil {
		return Record{}, "", err
	} else if revoked {
		_ = s.kv.Delete(ctx, keyToken+h)
		return Record{}, "", ErrReused
	}

	// Consume: deactivate this token and remember it as used.
	if err := s.kv.Delete(ctx, keyToken+h); err != nil {
		return Record{}, "", err
	}
	if err := s.kv.Set(ctx, keyUsed+h, []byte(rec.FamilyID), s.ttl); err != nil {
		return Record{}, "", err
	}

	next, err := s.issue(ctx, rec)
	if err != nil {
		return Record{}, "", err
	}
	return rec, next, nil
}

// Issuer bundles JWT access-token generation with the refresh store.
type Issuer struct {
	JWT     *jwt.JWTManager
	Refresh *Store
}

// tenantPermissionPrefix carries the tenant ID inside the JWT authorization
// claims; guard middleware lifts it onto authn.Principal.TenantID.
const tenantPermissionPrefix = "tenant:"

// IssueTokens returns an access token plus a refresh token in a new family.
func (i *Issuer) IssueTokens(ctx context.Context, userID, tenantID string, role access.Role) (TokenSet, error) {
	accessToken, expiresIn, err := i.IssueAccess(ctx, userID, tenantID, role)
	if err != nil {
		return TokenSet{}, err
	}
	refresh, err := i.Refresh.Issue(ctx, userID, tenantID)
	if err != nil {
		return TokenSet{}, err
	}
	return TokenSet{
		AccessToken:  accessToken,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// IssueAccess returns a short-lived access JWT carrying the tenant and role.
func (i *Issuer) IssueAccess(ctx context.Context, userID, tenantID string, role access.Role) (string, int64, error) {
	pair, err := i.JWT.GenerateTokenPair(ctx,
		jwt.IdentityClaims{Subject: userID},
		jwt.AuthorizationClaims{
			Roles:       []string{string(role)},
			Permissions: []string{tenantPermissionPrefix + tenantID},
		})
	if err != nil {
		return "", 0, err
	}
	// The manager's own refresh JWT is intentionally discarded: refresh
	// lifecycle is owned by Store so rotation and reuse detection apply.
	return pair.AccessToken, pair.ExpiresIn, nil
}

// RotateRefresh consumes a refresh token; the handler re-derives the caller's
// current role from the membership store before issuing a new access token.
func (i *Issuer) RotateRefresh(ctx context.Context, token string) (Record, string, error) {
	return i.Refresh.Rotate(ctx, token)
}

// TenantFromClaims extracts the tenant ID carried in the token's permissions.
func TenantFromClaims(claims *jwt.TokenClaims) string {
	if claims == nil {
		return ""
	}
	for _, p := range claims.Authorization.Permissions {
		if len(p) > len(tenantPermissionPrefix) && p[:len(tenantPermissionPrefix)] == tenantPermissionPrefix {
			return p[len(tenantPermissionPrefix):]
		}
	}
	return ""
}

// RoleFromClaims extracts the membership role carried in the token.
func RoleFromClaims(claims *jwt.TokenClaims) access.Role {
	if claims == nil || len(claims.Authorization.Roles) == 0 {
		return ""
	}
	return access.Role(claims.Authorization.Roles[0])
}

func newToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("session: generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
