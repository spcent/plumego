package websocket

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateTextMessage(t *testing.T) {
	cfg := DefaultMessageValidationConfig()

	tests := []struct {
		name    string
		data    []byte
		cfg     MessageValidationConfig
		wantErr bool
	}{
		{
			name:    "Valid message",
			data:    []byte("Hello, World!"),
			cfg:     cfg,
			wantErr: false,
		},
		{
			name:    "Valid with newlines",
			data:    []byte("Hello\nWorld\n"),
			cfg:     cfg,
			wantErr: false,
		},
		{
			name:    "Valid with tabs",
			data:    []byte("Hello\tWorld"),
			cfg:     cfg,
			wantErr: false,
		},
		{
			name:    "Null byte rejected",
			data:    []byte("Hello\x00World"),
			cfg:     cfg,
			wantErr: true,
		},
		{
			name:    "Bell character rejected",
			data:    []byte("Hello\x07World"),
			cfg:     cfg,
			wantErr: true,
		},
		{
			name:    "Escape sequence rejected",
			data:    []byte("Hello\x1b[31mWorld"),
			cfg:     cfg,
			wantErr: true,
		},
		{
			name:    "DEL character rejected",
			data:    []byte("Hello\x7FWorld"),
			cfg:     cfg,
			wantErr: true,
		},
		{
			name:    "Carriage return rejected when configured",
			data:    []byte("Hello\rWorld"),
			cfg:     MessageValidationConfig{RejectControlCharacters: true, AllowedNewlines: false},
			wantErr: true,
		},
		{
			name:    "Empty message rejected when not allowed",
			data:    []byte(""),
			cfg:     MessageValidationConfig{AllowEmpty: false},
			wantErr: true,
		},
		{
			name:    "Empty message allowed when configured",
			data:    []byte(""),
			cfg:     MessageValidationConfig{AllowEmpty: true},
			wantErr: false,
		},
		{
			name:    "Message too long",
			data:    []byte(strings.Repeat("a", 1001)),
			cfg:     MessageValidationConfig{MaxLength: 1000},
			wantErr: true,
		},
		{
			name:    "Invalid UTF-8 rejected",
			data:    []byte{0xFF, 0xFE, 0xFD},
			cfg:     MessageValidationConfig{RequireValidUTF8: true},
			wantErr: true,
		},
		{
			name:    "Valid UTF-8 accepted",
			data:    []byte("Hello 世界 🌍"),
			cfg:     cfg,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTextMessage(tc.data, tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateTextMessage() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && err != nil {
				// Verify we get specific error types
				if !errors.Is(err, ErrInvalidUTF8) &&
					!errors.Is(err, ErrControlCharacters) &&
					!errors.Is(err, ErrMessageTooLong) &&
					!errors.Is(err, ErrEmptyMessage) {
					t.Fatalf("expected validation error, got %v", err)
				}
			}
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		maxLen  int
		want    string
		wantLen int
	}{
		{
			name:    "Normal text",
			data:    []byte("Hello, World!"),
			maxLen:  100,
			want:    "Hello, World!",
			wantLen: 13,
		},
		{
			name:    "With control characters",
			data:    []byte("Hello\x00\x07World"),
			maxLen:  100,
			want:    "Hello  World",
			wantLen: 12,
		},
		{
			name:    "Truncated message",
			data:    []byte(strings.Repeat("a", 100)),
			maxLen:  10,
			want:    strings.Repeat("a", 10) + "...",
			wantLen: 13,
		},
		{
			name:    "With ANSI escape",
			data:    []byte("Hello\x1b[31mWorld"),
			maxLen:  100,
			want:    "Hello [31mWorld",
			wantLen: 15,
		},
		{
			name:    "Invalid UTF-8 replaced",
			data:    []byte{0x48, 0x65, 0xFF, 0xFE, 0x6C, 0x6C, 0x6F}, // "He" + invalid + "llo"
			maxLen:  100,
			want:    "He�llo",
			wantLen: 8,
		},
		{
			name:    "Preserves newlines",
			data:    []byte("Hello\nWorld"),
			maxLen:  100,
			want:    "Hello\nWorld",
			wantLen: 11,
		},
		{
			name:    "Preserves tabs",
			data:    []byte("Hello\tWorld"),
			maxLen:  100,
			want:    "Hello\tWorld",
			wantLen: 11,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeForLogging(tc.data, tc.maxLen)
			if got != tc.want {
				t.Fatalf("SanitizeForLogging() = %q, want %q", got, tc.want)
			}
			if len(got) != tc.wantLen {
				t.Fatalf("SanitizeForLogging() length = %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

func TestContainsDangerousPatterns(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		dangerous bool
	}{
		{
			name:      "Safe text",
			data:      []byte("Hello, World!"),
			dangerous: false,
		},
		{
			name:      "ANSI escape sequence",
			data:      []byte("\x1b[31mRed Text"),
			dangerous: true,
		},
		{
			name:      "Script tag",
			data:      []byte("<script>alert('xss')</script>"),
			dangerous: true,
		},
		{
			name:      "Script tag uppercase",
			data:      []byte("<SCRIPT>alert('xss')</SCRIPT>"),
			dangerous: true,
		},
		{
			name:      "Iframe tag",
			data:      []byte("<iframe src='evil.com'>"),
			dangerous: true,
		},
		{
			name:      "JavaScript protocol",
			data:      []byte("javascript:alert('xss')"),
			dangerous: true,
		},
		{
			name:      "Onerror handler",
			data:      []byte("<img onerror='alert(1)'>"),
			dangerous: true,
		},
		{
			name:      "Onload handler",
			data:      []byte("<body onload='alert(1)'>"),
			dangerous: true,
		},
		{
			name:      "SQL UNION attack",
			data:      []byte("' UNION SELECT password FROM users--"),
			dangerous: true,
		},
		{
			name:      "SQL DROP attack",
			data:      []byte("'; DROP TABLE users; --"),
			dangerous: true,
		},
		{
			name:      "SQL OR 1=1 attack",
			data:      []byte("admin' OR '1'='1"),
			dangerous: true,
		},
		{
			name:      "SQL comment injection",
			data:      []byte("'; --"),
			dangerous: true,
		},
		{
			name:      "Safe HTML mention",
			data:      []byte("I like HTML and scripts"),
			dangerous: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsDangerousPatterns(tc.data)
			if got != tc.dangerous {
				t.Fatalf("ContainsDangerousPatterns() = %v, want %v", got, tc.dangerous)
			}
		})
	}
}

func TestDefaultMessageValidationConfig(t *testing.T) {
	cfg := DefaultMessageValidationConfig()

	if cfg.MaxLength <= 0 {
		t.Fatal("default should have max length")
	}
	if cfg.AllowEmpty {
		t.Fatal("default should not allow empty")
	}
	if !cfg.RejectControlCharacters {
		t.Fatal("default should reject control characters")
	}
	if !cfg.RequireValidUTF8 {
		t.Fatal("default should require valid UTF-8")
	}
	if !cfg.AllowedNewlines {
		t.Fatal("default should allow newlines")
	}
	if !cfg.AllowedTabs {
		t.Fatal("default should allow tabs")
	}
}

func TestValidateTextMessage_AllControlCharacters(t *testing.T) {
	cfg := MessageValidationConfig{
		RejectControlCharacters: true,
		AllowedNewlines:         false,
		AllowedTabs:             false,
	}

	// Test all ASCII control characters (0x00-0x1F and 0x7F)
	for i := 0; i < 32; i++ {
		data := []byte{byte(i)}
		err := ValidateTextMessage(data, cfg)
		if err == nil {
			t.Fatalf("expected error for control character 0x%02X", i)
		}
		if !errors.Is(err, ErrControlCharacters) {
			t.Fatalf("expected ErrControlCharacters for 0x%02X, got %v", i, err)
		}
	}

	// Test DEL (0x7F)
	data := []byte{0x7F}
	err := ValidateTextMessage(data, cfg)
	if err == nil {
		t.Fatal("expected error for DEL character")
	}
	if !errors.Is(err, ErrControlCharacters) {
		t.Fatalf("expected ErrControlCharacters for DEL, got %v", err)
	}
}
