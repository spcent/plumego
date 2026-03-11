package pubsub

import (
	"errors"
	"fmt"
)

// ErrorCode is the machine-readable error identifier.
type ErrorCode string

const (
	ErrCodeClosed         ErrorCode = "PUBSUB_CLOSED"
	ErrCodeInvalidTopic   ErrorCode = "INVALID_TOPIC"
	ErrCodeInvalidPattern ErrorCode = "INVALID_PATTERN"
	ErrCodeInvalidOptions ErrorCode = "INVALID_OPTIONS"
	ErrCodeBufferFull     ErrorCode = "BUFFER_FULL"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeCancelled      ErrorCode = "CANCELLED"
	ErrCodeBackpressure   ErrorCode = "BACKPRESSURE"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeDuplicate      ErrorCode = "DUPLICATE"
	ErrCodeExpired        ErrorCode = "EXPIRED"
	ErrCodeInternal       ErrorCode = "INTERNAL"
)

// Error is the single structured error type for the pubsub package.
//
// Use errors.As to inspect code and cause:
//
//	var e *pubsub.Error
//	if errors.As(err, &e) && e.Code == pubsub.ErrCodeClosed {
//		// broker is shut down
//	}
type Error struct {
	// Code is the machine-readable error code.
	Code ErrorCode

	// Op is the operation that failed (e.g. "publish", "subscribe").
	Op string

	// Topic is the topic involved, if applicable.
	Topic string

	// Message is a human-readable description.
	Message string

	// Cause is the underlying error, if any.
	Cause error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Topic != "" {
		if e.Cause != nil {
			return fmt.Sprintf("%s: %s [topic=%s]: %v", e.Op, e.Message, e.Topic, e.Cause)
		}
		return fmt.Sprintf("%s: %s [topic=%s]", e.Op, e.Message, e.Topic)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying cause for errors.Is / errors.As chaining.
func (e *Error) Unwrap() error { return e.Cause }

// Is compares by Code when target is also a *Error.
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// newErr constructs an *Error. Internal helper — not exported.
func newErr(code ErrorCode, op, topic, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Op:      op,
		Topic:   topic,
		Message: message,
		Cause:   cause,
	}
}
