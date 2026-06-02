package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud-vault/internal/database"
)

// --- helper shared with auth_test.go (same package) ---

func makeService(t *testing.T) *Service {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open DB: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)
	return NewService(repo, Config{
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LoginFailureWindowMinutes: 15,
		SessionTTLHours:           24,
		CookieName:                "session",
	})
}

func makeHandler(t *testing.T, svc *Service) *Handler {
	t.Helper()
	return NewHandler(svc, Config{
		CookieName:      "session",
		SessionTTLHours: 24,
	}, nil)
}

// --- Middleware tests ---

func TestMiddleware_NoCookie_RequireAuth(t *testing.T) {
	svc := makeService(t)
	mw := RequireAuth(svc, "session")
	reached := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("No cookie: got %d, want 401", w.Code)
	}
	if reached {
		t.Error("Handler should not be called when not authenticated")
	}
}

func TestMiddleware_NoCookie_OptionalAuth(t *testing.T) {
	svc := makeService(t)
	mw := OptionalAuth(svc, "session")
	reached := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Optional auth no cookie: got %d, want 200", w.Code)
	}
	if !reached {
		t.Error("Handler should be called for optional auth without cookie")
	}
}

func TestMiddleware_EmptyCookie_RequireAuth(t *testing.T) {
	svc := makeService(t)
	mw := RequireAuth(svc, "session")
	reached := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: ""})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Empty cookie: got %d, want 401", w.Code)
	}
	if reached {
		t.Error("Handler should not be called when cookie is empty")
	}
}

func TestMiddleware_ValidSession_PassesThrough(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	_, err := svc.Setup(ctx, "mwuser", "mw@example.com", "StrongPass123")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	_, _, token, err := svc.Login(ctx, "mwuser", "StrongPass123", "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	mw := RequireAuth(svc, "session")
	var contextUser *User
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Valid session: got %d, want 200", w.Code)
	}
	if contextUser == nil {
		t.Error("User not set in context for valid session")
	}
	if contextUser != nil && contextUser.Username != "mwuser" {
		t.Errorf("Context user: got %q, want mwuser", contextUser.Username)
	}
}

func TestMiddleware_InvalidToken_RequireAuth(t *testing.T) {
	svc := makeService(t)
	mw := RequireAuth(svc, "session")
	reached := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token-xyz"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Invalid token: got %d, want 401", w.Code)
	}
	if reached {
		t.Error("Handler should not be called for invalid token")
	}

	// Middleware should clear the invalid cookie.
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" && c.Value == "" {
			found = true
		}
	}
	if !found {
		t.Error("Invalid token: middleware should clear the cookie")
	}
}

func TestMiddleware_InvalidToken_OptionalAuth(t *testing.T) {
	svc := makeService(t)
	mw := OptionalAuth(svc, "session")
	reached := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		// Optional auth should pass through with no user in context.
		user := UserFromContext(r.Context())
		if user != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Optional auth invalid token: got %d, want 200", w.Code)
	}
	if !reached {
		t.Error("Handler should be called for optional auth with invalid token")
	}
}

func TestMiddleware_SessionInContext(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	_, err := svc.Setup(ctx, "sessctx", "sessctx@example.com", "Password1234")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	_, session, token, err := svc.Login(ctx, "sessctx", "Password1234", "agent", "1.2.3.4")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	mw := RequireAuth(svc, "session")
	var ctxSession *Session
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxSession = SessionFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if ctxSession == nil {
		t.Fatal("Session not set in context")
	}
	if ctxSession.ID != session.ID {
		t.Errorf("Context session ID: got %q, want %q", ctxSession.ID, session.ID)
	}
}

// --- Handler / Login tests ---

func TestHandler_Login_Success(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "loginuser", "login@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	body := `{"username":"loginuser","password":"StrongPassword123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Login success: got %d, want 200", w.Code)
	}

	// Response should have user and session.
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	if _, ok := resp["user"]; !ok {
		t.Error("Response missing 'user' field")
	}
	if _, ok := resp["session"]; !ok {
		t.Error("Response missing 'session' field")
	}

	// Cookie should be set.
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("Session cookie not set after login")
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "user2", "user2@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	body := `{"username":"user2","password":"WrongPassword"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login invalid creds: got %d, want 401", w.Code)
	}
}

func TestHandler_Login_EmptyFields(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	body := `{"username":"","password":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Login empty fields: got %d, want 400", w.Code)
	}
}

func TestHandler_Login_MissingPassword(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	body := `{"username":"admin"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Login missing password: got %d, want 400", w.Code)
	}
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Login invalid JSON: got %d, want 400", w.Code)
	}
}

func TestHandler_Login_RateLimited(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "rluser", "rl@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	// Exhaust the rate limit.
	for i := 0; i < 5; i++ {
		body := `{"username":"rluser","password":"WrongPass123"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Login(w, req)
	}

	// 6th attempt should be rate limited.
	body := `{"username":"rluser","password":"StrongPassword123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Rate limited: got %d, want 429", w.Code)
	}
}

// --- ChangePassword service tests ---

func TestService_ChangePassword_WeakNewPassword(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	user, err := svc.Setup(ctx, "cpuser", "cp@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	err = svc.ChangePassword(ctx, user.ID, "StrongPassword123", "short")
	if err == nil {
		t.Error("ChangePassword weak new password: expected error")
	}
}

func TestService_ChangePassword_UserNotFound(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	err := svc.ChangePassword(ctx, "nonexistent-id", "any-password", "NewStrongPassword123")
	if err == nil {
		t.Error("ChangePassword nonexistent user: expected error")
	}
}

// --- RevokeSession service tests ---

func TestService_RevokeSession_OtherUserSession(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	// Setup first user.
	user1, err := svc.Setup(ctx, "user1rv", "u1rv@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Setup user1: %v", err)
	}

	// Register second user.
	user2, err := svc.Register(ctx, "user2rv", "u2rv@example.com", "StrongPassword456")
	if err != nil {
		t.Fatalf("Register user2: %v", err)
	}

	// Login user2 to create a session.
	_, session2, _, err := svc.Login(ctx, "user2rv", "StrongPassword456", "agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login user2: %v", err)
	}

	// user1 tries to revoke user2's session.
	err = svc.RevokeSession(ctx, user1.ID, session2.ID)
	if err != ErrUnauthorized {
		t.Errorf("RevokeSession other user: got %v, want ErrUnauthorized", err)
	}
	_ = user2
}

func TestService_RevokeSession_SessionNotFound(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	user, err := svc.Setup(ctx, "rv2user", "rv2@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	err = svc.RevokeSession(ctx, user.ID, "nonexistent-session-id")
	if err == nil {
		t.Error("RevokeSession nonexistent: expected error")
	}
}

// --- BootstrapAdmin tests ---

func TestService_BootstrapAdmin_Disabled(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open DB: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, Config{
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LoginFailureWindowMinutes: 15,
		SessionTTLHours:           24,
		BootstrapAdminEnabled:     false,
	})

	// Should be a no-op.
	if err := svc.BootstrapAdmin(ctx); err != nil {
		t.Errorf("BootstrapAdmin disabled: %v", err)
	}

	count, err := repo.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount: %v", err)
	}
	if count != 0 {
		t.Errorf("BootstrapAdmin disabled: created %d users, want 0", count)
	}
}

func TestService_BootstrapAdmin_Enabled(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open DB: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, Config{
		PasswordMinLength:         8,
		MaxLoginFailures:          5,
		LoginFailureWindowMinutes: 15,
		SessionTTLHours:           24,
		BootstrapAdminEnabled:     true,
		BootstrapAdminUsername:    "bootstrap-admin",
		BootstrapAdminEmail:       "admin@bootstrap.com",
		BootstrapAdminPassword:    "BootstrapPass1",
	})

	if err := svc.BootstrapAdmin(ctx); err != nil {
		t.Fatalf("BootstrapAdmin: %v", err)
	}

	count, err := repo.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount: %v", err)
	}
	if count != 1 {
		t.Errorf("BootstrapAdmin: created %d users, want 1", count)
	}

	// Running again should be a no-op.
	if err := svc.BootstrapAdmin(ctx); err != nil {
		t.Fatalf("BootstrapAdmin second call: %v", err)
	}
	count, err = repo.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount second: %v", err)
	}
	if count != 1 {
		t.Errorf("BootstrapAdmin second: count = %d, want 1 (no duplicates)", count)
	}
}

func TestService_BootstrapAdmin_MissingCredentials(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open DB: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, Config{
		PasswordMinLength:      10,
		BootstrapAdminEnabled:  true,
		BootstrapAdminUsername: "",
		BootstrapAdminPassword: "",
	})

	err = svc.BootstrapAdmin(ctx)
	if err == nil {
		t.Error("BootstrapAdmin empty credentials: expected error")
	}
}

// --- Register tests ---

func TestService_Register_Success(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	user, err := svc.Register(ctx, "reguser", "reg@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.ID == "" {
		t.Error("Register: ID is empty")
	}
	if user.Username != "reguser" {
		t.Errorf("Register: Username = %q, want reguser", user.Username)
	}
	if user.Role != "user" {
		t.Errorf("Register: Role = %q, want user", user.Role)
	}
	// Password hash should not be empty.
	if user.PasswordHash == "" {
		t.Error("Register: PasswordHash is empty")
	}
}

func TestService_Register_WeakPassword(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	_, err := svc.Register(ctx, "weakuser", "weak@example.com", "weak")
	if err == nil {
		t.Error("Register weak password: expected error")
	}
}

func TestService_Register_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)

	if _, err := svc.Register(ctx, "dupuser", "dup1@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Register first: %v", err)
	}

	_, err := svc.Register(ctx, "dupuser", "dup2@example.com", "StrongPassword456")
	if err == nil {
		t.Error("Register duplicate username: expected error")
	}
}

// --- Context helpers tests ---

func TestUserFromContext_Nil(t *testing.T) {
	ctx := context.Background()
	if UserFromContext(ctx) != nil {
		t.Error("UserFromContext empty context: expected nil")
	}
}

func TestSessionFromContext_Nil(t *testing.T) {
	ctx := context.Background()
	if SessionFromContext(ctx) != nil {
		t.Error("SessionFromContext empty context: expected nil")
	}
}

func TestWithUser_WithSession_RoundTrip(t *testing.T) {
	ctx := context.Background()

	user := &User{ID: "test-user-id", Username: "testuser"}
	session := &Session{ID: "test-session-id"}

	ctx = WithUser(ctx, user)
	ctx = WithSession(ctx, session)

	gotUser := UserFromContext(ctx)
	if gotUser == nil || gotUser.ID != user.ID {
		t.Errorf("WithUser roundtrip: got %v, want %v", gotUser, user)
	}

	gotSession := SessionFromContext(ctx)
	if gotSession == nil || gotSession.ID != session.ID {
		t.Errorf("WithSession roundtrip: got %v, want %v", gotSession, session)
	}
}

// --- Handler other methods ---

func TestHandler_GetStatus_Uninitialized(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("GET", "/api/v1/auth/status", nil)
	w := httptest.NewRecorder()

	h.GetStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetStatus: got %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if initialized, ok := resp["initialized"].(bool); !ok || initialized {
		t.Errorf("GetStatus uninitialized: initialized = %v, want false", resp["initialized"])
	}
}

func TestHandler_GetStatus_Initialized(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "statususer", "st@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/auth/status", nil)
	w := httptest.NewRecorder()

	h.GetStatus(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if initialized, ok := resp["initialized"].(bool); !ok || !initialized {
		t.Errorf("GetStatus initialized: initialized = %v, want true", resp["initialized"])
	}
}

func TestHandler_GetCurrentUser_NoAuth(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCurrentUser no auth: got %d, want 401", w.Code)
	}
}

func TestHandler_GetCurrentUser_Authenticated(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	user, err := svc.Setup(ctx, "meuser", "me@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req = req.WithContext(WithUser(req.Context(), user))
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetCurrentUser authenticated: got %d, want 200", w.Code)
	}
}

func TestHandler_Logout_NoSession(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Logout no session: got %d, want 401", w.Code)
	}
}

func TestHandler_Logout_Success(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "lout", "lout@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	_, session, token, err := svc.Login(ctx, "lout", "StrongPassword123", "agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req = req.WithContext(WithSession(req.Context(), session))
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Logout success: got %d, want 204", w.Code)
	}

	// Token should be invalid after logout.
	if _, _, err := svc.ValidateSession(ctx, token); err == nil {
		t.Error("Token should be invalid after logout")
	}
}

func TestHandler_ListSessions_NoAuth(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("GET", "/api/v1/auth/sessions", nil)
	w := httptest.NewRecorder()

	h.ListSessions(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ListSessions no auth: got %d, want 401", w.Code)
	}
}

func TestHandler_RevokeAllSessions_NoAuth(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	req := httptest.NewRequest("POST", "/api/v1/auth/sessions/revoke-all", nil)
	w := httptest.NewRecorder()

	h.RevokeAllSessions(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RevokeAllSessions no auth: got %d, want 401", w.Code)
	}
}

func TestHandler_ChangePassword_NoAuth(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	body := `{"current_password":"old","new_password":"new"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ChangePassword no auth: got %d, want 401", w.Code)
	}
}

func TestHandler_ChangePassword_EmptyFields(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	user, err := svc.Setup(ctx, "cphandler", "cphandler@example.com", "StrongPassword123")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	body := `{"current_password":"","new_password":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithUser(req.Context(), user))
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword empty fields: got %d, want 400", w.Code)
	}
}

func TestHandler_Setup_Success(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	body := `{"username":"setuptest","email":"setup@test.com","password":"SetupPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Setup success: got %d, want 201", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := resp["user"]; !ok {
		t.Error("Setup response missing 'user'")
	}
}

func TestHandler_Setup_AlreadyDone(t *testing.T) {
	ctx := context.Background()
	svc := makeService(t)
	h := makeHandler(t, svc)

	if _, err := svc.Setup(ctx, "first-admin", "first@example.com", "StrongPassword123"); err != nil {
		t.Fatalf("First setup: %v", err)
	}

	body := `{"username":"second-admin","email":"second@example.com","password":"StrongPassword456"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Setup(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Setup already done: got %d, want 409", w.Code)
	}
}

func TestHandler_Setup_EmptyFields(t *testing.T) {
	svc := makeService(t)
	h := makeHandler(t, svc)

	body := `{"username":"","password":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Setup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Setup empty fields: got %d, want 400", w.Code)
	}
}

// --- Repository edge case tests ---

func TestRepository_GetUserByID_NotFound(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err = repo.GetUserByID(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByID missing: got %v, want ErrUserNotFound", err)
	}
}

func TestRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err = repo.GetUserByEmail(ctx, "noone@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByEmail missing: got %v, want ErrUserNotFound", err)
	}
}

func TestRepository_GetSessionByID_NotFound(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err = repo.GetSessionByID(ctx, "nosession")
	if err != ErrSessionNotFound {
		t.Errorf("GetSessionByID missing: got %v, want ErrSessionNotFound", err)
	}
}

func TestRepository_DeleteExpiredSessions(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// No sessions → delete 0.
	count, err := repo.DeleteExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	if count != 0 {
		t.Errorf("DeleteExpiredSessions empty: got %d, want 0", count)
	}
}

func TestRepository_GetUserCount_Empty(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount: %v", err)
	}
	if count != 0 {
		t.Errorf("GetUserCount empty: got %d, want 0", count)
	}
}

func TestRepository_ListSecurityEvents_Empty(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	events, err := repo.ListSecurityEvents(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListSecurityEvents: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("ListSecurityEvents empty: got %d", len(events))
	}
}
