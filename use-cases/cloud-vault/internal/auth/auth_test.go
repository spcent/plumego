package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cloud-vault/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	db, err := database.Open(":memory:")
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}

func createTestService(t *testing.T) *Service {
	db := setupTestDB(t)
	repo := NewRepository(db)
	config := Config{
		PasswordMinLength:         10,
		MaxLoginFailures:          5,
		LoginFailureWindowMinutes: 15,
		SessionTTLHours:           24,
	}

	return NewService(repo, config)
}

func TestService_Setup_Success(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// First setup should succeed
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "admin@example.com", user.Email)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "admin", user.Role)
}

func TestService_Setup_AlreadyInitialized(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// First setup
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Second setup should fail
	_, err = svc.Setup(ctx, "admin2", "admin2@example.com", "StrongPassword456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already")
}

func TestService_Setup_WeakPassword(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Password too short
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "short")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestService_Login_Success(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Login with username
	loginUser, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)
	assert.NotNil(t, loginUser)
	assert.NotNil(t, session)
	assert.NotEmpty(t, token)
	assert.Equal(t, user.ID, loginUser.ID)
	assert.Equal(t, user.ID, session.UserID)

	// Login with email
	loginUser, session, token, err = svc.Login(ctx, "admin@example.com", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)
	assert.NotNil(t, loginUser)
	assert.NotNil(t, session)
	assert.NotEmpty(t, token)
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Wrong password
	_, _, _, err = svc.Login(ctx, "admin", "WrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Wrong username
	_, _, _, err = svc.Login(ctx, "wronguser", "StrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestService_Login_RateLimiting(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Fail 5 times
	for i := 0; i < 5; i++ {
		_, _, _, err = svc.Login(ctx, "admin", "WrongPassword123", "test-agent", "127.0.0.1")
		assert.Error(t, err)
	}

	// 6th attempt should be rate limited
	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err)
	assert.Equal(t, ErrRateLimited, err)
}

func TestService_ValidateSession_Success(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Validate session
	validatedUser, validatedSession, err := svc.ValidateSession(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, session.ID, validatedSession.ID)
}

func TestService_ValidateSession_InvalidToken(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	_, _, err := svc.ValidateSession(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestService_ValidateSession_RevokedSession(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Revoke session
	err = svc.RevokeSession(ctx, session.UserID, session.ID)
	require.NoError(t, err)

	// Validate should fail
	_, _, err = svc.ValidateSession(ctx, token)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestService_Logout(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Logout
	err = svc.Logout(ctx, session.ID)
	require.NoError(t, err)

	// Session should be invalid
	_, _, err = svc.ValidateSession(ctx, token)
	assert.Error(t, err)
}

func TestService_ChangePassword_Success(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Change password
	err = svc.ChangePassword(ctx, user.ID, "StrongPassword123", "NewStrongPassword456")
	require.NoError(t, err)

	// Old password should fail
	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// New password should work
	_, _, _, err = svc.Login(ctx, "admin", "NewStrongPassword456", "test-agent", "127.0.0.1")
	require.NoError(t, err)
}

func TestService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Change with wrong current password
	err = svc.ChangePassword(ctx, user.ID, "WrongPassword123", "NewStrongPassword456")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestService_ChangePassword_RevokeAllSessions(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, _, token1, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	_, _, token2, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Change password
	err = svc.ChangePassword(ctx, user.ID, "StrongPassword123", "NewStrongPassword456")
	require.NoError(t, err)

	// Both sessions should be invalid
	_, _, err = svc.ValidateSession(ctx, token1)
	assert.Error(t, err)

	_, _, err = svc.ValidateSession(ctx, token2)
	assert.Error(t, err)
}

func TestService_UpdateProfile(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Update profile
	err = svc.UpdateProfile(ctx, user.ID, "New Name", "new@example.com", "zh-CN", "dark")
	require.NoError(t, err)

	// Verify changes
	updatedUser, err := svc.repo.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updatedUser.DisplayName)
	assert.Equal(t, "new@example.com", updatedUser.Email)
	assert.Equal(t, "zh-CN", updatedUser.Locale)
	assert.Equal(t, "dark", updatedUser.Theme)
}

func TestService_ListSessions(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Create multiple sessions
	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// List sessions
	sessions, err := svc.ListSessions(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestService_RevokeSession(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, session, token, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Revoke session
	err = svc.RevokeSession(ctx, session.UserID, session.ID)
	require.NoError(t, err)

	// Session should be invalid
	_, _, err = svc.ValidateSession(ctx, token)
	assert.Error(t, err)
}

func TestService_RevokeAllSessions(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup and login
	user, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	_, _, token1, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	_, _, token2, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	_, _, token3, err := svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Revoke all sessions
	err = svc.RevokeAllSessions(ctx, user.ID)
	require.NoError(t, err)

	// All sessions should be invalid
	_, _, err = svc.ValidateSession(ctx, token1)
	assert.Error(t, err)

	_, _, err = svc.ValidateSession(ctx, token2)
	assert.Error(t, err)

	_, _, err = svc.ValidateSession(ctx, token3)
	assert.Error(t, err)

	// List sessions - should still have 3 but all revoked
	sessions, err := svc.ListSessions(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
	for _, s := range sessions {
		assert.NotNil(t, s.RevokedAt)
	}
}

func TestService_ListSecurityEvents(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Setup creates security events
	_, err := svc.Setup(ctx, "admin", "admin@example.com", "StrongPassword123")
	require.NoError(t, err)

	// Login creates security events
	_, _, _, err = svc.Login(ctx, "admin", "StrongPassword123", "test-agent", "127.0.0.1")
	require.NoError(t, err)

	// Failed login creates login_attempts (not security events)
	_, _, _, _ = svc.Login(ctx, "admin", "WrongPassword123", "test-agent", "127.0.0.1")

	// List events (setup + login_success = 2 security events)
	events, err := svc.repo.ListSecurityEvents(ctx, 50, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 2) // setup + login
}

func TestGenerateSessionToken(t *testing.T) {
	token1, err := GenerateSessionToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := GenerateSessionToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestHashSessionToken(t *testing.T) {
	token := "test-token-123"

	hash1 := HashSessionToken(token)
	hash2 := HashSessionToken(token)

	// Same token should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different tokens should produce different hashes
	hash3 := HashSessionToken("different-token")
	assert.NotEqual(t, hash1, hash3)
}
