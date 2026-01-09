package websocket

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	InvalidJWTSecrets         uint64 `json:"invalid_jwt_secrets"`
	WeakRoomPasswords         uint64 `json:"weak_room_passwords"`
	InvalidWebSocketKeys      uint64 `json:"invalid_websocket_keys"`
	BroadcastQueueFull        uint64 `json:"broadcast_queue_full"`
	RejectedConnections       uint64 `json:"rejected_connections"`
	SuccessfulAuthentications uint64 `json:"successful_authentications"`
}

var (
	// ErrWeakJWTSecret is returned when JWT secret is too short
	ErrWeakJWTSecret = errors.New("jwt secret too weak: minimum 32 bytes required")

	// ErrWeakRoomPassword is returned when room password doesn't meet strength requirements
	ErrWeakRoomPassword = errors.New("room password does not meet strength requirements")

	// ErrInvalidWebSocketKey is returned when Sec-WebSocket-Key is malformed
	ErrInvalidWebSocketKey = errors.New("invalid websocket key format")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid security configuration")
)

// global security metrics (thread-safe)
var securityMetrics SecurityMetrics
var metricsMutex sync.RWMutex

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
		log.Printf("WARNING: JWT secret contains common weak patterns")
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
		return errors.New("password cannot be empty")
	}

	// Check strength - use the password package's function
	isStrong := password.ValidatePasswordStrength(pwd, config)
	if !isStrong {
		if enforce {
			return ErrWeakRoomPassword
		}
		log.Printf("WARNING: Weak room password used (enforcement disabled)")
	}

	return nil
}

// SecureRoomAuth extends simpleRoomAuth with security features
type SecureRoomAuth struct {
	*simpleRoomAuth
	securityConfig SecurityConfig
}

// NewSecureRoomAuth creates a secure room auth with validation
func NewSecureRoomAuth(secret []byte, cfg SecurityConfig) (*SecureRoomAuth, error) {
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
		simpleRoomAuth: NewSimpleRoomAuth(secret),
		securityConfig: cfg,
	}, nil
}

// SetRoomPassword overrides with security validation
func (s *SecureRoomAuth) SetRoomPassword(room, pwd string) error {
	// Validate password strength
	if err := ValidateRoomPassword(pwd, s.securityConfig.RoomPasswordConfig, s.securityConfig.EnforcePasswordStrength); err != nil {
		metricsMutex.Lock()
		securityMetrics.WeakRoomPasswords++
		metricsMutex.Unlock()
		return err
	}

	s.simpleRoomAuth.SetRoomPassword(room, pwd)
	return nil
}

// VerifyJWT overrides with additional logging
func (s *SecureRoomAuth) VerifyJWT(token string) (map[string]any, error) {
	payload, err := s.simpleRoomAuth.VerifyJWT(token)
	if err != nil {
		if s.securityConfig.EnableDebugLogging {
			log.Printf("JWT verification failed: %v", err)
		}
		metricsMutex.Lock()
		securityMetrics.InvalidJWTSecrets++
		metricsMutex.Unlock()
		return nil, err
	}

	metricsMutex.Lock()
	securityMetrics.SuccessfulAuthentications++
	metricsMutex.Unlock()

	return payload, nil
}

// GetSecurityMetrics returns current security metrics
func GetSecurityMetrics() SecurityMetrics {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()
	return securityMetrics
}

// ResetSecurityMetrics resets all security metrics
func ResetSecurityMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	securityMetrics = SecurityMetrics{}
}

// GenerateSecureSecret generates a cryptographically secure random secret
func GenerateSecureSecret(length int) ([]byte, error) {
	if length < 32 {
		return nil, errors.New("secret length should be at least 32 bytes")
	}
	secret := make([]byte, length)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// LogSecurityEvent logs security events based on configuration
func LogSecurityEvent(event string, details map[string]any, cfg SecurityConfig) {
	if !cfg.EnableDebugLogging {
		return
	}

	// In production, this should use structured logging
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(os.Stderr, "[SECURITY] %s | Event: %s | Details: %v\n", timestamp, event, details)
}
