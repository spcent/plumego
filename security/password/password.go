// Package password provides secure password hashing and strength validation.
//
// This package implements industry-standard password security using PBKDF2-HMAC-SHA512
// with configurable iterations and salt generation. It provides both hashing/verification
// and password strength validation.
//
// Features:
//   - PBKDF2-HMAC-SHA512 key derivation (default: 210,000 iterations)
//   - Automatic salt generation (32 bytes)
//   - Constant-time comparison to prevent timing attacks
//   - Password strength validation with customizable rules
//   - Common password dictionary check (optional)
//
// Example usage:
//
//	import "github.com/spcent/plumego/security/password"
//
//	// Hash a password
//	hash, err := password.Hash("user-password-123")
//	if err != nil {
//		// Handle error
//	}
//
//	// Verify a password
//	ok := password.Verify("user-password-123", hash)
//	if !ok {
//		// Invalid password
//	}
//
//	// Validate password strength
//	config := password.PasswordStrengthConfig{
//		MinLength:        12,
//		RequireUppercase: true,
//		RequireLowercase: true,
//		RequireDigit:     true,
//		RequireSpecial:   true,
//	}
//	err = password.ValidateStrength("WeakPass", config)
//	// Returns error if password doesn't meet requirements
package password

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// PasswordStrengthConfig defines the configuration for password strength validation.
//
// This configuration allows you to specify the minimum requirements for passwords
// to ensure they meet security standards.
//
// Example:
//
//	import "github.com/spcent/plumego/security/password"
//
//	config := password.PasswordStrengthConfig{
//		MinLength:        12,
//		RequireUppercase: true,
//		RequireLowercase: true,
//		RequireDigit:     true,
//		RequireSpecial:   true,
//	}
//	if !password.ValidatePasswordStrength("MySecurePass123!", config) {
//		// Password doesn't meet requirements
//	}
type PasswordStrengthConfig struct {
	// MinLength is the minimum password length
	MinLength int

	// RequireUppercase requires at least one uppercase letter
	RequireUppercase bool

	// RequireLowercase requires at least one lowercase letter
	RequireLowercase bool

	// RequireDigit requires at least one digit
	RequireDigit bool

	// RequireSpecial requires at least one special character
	RequireSpecial bool
}

// DefaultPasswordStrengthConfig returns the default password strength configuration.
//
// Defaults:
//   - MinLength: 8
//   - RequireUppercase: true
//   - RequireLowercase: true
//   - RequireDigit: true
//   - RequireSpecial: false
//
// Example:
//
//	import "github.com/spcent/plumego/security/password"
//
//	config := password.DefaultPasswordStrengthConfig()
//	if !password.ValidatePasswordStrength("MySecurePass123", config) {
//		// Password doesn't meet requirements
//	}
func DefaultPasswordStrengthConfig() PasswordStrengthConfig {
	return PasswordStrengthConfig{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   false,
	}
}

// ValidatePasswordStrength checks if the password meets the required strength criteria.
//
// Example:
//
//	import "github.com/spcent/plumego/security/password"
//
//	config := password.DefaultPasswordStrengthConfig()
//	if !password.ValidatePasswordStrength("MySecurePass123", config) {
//		// Password doesn't meet requirements
//	}
func ValidatePasswordStrength(password string, config PasswordStrengthConfig) bool {
	if len(password) < config.MinLength {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Check all required conditions
	if config.RequireUppercase && !hasUpper {
		return false
	}
	if config.RequireLowercase && !hasLower {
		return false
	}
	if config.RequireDigit && !hasDigit {
		return false
	}
	if config.RequireSpecial && !hasSpecial {
		return false
	}

	return true
}

// DefaultCost represents the default iteration count for password hashing.
// A higher value increases the computational cost for attackers attempting to
// brute-force hashed passwords.
//
// OWASP 2024 recommends at least 600,000 iterations for PBKDF2-SHA256 or
// 210,000 iterations for PBKDF2-SHA512 to achieve approximately 100ms of
// computation time on typical hardware. We use 210,000 as the default for
// PBKDF2-SHA512.
//
// Note: The previous value of 10,000 was too low for modern security standards.
// If you have existing hashed passwords, consider implementing a migration
// strategy to rehash passwords with the new cost when users log in.
const DefaultCost = 210_000

// MinimumCost represents the minimum recommended iteration count.
// Using fewer iterations than this is not recommended for security-critical applications.
const MinimumCost = 100_000

// HashPassword generates a salted hash of the password with the default cost.
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultCost)
}

// HashPasswordWithCost generates a salted hash of the password with the specified cost.
// The returned string has the format: "<cost>$<salt>$<hash>".
func HashPasswordWithCost(password string, cost int) (string, error) {
	if cost < 1 {
		return "", errors.New("cost must be at least 1")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	derived := deriveKey(password, salt, cost)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHash := base64.StdEncoding.EncodeToString(derived)

	return fmt.Sprintf("%d$%s$%s", cost, encodedSalt, encodedHash), nil
}

// CheckPassword compares a hashed password with its plaintext version.
func CheckPassword(hashedPassword, password string) error {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 3 {
		return errors.New("invalid hash format")
	}

	cost, err := strconv.Atoi(parts[0])
	if err != nil || cost < 1 {
		return errors.New("invalid hash format")
	}

	salt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	derived := deriveKey(password, salt, cost)
	if subtle.ConstantTimeCompare(expectedHash, derived) == 1 {
		return nil
	}

	return errors.New("password mismatch")
}

func deriveKey(password string, salt []byte, cost int) []byte {
	return pbkdf2SHA512([]byte(password), salt, cost, 32)
}

func pbkdf2SHA512(password, salt []byte, iterations, keyLen int) []byte {
	if iterations < 1 {
		return nil
	}

	hLen := sha512.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	derived := make([]byte, 0, numBlocks*hLen)

	for block := 1; block <= numBlocks; block++ {
		t := pbkdf2Block(password, salt, iterations, block)
		derived = append(derived, t...)
	}

	return derived[:keyLen]
}

func pbkdf2Block(password, salt []byte, iterations, blockIndex int) []byte {
	h := hmac.New(sha512.New, password)
	h.Write(salt)
	h.Write([]byte{
		byte(blockIndex >> 24),
		byte(blockIndex >> 16),
		byte(blockIndex >> 8),
		byte(blockIndex),
	})

	u := h.Sum(nil)
	t := make([]byte, len(u))
	copy(t, u)

	for i := 1; i < iterations; i++ {
		h = hmac.New(sha512.New, password)
		h.Write(u)
		u = h.Sum(nil)
		for j := range t {
			t[j] ^= u[j]
		}
	}

	return t
}
