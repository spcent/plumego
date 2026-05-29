package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/security/password"

	"cloud-vault/internal/idgen"
)

// Service handles authentication business logic.
type Service struct {
	repo   *Repository
	config Config
}

// NewService creates a new authentication service.
func NewService(repo *Repository, config Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

// validatePasswordStrength wraps password.ValidatePasswordStrength to return an error.
func (s *Service) validatePasswordStrength(pwd string) error {
	cfg := password.PasswordStrengthConfig{
		MinLength:        s.config.PasswordMinLength,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   false,
	}
	if !password.ValidatePasswordStrength(pwd, cfg) {
		return fmt.Errorf("auth: password does not meet strength requirements (min length: %d)", s.config.PasswordMinLength)
	}
	return nil
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, username, email, pwd string) (*User, error) {
	if err := s.validatePasswordStrength(pwd); err != nil {
		return nil, err
	}

	hash, err := password.HashPassword(pwd)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           idgen.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         "user",
		Locale:       "en-US",
		Theme:        "system",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and creates a session.
func (s *Service) Login(ctx context.Context, username, pwd, userAgent, ipAddress string) (*User, *Session, string, error) {
	// Check rate limiting
	failures, err := s.repo.CountRecentLoginFailures(ctx, username, s.config.LoginFailureWindowMinutes)
	if err != nil {
		return nil, nil, "", err
	}

	if failures >= s.config.MaxLoginFailures {
		// Record the failed attempt
		_ = s.repo.RecordLoginAttempt(ctx, &LoginAttempt{
			ID:         idgen.New(),
			Identifier: username,
			IPAddress:  ipAddress,
			Success:    false,
			CreatedAt:  time.Now().UTC(),
		})

		return nil, nil, "", ErrRateLimited
	}

	// Look up user by username or email
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err == ErrUserNotFound {
		user, err = s.repo.GetUserByEmail(ctx, username)
	}

	if err != nil {
		// Record failed attempt
		_ = s.repo.RecordLoginAttempt(ctx, &LoginAttempt{
			ID:         idgen.New(),
			Identifier: username,
			IPAddress:  ipAddress,
			Success:    false,
			CreatedAt:  time.Now().UTC(),
		})

		return nil, nil, "", ErrInvalidCredentials
	}

	// Verify password
	if err := password.CheckPassword(user.PasswordHash, pwd); err != nil {
		// Record failed attempt
		_ = s.repo.RecordLoginAttempt(ctx, &LoginAttempt{
			ID:         idgen.New(),
			Identifier: username,
			IPAddress:  ipAddress,
			Success:    false,
			CreatedAt:  time.Now().UTC(),
		})

		return nil, nil, "", ErrInvalidCredentials
	}

	// Record successful login
	_ = s.repo.RecordLoginAttempt(ctx, &LoginAttempt{
		ID:         idgen.New(),
		Identifier: username,
		IPAddress:  ipAddress,
		Success:    true,
		CreatedAt:  time.Now().UTC(),
	})

	// Generate session
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, nil, "", err
	}

	tokenHash := HashSessionToken(token)
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.SessionTTLHours) * time.Hour)

	session := &Session{
		ID:        idgen.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, nil, "", err
	}

	// Record security event
	_ = s.repo.CreateSecurityEvent(ctx, &SecurityEvent{
		ID:        idgen.New(),
		UserID:    user.ID,
		EventType: "login_success",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   fmt.Sprintf("User %s logged in successfully", user.Username),
		CreatedAt: time.Now().UTC(),
	})

	return user, session, token, nil
}

// Logout revokes a session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		return err
	}

	// Record security event
	_ = s.repo.CreateSecurityEvent(ctx, &SecurityEvent{
		ID:        idgen.New(),
		UserID:    session.UserID,
		EventType: "logout",
		IPAddress: session.IPAddress,
		UserAgent: session.UserAgent,
		Details:   "User logged out",
		CreatedAt: time.Now().UTC(),
	})

	return nil
}

// ValidateSession checks if a session token is valid and returns the user.
func (s *Service) ValidateSession(ctx context.Context, token string) (*User, *Session, error) {
	tokenHash := HashSessionToken(token)

	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}

	// Check if session is revoked
	if session.RevokedAt != nil {
		return nil, nil, ErrUnauthorized
	}

	// Check if session is expired
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, nil, ErrSessionExpired
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}

	// Update last used timestamp
	_ = s.repo.UpdateSessionLastUsed(ctx, session.ID)

	return user, session, nil
}

// UpdateProfile updates user profile information.
func (s *Service) UpdateProfile(ctx context.Context, userID string, displayName, email, locale, theme string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if displayName != "" {
		user.DisplayName = displayName
	}
	if email != "" {
		user.Email = email
	}
	if locale != "" {
		user.Locale = locale
	}
	if theme != "" {
		user.Theme = theme
	}

	return s.repo.UpdateUser(ctx, user)
}

// ChangePassword updates a user's password.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := password.CheckPassword(user.PasswordHash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := s.validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := password.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.repo.UpdateUserPassword(ctx, userID, hash); err != nil {
		return err
	}

	// Revoke all existing sessions
	if err := s.repo.RevokeAllUserSessions(ctx, userID); err != nil {
		return err
	}

	// Record security event
	_ = s.repo.CreateSecurityEvent(ctx, &SecurityEvent{
		ID:        idgen.New(),
		UserID:    userID,
		EventType: "password_changed",
		Details:   "Password changed successfully, all sessions revoked",
		CreatedAt: time.Now().UTC(),
	})

	return nil
}

// ListSessions returns all sessions for a user.
func (s *Service) ListSessions(ctx context.Context, userID string) ([]*Session, error) {
	return s.repo.ListUserSessions(ctx, userID)
}

// RevokeSession revokes a specific session.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Ensure the session belongs to the user
	if session.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.RevokeSession(ctx, sessionID)
}

// RevokeAllSessions revokes all sessions for a user.
func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	return s.repo.RevokeAllUserSessions(ctx, userID)
}

// BootstrapAdmin creates the initial admin user if no users exist.
func (s *Service) BootstrapAdmin(ctx context.Context) error {
	if !s.config.BootstrapAdminEnabled {
		return nil
	}

	count, err := s.repo.GetUserCount(ctx)
	if err != nil {
		return err
	}

	// Only bootstrap if no users exist
	if count > 0 {
		return nil
	}

	// Validate configuration
	if s.config.BootstrapAdminUsername == "" || s.config.BootstrapAdminPassword == "" {
		return fmt.Errorf("auth: bootstrap admin username and password are required")
	}

	// Validate password strength
	if err := s.validatePasswordStrength(s.config.BootstrapAdminPassword); err != nil {
		return err
	}

	// Create admin user
	hash, err := password.HashPassword(s.config.BootstrapAdminPassword)
	if err != nil {
		return err
	}

	admin := &User{
		ID:           idgen.New(),
		Username:     s.config.BootstrapAdminUsername,
		Email:        s.config.BootstrapAdminEmail,
		PasswordHash: hash,
		Role:         "admin",
		Locale:       "en-US",
		Theme:        "system",
	}

	if err := s.repo.CreateUser(ctx, admin); err != nil {
		return err
	}

	// Record security event
	_ = s.repo.CreateSecurityEvent(ctx, &SecurityEvent{
		ID:        idgen.New(),
		UserID:    admin.ID,
		EventType: "admin_bootstrap",
		Details:   fmt.Sprintf("Admin user %s created via bootstrap", admin.Username),
		CreatedAt: time.Now().UTC(),
	})

	return nil
}
