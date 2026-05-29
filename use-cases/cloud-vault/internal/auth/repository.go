package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloud-vault/internal/database"
)

// Repository handles auth-related database operations.
type Repository struct {
	db *database.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user into the database.
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var email, displayName sql.NullString
	if user.Email != "" {
		email = sql.NullString{String: user.Email, Valid: true}
	}
	if user.DisplayName != "" {
		displayName = sql.NullString{String: user.DisplayName, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, display_name, password_hash, role, locale, theme, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, email, displayName, user.PasswordHash,
		user.Role, user.Locale, user.Theme, now, now,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("auth: create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID.
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, display_name, password_hash, role, locale, theme, created_at, updated_at
		FROM users
		WHERE id = ?`, id)

	var userRow userRow
	err := row.Scan(
		&userRow.ID, &userRow.Username, &userRow.Email, &userRow.DisplayName,
		&userRow.PasswordHash, &userRow.Role, &userRow.Locale, &userRow.Theme,
		&userRow.CreatedAt, &userRow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("auth: get user by id: %w", err)
	}

	return userRow.toUser(), nil
}

// GetUserByUsername retrieves a user by username.
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, display_name, password_hash, role, locale, theme, created_at, updated_at
		FROM users
		WHERE username = ?`, username)

	var userRow userRow
	err := row.Scan(
		&userRow.ID, &userRow.Username, &userRow.Email, &userRow.DisplayName,
		&userRow.PasswordHash, &userRow.Role, &userRow.Locale, &userRow.Theme,
		&userRow.CreatedAt, &userRow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("auth: get user by username: %w", err)
	}

	return userRow.toUser(), nil
}

// GetUserByEmail retrieves a user by email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, display_name, password_hash, role, locale, theme, created_at, updated_at
		FROM users
		WHERE email = ?`, email)

	var userRow userRow
	err := row.Scan(
		&userRow.ID, &userRow.Username, &userRow.Email, &userRow.DisplayName,
		&userRow.PasswordHash, &userRow.Role, &userRow.Locale, &userRow.Theme,
		&userRow.CreatedAt, &userRow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("auth: get user by email: %w", err)
	}

	return userRow.toUser(), nil
}

// GetUserCount returns the total number of users.
func (r *Repository) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("auth: get user count: %w", err)
	}
	return count, nil
}

// UpdateUser updates user profile fields.
func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var email, displayName sql.NullString
	if user.Email != "" {
		email = sql.NullString{String: user.Email, Valid: true}
	}
	if user.DisplayName != "" {
		displayName = sql.NullString{String: user.DisplayName, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET email = ?, display_name = ?, locale = ?, theme = ?, updated_at = ?
		WHERE id = ?`,
		email, displayName, user.Locale, user.Theme, now, user.ID,
	)

	if err != nil {
		return fmt.Errorf("auth: update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: update user rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserPassword updates a user's password hash.
func (r *Repository) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = ?, updated_at = ?
		WHERE id = ?`,
		passwordHash, now, userID,
	)

	if err != nil {
		return fmt.Errorf("auth: update user password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: update password rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// CreateSession inserts a new session into the database.
func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var userAgent, ipAddress sql.NullString
	if session.UserAgent != "" {
		userAgent = sql.NullString{String: session.UserAgent, Valid: true}
	}
	if session.IPAddress != "" {
		ipAddress = sql.NullString{String: session.IPAddress, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_sessions (id, user_id, token_hash, user_agent, ip_address, expires_at, last_used_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.TokenHash,
		userAgent, ipAddress,
		session.ExpiresAt.UTC().Format(time.RFC3339),
		now, now,
	)

	if err != nil {
		return fmt.Errorf("auth: create session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID.
func (r *Repository) GetSessionByID(ctx context.Context, id string) (*Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, last_used_at, created_at
		FROM user_sessions
		WHERE id = ?`, id)

	var sessionRow sessionRow
	err := row.Scan(
		&sessionRow.ID, &sessionRow.UserID, &sessionRow.TokenHash,
		&sessionRow.UserAgent, &sessionRow.IPAddress,
		&sessionRow.ExpiresAt, &sessionRow.RevokedAt,
		&sessionRow.LastUsedAt, &sessionRow.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("auth: get session by id: %w", err)
	}

	return sessionRow.toSession(), nil
}

// GetSessionByTokenHash retrieves a session by token hash.
func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, last_used_at, created_at
		FROM user_sessions
		WHERE token_hash = ?`, tokenHash)

	var sessionRow sessionRow
	err := row.Scan(
		&sessionRow.ID, &sessionRow.UserID, &sessionRow.TokenHash,
		&sessionRow.UserAgent, &sessionRow.IPAddress,
		&sessionRow.ExpiresAt, &sessionRow.RevokedAt,
		&sessionRow.LastUsedAt, &sessionRow.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("auth: get session by token hash: %w", err)
	}

	return sessionRow.toSession(), nil
}

// UpdateSessionLastUsed updates the last_used_at timestamp for a session.
func (r *Repository) UpdateSessionLastUsed(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions
		SET last_used_at = ?
		WHERE id = ?`,
		now, id,
	)

	if err != nil {
		return fmt.Errorf("auth: update session last used: %w", err)
	}

	return nil
}

// RevokeSession marks a session as revoked.
func (r *Repository) RevokeSession(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions
		SET revoked_at = ?
		WHERE id = ?`,
		now, id,
	)

	if err != nil {
		return fmt.Errorf("auth: revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: revoke session rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user.
func (r *Repository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions
		SET revoked_at = ?
		WHERE user_id = ? AND revoked_at IS NULL`,
		now, userID,
	)

	if err != nil {
		return fmt.Errorf("auth: revoke all user sessions: %w", err)
	}

	return nil
}

// ListUserSessions returns all sessions for a user.
func (r *Repository) ListUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, last_used_at, created_at
		FROM user_sessions
		WHERE user_id = ?
		ORDER BY created_at DESC`, userID)

	if err != nil {
		return nil, fmt.Errorf("auth: list user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sessionRow sessionRow
		err := rows.Scan(
			&sessionRow.ID, &sessionRow.UserID, &sessionRow.TokenHash,
			&sessionRow.UserAgent, &sessionRow.IPAddress,
			&sessionRow.ExpiresAt, &sessionRow.RevokedAt,
			&sessionRow.LastUsedAt, &sessionRow.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("auth: scan session: %w", err)
		}
		sessions = append(sessions, sessionRow.toSession())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("auth: iterate sessions: %w", err)
	}

	return sessions, nil
}

// DeleteExpiredSessions removes expired sessions from the database.
func (r *Repository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM user_sessions
		WHERE expires_at < ?`, now)

	if err != nil {
		return 0, fmt.Errorf("auth: delete expired sessions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("auth: delete expired sessions count: %w", err)
	}

	return count, nil
}

// RecordLoginAttempt records a login attempt.
func (r *Repository) RecordLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	var ipAddress sql.NullString
	if attempt.IPAddress != "" {
		ipAddress = sql.NullString{String: attempt.IPAddress, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO login_attempts (id, identifier, ip_address, success, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		attempt.ID, attempt.Identifier, ipAddress, attempt.Success,
		attempt.CreatedAt.UTC().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("auth: record login attempt: %w", err)
	}

	return nil
}

// CountRecentLoginFailures counts failed login attempts within a time window.
func (r *Repository) CountRecentLoginFailures(ctx context.Context, identifier string, windowMinutes int) (int, error) {
	since := time.Now().UTC().Add(-time.Duration(windowMinutes) * time.Minute).Format(time.RFC3339)

	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM login_attempts
		WHERE identifier = ? AND success = FALSE AND created_at >= ?`,
		identifier, since,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("auth: count recent login failures: %w", err)
	}

	return count, nil
}

// CreateSecurityEvent records a security event.
func (r *Repository) CreateSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	var userID, ipAddress, userAgent, details sql.NullString
	if event.UserID != "" {
		userID = sql.NullString{String: event.UserID, Valid: true}
	}
	if event.IPAddress != "" {
		ipAddress = sql.NullString{String: event.IPAddress, Valid: true}
	}
	if event.UserAgent != "" {
		userAgent = sql.NullString{String: event.UserAgent, Valid: true}
	}
	if event.Details != "" {
		details = sql.NullString{String: event.Details, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO security_events (id, user_id, event_type, ip_address, user_agent, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, userID, event.EventType, ipAddress, userAgent, details,
		event.CreatedAt.UTC().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("auth: create security event: %w", err)
	}

	return nil
}

// ListSecurityEvents returns security events with pagination.
func (r *Repository) ListSecurityEvents(ctx context.Context, limit, offset int) ([]*SecurityEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, event_type, ip_address, user_agent, details, created_at
		FROM security_events
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("auth: list security events: %w", err)
	}
	defer rows.Close()

	var events []*SecurityEvent
	for rows.Next() {
		var eventRow securityEventRow
		err := rows.Scan(
			&eventRow.ID, &eventRow.UserID, &eventRow.EventType,
			&eventRow.IPAddress, &eventRow.UserAgent, &eventRow.Details,
			&eventRow.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("auth: scan security event: %w", err)
		}
		events = append(events, eventRow.toSecurityEvent())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("auth: iterate security events: %w", err)
	}

	return events, nil
}

// isDuplicateKeyError checks if an error is a duplicate key violation.
func isDuplicateKeyError(err error) bool {
	// SQLite error code for UNIQUE constraint violation
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") ||
		contains(err.Error(), "PRIMARY KEY must be unique"))
}

// contains checks if a string contains a substring (simple implementation).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring finds a substring in a string.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
