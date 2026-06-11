package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testPassword = "Str0ng!Password#2026"

// newTestServer builds the fully wired app and returns its handler.
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv.Handler
}

func buildJSONRequest(t *testing.T, method, path, bearer string, body any) *http.Request {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return req
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var decoded map[string]any
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("decode response (%d): %v\n%s", rec.Code, err, rec.Body.String())
		}
	}
	return decoded
}

func doJSON(t *testing.T, h http.Handler, method, path, bearer string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := buildJSONRequest(t, method, path, bearer, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeBody(t, rec)
}

// signup registers a user+workspace and returns (accessToken, refreshToken).
func signup(t *testing.T, h http.Handler, email, slug string) (string, string) {
	t.Helper()
	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/signup", "", map[string]any{
		"email":          email,
		"name":           "Tester",
		"password":       testPassword,
		"workspace_name": "Test Space",
		"workspace_slug": slug,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	tokens := body["data"].(map[string]any)["tokens"].(map[string]any)
	return tokens["access_token"].(string), tokens["refresh_token"].(string)
}

func TestAcceptanceSignupLoginMeFlow(t *testing.T) {
	h := newTestServer(t)

	access, _ := signup(t, h, "alice@example.com", "alice-space")

	// /me with the signup access token.
	rec, body := doJSON(t, h, http.MethodGet, "/api/v1/me", access, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("me: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	data := body["data"].(map[string]any)
	if data["user"].(map[string]any)["email"] != "alice@example.com" {
		t.Fatalf("me returned wrong user: %v", data["user"])
	}
	if data["membership"].(map[string]any)["role"] != "owner" {
		t.Fatalf("signup owner role missing: %v", data["membership"])
	}

	// Login returns a fresh pair.
	rec, body = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    "alice@example.com",
		"password": testPassword,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	loginAccess := body["data"].(map[string]any)["tokens"].(map[string]any)["access_token"].(string)
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/me", loginAccess, nil); rec.Code != http.StatusOK {
		t.Fatalf("me with login token: status = %d", rec.Code)
	}
}

func TestAcceptanceRefreshRotation(t *testing.T) {
	h := newTestServer(t)
	_, refresh := signup(t, h, "bob@example.com", "bob-space")

	// First rotation succeeds.
	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": refresh})
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	tokens := body["data"].(map[string]any)["tokens"].(map[string]any)
	next := tokens["refresh_token"].(string)
	newAccess := tokens["access_token"].(string)
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/me", newAccess, nil); rec.Code != http.StatusOK {
		t.Fatalf("me with refreshed access token: status = %d", rec.Code)
	}

	// Reusing the consumed refresh token is rejected …
	if rec, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": refresh}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh reuse: status = %d, want 401", rec.Code)
	}
	// … and revokes the whole family, killing the successor too.
	if rec, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": next}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked family successor: status = %d, want 401", rec.Code)
	}
}

func TestAuthNegativeMatrix(t *testing.T) {
	h := newTestServer(t)
	signup(t, h, "carol@example.com", "carol-space")

	cases := []struct {
		name   string
		method string
		path   string
		bearer string
		body   any
		want   int
	}{
		{"me without token", http.MethodGet, "/api/v1/me", "", nil, http.StatusUnauthorized},
		{"me with garbage token", http.MethodGet, "/api/v1/me", "garbage.token.here", nil, http.StatusUnauthorized},
		{"login wrong password", http.MethodPost, "/api/v1/auth/login", "", map[string]any{"email": "carol@example.com", "password": "Wrong!Pass1"}, http.StatusUnauthorized},
		{"login unknown email", http.MethodPost, "/api/v1/auth/login", "", map[string]any{"email": "ghost@example.com", "password": testPassword}, http.StatusUnauthorized},
		{"signup duplicate email", http.MethodPost, "/api/v1/auth/signup", "", map[string]any{"email": "carol@example.com", "name": "X", "password": testPassword, "workspace_name": "X", "workspace_slug": "other-space"}, http.StatusConflict},
		{"signup weak password", http.MethodPost, "/api/v1/auth/signup", "", map[string]any{"email": "new@example.com", "name": "X", "password": "weak", "workspace_name": "X", "workspace_slug": "new-space"}, http.StatusBadRequest},
		{"signup taken slug", http.MethodPost, "/api/v1/auth/signup", "", map[string]any{"email": "new2@example.com", "name": "X", "password": testPassword, "workspace_name": "X", "workspace_slug": "carol-space"}, http.StatusConflict},
		{"refresh garbage token", http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": "nope"}, http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec, _ := doJSON(t, h, c.method, c.path, c.bearer, c.body)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, c.want, rec.Body.String())
			}
		})
	}
}

func TestPasswordHashNeverInResponse(t *testing.T) {
	h := newTestServer(t)
	rec, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/signup", "", map[string]any{
		"email": "dave@example.com", "name": "Dave", "password": testPassword,
		"workspace_name": "D", "workspace_slug": "dave-space",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup: %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("password")) && bytes.Contains(rec.Body.Bytes(), []byte("$2")) {
		t.Fatal("response leaks a bcrypt hash")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("password_hash")) {
		t.Fatal("response contains password_hash field")
	}
}
