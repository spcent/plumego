package websocket

import (
	"errors"
	"fmt"
)

// Sentinel errors for common websocket conditions
var (
	// Connection errors
	ErrConnClosed = errors.New("websocket: connection closed")
	ErrNilNetConn = errors.New("websocket: net connection is nil")

	// Queue errors
	ErrQueueFull       = errors.New("websocket: send queue full")
	ErrQueueFullClosed = errors.New("websocket: send queue full, connection closed")

	// Frame/protocol errors (returned from readFrame)
	ErrPayloadTooLarge    = errors.New("websocket: payload too large")
	ErrProtocolError      = errors.New("websocket: protocol error")
	ErrUnmaskedFrame      = errors.New("websocket: unmasked client frame")
	// ErrUnexpectedMaskedFrame is returned when a client-side Conn receives a
	// masked frame from the server. RFC 6455 §5.1 forbids server→client frames
	// from being masked.
	ErrUnexpectedMaskedFrame = errors.New("websocket: unexpected masked server frame")
	ErrFragmentedControl  = errors.New("websocket: fragmented control frame")
	ErrControlTooLarge    = errors.New("websocket: control frame too large")
	ErrInvalidOpcode      = errors.New("websocket: invalid opcode")
	ErrInvalidCloseCode   = errors.New("websocket: invalid close code")
	ErrCloseReasonTooLong = errors.New("websocket: close reason too long")

	// Hub errors
	ErrHubFull           = errors.New("websocket: hub at capacity")
	ErrRoomFull          = errors.New("websocket: room at capacity")
	ErrHubStopped        = errors.New("websocket: hub stopped")
	ErrRateLimitExceeded = errors.New("websocket: rate limit exceeded")
	ErrNilConn           = errors.New("websocket: connection is nil")
	ErrInvalidRoomName   = errors.New("websocket: invalid room name")

	// Auth errors
	ErrInvalidToken = errors.New("websocket: invalid token")
	ErrTokenExpired = errors.New("websocket: token expired")

	// Security configuration errors
	ErrWeakJWTSecret       = errors.New("jwt secret too weak: minimum 32 bytes required")
	ErrWeakRoomPassword    = errors.New("room password does not meet strength requirements")
	ErrInvalidWebSocketKey = errors.New("invalid websocket key format")
	ErrInvalidConfig       = errors.New("invalid security configuration")

	// Validation errors
	ErrInvalidUTF8       = errors.New("websocket: invalid utf-8 in message")
	ErrControlCharacters = errors.New("websocket: message contains control characters")
	ErrMessageTooLong    = errors.New("websocket: message too long")
	ErrEmptyMessage      = errors.New("websocket: empty message")

	// Server configuration errors
	ErrNilHub              = errors.New("websocket: hub is nil")
	ErrNilAuthenticator    = errors.New("websocket: authenticator is nil")
	ErrNegativeQueueSize   = errors.New("websocket: queue size cannot be negative")
	ErrNegativeWorkerCount = errors.New("websocket: worker count cannot be negative")
	ErrNegativeJobQueue    = errors.New("websocket: job queue size cannot be negative")
	ErrInvalidSendBehavior = errors.New("websocket: invalid send behavior")
	ErrNegativeReadLimit   = errors.New("websocket: read limit cannot be negative")
	ErrInvalidReadLimit    = errors.New("websocket: read limit must be positive")
	ErrInvalidWriteTimeout = errors.New("websocket: write timeout must be positive")
	ErrNegativeLimit       = errors.New("websocket: limit cannot be negative")
	ErrInvalidPingPeriod   = errors.New("websocket: ping period must be positive")
	ErrInvalidPongWait     = errors.New("websocket: pong wait must be positive")
)

const (
	codeWebSocketInvalidConfig      = "WEBSOCKET_INVALID_CONFIG"
	codeWebSocketBadUpgrade         = "WEBSOCKET_BAD_UPGRADE"
	codeWebSocketBadVersion         = "WEBSOCKET_BAD_VERSION"
	codeWebSocketKeyMissing         = "WEBSOCKET_KEY_MISSING"
	codeWebSocketKeyInvalid         = "WEBSOCKET_KEY_INVALID"
	codeWebSocketForbiddenOrigin    = "WEBSOCKET_FORBIDDEN_ORIGIN"
	codeWebSocketRoomForbidden      = "WEBSOCKET_ROOM_FORBIDDEN"
	codeWebSocketInvalidRoom        = "WEBSOCKET_INVALID_ROOM"
	codeWebSocketJoinDenied         = "WEBSOCKET_JOIN_DENIED"
	codeWebSocketTokenRequired      = "WEBSOCKET_TOKEN_REQUIRED"
	codeWebSocketInvalidToken       = "WEBSOCKET_INVALID_TOKEN"
	codeWebSocketHijackUnsupported  = "WEBSOCKET_HIJACK_UNSUPPORTED"
	codeWebSocketHandshakeFailed    = "WEBSOCKET_HANDSHAKE_FAILED"
	codeWebSocketRequestReadFailure = "WEBSOCKET_REQUEST_READ_FAILED"
	codeWebSocketRequestTooLarge    = "WEBSOCKET_REQUEST_TOO_LARGE"
	codeWebSocketBroadcastRejected  = "WEBSOCKET_BROADCAST_REJECTED"
)

// Error types for more detailed error information

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("websocket: validation error on %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// CloseError lets a MessageHandler choose the WebSocket close frame sent to the
// client when it rejects or cannot process a message.
type CloseError struct {
	Code   uint16
	Reason string
	Err    error
}

func (e *CloseError) Error() string {
	if e == nil {
		return "websocket: close error"
	}
	if e.Err != nil {
		return fmt.Sprintf("websocket: close %d %q: %v", e.Code, e.Reason, e.Err)
	}
	return fmt.Sprintf("websocket: close %d %q", e.Code, e.Reason)
}

func (e *CloseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewCloseError creates a handler error that maps to a WebSocket close frame.
func NewCloseError(code uint16, reason string, err error) *CloseError {
	return &CloseError{Code: code, Reason: reason, Err: err}
}
