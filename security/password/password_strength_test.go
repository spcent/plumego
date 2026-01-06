package password

import (
	"strings"
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		config   PasswordStrengthConfig
		expected bool
	}{
		{
			name:     "valid password with all requirements",
			password: "StrongPass1!",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: true,
		},
		{
			name:     "password too short",
			password: "Short1!",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "missing uppercase",
			password: "lowercase1!",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "missing lowercase",
			password: "UPPERCASE1!",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "missing digit",
			password: "StrongPass!",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "missing special (but allowed)",
			password: "StrongPass1",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   false,
			},
			expected: true,
		},
		{
			name:     "missing special (required)",
			password: "StrongPass1",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: false,
		},
		{
			name:     "no requirements",
			password: "any",
			config: PasswordStrengthConfig{
				MinLength:        0,
				RequireUppercase: false,
				RequireLowercase: false,
				RequireDigit:     false,
				RequireSpecial:   false,
			},
			expected: true,
		},
		{
			name:     "unicode characters",
			password: "StrongPass1!中文",
			config: PasswordStrengthConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePasswordStrength(tt.password, tt.config)
			if result != tt.expected {
				t.Errorf("ValidatePasswordStrength(%q, %+v) = %v, expected %v",
					tt.password, tt.config, result, tt.expected)
			}
		})
	}
}

func TestDefaultPasswordStrengthConfig(t *testing.T) {
	config := DefaultPasswordStrengthConfig()

	if config.MinLength != 8 {
		t.Errorf("MinLength = %d, expected 8", config.MinLength)
	}
	if !config.RequireUppercase {
		t.Error("RequireUppercase should be true")
	}
	if !config.RequireLowercase {
		t.Error("RequireLowercase should be true")
	}
	if !config.RequireDigit {
		t.Error("RequireDigit should be true")
	}
	if config.RequireSpecial {
		t.Error("RequireSpecial should be false")
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	// Test with invalid cost
	_, err := HashPasswordWithCost("password", 0)
	if err == nil {
		t.Error("Expected error for cost < 1")
	}

	// Test with negative cost
	_, err = HashPasswordWithCost("password", -1)
	if err == nil {
		t.Error("Expected error for negative cost")
	}

	// Test with valid cost
	hash, err := HashPasswordWithCost("password", 5)
	if err != nil {
		t.Errorf("HashPasswordWithCost failed: %v", err)
	}
	if hash == "" {
		t.Error("HashPasswordWithCost returned empty hash")
	}

	// Verify the hash format
	parts := strings.Split(hash, "$")
	if len(parts) != 3 {
		t.Errorf("Invalid hash format: %s", hash)
	}
}

func TestCheckPasswordInvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{
			name: "missing parts",
			hash: "10$salt",
		},
		{
			name: "invalid cost",
			hash: "abc$salt$hash",
		},
		{
			name: "invalid salt base64",
			hash: "10$invalid!!!$hash",
		},
		{
			name: "invalid hash base64",
			hash: "10$salt$invalid!!!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hash, "password")
			if err == nil {
				t.Errorf("Expected error for invalid hash format: %s", tt.hash)
			}
		})
	}
}

func TestPBKDF2Implementation(t *testing.T) {
	// Test pbkdf2SHA512 with various parameters
	password := []byte("test")
	salt := []byte("salt")

	// Test with different iterations
	result1 := pbkdf2SHA512(password, salt, 1, 32)
	result2 := pbkdf2SHA512(password, salt, 100, 32)

	if len(result1) != 32 {
		t.Errorf("pbkdf2SHA512 returned length %d, expected 32", len(result1))
	}
	if len(result2) != 32 {
		t.Errorf("pbkdf2SHA512 returned length %d, expected 32", len(result2))
	}

	// Results should be different with different iterations
	if string(result1) == string(result2) {
		t.Error("pbkdf2SHA512 with different iterations should produce different results")
	}

	// Test with invalid iterations
	result3 := pbkdf2SHA512(password, salt, 0, 32)
	if result3 != nil {
		t.Error("pbkdf2SHA512 with iterations=0 should return nil")
	}

	// Test with different key lengths
	result4 := pbkdf2SHA512(password, salt, 10, 16)
	if len(result4) != 16 {
		t.Errorf("pbkdf2SHA512 with keyLen=16 returned length %d", len(result4))
	}
}

func TestPBKDF2Block(t *testing.T) {
	password := []byte("test")
	salt := []byte("salt")

	// Test block with different indices
	block1 := pbkdf2Block(password, salt, 10, 1)
	block2 := pbkdf2Block(password, salt, 10, 2)

	// Verify blocks are non-nil and have expected length
	if block1 == nil || len(block1) == 0 {
		t.Error("pbkdf2Block should return non-empty result")
	}
	if block2 == nil || len(block2) == 0 {
		t.Error("pbkdf2Block should return non-empty result")
	}

	// Blocks with different indices should be different
	if string(block1) == string(block2) {
		t.Error("pbkdf2Block with different indices should produce different results")
	}
}
