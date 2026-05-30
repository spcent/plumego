package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cloud-vault/internal/auth"
	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
)

// TestSecurityRegression_UnauthenticatedAccess verifies that protected endpoints
// return 401 when accessed without a valid session.
func TestSecurityRegression_UnauthenticatedAccess(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	cfg := config.Config{
		Auth: config.AuthConfig{
			Enabled:               true,
			CookieName:            "session",
			SessionTTLHours:       24,
			PasswordMinLength:     10,
			MaxLoginFailures:      5,
			LockoutMinutes:        30,
			LoginFailureWindowMinutes: 15,
		},
	}

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   cfg.Auth.Enabled,
		CookieName:                cfg.Auth.CookieName,
		SessionTTLHours:           cfg.Auth.SessionTTLHours,
		PasswordMinLength:         cfg.Auth.PasswordMinLength,
		MaxLoginFailures:          cfg.Auth.MaxLoginFailures,
		LockoutMinutes:            cfg.Auth.LockoutMinutes,
		LoginFailureWindowMinutes: cfg.Auth.LoginFailureWindowMinutes,
	})

	handler := auth.NewHandler(svc, auth.Config{
		Enabled:    cfg.Auth.Enabled,
		CookieName: cfg.Auth.CookieName,
	}, nil)

	// Setup initial user
	setupReq := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(`{
		"username": "admin",
		"email": "admin@example.com",
		"password": "StrongPassword123"
	}`))
	setupReq.Header.Set("Content-Type", "application/json")
	setupW := httptest.NewRecorder()
	handler.Setup(setupW, setupReq)
	require.Equal(t, http.StatusCreated, setupW.Code)

	// Test protected endpoint without auth
	protectedReq := httptest.NewRequest("GET", "/api/v1/documents", nil)
	protectedW := httptest.NewRecorder()

	// Simulate middleware check
	if cfg.Auth.Enabled {
		middleware := auth.RequireAuth(svc, cfg.Auth.CookieName)
		wrappedHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		wrappedHandler.ServeHTTP(protectedW, protectedReq)
		assert.Equal(t, http.StatusUnauthorized, protectedW.Code)
	}
}

// TestSecurityRegression_SessionHijackingPrevention verifies that session tokens
// are properly hashed and cannot be used directly from database.
func TestSecurityRegression_SessionHijackingPrevention(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Login
	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is not stored in plain text
	var storedHash string
	err = db.QueryRowContext(ctx,
		"SELECT session_hash FROM user_sessions WHERE id = ?",
		session.ID,
	).Scan(&storedHash)
	require.NoError(t, err)

	// Token should not match stored hash (token is raw, storedHash is SHA-256)
	assert.NotEqual(t, token, storedHash)
	assert.Equal(t, 44, len(storedHash), "SHA-256 base64 length should be 44") // SHA-256 base64 encoded
}

// TestSecurityRegression_PasswordNotInResponse verifies that password hashes
// are never returned in API responses.
func TestSecurityRegression_PasswordNotInResponse(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	handler := auth.NewHandler(svc, auth.Config{
		Enabled:    true,
		CookieName: "session",
	}, nil)

	// Setup user
	setupReq := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(`{
		"username": "admin",
		"email": "admin@example.com",
		"password": "StrongPassword123"
	}`))
	setupReq.Header.Set("Content-Type", "application/json")
	setupW := httptest.NewRecorder()
	handler.Setup(setupW, setupReq)
	require.Equal(t, http.StatusCreated, setupW.Code)

	// Verify response does not contain password hash
	var response map[string]interface{}
	err = json.Unmarshal(setupW.Body.Bytes(), &response)
	require.NoError(t, err)

	userData, ok := response["user"].(map[string]interface{})
	require.True(t, ok)

	// Should not contain password_hash field
	_, hasPasswordHash := userData["password_hash"]
	assert.False(t, hasPasswordHash, "response should not contain password_hash")

	// Should not contain password field
	_, hasPassword := userData["password"]
	assert.False(t, hasPassword, "response should not contain password")
}

// TestSecurityRegression_SessionInvalidationOnPasswordChange verifies that all
// sessions are invalidated when a user changes their password.
func TestSecurityRegression_SessionInvalidationOnPasswordChange(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Create multiple sessions
	_, session1, token1, err := svc.Login(ctx, "admin", "StrongPassword123", "agent1", "127.0.0.1")
	require.NoError(t, err)

	_, session2, token2, err := svc.Login(ctx, "admin", "StrongPassword123", "agent2", "127.0.0.2")
	require.NoError(t, err)

	// Verify both sessions are valid
	_, _, err = svc.ValidateSession(ctx, token1)
	require.NoError(t, err)

	_, _, err = svc.ValidateSession(ctx, token2)
	require.NoError(t, err)

	// Change password
	err = svc.ChangePassword(ctx, user.ID, "StrongPassword123", "NewStrongPassword456")
	require.NoError(t, err)

	// Verify both sessions are now invalid
	_, _, err = svc.ValidateSession(ctx, token1)
	assert.Error(t, err, "session1 should be invalidated after password change")

	_, _, err = svc.ValidateSession(ctx, token2)
	assert.Error(t, err, "session2 should be invalidated after password change")

	_ = session1
	_ = session2
}

// TestSecurityRegression_BruteForceProtection verifies that accounts are locked
// after too many failed login attempts.
func TestSecurityRegression_BruteForceProtection(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Attempt 5 failed logins
	for i := 0; i < 5; i++ {
		_, _, _, err := svc.Login(ctx, "admin", "WrongPassword", "test-agent", "127.0.0.1")
		assert.Error(t, err)
	}

	// 6th attempt should fail even with correct password (account locked)
	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limited")
}

// TestSecurityRegression_SessionExpiration verifies that expired sessions
// cannot be used for authentication.
func TestSecurityRegression_SessionExpiration(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           1, // Short TTL for testing
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Login
	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Manually expire the session in database
	_, err = db.ExecContext(ctx,
		"UPDATE user_sessions SET expires_at = datetime('now', '-1 hour') WHERE id = ?",
		session.ID,
	)
	require.NoError(t, err)

	// Validate should fail
	_, _, err = svc.ValidateSession(ctx, token)
	assert.Error(t, err, "expired session should not be valid")
}

// TestSecurityRegression_ConcurrentSessionLimit verifies that there's no
// hard limit on concurrent sessions (users can be logged in from multiple devices).
func TestSecurityRegression_ConcurrentSessionLimit(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Create 10 concurrent sessions
	tokens := make([]string, 10)
	for i := 0; i < 10; i++ {
		_, _, token, err := svc.Login(ctx, "admin", "StrongPassword123", "agent", "127.0.0.1")
		require.NoError(t, err)
		tokens[i] = token
	}

	// All sessions should be valid
	for i, token := range tokens {
		_, _, err := svc.ValidateSession(ctx, token)
		assert.NoError(t, err, "session %d should be valid", i)
	}
}

// TestSecurityRegression_SQLInjectionPrevention verifies that login credentials
// are properly parameterized to prevent SQL injection.
func TestSecurityRegression_SQLInjectionPrevention(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Attempt SQL injection in username
	_, _, _, err = svc.Login(ctx, "admin' OR '1'='1", "StrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err, "SQL injection in username should fail")

	// Attempt SQL injection in password
	_, _, _, err = svc.Login(ctx, "admin", "' OR '1'='1", "test-agent", "127.0.0.1")
	assert.Error(t, err, "SQL injection in password should fail")
}

// TestSecurityRegression_CookieHttpOnlyFlag verifies that session cookies
// are marked as HttpOnly to prevent XSS attacks.
func TestSecurityRegression_CookieHttpOnlyFlag(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	handler := auth.NewHandler(svc, auth.Config{
		Enabled:    true,
		CookieName: "session",
	}, nil)

	// Setup user
	setupReq := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBufferString(`{
		"username": "admin",
		"email": "admin@example.com",
		"password": "StrongPassword123"
	}`))
	setupReq.Header.Set("Content-Type", "application/json")
	setupW := httptest.NewRecorder()
	handler.Setup(setupW, setupReq)
	require.Equal(t, http.StatusCreated, setupW.Code)

	// Login
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(`{
		"username": "admin",
		"password": "StrongPassword123"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	handler.Login(loginW, loginReq)
	require.Equal(t, http.StatusOK, loginW.Code)

	// Check Set-Cookie header
	cookies := loginW.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "session", cookie.Name)
	assert.True(t, cookie.HttpOnly, "session cookie must be HttpOnly")
	assert.Equal(t, "/", cookie.Path)
}

// TestSecurityRegression_SessionTokenRandomness verifies that session tokens
// are cryptographically random and unique.
func TestSecurityRegression_SessionTokenRandomness(t *testing.T) {
	db, err := database.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.Migrate())

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo, auth.Config{
		Enabled:                   true,
		CookieName:                "session",
		SessionTTLHours:           24,
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LockoutMinutes:            30,
		LoginFailureWindowMinutes: 15,
	})

	ctx := t.Context()

	// Setup user
	_, err = svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Generate 100 session tokens
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		_, _, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
		require.NoError(t, err)

		// Check uniqueness
		_, exists := tokens[token]
		assert.False(t, exists, "duplicate token generated")
		tokens[token] = true

		// Check token length (32 bytes base64 encoded = 43 chars)
		assert.GreaterOrEqual(t, len(token), 40, "token too short")
	}

	// All tokens should be unique
	assert.Equal(t, 100, len(tokens), "not all tokens are unique")
}
