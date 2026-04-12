// Package jwt provides JSON Web Token (JWT) generation, verification, and management
// with key rotation support.
//
// This package implements a production-ready JWT system supporting multiple token types:
//   - Access tokens: Short-lived tokens for API authentication (default: 15 minutes)
//   - Refresh tokens: Long-lived tokens for obtaining new access tokens (default: 7 days)
//   - API tokens: Tokens for programmatic API access with custom expiration
//
// Features:
//   - HMAC-SHA256 signing with automatic key rotation
//   - EdDSA (Ed25519) signing for enhanced security
//   - transport-agnostic token validation with explicit middleware adapters
//   - thread-safe key management with minimal lock contention
//
// Example usage:
//
//	import (
//		"context"
//
//		authmw "github.com/spcent/plumego/middleware/auth"
//		"github.com/spcent/plumego/security/jwt"
//		kvstore "github.com/spcent/plumego/store/kv"
//	)
//
//	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: "/tmp/plumego-jwt"})
//	if err != nil {
//		// Handle store initialization error
//	}
//
//	config := jwt.DefaultJWTConfig()
//	manager, err := jwt.NewJWTManager(store, config)
//	if err != nil {
//		// Handle manager initialization error
//	}
//
//	identity := jwt.IdentityClaims{Subject: "user123"}
//	authz := jwt.AuthorizationClaims{Roles: []string{"user"}}
//	pair, err := manager.GenerateTokenPair(context.Background(), identity, authz)
//
//	// Verify a token
//	verified, err := manager.VerifyToken(context.Background(), pair.AccessToken, jwt.TokenTypeAccess)
//	if err != nil {
//		// Handle invalid or expired token
//	}
//
//	// Use with the canonical transport adapter
//	app.Use(authmw.Authenticate(manager.Authenticator(jwt.TokenTypeAccess)))
package jwt

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/security/authn"
	kvstore "github.com/spcent/plumego/store/kv"
)

// TokenType represents the semantic purpose of a JWT.
//
// JWTs can serve different purposes in an authentication system:
//   - Access tokens: Short-lived tokens for API access
//   - Refresh tokens: Long-lived tokens for obtaining new access tokens
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	// Verify an access token
//	claims, err := manager.VerifyToken(ctx, token, jwt.TokenTypeAccess)
type TokenType string

const (
	// TokenTypeAccess is used for short-lived access tokens.
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh is used for long-lived refresh tokens.
	TokenTypeRefresh TokenType = "refresh"
)

// Algorithm represents a supported signing algorithm.
//
// Supported algorithms:
//   - HS256: HMAC with SHA-256 (symmetric, uses shared secret)
//   - EdDSA: Ed25519 (asymmetric, uses public/private key pair)
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	config := jwt.DefaultJWTConfig()
//	config.Algorithm = jwt.AlgorithmEdDSA
type Algorithm string

const (
	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmEdDSA Algorithm = "EdDSA"
)

// Errors returned by JWT operations.
var (
	ErrInvalidToken     = authn.ErrInvalidToken
	ErrTokenExpired     = authn.ErrExpiredToken
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrUnknownKey       = errors.New("unknown signing key")
	ErrMissingSubject   = errors.New("subject is required")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience  = errors.New("invalid audience")
)

const (
	keyPrefix    = "jwt:keys:"
	activeKeyKey = "jwt:active"
)

// IdentityClaims captures authentication (who the subject is).
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	identity := jwt.IdentityClaims{
//		Subject: "user-123",
//		Version: 1,
//	}
type IdentityClaims struct {
	Subject string `json:"sub"`
	Version int64  `json:"ver"`
}

// AuthorizationClaims captures authorization data (what the subject can do).
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	authz := jwt.AuthorizationClaims{
//		Roles:       []string{"admin", "user"},
//		Permissions: []string{"read:users", "write:users"},
//	}
type AuthorizationClaims struct {
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// TokenClaims represents a full JWT payload.
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	claims := jwt.TokenClaims{
//		TokenID:   "token-123",
//		TokenType: jwt.TokenTypeAccess,
//		Identity: jwt.IdentityClaims{
//			Subject: "user-123",
//			Version: 1,
//		},
//		Authorization: jwt.AuthorizationClaims{
//			Roles: []string{"admin"},
//		},
//		Issuer:    "plumego",
//		Audience:  "plumego-client",
//		IssuedAt:  time.Now().Unix(),
//		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
//	}
type TokenClaims struct {
	TokenID       string              `json:"jti"`
	TokenType     TokenType           `json:"token_type"`
	Identity      IdentityClaims      `json:"identity"`
	Authorization AuthorizationClaims `json:"authorization"`
	Issuer        string              `json:"iss"`
	Audience      string              `json:"aud"`
	IssuedAt      int64               `json:"iat"`
	NotBefore     int64               `json:"nbf"`
	ExpiresAt     int64               `json:"exp"`
	KeyID         string              `json:"kid"`
}

// JWTConfig holds JWT configuration.
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	config := jwt.DefaultJWTConfig()
//	config.Issuer = "my-app"
//	config.AccessExpiration = 15 * time.Minute
//	config.RefreshExpiration = 7 * 24 * time.Hour
//	config.Algorithm = jwt.AlgorithmHS256
type JWTConfig struct {
	// Issuer is the JWT issuer (iss claim)
	Issuer string

	// Audience is the JWT audience (aud claim)
	Audience string

	// AccessExpiration is the lifetime of access tokens
	AccessExpiration time.Duration

	// RefreshExpiration is the lifetime of refresh tokens
	RefreshExpiration time.Duration

	// RotationInterval is how often to rotate signing keys
	RotationInterval time.Duration

	// Algorithm is the signing algorithm
	Algorithm Algorithm

	// ClockSkew is the tolerance for clock skew
	ClockSkew time.Duration
}

// DefaultJWTConfig returns sane defaults.
//
// Defaults:
//   - Issuer: "plumego"
//   - Audience: "plumego-client"
//   - AccessExpiration: 15 minutes
//   - RefreshExpiration: 7 days
//   - RotationInterval: 24 hours
//   - Algorithm: HS256
//   - ClockSkew: 5 seconds
//
// Example:
//
//	config := jwt.DefaultJWTConfig()
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Issuer:            "plumego",
		Audience:          "plumego-client",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		RotationInterval:  24 * time.Hour,
		Algorithm:         AlgorithmHS256,
		ClockSkew:         5 * time.Second,
	}
}

// JWTSigningKey represents a signing key with metadata.
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	key := jwt.JWTSigningKey{
//		ID:        "key-123",
//		Algorithm: jwt.AlgorithmHS256,
//		Secret:    []byte("my-secret-key"),
//		CreatedAt: time.Now(),
//	}
type JWTSigningKey struct {
	ID        string    `json:"id"`
	Algorithm Algorithm `json:"alg"`
	Secret    []byte    `json:"secret,omitempty"`
	Public    []byte    `json:"public,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenPair contains generated access and refresh tokens.
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//
//	pair, err := manager.GenerateTokenPair(ctx, identity, authz)
//	if err != nil {
//		// handle error
//	}
//	// Send pair.AccessToken and pair.RefreshToken to client
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // expiration time in seconds
	TokenType    string `json:"token_type"` // always "Bearer"
}

// JWTManager handles JWT token generation and verification.
//
// JWTManager provides a JWT signing and verification primitive with:
//   - Token generation (access and refresh tokens)
//   - Token verification and validation
//   - Key rotation
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//	import "github.com/spcent/plumego/store/kv"
//
//	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: "/tmp/plumego-jwt"})
//	if err != nil {
//		// handle error
//	}
//	config := jwt.DefaultJWTConfig()
//	manager, err := jwt.NewJWTManager(store, config)
//	if err != nil {
//		// handle error
//	}
//
//	// Generate token pair
//	identity := jwt.IdentityClaims{Subject: "user-123"}
//	authz := jwt.AuthorizationClaims{Roles: []string{"admin"}}
//	pair, err := manager.GenerateTokenPair(ctx, identity, authz)
//
//	// Verify token
//	claims, err := manager.VerifyToken(ctx, pair.AccessToken, jwt.TokenTypeAccess)
//
//	// Use with middleware/auth.Authenticate(manager.Authenticator(jwt.TokenTypeAccess))
type JWTManager struct {
	config JWTConfig
	store  *kvstore.KVStore

	mu       sync.RWMutex
	keyCache map[string]JWTSigningKey
	active   string
}

// NewJWTManager creates a new JWT manager with the given configuration and backing store.
//
// Example:
//
//	import "github.com/spcent/plumego/security/jwt"
//	import "github.com/spcent/plumego/store/kv"
//
//	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: "/tmp/plumego-jwt"})
//	if err != nil {
//		// handle error
//	}
//	config := jwt.DefaultJWTConfig()
//	manager, err := jwt.NewJWTManager(store, config)
//	if err != nil {
//		// handle error
//	}
//	defer manager.Stop()
func NewJWTManager(store *kvstore.KVStore, config JWTConfig) (*JWTManager, error) {
	if store == nil {
		return nil, errors.New("kv store is required")
	}

	mgr := &JWTManager{
		config:   config,
		store:    store,
		keyCache: make(map[string]JWTSigningKey),
	}

	if err := mgr.loadKeys(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// loadKeys reads signing keys from the KV store and ensures an active key exists.
func (m *JWTManager) loadKeys() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range m.store.Keys() {
		if strings.HasPrefix(key, keyPrefix) {
			raw, err := m.store.Get(key)
			if err != nil {
				return err
			}
			var signingKey JWTSigningKey
			if err := json.Unmarshal(raw, &signingKey); err != nil {
				return fmt.Errorf("failed to decode signing key %s: %w", key, err)
			}
			m.keyCache[signingKey.ID] = signingKey
		}
	}

	if activeRaw, err := m.store.Get(activeKeyKey); err == nil {
		m.active = string(activeRaw)
	}

	if m.active == "" {
		key, err := m.generateKeyUnsafe(m.config.Algorithm)
		if err != nil {
			return err
		}
		if err := m.persistKeyUnsafe(key); err != nil {
			return err
		}
		m.active = key.ID
		if err := m.store.Set(activeKeyKey, []byte(key.ID), 0); err != nil {
			return err
		}
	}

	return nil
}

// RotateKey generates and activates a new signing key while keeping the previous keys for verification.
func (m *JWTManager) RotateKey() (JWTSigningKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.rotateKeyUnsafe() // rotate key without lock contention
}

// rotateKeyUnsafe is the unsafe version of RotateKey, assuming the caller holds the lock.
func (m *JWTManager) rotateKeyUnsafe() (JWTSigningKey, error) {
	key, err := m.generateKeyUnsafe(m.config.Algorithm)
	if err != nil {
		return JWTSigningKey{}, err
	}

	if err := m.persistKeyUnsafe(key); err != nil {
		return JWTSigningKey{}, err
	}
	m.active = key.ID
	if err := m.store.Set(activeKeyKey, []byte(key.ID), 0); err != nil {
		return JWTSigningKey{}, err
	}
	return key, nil
}

// persistKeyUnsafe is the unsafe version of persistKeyUnsafe, assuming the caller holds the lock.
func (m *JWTManager) persistKeyUnsafe(key JWTSigningKey) error {
	encoded, err := json.Marshal(key)
	if err != nil {
		return err
	}
	if err := m.store.Set(keyPrefix+key.ID, encoded, 0); err != nil {
		return err
	}
	m.keyCache[key.ID] = key
	return nil
}

// generateKeyUnsafe is the unsafe version of generateKeyUnsafe, assuming the caller holds the lock.
func (m *JWTManager) generateKeyUnsafe(alg Algorithm) (JWTSigningKey, error) {
	kid, err := randomID()
	if err != nil {
		return JWTSigningKey{}, err
	}
	key := JWTSigningKey{ID: kid, Algorithm: alg, CreatedAt: time.Now().UTC()}
	switch alg {
	case AlgorithmHS256:
		secret := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, secret); err != nil {
			return JWTSigningKey{}, fmt.Errorf("generate hs256 secret: %w", err)
		}
		key.Secret = secret
	case AlgorithmEdDSA:
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return JWTSigningKey{}, fmt.Errorf("generate eddsa key: %w", err)
		}
		key.Secret = priv
		key.Public = pub
	default:
		return JWTSigningKey{}, fmt.Errorf("unsupported algorithm: %s", alg)
	}
	return key, nil
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// GenerateTokenPair issues a new access/refresh token pair.
func (m *JWTManager) GenerateTokenPair(ctx context.Context, identity IdentityClaims, authz AuthorizationClaims) (TokenPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// ensureRotationUnsafe ensures the key cache is up-to-date before issuing tokens.
	if err := m.ensureRotationUnsafe(); err != nil {
		return TokenPair{}, err
	}

	activeKey, ok := m.keyCache[m.active]
	if !ok {
		return TokenPair{}, ErrUnknownKey
	}

	if identity.Subject == "" {
		return TokenPair{}, ErrMissingSubject
	}

	now := time.Now().UTC()
	access, err := m.buildToken(activeKey, TokenTypeAccess, identity, authz, now, m.config.AccessExpiration)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.buildToken(activeKey, TokenTypeRefresh, identity, authz, now, m.config.RefreshExpiration)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(m.config.AccessExpiration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (m *JWTManager) buildToken(key JWTSigningKey, tokenType TokenType, identity IdentityClaims, authz AuthorizationClaims, now time.Time, ttl time.Duration) (string, error) {
	jti, err := randomID()
	if err != nil {
		return "", err
	}
	claims := TokenClaims{
		TokenID:       jti,
		TokenType:     tokenType,
		Identity:      identity,
		Authorization: authz,
		Issuer:        m.config.Issuer,
		Audience:      m.config.Audience,
		IssuedAt:      now.Unix(),
		NotBefore:     now.Unix(),
		ExpiresAt:     now.Add(ttl).Unix(),
		KeyID:         key.ID,
	}
	return signJWT(key, claims)
}

// VerifyToken verifies token signature and semantic checks.
func (m *JWTManager) VerifyToken(ctx context.Context, token string, expectedType TokenType) (*TokenClaims, error) {
	claims, err := m.parseAndVerify(token)
	if err != nil {
		return nil, err
	}

	// ensureRotationUnsafe ensures the active signing key state is current.
	if err = m.ensureRotationUnsafe(); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	skew := int64(m.config.ClockSkew.Seconds())

	if claims.ExpiresAt > 0 && now > claims.ExpiresAt+skew {
		return nil, ErrTokenExpired
	}
	if claims.NotBefore > 0 && now < claims.NotBefore-skew {
		return nil, ErrTokenNotYetValid
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *JWTManager) parseAndVerify(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header map[string]any
	if err = json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}

	kid, _ := header["kid"].(string)
	algStr, _ := header["alg"].(string)

	m.mu.RLock()
	key, ok := m.keyCache[kid]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrUnknownKey
	}
	if string(key.Algorithm) != algStr {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if err := verifySignature(key, parts[0], parts[1], parts[2]); err != nil {
		return nil, err
	}

	if claims.Issuer != "" && m.config.Issuer != "" && claims.Issuer != m.config.Issuer {
		return nil, ErrInvalidIssuer
	}
	if claims.Audience != "" && m.config.Audience != "" && claims.Audience != m.config.Audience {
		return nil, ErrInvalidAudience
	}

	return &claims, nil
}

func verifySignature(key JWTSigningKey, header, payload, sigPart string) error {
	signature, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return ErrInvalidToken
	}
	signed := header + "." + payload
	switch key.Algorithm {
	case AlgorithmHS256:
		mac := hmac.New(sha256.New, key.Secret)
		mac.Write([]byte(signed))
		if !hmac.Equal(mac.Sum(nil), signature) {
			return ErrInvalidToken
		}
	case AlgorithmEdDSA:
		if !ed25519.Verify(key.Public, []byte(signed), signature) {
			return ErrInvalidToken
		}
	default:
		return ErrInvalidToken
	}
	return nil
}

func signJWT(key JWTSigningKey, claims TokenClaims) (string, error) {
	header := map[string]any{
		"alg": key.Algorithm,
		"typ": "JWT",
		"kid": key.ID,
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart

	var signature []byte
	switch key.Algorithm {
	case AlgorithmHS256:
		mac := hmac.New(sha256.New, key.Secret)
		mac.Write([]byte(signingInput))
		signature = mac.Sum(nil)
	case AlgorithmEdDSA:
		signature = ed25519.Sign(ed25519.PrivateKey(key.Secret), []byte(signingInput))
	default:
		return "", ErrInvalidToken
	}

	sigPart := base64.RawURLEncoding.EncodeToString(signature)
	return signingInput + "." + sigPart, nil
}

// ensureRotationUnsafe rotates the active signing key if the configured interval has elapsed.
func (m *JWTManager) ensureRotationUnsafe() error {
	activeKey, ok := m.keyCache[m.active]
	if !ok {
		return ErrUnknownKey
	}
	if m.config.RotationInterval <= 0 {
		return nil
	}
	if time.Since(activeKey.CreatedAt) >= m.config.RotationInterval {
		_, err := m.rotateKeyUnsafe() // rotate key without lock contention
		return err
	}
	return nil
}

// AuthZPolicy defines role/permission requirements.
type AuthZPolicy struct {
	AnyRole        []string
	AllRoles       []string
	AnyPermission  []string
	AllPermissions []string
}

func checkPolicy(policy AuthZPolicy, auth AuthorizationClaims) bool {
	hasAll := func(required []string, actual []string) bool {
		for _, r := range required {
			if !contains(actual, r) {
				return false
			}
		}
		return true
	}
	hasAny := func(required []string, actual []string) bool {
		if len(required) == 0 {
			return true
		}
		for _, r := range required {
			if contains(actual, r) {
				return true
			}
		}
		return false
	}

	if len(policy.AllRoles) > 0 && !hasAll(policy.AllRoles, auth.Roles) {
		return false
	}
	if len(policy.AllPermissions) > 0 && !hasAll(policy.AllPermissions, auth.Permissions) {
		return false
	}
	if !hasAny(policy.AnyRole, auth.Roles) {
		return false
	}
	if !hasAny(policy.AnyPermission, auth.Permissions) {
		return false
	}
	return true
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

// WithTokenClaims stores JWT claims in request context for downstream handlers.
func WithTokenClaims(ctx context.Context, claims *TokenClaims) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, tokenClaimsContextKey{}, claims)
}

// TokenClaimsFromContext returns JWT claims from request context when present.
func TokenClaimsFromContext(ctx context.Context) *TokenClaims {
	if ctx == nil {
		return nil
	}
	claims, _ := ctx.Value(tokenClaimsContextKey{}).(*TokenClaims)
	return claims
}

type tokenClaimsContextKey struct{}
