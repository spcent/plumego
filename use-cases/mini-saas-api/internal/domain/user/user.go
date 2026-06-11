// Package user owns the user account model, repository, and service.
package user

import (
	"errors"
	"time"
)

// Sentinel errors translated to HTTP status codes by the handler layer.
var (
	ErrEmailTaken = errors.New("user: email already registered")
	ErrNotFound   = errors.New("user: not found")
	// ErrInvalidCredentials covers both unknown email and wrong password so
	// responses do not reveal which one failed.
	ErrInvalidCredentials = errors.New("user: invalid credentials")
	ErrWeakPassword       = errors.New("user: password does not meet strength requirements")
)

// User is an account that can hold memberships in one or more tenants.
// PasswordHash must never be serialized into responses or logs.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
