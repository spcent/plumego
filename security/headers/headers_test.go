package headers

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	if got := w.Result().Header.Get("Content-Security-Policy"); got == "" {
		t.Fatalf("expected CSP to be set for strict policy")
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

func TestPolicyValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		policy := StrictPolicy()
		policy.Additional = map[string]string{"X-Trace-Mode": "sampled"}
		if err := policy.Validate(); err != nil {
			t.Fatalf("Validate: %v", err)
		}
	})

	t.Run("invalid values", func(t *testing.T) {
		policy := Policy{
			FrameOptions:              "DENY\nX-Injected: yes",
			ContentSecurityPolicy:     "default-src 'self'\r\nscript-src *",
			StrictTransportSecurity:   &HSTSOptions{MaxAge: -1 * time.Second},
			CrossOriginResourcePolicy: "same-origin",
			Additional: map[string]string{
				"Bad Header": "value",
				"X-Unsafe":   "bad\nvalue",
			},
		}

		err := policy.Validate()
		if !errors.Is(err, ErrInvalidPolicy) {
			t.Fatalf("Validate error = %v, want ErrInvalidPolicy", err)
		}
		message := err.Error()
		for _, want := range []string{
			"X-Frame-Options",
			"Content-Security-Policy",
			"Strict-Transport-Security",
			"Bad Header",
			"X-Unsafe",
		} {
			if !strings.Contains(message, want) {
				t.Fatalf("Validate error %q missing %q", message, want)
			}
		}
	})

	t.Run("invalid standard header semantics", func(t *testing.T) {
		policy := Policy{
			FrameOptions:              "ALLOWALL",
			ContentTypeOptions:        "sniff",
			ReferrerPolicy:            "send-everything",
			CrossOriginOpenerPolicy:   "isolated",
			CrossOriginResourcePolicy: "private",
			CrossOriginEmbedderPolicy: "required",
		}

		err := policy.Validate()
		if !errors.Is(err, ErrInvalidPolicy) {
			t.Fatalf("Validate error = %v, want ErrInvalidPolicy", err)
		}
		message := err.Error()
		for _, want := range []string{
			"X-Frame-Options",
			"X-Content-Type-Options",
			"Referrer-Policy",
			"Cross-Origin-Opener-Policy",
			"Cross-Origin-Resource-Policy",
			"Cross-Origin-Embedder-Policy",
		} {
			if !strings.Contains(message, want) {
				t.Fatalf("Validate error %q missing %q", message, want)
			}
		}
	})

	t.Run("valid standard header semantics are case insensitive", func(t *testing.T) {
		policy := Policy{
			FrameOptions:              "deny",
			ContentTypeOptions:        "NoSniff",
			ReferrerPolicy:            "Strict-Origin",
			CrossOriginOpenerPolicy:   "Same-Origin-Allow-Popups",
			CrossOriginResourcePolicy: "Cross-Origin",
			CrossOriginEmbedderPolicy: "Credentialless",
		}
		if err := policy.Validate(); err != nil {
			t.Fatalf("Validate: %v", err)
		}
	})
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
			name: "bare sandbox",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'").
					Sandbox().
					Build()
			},
			expected: "default-src 'self'; sandbox",
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
		{
			name: "drops directive injection values",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'self'", "'none'; script-src *").
					ScriptSrc("https://cdn.example.com", "bad\nvalue").
					ReportURI("/csp-report").
					Sandbox("allow-scripts", "allow-forms; allow-same-origin").
					Build()
			},
			expected: "default-src 'self'; script-src https://cdn.example.com; report-uri /csp-report; sandbox allow-scripts",
		},
		{
			name: "drops directive with only unsafe values",
			build: func() string {
				return NewCSPBuilder().
					DefaultSrc("'none'; script-src *").
					ReportURI("/bad; report-to x").
					UpgradeInsecureRequests().
					Build()
			},
			expected: "upgrade-insecure-requests",
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
	for _, part := range strings.Split(csp, ";") {
		part = strings.TrimSpace(part)
		if part == directive || strings.HasPrefix(part, directive+" ") {
			return true
		}
	}
	return false
}
