package input

import (
	"strings"
	"testing"
)

func TestIsToken(t *testing.T) {
	cases := []struct {
		value string
		ok    bool
	}{
		{"X-Test", true},
		{"content-type", true},
		{"abc123", true},
		{"", false},
		{"bad value", false},
		{"bad\tvalue", false},
		{"bad,value", false},
		{"bad:value", false},
	}

	for _, tc := range cases {
		if got := IsToken(tc.value); got != tc.ok {
			t.Fatalf("IsToken(%q) = %v, want %v", tc.value, got, tc.ok)
		}
	}
}

func TestIsHeaderValue(t *testing.T) {
	if !IsHeaderValue("nosniff") {
		t.Fatalf("expected safe header value")
	}

	if !IsHeaderValue("value\tcontinued") {
		t.Fatalf("expected horizontal tab to be allowed")
	}

	if IsHeaderValue("bad\nvalue") {
		t.Fatalf("expected newline to be rejected")
	}

	if IsHeaderValue("bad\rvalue") {
		t.Fatalf("expected carriage return to be rejected")
	}

	if IsHeaderValue(string([]byte{0xff})) {
		t.Fatalf("expected invalid utf-8 to be rejected")
	}

	if IsHeaderValue("bad\x1bvalue") {
		t.Fatalf("expected escape control to be rejected")
	}

	if IsHeaderValue("bad\x7fvalue") {
		t.Fatalf("expected DEL control to be rejected")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid basic", "user@example.com", true},
		{"valid with subdomain", "user@mail.example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with dots", "first.last@example.com", true},
		{"empty", "", false},
		{"no @", "userexample.com", false},
		{"multiple @", "user@@example.com", false},
		{"no domain", "user@", false},
		{"no local", "@example.com", false},
		{"no tld", "user@example", false},
		{"double dots", "user..name@example.com", false},
		{"too long", strings.Repeat("a", 256) + "@example.com", false},
		{"local too long", strings.Repeat("a", 65) + "@example.com", false},
		{"spaces", "user name@example.com", false},
		{"special chars", "user<>@example.com", false},
		{"valid with numbers", "user123@example456.com", true},
		{"valid with hyphen", "user-name@ex-ample.com", true},
		{"domain underscore", "user@bad_domain.com", false},
		{"domain unicode", "user@例子.com", false},
		{"domain label too long", "user@" + strings.Repeat("a", 64) + ".com", false},
		{"domain label bad char", "user@example!.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"valid http", "http://example.com", true},
		{"valid https", "https://example.com", true},
		{"valid with path", "https://example.com/path", true},
		{"valid with query", "https://example.com?key=value", true},
		{"empty", "", false},
		{"javascript scheme", "javascript:alert(1)", false},
		{"data scheme", "data:text/html,<script>alert(1)</script>", false},
		{"file scheme", "file:///etc/passwd", false},
		{"vbscript scheme", "vbscript:msgbox", false},
		{"too long", "https://" + strings.Repeat("a", 2100), false},
		{"null byte", "https://example.com\x00", false},
		{"relative path", "/path/to/resource", true},
		{"relative network path", "//example.com/path", false},
		{"relative path newline", "/path\nLocation: https://evil.example", false},
		{"relative path carriage return", "/path\rLocation: https://evil.example", false},
		{"relative path tab", "/path\tto", false},
		{"relative path backslash", `/path\to`, false},
		{"valid with port", "https://example.com:8080/path", true},
		{"userinfo credentials rejected", "https://user:pass@example.com", false},
		{"valid with fragment", "https://example.com#section", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateURL(tt.url)
			if got != tt.valid {
				t.Errorf("ValidateURL(%q) = %v, want %v", tt.url, got, tt.valid)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		valid bool
	}{
		{"valid US", "+1234567890", true},
		{"valid international", "+441234567890", true},
		{"valid with spaces", "+1 234 567 890", true},
		{"valid with dashes", "123-456-7890", true},
		{"valid with parens", "(123) 456-7890", true},
		{"valid with dots", "123.456.7890", true},
		{"empty", "", false},
		{"too short", "123", false},
		{"too long", "+12345678901234567", false},
		{"letters", "123-ABC-7890", false},
		{"special chars", "123@456-7890", false},
		{"just plus", "+", false},
		{"valid minimal", "1234567", true},
		{"valid E164", "+12345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePhone(tt.phone)
			if got != tt.valid {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tt.phone, got, tt.valid)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    string
		notContains string
	}{
		{
			name:        "remove script tag",
			input:       "Hello<script>alert(1)</script>World",
			contains:    "HelloWorld",
			notContains: "script",
		},
		{
			name:        "remove mixed case script tag",
			input:       `Hello<ScRiPt type="text/javascript">alert(1)</sCrIpT>World`,
			contains:    "HelloWorld",
			notContains: "ScRiPt",
		},
		{
			name:        "remove multiline script tag",
			input:       "Hello<script>\nalert(1)\n</script>World",
			contains:    "HelloWorld",
			notContains: "alert(1)",
		},
		{
			name:        "remove onclick",
			input:       `<div onclick="alert(1)">Click</div>`,
			contains:    "<div>Click</div>",
			notContains: "onclick",
		},
		{
			name:        "remove unquoted event handler",
			input:       `<button onmouseover=alert(1) class="safe">Click</button>`,
			contains:    `<button class="safe">Click</button>`,
			notContains: "onmouseover",
		},
		{
			name:        "remove mixed case event handler",
			input:       `<img OnLoad='alert(1)' src="/safe.png">`,
			contains:    `<img src="/safe.png">`,
			notContains: "OnLoad",
		},
		{
			name:        "remove javascript URL",
			input:       `<a href="javascript:alert(1)">Link</a>`,
			notContains: "javascript:",
		},
		{
			name:        "remove mixed case javascript URL",
			input:       `<a href="JaVaScRiPt:alert(1)">Link</a>`,
			notContains: "JaVaScRiPt:",
		},
		{
			name:        "remove data URL",
			input:       `<img src="data:image/png,base64...">`,
			notContains: "data:",
		},
		{
			name:        "remove mixed case data URL",
			input:       `<img src="DaTa:image/png,base64...">`,
			notContains: "DaTa:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTML(tt.input)
			if tt.contains != "" {
				if !strings.Contains(got, tt.contains) {
					t.Fatalf("expected %q to contain %q", got, tt.contains)
				}
			}
			if tt.notContains != "" {
				if strings.Contains(got, tt.notContains) {
					t.Fatalf("expected %q not to contain %q", got, tt.notContains)
				}
			}
		})
	}
}

func TestSanitizeSQL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		notContains string
	}{
		{"remove semicolon", "test; DROP TABLE", ";"},
		{"remove SQL comment", "test -- comment", "--"},
		{"remove block comment", "test /* comment */", "/*"},
		{"remove trailing SQL comment body", "test -- comment", "comment"},
		{"remove block comment body", "test /* comment */ value", "comment"},
		{"remove multiline block comment body", "test /* comment\nsecond line */ value", "second line"},
		{"remove UNION", "test UNION SELECT", "UNION"},
		{"remove SELECT", "test SELECT * FROM", "SELECT"},
		{"remove mixed case keywords", "test UnIoN SeLeCt value", "UnIoN"},
		{"remove mixed case select", "test UnIoN SeLeCt value", "SeLeCt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeSQL(tt.input)
			if strings.Contains(got, tt.notContains) {
				t.Fatalf("expected %q not to contain %q", got, tt.notContains)
			}
		})
	}
}

func TestStripControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"keep newline", "hello\nworld", "hello\nworld"},
		{"keep tab", "hello\tworld", "hello\tworld"},
		{"remove null", "hello\x00world", "helloworld"},
		{"remove bell", "hello\x07world", "helloworld"},
		{"remove ESC", "hello\x1bworld", "helloworld"},
		{"normal text", "hello world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripControlChars(tt.input)
			if got != tt.expected {
				t.Errorf("StripControlChars(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContainsDangerousChars(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		dangerous bool
	}{
		{"safe text", "hello world", false},
		{"null byte", "hello\x00world", true},
		{"carriage return", "hello\rworld", true},
		{"angle brackets", "hello<script>", true},
		{"quotes", `hello"world`, true},
		{"backtick", "hello`world", true},
		{"backslash", `hello\world`, true},
		{"curly braces", "hello{world}", true},
		{"normal punctuation", "hello, world!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsDangerousChars(tt.input)
			if got != tt.dangerous {
				t.Errorf("ContainsDangerousChars(%q) = %v, want %v", tt.input, got, tt.dangerous)
			}
		})
	}
}

func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"leading spaces", "  hello", "hello"},
		{"trailing spaces", "hello  ", "hello"},
		{"multiple spaces", "hello    world", "hello world"},
		{"mixed whitespace", "  hello  \t  world  ", "hello world"},
		{"newlines and spaces", "hello\n\nworld", "hello world"},
		{"already clean", "hello world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimWhitespace(tt.input)
			if got != tt.expected {
				t.Errorf("TrimWhitespace(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
