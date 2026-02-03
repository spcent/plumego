package websocket

import (
	"errors"
	"fmt"
)

// Sentinel errors for common websocket conditions
var (
	// Connection errors
	ErrConnClosed  = errors.New("websocket: connection closed")
	ErrConnClosing = errors.New("websocket: connection closing")
	ErrInvalidConn = errors.New("websocket: invalid connection")

	// Queue errors
	ErrQueueFull       = errors.New("websocket: send queue full")
	ErrQueueFullClosed = errors.New("websocket: send queue full, connection closed")

	// Protocol errors
	ErrBadHandshake      = errors.New("websocket: bad handshake")
	ErrBadMethod         = errors.New("websocket: method not GET")
	ErrBadUpgrade        = errors.New("websocket: missing or bad upgrade header")
	ErrBadConnection     = errors.New("websocket: missing or bad connection header")
	ErrBadSecKey         = errors.New("websocket: missing or bad Sec-WebSocket-Key header")
	ErrBadSecAccept      = errors.New("websocket: bad Sec-WebSocket-Accept header")
	ErrPayloadTooLarge   = errors.New("websocket: payload too large")
	ErrProtocolError     = errors.New("websocket: protocol error")
	ErrUnmaskedFrame     = errors.New("websocket: unmasked client frame")
	ErrFragmentedControl = errors.New("websocket: fragmented control frame")
	ErrControlTooLarge   = errors.New("websocket: control frame too large")

	// Hub errors
	ErrHubFull           = errors.New("websocket: hub at capacity")
	ErrRoomFull          = errors.New("websocket: room at capacity")
	ErrRateLimitExceeded = errors.New("websocket: rate limit exceeded")

	// Auth errors
	ErrInvalidToken    = errors.New("websocket: invalid token")
	ErrTokenExpired    = errors.New("websocket: token expired")
	ErrInvalidPassword = errors.New("websocket: invalid password")
	ErrForbiddenOrigin = errors.New("websocket: forbidden origin")
	ErrUnauthorized    = errors.New("websocket: unauthorized")
)

// Error types for more detailed error information

// CloseError represents a WebSocket close frame.
type CloseError struct {
	Code   int
	Text   string
	Reason string
}

func (e *CloseError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("websocket: close %d: %s", e.Code, e.Reason)
	}
	return fmt.Sprintf("websocket: close %d", e.Code)
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("websocket: validation error on %s: %s", e.Field, e.Message)
}

// SecurityError represents a security-related error.
type SecurityError struct {
	Type    string
	Message string
	Details map[string]any
}

func (e *SecurityError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("websocket: security error (%s): %s", e.Type, e.Message)
	}
	return fmt.Sprintf("websocket: security error (%s)", e.Type)
}

// Helper functions to create specific errors

// NewCloseError creates a new CloseError
func NewCloseError(code int, reason string) *CloseError {
	return &CloseError{
		Code:   code,
		Reason: reason,
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewSecurityError creates a new SecurityError
func NewSecurityError(typ, message string) *SecurityError {
	return &SecurityError{
		Type:    typ,
		Message: message,
		Details: make(map[string]any),
	}
}

// IsTemporary checks if an error is temporary
func IsTemporary(err error) bool {
	var closeErr *CloseError
	if errors.As(err, &closeErr) {
		return false // Close errors are permanent
	}

	switch err {
	case ErrQueueFull, ErrRateLimitExceeded:
		return true // These are temporary
	default:
		return false
	}
}
