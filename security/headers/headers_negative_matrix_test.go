package headers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdditionalHeadersNegativeMatrix(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		shouldBeSet bool
	}{
		{
			name:        "invalid name with space",
			key:         "Bad Header",
			value:       "x",
			shouldBeSet: false,
		},
		{
			name:        "empty header name",
			key:         "",
			value:       "x",
			shouldBeSet: false,
		},
		{
			name:        "invalid value newline",
			key:         "X-Bad-Newline",
			value:       "bad\nvalue",
			shouldBeSet: false,
		},
		{
			name:        "invalid value carriage return",
			key:         "X-Bad-CR",
			value:       "bad\rvalue",
			shouldBeSet: false,
		},
		{
			name:        "invalid value escape control",
			key:         "X-Bad-ESC",
			value:       "bad\x1bvalue",
			shouldBeSet: false,
		},
		{
			name:        "invalid value utf8",
			key:         "X-Bad-UTF8",
			value:       string([]byte{0xff}),
			shouldBeSet: false,
		},
		{
			name:        "valid custom header",
			key:         "X-Trace-Mode",
			value:       "strict",
			shouldBeSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := Policy{
				Additional: map[string]string{
					tt.key: tt.value,
				},
			}

			req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
			req.TLS = &tls.ConnectionState{}
			rec := httptest.NewRecorder()
			policy.Apply(rec, req)

			got := rec.Header().Get(tt.key)
			if tt.shouldBeSet && got != tt.value {
				t.Fatalf("expected header %q=%q, got %q", tt.key, tt.value, got)
			}
			if !tt.shouldBeSet && got != "" {
				t.Fatalf("expected header %q to be dropped, got %q", tt.key, got)
			}
		})
	}
}

func TestHTTPSDetectionProxyHeadersNegativeMatrix(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    bool
	}{
		{
			name: "all forwarded proto values https",
			headers: map[string]string{
				"X-Forwarded-Proto": "https, https",
			},
			want: true,
		},
		{
			name: "mixed forwarded proto chain",
			headers: map[string]string{
				"X-Forwarded-Proto": "https, http",
			},
			want: false,
		},
		{
			name: "malformed forwarded proto chain",
			headers: map[string]string{
				"X-Forwarded-Proto": "https,",
			},
			want: false,
		},
		{
			name: "forwarded exact proto https",
			headers: map[string]string{
				"Forwarded": `for=192.0.2.1; proto=https; host=example.com`,
			},
			want: true,
		},
		{
			name: "forwarded quoted proto https",
			headers: map[string]string{
				"Forwarded": `for=192.0.2.1; proto="https"`,
			},
			want: true,
		},
		{
			name: "forwarded proto substring is not enough",
			headers: map[string]string{
				"Forwarded": `for=192.0.2.1; proto=httpsx`,
			},
			want: false,
		},
		{
			name: "forwarded mixed chain",
			headers: map[string]string{
				"Forwarded": `for=192.0.2.1; proto=https, for=192.0.2.2; proto=http`,
			},
			want: false,
		},
		{
			name: "forwarded missing proto",
			headers: map[string]string{
				"Forwarded": `for=192.0.2.1; host=example.com`,
			},
			want: false,
		},
		{
			name: "forwarded ssl on",
			headers: map[string]string{
				"X-Forwarded-Ssl": "on",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}
			if got := isHTTPSRequest(req); got != tt.want {
				t.Fatalf("isHTTPSRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
