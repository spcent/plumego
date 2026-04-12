package session

import "errors"

var (
	ErrSessionRevoked       = errors.New("session revoked")
	ErrSessionExpired       = errors.New("session expired")
	ErrRefreshReused        = errors.New("refresh token reuse detected")
	ErrTokenVersionMismatch = errors.New("token version mismatch")
)
