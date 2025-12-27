package jwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/middleware"
	kvstore "github.com/spcent/plumego/store/kv"
)

func newTestStore(t *testing.T) *kvstore.KVStore {
	t.Helper()
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("failed to create kv store: %v", err)
	}
	return store
}

// ========== basic function test ==========

func TestGenerateAndVerifyTokenPair(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.AccessExpiration = time.Minute
	cfg.RefreshExpiration = time.Hour

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity := IdentityClaims{Subject: "user-1"}
	authz := AuthorizationClaims{
		Roles:       []string{"user"},
		Permissions: []string{"read"},
	}

	pair, err := mgr.GenerateTokenPair(context.Background(), identity, authz)
	if err != nil {
		t.Fatalf("failed to generate pair: %v", err)
	}

	// verify returned TokenPair structure
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token_type=Bearer, got %s", pair.TokenType)
	}
	if pair.ExpiresIn != 60 {
		t.Errorf("expected expires_in=60, got %d", pair.ExpiresIn)
	}

	// verify access token claims
	accessClaims, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}
	if accessClaims.Identity.Subject != "user-1" {
		t.Errorf("expected subject=user-1, got %s", accessClaims.Identity.Subject)
	}
	if accessClaims.TokenType != TokenTypeAccess {
		t.Errorf("expected token_type=access, got %s", accessClaims.TokenType)
	}
	if len(accessClaims.Authorization.Roles) != 1 || accessClaims.Authorization.Roles[0] != "user" {
		t.Errorf("unexpected roles: %v", accessClaims.Authorization.Roles)
	}

	// verify refresh token claims
	refreshClaims, err := mgr.VerifyToken(context.Background(), pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("failed to verify refresh token: %v", err)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Errorf("expected token_type=refresh, got %s", refreshClaims.TokenType)
	}
}

func TestMissingSubject(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// missing subject should fail
	_, err = mgr.GenerateTokenPair(context.Background(), IdentityClaims{}, AuthorizationClaims{})
	if err != ErrMissingSubject {
		t.Errorf("expected ErrMissingSubject, got %v", err)
	}
}

// ========== key rotation test ==========

func TestKeyRotationAndVerification(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.RotationInterval = time.Second

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}
	originalKid := mustExtractKid(t, pair.AccessToken)

	// force rotation of keys
	rotatedKey, err := mgr.RotateKey()
	if err != nil {
		t.Fatalf("rotate key: %v", err)
	}

	// old token should still be valid (old key is retained)
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("old token failed verification after rotation: %v", err)
	}

	// new token should use new key
	newPair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate new pair: %v", err)
	}
	newKid := mustExtractKid(t, newPair.AccessToken)

	if newKid == originalKid {
		t.Errorf("expected new key id after rotation, got same id: %s", newKid)
	}
	if newKid != rotatedKey.ID {
		t.Errorf("expected kid=%s, got %s", rotatedKey.ID, newKid)
	}
}

func TestAutomaticRotation(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.RotationInterval = 100 * time.Millisecond

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair1, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid1 := mustExtractKid(t, pair1.AccessToken)

	// wait for rotation interval
	time.Sleep(150 * time.Millisecond)

	// generate new token pair should trigger automatic rotation
	pair2, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid2 := mustExtractKid(t, pair2.AccessToken)

	if kid1 == kid2 {
		t.Errorf("expected automatic rotation, but kid remained the same")
	}
}

// ========== blacklist and versioning test ==========

func TestBlacklistAndVersioning(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.AccessExpiration = time.Minute

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate token pair
	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-3"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	// token should be valid
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("token should be valid: %v", err)
	}

	// revoke token
	if err := mgr.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("revoke token: %v", err)
	}

	// token should be revoked
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != ErrTokenRevoked {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}

	// generate new token pair and increment version
	pair, err = mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-3"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	// increment identity version
	if err := mgr.IncrementIdentityVersion("user-3"); err != nil {
		t.Fatalf("increment version: %v", err)
	}

	// old token should be invalidated
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != ErrVersionMismatch {
		t.Errorf("expected ErrVersionMismatch, got %v", err)
	}

	// new token should have new version
	newPair, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-3"}, AuthorizationClaims{})
	newClaims, err := mgr.VerifyToken(context.Background(), newPair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("new token should be valid: %v", err)
	}
	if newClaims.Identity.Version != 1 {
		t.Errorf("expected version=1, got %d", newClaims.Identity.Version)
	}
}

// ========== clock skew tolerance test ==========

func TestClockSkewTolerance(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.AccessExpiration = 1 * time.Second
	cfg.ClockSkew = 2 * time.Second // allow 2 second clock skew

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-skew"}, AuthorizationClaims{})

	// wait for token expiration
	time.Sleep(1100 * time.Millisecond)

	// verify token is valid within clock skew
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Errorf("token should be valid within clock skew: %v", err)
	}

	// wait for token expiration
	time.Sleep(1100 * time.Millisecond)

	// token should be expired
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestNotBeforeWithClockSkew(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.ClockSkew = 2 * time.Second

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// manually build a token with nbf in the future
	mgr.mu.Lock()
	activeKey := mgr.keyCache[mgr.active]
	identity := IdentityClaims{Subject: "user-nbf", Version: 0}
	authz := AuthorizationClaims{}

	jti, _ := randomID()
	now := time.Now().UTC()
	claims := TokenClaims{
		TokenID:       jti,
		TokenType:     TokenTypeAccess,
		Identity:      identity,
		Authorization: authz,
		Issuer:        cfg.Issuer,
		Audience:      cfg.Audience,
		IssuedAt:      now.Unix(),
		NotBefore:     now.Add(3 * time.Second).Unix(), // nbf=3s
		ExpiresAt:     now.Add(time.Hour).Unix(),
		KeyID:         activeKey.ID,
	}
	token, _ := signJWT(activeKey, claims)
	mgr.mu.Unlock()

	// verify token should fail immediately (nbf=3s, elapsed=0s, skew=2s, 0 < 3-2)
	_, err = mgr.VerifyToken(context.Background(), token, TokenTypeAccess)
	if err != ErrTokenNotYetValid {
		t.Errorf("expected ErrTokenNotYetValid, got %v", err)
	}

	// wait for token to become valid (nbf=3s, elapsed=1.5s, skew=2s, 1.5 < 3-2)
	time.Sleep(1500 * time.Millisecond)

	// token should be valid now (nbf=3s, elapsed=1.5s, skew=2s, 1.5 < 3-2)
	_, err = mgr.VerifyToken(context.Background(), token, TokenTypeAccess)
	if err != nil {
		t.Errorf("token should be valid within clock skew: %v", err)
	}
}

// ========== EdDSA Algorithm Test ==========

func TestEdDSAAlgorithm(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.Algorithm = AlgorithmEdDSA

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-eddsa"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair with EdDSA: %v", err)
	}

	// verify header algorithm is EdDSA
	header := mustExtractHeader(t, pair.AccessToken)
	if header["alg"] != string(AlgorithmEdDSA) {
		t.Errorf("expected alg=EdDSA, got %v", header["alg"])
	}

	// verify token claims
	claims, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to verify EdDSA token: %v", err)
	}
	if claims.Identity.Subject != "user-eddsa" {
		t.Errorf("unexpected subject: %s", claims.Identity.Subject)
	}
}

// ========== Error Scenario Test ==========

func TestInvalidToken(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	mgr, _ := NewJWTManager(store, cfg)

	tests := []struct {
		name  string
		token string
		want  error
	}{
		{"empty", "", ErrInvalidToken},
		{"malformed", "not.a.jwt", ErrInvalidToken},
		{"invalid base64", "aaa.bbb.ccc", ErrInvalidToken},
		{"missing parts", "aaa.bbb", ErrInvalidToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.VerifyToken(context.Background(), tt.token, TokenTypeAccess)
			if err != tt.want {
				t.Errorf("expected error %v, got %v", tt.want, err)
			}
		})
	}
}

func TestTokenTypeMismatch(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	mgr, _ := NewJWTManager(store, cfg)

	pair, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-type"}, AuthorizationClaims{})

	// verify using access token to verify refresh token should fail
	_, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeRefresh)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for type mismatch, got %v", err)
	}
}

func TestInvalidIssuerAudience(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.Issuer = "plumego"
	cfg.Audience = "plumego-client"

	mgr, _ := NewJWTManager(store, cfg)

	// generate normal token
	pair, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-iss"}, AuthorizationClaims{})

	// verify with modified issuer should fail
	mgr.config.Issuer = "different-issuer"
	_, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != ErrInvalidIssuer {
		t.Errorf("expected ErrInvalidIssuer, got %v", err)
	}

	// restore issuer, modify audience should fail
	mgr.config.Issuer = "plumego"
	mgr.config.Audience = "different-audience"
	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != ErrInvalidAudience {
		t.Errorf("expected ErrInvalidAudience, got %v", err)
	}
}

// ========== Authorization Policy Test ==========

func TestCheckPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy AuthZPolicy
		auth   AuthorizationClaims
		want   bool
	}{
		{
			name: "all roles and permissions match",
			policy: AuthZPolicy{
				AllRoles:       []string{"admin"},
				AllPermissions: []string{"read", "write"},
			},
			auth: AuthorizationClaims{
				Roles:       []string{"admin", "user"},
				Permissions: []string{"read", "write", "delete"},
			},
			want: true,
		},
		{
			name: "missing required role",
			policy: AuthZPolicy{
				AllRoles: []string{"admin"},
			},
			auth: AuthorizationClaims{
				Roles: []string{"user"},
			},
			want: false,
		},
		{
			name: "any role match",
			policy: AuthZPolicy{
				AnyRole: []string{"admin", "moderator"},
			},
			auth: AuthorizationClaims{
				Roles: []string{"moderator", "user"},
			},
			want: true,
		},
		{
			name: "any permission match",
			policy: AuthZPolicy{
				AnyPermission: []string{"write", "delete"},
			},
			auth: AuthorizationClaims{
				Permissions: []string{"read", "delete"},
			},
			want: true,
		},
		{
			name:   "empty policy allows all",
			policy: AuthZPolicy{},
			auth: AuthorizationClaims{
				Roles: []string{"user"},
			},
			want: true,
		},
		{
			name: "case insensitive matching",
			policy: AuthZPolicy{
				AllRoles: []string{"Admin"},
			},
			auth: AuthorizationClaims{
				Roles: []string{"admin"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPolicy(tt.policy, tt.auth)
			if got != tt.want {
				t.Errorf("checkPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ========== Middleware Test ==========

func TestJWTAuthenticatorMiddleware(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	mgr, _ := NewJWTManager(store, cfg)

	pair, _ := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-mw"}, AuthorizationClaims{Roles: []string{"user"}})

	handler := middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaimsFromContext(r)
		if err != nil {
			t.Errorf("failed to get claims from context: %v", err)
			return
		}
		if claims.Identity.Subject != "user-mw" {
			t.Errorf("unexpected subject in claims: %s", claims.Identity.Subject)
		}
		w.WriteHeader(http.StatusOK)
	})

	authenticator := mgr.JWTAuthenticator(TokenTypeAccess)
	wrapped := authenticator(handler)

	// test valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// test missing token
	req = httptest.NewRequest("GET", "/protected", nil)
	rec = httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	// test invalid token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec = httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ========== Authorization Middleware Test ==========

func TestAuthorizeMiddleware(t *testing.T) {
	handler := middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	policy := AuthZPolicy{
		AllRoles: []string{"admin"},
	}
	authMiddleware := AuthorizeMiddleware(policy)
	wrapped := authMiddleware(handler)

	// test valid claims
	claims := &TokenClaims{
		Identity: IdentityClaims{Subject: "user-authz"},
		Authorization: AuthorizationClaims{
			Roles: []string{"admin"},
		},
	}

	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// test invalid claims
	claims.Authorization.Roles = []string{"user"}
	req = httptest.NewRequest("GET", "/admin", nil)
	ctx = context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)
	rec = httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// ========== Debug Mode Test ==========

func TestDebugMode(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.DebugMode = true
	mgr, _ := NewJWTManager(store, cfg)

	handler := middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authenticator := mgr.JWTAuthenticator(TokenTypeAccess)
	wrapped := authenticator(handler)

	// test expired token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer expired.token.here")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// debug mode should return detailed error
	body := rec.Body.String()
	if !strings.Contains(body, "verification failed") {
		t.Errorf("expected detailed error message in debug mode, got: %s", body)
	}
}

// ========== URL Token Test ==========

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		queryParam string
		allowQuery bool
		want       string
	}{
		{
			name:   "valid bearer header",
			header: "Bearer token123",
			want:   "token123",
		},
		{
			name:   "bearer with extra spaces",
			header: "Bearer   token456  ",
			want:   "token456",
		},
		{
			name:       "query param allowed",
			queryParam: "querytoken",
			allowQuery: true,
			want:       "querytoken",
		},
		{
			name:       "query param not allowed",
			queryParam: "querytoken",
			allowQuery: false,
			want:       "",
		},
		{
			name:       "header takes precedence",
			header:     "Bearer headertoken",
			queryParam: "querytoken",
			allowQuery: true,
			want:       "headertoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/test"
			if tt.queryParam != "" {
				url += "?token=" + tt.queryParam
			}

			req := httptest.NewRequest("GET", url, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := extractBearerToken(req, tt.allowQuery)
			if got != tt.want {
				t.Errorf("extractBearerToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ========== Performance Test ==========

func BenchmarkGenerateTokenPair(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig(nil)
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	authz := AuthorizationClaims{Roles: []string{"user"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GenerateTokenPair(context.Background(), identity, authz)
	}
}

func BenchmarkVerifyToken(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig(nil)
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	pair, _ := mgr.GenerateTokenPair(context.Background(), identity, AuthorizationClaims{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	}
}

func BenchmarkGenerateTokenPairEdDSA(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig(nil)
	cfg.Algorithm = AlgorithmEdDSA
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	authz := AuthorizationClaims{Roles: []string{"user"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GenerateTokenPair(context.Background(), identity, authz)
	}
}

// ========== Helper Functions ==========

func mustExtractKid(t *testing.T, token string) string {
	t.Helper()
	header := mustExtractHeader(t, token)
	kid, _ := header["kid"].(string)
	return kid
}

func mustExtractHeader(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid token format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var hdr map[string]any
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	return hdr
}
