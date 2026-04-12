package authn

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticToken(t *testing.T) {
	tests := []struct {
		name          string
		configured    string
		authorization string
		xToken        string
		wantErr       error
	}{
		{
			name:       "valid x-token header",
			configured: "secret123",
			xToken:     "secret123",
		},
		{
			name:          "valid bearer token",
			configured:    "secret123",
			authorization: "Bearer secret123",
		},
		{
			name:    "missing configured token fails closed",
			wantErr: ErrUnauthenticated,
		},
		{
			name:       "missing credentials",
			configured: "secret123",
			wantErr:    ErrUnauthenticated,
		},
		{
			name:          "malformed authorization header",
			configured:    "secret123",
			authorization: "Bearer",
			wantErr:       ErrUnauthenticated,
		},
		{
			name:       "query token ignored",
			configured: "secret123",
			wantErr:    ErrUnauthenticated,
		},
		{
			name:       "invalid token",
			configured: "secret123",
			xToken:     "wrong-token",
			wantErr:    ErrUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?token=secret123", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			if tt.xToken != "" {
				req.Header.Set("X-Token", tt.xToken)
			}

			principal, err := StaticToken(tt.configured).Authenticate(req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil {
				if principal != nil {
					t.Fatalf("expected nil principal on auth failure, got %#v", principal)
				}
				return
			}
			if principal == nil || principal.Subject != staticTokenSubject {
				t.Fatalf("expected static-token principal, got %#v", principal)
			}
		})
	}
}
