package websocket

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/spcent/plumego/security/password"
)

// SecurityConfig defines security-related configurations for WebSocket
type SecurityConfig struct {
	// JWTSecret is the secret key for JWT verification (HS256)
	// Minimum length: 32 bytes (256 bits) recommended
	JWTSecret []byte

	// MinJWTSecretLength enforces minimum JWT secret length
	// Default: 32 bytes
	MinJWTSecretLength int

	// RoomPasswordConfig defines strength requirements for room passwords
	RoomPasswordConfig password.PasswordStrengthConfig

	// EnforcePasswordStrength if true, rejects weak passwords
	// Default: true in production
	EnforcePasswordStrength bool

	// MaxMessageSize limits incoming message size (bytes)
	// Default: 16MB
	MaxMessageSize int64

	// EnableDebugLogging enables detailed logging for debugging
	// Should be false in production
	EnableDebugLogging bool

	// Logger is optional and caller-provided. When nil, auth/security helpers do
	// not emit logs.
	Logger *log.Logger

	// RejectOnQueueFull determines behavior when broadcast queue is full
	// true: reject message and log error
	// false: drop message silently (current behavior)
	RejectOnQueueFull bool

	// MaxConnectionRate limits new connections per second
	// 0 means no limit
	MaxConnectionRate int

	// EnableMetrics enables security metrics collection
	EnableMetrics bool
}

// SecurityMetrics tracks security-related metrics for a SecureRoomAuth instance.
type SecurityMetrics struct {
	// InvalidJWTSecrets counts JWT verifications that failed.
	InvalidJWTSecrets uint64 `json:"invalid_jwt_secrets"`
	// WeakRoomPasswords counts rejected weak passwords.
	WeakRoomPasswords uint64 `json:"weak_room_passwords"`
	// SuccessfulAuthentications counts successful JWT verifications.
	SuccessfulAuthentications uint64 `json:"successful_authentications"`
}

// ValidateSecurityConfig validates the security configuration
func ValidateSecurityConfig(cfg SecurityConfig) error {
	// Set default minimum if not specified
	if cfg.MinJWTSecretLength == 0 {
		cfg.MinJWTSecretLength = 32
	}

	if len(cfg.JWTSecret) < cfg.MinJWTSecretLength {
		return fmt.Errorf("%w: got %d bytes, minimum %d bytes required",
			ErrWeakJWTSecret, len(cfg.JWTSecret), cfg.MinJWTSecretLength)
	}

	// Check if secret is not using common weak patterns (warning only)
	secretStr := string(cfg.JWTSecret)
	if strings.Contains(secretStr, "secret") ||
		strings.Contains(secretStr, "password") ||
		strings.Contains(secretStr, "123456") {
		if cfg.Logger != nil {
			cfg.Logger.Printf("websocket: JWT secret contains common weak patterns")
		}
	}

	return nil
}

// ValidateWebSocketKey validates the Sec-WebSocket-Key header
func ValidateWebSocketKey(key string) error {
	if key == "" {
		return ErrInvalidWebSocketKey
	}

	// Key must be base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("%w: not valid base64", ErrInvalidWebSocketKey)
	}

	// Decoded key must be exactly 16 bytes
	if len(decoded) != 16 {
		return fmt.Errorf("%w: decoded length must be 16 bytes, got %d", ErrInvalidWebSocketKey, len(decoded))
	}

	return nil
}

// ValidateRoomPassword checks password strength
func ValidateRoomPassword(pwd string, config password.PasswordStrengthConfig, enforce bool) error {
	if pwd == "" {
		return NewValidationError("password", "cannot be empty")
	}

	// Check strength - use the password package's function
	isStrong := password.ValidatePasswordStrength(pwd, config)
	if !isStrong {
		if enforce {
			return ErrWeakRoomPassword
		}
	}

	return nil
}

// SecureRoomAuth extends SimpleRoomAuth with security validation and per-instance metrics.
type SecureRoomAuth struct {
	*SimpleRoomAuth
	securityConfig SecurityConfig

	// Per-instance metrics (lock-free atomics)
	invalidJWTSecrets         atomic.Uint64
	weakRoomPasswords         atomic.Uint64
	successfulAuthentications atomic.Uint64
}

// NewSecureRoomAuth creates a secure room auth with validation
func NewSecureRoomAuth(secret []byte, cfg SecurityConfig) (*SecureRoomAuth, error) {
	effectiveSecret := secret
	if len(cfg.JWTSecret) > 0 {
		if len(secret) > 0 && !bytes.Equal(secret, cfg.JWTSecret) {
			return nil, fmt.Errorf("%w: provided secret and config JWTSecret do not match", ErrInvalidConfig)
		}
		effectiveSecret = cfg.JWTSecret
	}
	cfg.JWTSecret = effectiveSecret

	// Validate config
	if err := ValidateSecurityConfig(cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.MinJWTSecretLength == 0 {
		cfg.MinJWTSecretLength = 32
	}
	if cfg.MaxMessageSize == 0 {
		cfg.MaxMessageSize = 16 << 20 // 16MB
	}
	if cfg.RoomPasswordConfig.MinLength == 0 {
		cfg.RoomPasswordConfig = password.DefaultPasswordStrengthConfig()
	}

	return &SecureRoomAuth{
		SimpleRoomAuth: NewSimpleRoomAuth(effectiveSecret),
		securityConfig: cfg,
	}, nil
}

// MaxMessageSize returns the configured max inbound message size for optional
// server-side read-limit enforcement.
func (s *SecureRoomAuth) MaxMessageSize() int64 {
	return s.securityConfig.MaxMessageSize
}

// SetRoomPassword overrides with security validation
func (s *SecureRoomAuth) SetRoomPassword(room, pwd string) error {
	if err := ValidateRoomPassword(pwd, s.securityConfig.RoomPasswordConfig, s.securityConfig.EnforcePasswordStrength); err != nil {
		s.weakRoomPasswords.Add(1)
		return err
	}
	isStrong := password.ValidatePasswordStrength(pwd, s.securityConfig.RoomPasswordConfig)
	if !isStrong && s.securityConfig.Logger != nil {
		s.securityConfig.Logger.Printf("websocket: weak room password accepted for room %q", room)
	}
	return s.SimpleRoomAuth.SetRoomPassword(room, pwd)
}

// VerifyJWT overrides with additional logging and per-instance metrics
func (s *SecureRoomAuth) VerifyJWT(token string) (map[string]any, error) {
	payload, err := s.SimpleRoomAuth.VerifyJWT(token)
	if err != nil {
		if s.securityConfig.EnableDebugLogging && s.securityConfig.Logger != nil {
			s.securityConfig.Logger.Printf("websocket: JWT verification failed: %v", err)
		}
		s.invalidJWTSecrets.Add(1)
		return nil, err
	}
	s.successfulAuthentications.Add(1)
	return payload, nil
}

// GetMetrics returns a snapshot of this instance's security metrics.
func (s *SecureRoomAuth) GetMetrics() SecurityMetrics {
	return SecurityMetrics{
		InvalidJWTSecrets:         s.invalidJWTSecrets.Load(),
		WeakRoomPasswords:         s.weakRoomPasswords.Load(),
		SuccessfulAuthentications: s.successfulAuthentications.Load(),
	}
}

// ResetMetrics resets all per-instance security metrics to zero.
func (s *SecureRoomAuth) ResetMetrics() {
	s.invalidJWTSecrets.Store(0)
	s.weakRoomPasswords.Store(0)
	s.successfulAuthentications.Store(0)
}

// GenerateSecureSecret generates a cryptographically secure random secret
func GenerateSecureSecret(length int) ([]byte, error) {
	if length < 32 {
		return nil, NewValidationError("secret", "length should be at least 32 bytes")
	}
	secret := make([]byte, length)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}
