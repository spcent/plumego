package jwt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestJWTAuthenticatorNegativeMatrix(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	accessPair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "neg-user"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("failed to generate access pair: %v", err)
	}
	refreshPair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "neg-user"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("failed to generate refresh pair: %v", err)
	}

	handlerCalled := false
	protected := mgr.JWTAuthenticator(TokenTypeAccess)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name            string
		authzHeader     string
		expectedCode    string
		expectedMessage string
	}{
		{
			name:            "missing authorization header",
			authzHeader:     "",
			expectedCode:    "missing_token",
			expectedMessage: "missing authorization header",
		},
		{
			name:            "empty bearer token",
			authzHeader:     "Bearer   ",
			expectedCode:    "missing_token",
			expectedMessage: "missing authorization header",
		},
		{
			name:            "malformed token",
			authzHeader:     "Bearer invalid.token",
			expectedCode:    "invalid_token",
			expectedMessage: "invalid token",
		},
		{
			name:            "tampered signature",
			authzHeader:     "Bearer " + tamperJWT(accessPair.AccessToken),
			expectedCode:    "invalid_token",
			expectedMessage: "invalid token",
		},
		{
			name:            "wrong token type",
			authzHeader:     "Bearer " + refreshPair.RefreshToken,
			expectedCode:    "invalid_token",
			expectedMessage: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled = false

			req := httptest.NewRequest(http.MethodGet, "/secure", nil)
			if tt.authzHeader != "" {
				req.Header.Set("Authorization", tt.authzHeader)
			}

			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
			if handlerCalled {
				t.Fatalf("protected handler must not run for negative auth cases")
			}

			var payload contract.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("failed to decode error payload: %v", err)
			}

			if payload.Error.Code != tt.expectedCode {
				t.Fatalf("error code = %q, want %q", payload.Error.Code, tt.expectedCode)
			}
			if payload.Error.Message != tt.expectedMessage {
				t.Fatalf("error message = %q, want %q", payload.Error.Message, tt.expectedMessage)
			}
		})
	}
}

func tamperJWT(token string) string {
	if token == "" {
		return token
	}
	last := token[len(token)-1]
	replacement := "A"
	if last == 'A' {
		replacement = "B"
	}
	return token[:len(token)-1] + replacement
}

func TestJWTAuthenticatorRejectsNonBearerSchemes(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	protected := mgr.JWTAuthenticator(TokenTypeAccess)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing authorization header") {
		t.Fatalf("expected missing authorization header response, got %q", rec.Body.String())
	}
}
