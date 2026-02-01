package headers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	policy.Apply(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("expected X-Frame-Options SAMEORIGIN, got %q", got)
	}
	if got := resp.Header.Get("Referrer-Policy"); got == "" {
		t.Fatalf("expected Referrer-Policy to be set")
	}
}

func TestStrictPolicyHSTS(t *testing.T) {
	policy := StrictPolicy()

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	policy.Apply(w, req)
	if got := w.Result().Header.Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("expected no HSTS on non-TLS request, got %q", got)
	}

	tlsReq := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	w = httptest.NewRecorder()
	policy.Apply(w, tlsReq)
	if got := w.Result().Header.Get("Strict-Transport-Security"); got == "" {
		t.Fatalf("expected HSTS on TLS request")
	}
}

func TestAdditionalHeadersValidation(t *testing.T) {
	policy := Policy{
		Additional: map[string]string{
			"X-Test":       "ok",
			"Bad Header":   "nope",
			"X-Injected":   "bad\nvalue",
			"X-Also-Valid": "still-ok",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	policy.Apply(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Test"); got != "ok" {
		t.Fatalf("expected X-Test to be set, got %q", got)
	}
	if got := resp.Header.Get("Bad Header"); got != "" {
		t.Fatalf("expected invalid header name to be skipped")
	}
	if got := resp.Header.Get("X-Injected"); got != "" {
		t.Fatalf("expected invalid header value to be skipped")
	}
	if got := resp.Header.Get("X-Also-Valid"); got != "still-ok" {
		t.Fatalf("expected X-Also-Valid to be set, got %q", got)
	}
}

func TestMiddleware(t *testing.T) {
	policy := DefaultPolicy()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	middleware := policy.Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options SAMEORIGIN, got %q", got)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestCSPBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() string
		expected string
	}{
		{
			name: "simple CSP",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'").
					Build()
			},
			expected: "default-src 'self'",
		},
		{
			name: "multiple directives",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'").
					ScriptSrc("'self'", "https://cdn.example.com").
					StyleSrc("'self'", "'unsafe-inline'").
					Build()
			},
			expected: "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'self' 'unsafe-inline'",
		},
		{
			name: "directive without sources",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'").
					UpgradeInsecureRequests().
					Build()
			},
			expected: "default-src 'self'; upgrade-insecure-requests",
		},
		{
			name: "frame ancestors",
			build: func() string {
				return NewCSPBuilder().
					FrameAncestors("'none'").
					Build()
			},
			expected: "frame-ancestors 'none'",
		},
		{
			name: "sandbox",
			build: func() string {
				return NewCSPBuilder().
					Sandbox("allow-scripts", "allow-forms").
					Build()
			},
			expected: "sandbox allow-scripts allow-forms",
		},
		{
			name: "report-uri",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'").
					ReportURI("/csp-report").
					Build()
			},
			expected: "default-src 'self'; report-uri /csp-report",
		},
		{
			name:     "empty builder",
			build:    func() string { return NewCSPBuilder().Build() },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.build()
			if got != tt.expected {
				t.Errorf("CSP mismatch\ngot:  %q\nwant: %q", got, tt.expected)
			}
		})
	}
}

func TestStrictCSP(t *testing.T) {
	csp := StrictCSP()

	if csp == "" {
		t.Fatal("StrictCSP() returned empty string")
	}

	// Check for key directives
	if !containsDirective(csp, "default-src 'self'") {
		t.Error("expected default-src 'self'")
	}
	if !containsDirective(csp, "frame-src 'none'") {
		t.Error("expected frame-src 'none'")
	}
	if !containsDirective(csp, "object-src 'none'") {
		t.Error("expected object-src 'none'")
	}
	if !containsDirective(csp, "upgrade-insecure-requests") {
		t.Error("expected upgrade-insecure-requests")
	}
}

func TestCSPBuilderAllDirectives(t *testing.T) {
	csp := NewCSPBuilder().
		DefaultSrc("'self'").
		ScriptSrc("'self'").
		StyleSrc("'self'").
		ImgSrc("'self'").
		FontSrc("'self'").
		ConnectSrc("'self'").
		FrameSrc("'self'").
		ObjectSrc("'self'").
		MediaSrc("'self'").
		ChildSrc("'self'").
		FormAction("'self'").
		FrameAncestors("'self'").
		BaseURI("'self'").
		ManifestSrc("'self'").
		WorkerSrc("'self'").
		Build()

	if csp == "" {
		t.Fatal("expected non-empty CSP")
	}

	directives := []string{
		"default-src", "script-src", "style-src", "img-src", "font-src",
		"connect-src", "frame-src", "object-src", "media-src", "child-src",
		"form-action", "frame-ancestors", "base-uri", "manifest-src", "worker-src",
	}

	for _, directive := range directives {
		if !containsDirective(csp, directive+" 'self'") {
			t.Errorf("expected %s 'self'", directive)
		}
	}
}

func TestPolicyWithCSP(t *testing.T) {
	csp := NewCSPBuilder().
		DefaultSrc("'self'").
		ScriptSrc("'self'", "https://cdn.example.com").
		Build()

	policy := Policy{
		ContentSecurityPolicy: csp,
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	policy.Apply(w, req)

	resp := w.Result()
	got := resp.Header.Get("Content-Security-Policy")
	if got != csp {
		t.Errorf("CSP mismatch\ngot:  %q\nwant: %q", got, csp)
	}
}

func containsDirective(csp, directive string) bool {
	// Simple check if directive exists in CSP string
	parts := splitCSP(csp)
	for _, part := range parts {
		if part == directive || hasPrefix(part, directive+" ") {
			return true
		}
	}
	return false
}

func splitCSP(csp string) []string {
	var parts []string
	for _, part := range splitString(csp, ";") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
