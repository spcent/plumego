package auth

import (
	"database/sql"
	"time"
)

// userRow represents a user database row.
type userRow struct {
	ID           string
	Username     string
	Email        sql.NullString
	DisplayName  sql.NullString
	PasswordHash string
	Role         string
	Locale       string
	Theme        string
	CreatedAt    string
	UpdatedAt    string
}

// sessionRow represents a session database row.
type sessionRow struct {
	ID          string
	UserID      string
	SessionHash string
	UserAgent   sql.NullString
	IPAddress   sql.NullString
	ExpiresAt   string
	RevokedAt   sql.NullString
	CreatedAt   string
	UpdatedAt   string
}

// securityEventRow represents a security event database row.
type securityEventRow struct {
	ID        string
	UserID    sql.NullString
	EventType string
	IPAddress sql.NullString
	UserAgent sql.NullString
	Details   sql.NullString
	CreatedAt string
}

// loginAttemptRow represents a login attempt database row.
type loginAttemptRow struct {
	ID         string
	Identifier string
	IPAddress  sql.NullString
	Success    bool
	CreatedAt  string
}

// toUser converts a userRow to a User.
func (r *userRow) toUser() *User {
	user := &User{
		ID:           r.ID,
		Username:     r.Username,
		PasswordHash: r.PasswordHash,
		Role:         r.Role,
		Locale:       r.Locale,
		Theme:        r.Theme,
	}

	if r.Email.Valid {
		user.Email = r.Email.String
	}
	if r.DisplayName.Valid {
		user.DisplayName = r.DisplayName.String
	}

	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		user.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
		user.UpdatedAt = t
	}

	return user
}

// toSession converts a sessionRow to a Session.
func (r *sessionRow) toSession() *Session {
	session := &Session{
		ID:        r.ID,
		UserID:    r.UserID,
		TokenHash: r.TokenHash,
	}

	if r.UserAgent.Valid {
		session.UserAgent = r.UserAgent.String
	}
	if r.IPAddress.Valid {
		session.IPAddress = r.IPAddress.String
	}
	if r.RevokedAt.Valid {
		if t, err := time.Parse(time.RFC3339, r.RevokedAt.String); err == nil {
			session.RevokedAt = &t
		}
	}

	if t, err := time.Parse(time.RFC3339, r.ExpiresAt); err == nil {
		session.ExpiresAt = t
	}
	if t, err := time.Parse(time.RFC3339, r.LastUsedAt); err == nil {
		session.LastUsedAt = t
	}
	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		session.CreatedAt = t
	}

	return session
}

// toSecurityEvent converts a securityEventRow to a SecurityEvent.
func (r *securityEventRow) toSecurityEvent() *SecurityEvent {
	event := &SecurityEvent{
		ID:        r.ID,
		EventType: r.EventType,
	}

	if r.UserID.Valid {
		event.UserID = r.UserID.String
	}
	if r.IPAddress.Valid {
		event.IPAddress = r.IPAddress.String
	}
	if r.UserAgent.Valid {
		event.UserAgent = r.UserAgent.String
	}
	if r.Details.Valid {
		event.Details = r.Details.String
	}

	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		event.CreatedAt = t
	}

	return event
}

// toLoginAttempt converts a loginAttemptRow to a LoginAttempt.
func (r *loginAttemptRow) toLoginAttempt() *LoginAttempt {
	attempt := &LoginAttempt{
		ID:         r.ID,
		Identifier: r.Identifier,
		Success:    r.Success,
	}

	if r.IPAddress.Valid {
		attempt.IPAddress = r.IPAddress.String
	}

	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		attempt.CreatedAt = t
	}

	return attempt
}
