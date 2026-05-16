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

// Extension sentinel errors. Each group corresponds to an optional subsystem.

// Audit
var (
	ErrAuditClosed       = errors.New("audit log is closed")
	ErrAuditCorrupted    = errors.New("audit log corrupted - hash mismatch")
	ErrInvalidAuditQuery = errors.New("invalid audit query parameters")
)

// Backpressure
var (
	ErrBackpressureActive = errors.New("backpressure active - slow down")
	ErrBackpressureClosed = errors.New("backpressure controller is closed")
)

// Consumer groups
var (
	ErrGroupClosed       = errors.New("consumer group is closed")
	ErrGroupNotFound     = errors.New("consumer group not found")
	ErrConsumerNotFound  = errors.New("consumer not found")
	ErrRebalanceRequired = errors.New("rebalance required")
	ErrInvalidAssignment = errors.New("invalid assignment strategy")
)

// Distributed / cluster
var (
	ErrClusterNotJoined  = errors.New("not joined to cluster")
	ErrNodeNotFound      = errors.New("node not found")
	ErrNodeUnhealthy     = errors.New("node is unhealthy")
	ErrBroadcastFailed   = errors.New("broadcast failed")
	ErrConsensusTimeout  = errors.New("consensus timeout")
	ErrInvalidNodeConfig = errors.New("invalid node configuration")
)

// Dead-letter queue
var (
	ErrDLQClosed    = errors.New("dead letter queue is closed")
	ErrDLQNotFound  = errors.New("dead letter message not found")
	ErrInvalidQuery = errors.New("invalid query parameters")
)

// Multi-tenant
var (
	ErrQuotaExceeded       = errors.New("quota exceeded")
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenant       = errors.New("invalid tenant ID")
)

// Ordering
var (
	ErrOrderingClosed    = errors.New("ordering system is closed")
	ErrInvalidOrderLevel = errors.New("invalid order level")
	ErrSequenceGap       = errors.New("sequence number gap detected")
	ErrOutOfOrderMessage = errors.New("out of order message")
)

// Persistence
var (
	ErrPersistenceClosed   = errors.New("persistence layer is closed")
	ErrInvalidWALEntry     = errors.New("invalid WAL entry")
	ErrCorruptedWAL        = errors.New("corrupted WAL file")
	ErrSnapshotFailed      = errors.New("snapshot operation failed")
	ErrRestoreFailed       = errors.New("restore operation failed")
	ErrInvalidDurability   = errors.New("invalid durability level")
	ErrReplicationFailed   = errors.New("replication failed")
	ErrPersistenceDisabled = errors.New("persistence is not enabled")
)

// Rate limiting
var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidRateLimit  = errors.New("invalid rate limit configuration")
)

// Replay
var (
	ErrReplayClosed      = errors.New("replay store is closed")
	ErrMessageNotFound   = errors.New("message not found in replay store")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrArchiveNotEnabled = errors.New("archiving is not enabled")
)
