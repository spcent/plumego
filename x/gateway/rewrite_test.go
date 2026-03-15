package gateway

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

// --- StripPrefix ---

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		input  string
		want   string
	}{
		{"/api/v1", "/api/v1/users", "/users"},
		{"/api/v1", "/api/v1/", "/"},
		{"/api/v1", "/other/path", "/other/path"},
		{"/api/v1/", "/api/v1/users", "/users"}, // trailing slash in prefix
		{"/api", "/api", ""},
	}

	for _, tt := range tests {
		fn := StripPrefix(tt.prefix)
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("StripPrefix(%q)(%q) = %q, want %q", tt.prefix, tt.input, got, tt.want)
		}
	}
}

// --- AddPrefix ---

func TestAddPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		input  string
		want   string
	}{
		{"/internal", "/users", "/internal/users"},
		{"/internal", "users", "/internal/users"},
		{"/internal/", "/users", "/internal/users"}, // trailing slash trimmed
		{"", "/users", "/users"},
	}

	for _, tt := range tests {
		fn := AddPrefix(tt.prefix)
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("AddPrefix(%q)(%q) = %q, want %q", tt.prefix, tt.input, got, tt.want)
		}
	}
}

// --- ReplacePrefix ---

func TestReplacePrefix(t *testing.T) {
	tests := []struct {
		old, new string
		input    string
		want     string
	}{
		{"/api/v1", "/api/v2", "/api/v1/users", "/api/v2/users"},
		{"/api/v1", "/api/v2", "/other", "/other"},
		{"/api/v1/", "/api/v2/", "/api/v1/users", "/api/v2/users"},
	}

	for _, tt := range tests {
		fn := ReplacePrefix(tt.old, tt.new)
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("ReplacePrefix(%q,%q)(%q) = %q, want %q", tt.old, tt.new, tt.input, got, tt.want)
		}
	}
}

// --- RewriteMap ---

func TestRewriteMap(t *testing.T) {
	rules := map[string]string{
		"/old-api": "/new-api",
		"/legacy":  "/v2",
	}
	fn := RewriteMap(rules)

	tests := []struct {
		input string
		want  string
	}{
		{"/old-api/endpoint", "/new-api/endpoint"},
		{"/legacy/data", "/v2/data"},
		{"/unchanged/path", "/unchanged/path"},
	}

	for _, tt := range tests {
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("RewriteMap(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- RewriteRegex ---

func TestRewriteRegex(t *testing.T) {
	pattern := regexp.MustCompile(`^/api/v(\d+)`)
	fn := RewriteRegex(pattern, "/internal/v$1")

	tests := []struct {
		input string
		want  string
	}{
		{"/api/v1/users", "/internal/v1/users"},
		{"/api/v2/items", "/internal/v2/items"},
		{"/other/path", "/other/path"},
	}

	for _, tt := range tests {
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("RewriteRegex(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Chain ---

func TestChainRewriters(t *testing.T) {
	fn := Chain(
		StripPrefix("/api"),
		AddPrefix("/internal"),
	)

	tests := []struct {
		input string
		want  string
	}{
		{"/api/users", "/internal/users"},
		{"/api/items/123", "/internal/items/123"},
	}

	for _, tt := range tests {
		got := fn(tt.input)
		if got != tt.want {
			t.Errorf("Chain(StripPrefix,AddPrefix)(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestChainEmpty(t *testing.T) {
	fn := Chain()
	got := fn("/path")
	if got != "/path" {
		t.Errorf("Chain()(/path) = %q, want /path", got)
	}
}

// --- applyPathRewrite ---

func TestApplyPathRewriteNilRewriter(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/unchanged", nil)
	applyPathRewrite(r, nil) // should not panic
	if r.URL.Path != "/unchanged" {
		t.Errorf("path changed unexpectedly: %q", r.URL.Path)
	}
}

func TestApplyPathRewrite(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	applyPathRewrite(r, StripPrefix("/api"))
	if r.URL.Path != "/users" {
		t.Errorf("path = %q, want /users", r.URL.Path)
	}
}

func TestApplyPathRewriteRawPath(t *testing.T) {
	// httptest.NewRequest with a percent-encoded path sets RawPath.
	// applyPathRewrite copies the rewritten plain path into RawPath as well.
	r := httptest.NewRequest(http.MethodGet, "/api/users%2Fme", nil)
	// URL.Path = "/api/users/me", URL.RawPath = "/api/users%2Fme"
	applyPathRewrite(r, StripPrefix("/api"))
	// The implementation sets RawPath = newPath (the rewritten plain path)
	if r.URL.RawPath != "/users/me" {
		t.Errorf("RawPath = %q, want /users/me", r.URL.RawPath)
	}
	if r.URL.Path != "/users/me" {
		t.Errorf("Path = %q, want /users/me", r.URL.Path)
	}
}
