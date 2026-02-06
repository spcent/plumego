package pubsub

import (
	"errors"
	"fmt"
)

// ErrorCode represents a pubsub error code.
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

// Sentinel errors for backward compatibility
var (
	// ErrClosed is returned when the pubsub system is already closed
	ErrClosed = errors.New("pubsub is closed")

	// ErrInvalidTopic is returned for empty or whitespace-only topics
	ErrInvalidTopic = errors.New("invalid topic")

	// ErrInvalidOpts is returned for invalid subscription options
	ErrInvalidOpts = errors.New("invalid subscribe options")

	// ErrBufferTooSmall is returned when buffer size is less than 1
	ErrBufferTooSmall = errors.New("buffer size must be at least 1")

	// ErrPublishToClosed is returned when publishing to a closed pubsub
	ErrPublishToClosed = errors.New("cannot publish to closed pubsub")

	// ErrSubscribeToClosed is returned when subscribing to a closed pubsub
	ErrSubscribeToClosed = errors.New("cannot subscribe to closed pubsub")

	// ErrInvalidPattern is returned for invalid subscription patterns
	ErrInvalidPattern = errors.New("invalid pattern")

	// ErrBackpressure is returned when system is under backpressure
	ErrBackpressure = errors.New("system under backpressure")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("not found")

	// ErrDuplicate is returned for duplicate messages
	ErrDuplicate = errors.New("duplicate message")

	// ErrExpired is returned when a message has expired
	ErrExpired = errors.New("message expired")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timeout")

	// ErrNoResponse is returned when no response is received
	ErrNoResponse = errors.New("no response received")

	// ErrSchedulerDisabled is returned when delayed publish is attempted without enabling the scheduler
	ErrSchedulerDisabled = errors.New("message scheduler not enabled, use WithScheduler() option")

	// ErrTTLDisabled is returned when TTL publish is attempted without enabling the TTL manager
	ErrTTLDisabled = errors.New("TTL manager not enabled, use WithTTL() option")
)

// PubSubError is a structured error type for pubsub operations.
type PubSubError struct {
	// Code is the error code
	Code ErrorCode

	// Op is the operation that failed (e.g., "publish", "subscribe")
	Op string

	// Topic is the topic involved (if applicable)
	Topic string

	// Message is a human-readable error message
	Message string

	// Cause is the underlying error (if any)
	Cause error

	// Details contains additional context
	Details map[string]any
}

// Error implements the error interface.
func (e *PubSubError) Error() string {
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

// Unwrap returns the underlying error.
func (e *PubSubError) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is support.
func (e *PubSubError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check against sentinel errors based on code
	switch e.Code {
	case ErrCodeClosed:
		return target == ErrClosed || target == ErrPublishToClosed || target == ErrSubscribeToClosed
	case ErrCodeInvalidTopic:
		return target == ErrInvalidTopic
	case ErrCodeInvalidPattern:
		return target == ErrInvalidPattern
	case ErrCodeInvalidOptions:
		return target == ErrInvalidOpts || target == ErrBufferTooSmall
	case ErrCodeBackpressure:
		return target == ErrBackpressure
	case ErrCodeNotFound:
		return target == ErrNotFound
	case ErrCodeDuplicate:
		return target == ErrDuplicate
	case ErrCodeExpired:
		return target == ErrExpired
	case ErrCodeTimeout:
		return target == ErrTimeout
	}

	// Check if target is also a PubSubError with same code
	var pe *PubSubError
	if errors.As(target, &pe) {
		return e.Code == pe.Code
	}

	return false
}

// WithDetail adds a detail to the error.
func (e *PubSubError) WithDetail(key string, value any) *PubSubError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// NewError creates a new PubSubError.
func NewError(code ErrorCode, op, message string) *PubSubError {
	return &PubSubError{
		Code:    code,
		Op:      op,
		Message: message,
	}
}

// NewErrorWithTopic creates a new PubSubError with topic.
func NewErrorWithTopic(code ErrorCode, op, topic, message string) *PubSubError {
	return &PubSubError{
		Code:    code,
		Op:      op,
		Topic:   topic,
		Message: message,
	}
}

// NewErrorWithCause creates a new PubSubError with underlying cause.
func NewErrorWithCause(code ErrorCode, op, message string, cause error) *PubSubError {
	return &PubSubError{
		Code:    code,
		Op:      op,
		Message: message,
		Cause:   cause,
	}
}

// WrapError wraps an existing error with pubsub context.
func WrapError(err error, op, topic string) *PubSubError {
	if err == nil {
		return nil
	}

	// If already a PubSubError, update context
	var pe *PubSubError
	if errors.As(err, &pe) {
		newErr := *pe
		newErr.Op = op
		if topic != "" {
			newErr.Topic = topic
		}
		return &newErr
	}

	// Determine code from error
	code := determineErrorCode(err)

	return &PubSubError{
		Code:    code,
		Op:      op,
		Topic:   topic,
		Message: err.Error(),
		Cause:   err,
	}
}

// determineErrorCode determines the error code from an error.
func determineErrorCode(err error) ErrorCode {
	switch {
	case errors.Is(err, ErrClosed), errors.Is(err, ErrPublishToClosed), errors.Is(err, ErrSubscribeToClosed):
		return ErrCodeClosed
	case errors.Is(err, ErrInvalidTopic):
		return ErrCodeInvalidTopic
	case errors.Is(err, ErrInvalidPattern):
		return ErrCodeInvalidPattern
	case errors.Is(err, ErrInvalidOpts), errors.Is(err, ErrBufferTooSmall):
		return ErrCodeInvalidOptions
	case errors.Is(err, ErrBackpressure):
		return ErrCodeBackpressure
	case errors.Is(err, ErrNotFound):
		return ErrCodeNotFound
	case errors.Is(err, ErrDuplicate):
		return ErrCodeDuplicate
	case errors.Is(err, ErrExpired):
		return ErrCodeExpired
	case errors.Is(err, ErrTimeout):
		return ErrCodeTimeout
	default:
		return ErrCodeInternal
	}
}

// IsClosedError returns true if the error indicates the pubsub is closed.
func IsClosedError(err error) bool {
	if err == nil {
		return false
	}

	var pe *PubSubError
	if errors.As(err, &pe) {
		return pe.Code == ErrCodeClosed
	}

	return errors.Is(err, ErrClosed) ||
		errors.Is(err, ErrPublishToClosed) ||
		errors.Is(err, ErrSubscribeToClosed)
}

// IsTransientError returns true if the error is transient and the operation can be retried.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	var pe *PubSubError
	if errors.As(err, &pe) {
		switch pe.Code {
		case ErrCodeBackpressure, ErrCodeTimeout:
			return true
		}
	}

	return errors.Is(err, ErrBackpressure) || errors.Is(err, ErrTimeout)
}

// IsPermanentError returns true if the error is permanent and should not be retried.
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}

	var pe *PubSubError
	if errors.As(err, &pe) {
		switch pe.Code {
		case ErrCodeClosed, ErrCodeInvalidTopic, ErrCodeInvalidPattern, ErrCodeInvalidOptions:
			return true
		}
	}

	return errors.Is(err, ErrClosed) ||
		errors.Is(err, ErrInvalidTopic) ||
		errors.Is(err, ErrInvalidPattern) ||
		errors.Is(err, ErrInvalidOpts)
}
