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
