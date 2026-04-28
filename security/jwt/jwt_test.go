package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	authmw "github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/security/authn"
	kvstore "github.com/spcent/plumego/store/kv"
)

func newTestStore(t *testing.T) *kvstore.KVStore {
	t.Helper()
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("failed to create kv store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

// ========== basic function test ==========

func TestGenerateAndVerifyTokenPair(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
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

	pair, err := mgr.GenerateTokenPair(t.Context(), identity, authz)
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
	accessClaims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
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
	refreshClaims, err := mgr.VerifyToken(t.Context(), pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("failed to verify refresh token: %v", err)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Errorf("expected token_type=refresh, got %s", refreshClaims.TokenType)
	}
}

func TestMissingSubject(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// missing subject should fail
	_, err = mgr.GenerateTokenPair(t.Context(), IdentityClaims{}, AuthorizationClaims{})
	if err != ErrMissingSubject {
		t.Errorf("expected ErrMissingSubject, got %v", err)
	}
}

// ========== key rotation test ==========

func TestKeyRotationAndVerification(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = time.Second

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
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
	_, err = mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("old token failed verification after rotation: %v", err)
	}

	// new token should use new key
	newPair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
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

func TestNewJWTManagerRecoversMissingPersistedActiveKey(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-stale-active"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	staleKid := mustExtractKid(t, pair.AccessToken)
	if err := store.Delete(keyPrefix + staleKid); err != nil {
		t.Fatalf("delete active key: %v", err)
	}
	if err := store.Set(activeKeyKey, []byte(staleKid), 0); err != nil {
		t.Fatalf("set stale active key: %v", err)
	}

	recovered, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager with stale active key: %v", err)
	}
	if recovered.active == "" || recovered.active == staleKid {
		t.Fatalf("active key = %q, want new key different from %q", recovered.active, staleKid)
	}
	if _, ok := recovered.keyCache[recovered.active]; !ok {
		t.Fatalf("recovered active key %q missing from cache", recovered.active)
	}

	newPair, err := recovered.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-recovered"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair after recovery: %v", err)
	}
	if _, err := recovered.VerifyToken(t.Context(), newPair.AccessToken, TokenTypeAccess); err != nil {
		t.Fatalf("VerifyToken after recovery: %v", err)
	}
}

func TestAutomaticRotation(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = 100 * time.Millisecond

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair1, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid1 := mustExtractKid(t, pair1.AccessToken)

	// wait for rotation interval
	time.Sleep(150 * time.Millisecond)

	// generate new token pair should trigger automatic rotation
	pair2, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid2 := mustExtractKid(t, pair2.AccessToken)

	if kid1 == kid2 {
		t.Errorf("expected automatic rotation, but kid remained the same")
	}
}

// ========== caller-owned identity version test ==========

func TestIdentityVersionIsCallerOwned(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.AccessExpiration = time.Minute

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-3", Version: 7}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	claims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Identity.Version != 7 {
		t.Errorf("expected version=7, got %d", claims.Identity.Version)
	}
}

// ========== clock skew tolerance test ==========

func TestClockSkewTolerance(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.AccessExpiration = 1 * time.Second
	cfg.ClockSkew = 1 * time.Second // allow 1 second clock skew

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-skew"}, AuthorizationClaims{})

	// wait for token expiration
	time.Sleep(2 * time.Second)

	// verify token is valid within clock skew
	_, err = mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Errorf("token should be valid within clock skew: %v", err)
	}

	// wait for token expiration (total 3.1s > 1s expiration + 2s clock skew)
	time.Sleep(2100 * time.Millisecond)

	// token should be expired
	_, err = mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestNotBeforeWithClockSkew(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
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
	_, err = mgr.VerifyToken(t.Context(), token, TokenTypeAccess)
	if err != ErrTokenNotYetValid {
		t.Errorf("expected ErrTokenNotYetValid, got %v", err)
	}

	// wait for token to become valid (nbf=3s, elapsed=1.5s, skew=2s, 1.5 < 3-2)
	time.Sleep(1500 * time.Millisecond)

	// token should be valid now (nbf=3s, elapsed=1.5s, skew=2s, 1.5 < 3-2)
	_, err = mgr.VerifyToken(t.Context(), token, TokenTypeAccess)
	if err != nil {
		t.Errorf("token should be valid within clock skew: %v", err)
	}
}

// ========== EdDSA Algorithm Test ==========

func TestEdDSAAlgorithm(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Algorithm = AlgorithmEdDSA

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-eddsa"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair with EdDSA: %v", err)
	}

	// verify header algorithm is EdDSA
	header := mustExtractHeader(t, pair.AccessToken)
	if header["alg"] != string(AlgorithmEdDSA) {
		t.Errorf("expected alg=EdDSA, got %v", header["alg"])
	}

	// verify token claims
	claims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to verify EdDSA token: %v", err)
	}
	if claims.Identity.Subject != "user-eddsa" {
		t.Errorf("unexpected subject: %s", claims.Identity.Subject)
	}
}

func TestNewJWTManagerDefaultsEmptyAlgorithmToHS256(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Algorithm = ""

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager with empty algorithm: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-default-alg"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	header := mustExtractHeader(t, pair.AccessToken)
	if header["alg"] != string(AlgorithmHS256) {
		t.Fatalf("alg = %v, want %s", header["alg"], AlgorithmHS256)
	}
}

// ========== Error Scenario Test ==========

func TestInvalidToken(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
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
			_, err := mgr.VerifyToken(t.Context(), tt.token, TokenTypeAccess)
			if err != tt.want {
				t.Errorf("expected error %v, got %v", tt.want, err)
			}
		})
	}
}

func TestTokenTypeMismatch(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, _ := NewJWTManager(store, cfg)

	pair, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-type"}, AuthorizationClaims{})

	// verify using access token to verify refresh token should fail
	_, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeRefresh)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for type mismatch, got %v", err)
	}
}

func TestInvalidIssuerAudience(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Issuer = "plumego"
	cfg.Audience = "plumego-client"

	mgr, _ := NewJWTManager(store, cfg)

	// generate normal token
	pair, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-iss"}, AuthorizationClaims{})

	// verify with modified issuer should fail
	mgr.config.Issuer = "different-issuer"
	_, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != ErrInvalidIssuer {
		t.Errorf("expected ErrInvalidIssuer, got %v", err)
	}

	// restore issuer, modify audience should fail
	mgr.config.Issuer = "plumego"
	mgr.config.Audience = "different-audience"
	_, err = mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != ErrInvalidAudience {
		t.Errorf("expected ErrInvalidAudience, got %v", err)
	}
}

func TestVerifyTokenRequiresConfiguredIssuerAudienceAndSubject(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	baseClaims := TokenClaims{
		TokenID:       "token-semantic",
		TokenType:     TokenTypeAccess,
		Identity:      IdentityClaims{Subject: "user-semantic"},
		Authorization: AuthorizationClaims{},
		Issuer:        cfg.Issuer,
		Audience:      cfg.Audience,
		IssuedAt:      time.Now().UTC().Unix(),
		NotBefore:     time.Now().UTC().Unix(),
		ExpiresAt:     time.Now().UTC().Add(time.Hour).Unix(),
	}

	tests := []struct {
		name   string
		mutate func(*TokenClaims)
		want   error
	}{
		{
			name: "missing issuer",
			mutate: func(claims *TokenClaims) {
				claims.Issuer = ""
			},
			want: ErrInvalidIssuer,
		},
		{
			name: "missing audience",
			mutate: func(claims *TokenClaims) {
				claims.Audience = ""
			},
			want: ErrInvalidAudience,
		},
		{
			name: "missing subject",
			mutate: func(claims *TokenClaims) {
				claims.Identity.Subject = ""
			},
			want: ErrMissingSubject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := baseClaims
			tt.mutate(&claims)
			token := mustSignToken(t, mgr, claims)

			_, err := mgr.VerifyToken(t.Context(), token, TokenTypeAccess)
			if !errors.Is(err, tt.want) {
				t.Fatalf("VerifyToken error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestVerifyTokenRotationIsConcurrentSafe(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = time.Nanosecond
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-race"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				claims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
				if err != nil {
					t.Errorf("VerifyToken: %v", err)
					return
				}
				if claims.Identity.Subject != "user-race" {
					t.Errorf("subject = %q, want user-race", claims.Identity.Subject)
					return
				}
			}
		}()
	}
	wg.Wait()
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

func TestAuthenticatorWithMiddlewareAuth(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, _ := NewJWTManager(store, cfg)

	pair, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-mw"}, AuthorizationClaims{Roles: []string{"user"}})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := TokenClaimsFromContext(r.Context())
		if claims == nil {
			t.Errorf("expected claims in context")
			return
		}
		if claims.Identity.Subject != "user-mw" {
			t.Errorf("unexpected subject in claims: %s", claims.Identity.Subject)
		}
		w.WriteHeader(http.StatusOK)
	})

	authenticator := authmw.Authenticate(mgr.Authenticator(TokenTypeAccess))
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

func TestTokenClaimsContextCopiesMutableSlices(t *testing.T) {
	claims := &TokenClaims{
		Identity: IdentityClaims{Subject: "user-context-copy"},
		Authorization: AuthorizationClaims{
			Roles:       []string{"admin"},
			Permissions: []string{"read:all"},
		},
	}

	ctx := WithTokenClaims(t.Context(), claims)
	claims.Authorization.Roles[0] = "mutated-role"
	claims.Authorization.Permissions[0] = "mutated-permission"

	got := TokenClaimsFromContext(ctx)
	if got == nil {
		t.Fatal("expected claims from context")
	}
	if got.Authorization.Roles[0] != "admin" {
		t.Fatalf("stored role alias = %q, want admin", got.Authorization.Roles[0])
	}
	if got.Authorization.Permissions[0] != "read:all" {
		t.Fatalf("stored permission alias = %q, want read:all", got.Authorization.Permissions[0])
	}

	got.Authorization.Roles[0] = "returned-role"
	got.Authorization.Permissions[0] = "returned-permission"
	again := TokenClaimsFromContext(ctx)
	if again.Authorization.Roles[0] != "admin" {
		t.Fatalf("context role mutated through returned claims = %q", again.Authorization.Roles[0])
	}
	if again.Authorization.Permissions[0] != "read:all" {
		t.Fatalf("context permission mutated through returned claims = %q", again.Authorization.Permissions[0])
	}
}

// ========== Authorization Middleware Test ==========

func TestPolicyAuthorizerWithMiddlewareAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	policy := AuthZPolicy{
		AllRoles: []string{"admin"},
	}
	authMiddleware := authmw.Authorize(PolicyAuthorizer{Policy: policy}, "", "")
	wrapped := authMiddleware(handler)

	// test valid claims
	claims := &TokenClaims{
		Identity: IdentityClaims{Subject: "user-authz"},
		Authorization: AuthorizationClaims{
			Roles: []string{"admin"},
		},
	}

	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := authn.WithPrincipal(req.Context(), PrincipalFromClaims(claims))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// test invalid claims
	claims.Authorization.Roles = []string{"user"}
	req = httptest.NewRequest("GET", "/admin", nil)
	ctx = authn.WithPrincipal(req.Context(), PrincipalFromClaims(claims))
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
	cfg := DefaultJWTConfig()
	// DebugMode has been removed for security - tokens are always validated strictly
	mgr, _ := NewJWTManager(store, cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authenticator := authmw.Authenticate(mgr.Authenticator(TokenTypeAccess))
	wrapped := authenticator(handler)

	// test expired token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer expired.token.here")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Security fix: debug mode no longer returns detailed error messages
	// to prevent information leakage. Generic error is always returned.
	body := rec.Body.String()
	if !strings.Contains(body, "invalid token") {
		t.Errorf("expected generic error message, got: %s", body)
	}
}

// ========== URL Token Test ==========

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
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
			name:   "no authorization header",
			header: "",
			want:   "",
		},
		{
			name:   "non-bearer authorization",
			header: "Basic dXNlcjpwYXNz",
			want:   "",
		},
		{
			name:   "case insensitive bearer",
			header: "BEARER token789",
			want:   "token789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := authn.ExtractBearerToken(req)
			if got != tt.want {
				t.Errorf("ExtractBearerToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ========== Performance Test ==========

func BenchmarkGenerateTokenPair(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig()
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	authz := AuthorizationClaims{Roles: []string{"user"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GenerateTokenPair(b.Context(), identity, authz)
	}
}

func BenchmarkVerifyToken(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig()
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	pair, _ := mgr.GenerateTokenPair(b.Context(), identity, AuthorizationClaims{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.VerifyToken(b.Context(), pair.AccessToken, TokenTypeAccess)
	}
}

func BenchmarkGenerateTokenPairEdDSA(b *testing.B) {
	store, _ := kvstore.NewKVStore(kvstore.Options{DataDir: b.TempDir()})
	cfg := DefaultJWTConfig()
	cfg.Algorithm = AlgorithmEdDSA
	mgr, _ := NewJWTManager(store, cfg)

	identity := IdentityClaims{Subject: "bench-user"}
	authz := AuthorizationClaims{Roles: []string{"user"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GenerateTokenPair(b.Context(), identity, authz)
	}
}

// ========== Contract Adapter Test ==========

func TestContractAuthenticator(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-contract"}, AuthorizationClaims{Roles: []string{"user"}})

	authenticator := mgr.Authenticator(TokenTypeAccess)
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	principal, err := authenticator.Authenticate(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if principal == nil || principal.Subject != "user-contract" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if len(principal.Roles) != 1 || principal.Roles[0] != "user" {
		t.Fatalf("unexpected roles: %#v", principal.Roles)
	}
}

func TestContractAuthenticatorMissingToken(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	mgr, _ := NewJWTManager(store, cfg)

	authenticator := mgr.Authenticator(TokenTypeAccess)
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)

	_, err := authenticator.Authenticate(req)
	if err != authn.ErrUnauthenticated {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestPolicyAuthorizer(t *testing.T) {
	authorizer := PolicyAuthorizer{Policy: AuthZPolicy{AllRoles: []string{"admin"}}}
	principal := &authn.Principal{Roles: []string{"admin"}}

	if err := authorizer.Authorize(principal, "", ""); err != nil {
		t.Fatalf("expected authorized, got %v", err)
	}

	principal.Roles = []string{"user"}
	if err := authorizer.Authorize(principal, "", ""); err != authn.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
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

func mustSignToken(t *testing.T, mgr *JWTManager, claims TokenClaims) string {
	t.Helper()

	mgr.mu.RLock()
	key := mgr.keyCache[mgr.active]
	mgr.mu.RUnlock()
	claims.KeyID = key.ID

	token, err := signJWT(key, claims)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
