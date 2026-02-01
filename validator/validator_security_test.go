package validator

import (
	"strings"
	"testing"
)

func TestSecureEmail(t *testing.T) {
	rule := SecureEmail()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases
		{"valid basic", "user@example.com", false},
		{"valid subdomain", "user@mail.example.com", false},
		{"valid plus", "user+tag@example.com", false},
		{"valid dots", "first.last@example.com", false},
		{"valid numbers", "user123@example456.com", false},
		{"valid hyphen", "user-name@ex-ample.com", false},
		{"nil value", nil, false},
		{"empty string", "", false},

		// Invalid cases - format
		{"no @", "userexample.com", true},
		{"multiple @", "user@@example.com", true},
		{"no domain", "user@", true},
		{"no local", "@example.com", true},
		{"no tld", "user@example", true},
		{"spaces", "user name@example.com", true},
		{"special chars", "user<>@example.com", true},

		// Invalid cases - security
		{"double dots", "user..name@example.com", true},
		{"too long", strings.Repeat("a", 256) + "@example.com", true},
		{"local too long", strings.Repeat("a", 65) + "@example.com", true},
		{"domain too long", "user@" + strings.Repeat("a", 254) + ".com", true},

		// Type errors
		{"not a string", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureEmail().Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestSecureURL(t *testing.T) {
	rule := SecureURL()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/path", false},
		{"valid with query", "https://example.com?key=value", false},
		{"valid with port", "https://example.com:8080/path", false},
		{"valid with auth", "https://user:pass@example.com", false},
		{"valid with fragment", "https://example.com#section", false},
		{"relative path", "/path/to/resource", false},
		{"nil value", nil, false},
		{"empty string", "", false},

		// Invalid cases - dangerous schemes
		{"javascript scheme", "javascript:alert(1)", true},
		{"data scheme", "data:text/html,<script>alert(1)</script>", true},
		{"file scheme", "file:///etc/passwd", true},
		{"vbscript scheme", "vbscript:msgbox", true},

		// Invalid cases - security
		{"too long", "https://" + strings.Repeat("a", 2100), true},
		{"null byte", "https://example.com\x00", true},

		// Invalid cases - format
		{"invalid format", "ht!tp://example.com", true},

		// Type errors
		{"not a string", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureURL().Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestSecurePhone(t *testing.T) {
	rule := SecurePhone()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid cases
		{"valid US", "+1234567890", false},
		{"valid international", "+441234567890", false},
		{"valid with spaces", "+1 234 567 890", false},
		{"valid with dashes", "123-456-7890", false},
		{"valid with parens", "(123) 456-7890", false},
		{"valid with dots", "123.456.7890", false},
		{"valid minimal", "1234567", false},
		{"valid E164", "+12345678901", false},
		{"nil value", nil, false},
		{"empty string", "", false},

		// Invalid cases - format
		{"too short", "123", true},
		{"too long", "+12345678901234567", true},
		{"just plus", "+", true},

		// Invalid cases - security
		{"letters", "123-ABC-7890", true},
		{"special chars", "123@456-7890", true},
		{"injection attempt", "123';DROP TABLE--", true},

		// Type errors
		{"not a string", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecurePhone().Validate(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestSecureValidationInStruct(t *testing.T) {
	type UserInput struct {
		Email string `validate:"secureEmail"`
		URL   string `validate:"secureUrl"`
		Phone string `validate:"securePhone"`
	}

	tests := []struct {
		name    string
		input   UserInput
		wantErr bool
	}{
		{
			name: "all valid",
			input: UserInput{
				Email: "user@example.com",
				URL:   "https://example.com",
				Phone: "+1234567890",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: UserInput{
				Email: "user..name@example.com",
				URL:   "https://example.com",
				Phone: "+1234567890",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			input: UserInput{
				Email: "user@example.com",
				URL:   "javascript:alert(1)",
				Phone: "+1234567890",
			},
			wantErr: true,
		},
		{
			name: "invalid phone",
			input: UserInput{
				Email: "user@example.com",
				URL:   "https://example.com",
				Phone: "123-ABC-7890",
			},
			wantErr: true,
		},
	}

	validator := NewValidator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureValidationErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		value    any
		wantCode string
	}{
		{
			name:     "secureEmail error",
			rule:     SecureEmail(),
			value:    "invalid..email@example.com",
			wantCode: "secureEmail",
		},
		{
			name:     "secureUrl error",
			rule:     SecureURL(),
			value:    "javascript:alert(1)",
			wantCode: "secureUrl",
		},
		{
			name:     "securePhone error",
			rule:     SecurePhone(),
			value:    "123-ABC-7890",
			wantCode: "securePhone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate(tt.value)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != tt.wantCode {
				t.Errorf("error code = %v, want %v", err.Code, tt.wantCode)
			}

			if err.Message == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestSecureRulesRegistered(t *testing.T) {
	registry := NewRuleRegistry()

	rules := []string{"secureEmail", "secureUrl", "securePhone"}

	for _, ruleName := range rules {
		t.Run(ruleName, func(t *testing.T) {
			rule, exists := registry.Get(ruleName)
			if !exists {
				t.Errorf("rule %q not registered", ruleName)
			}
			if rule == nil {
				t.Errorf("rule %q is nil", ruleName)
			}
		})
	}
}
