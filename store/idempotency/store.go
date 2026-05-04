// Package idempotency defines stable, storage-agnostic contracts for
// idempotent request processing.
//
// This package contains only interface definitions, sentinel errors, and shared
// types. Concrete implementations (SQL-backed, KV-backed) live in x/data/idempotency.
//
// Typical wiring:
//
//	import (
//		stableidempotency "github.com/spcent/plumego/store/idempotency"
//		dataidempotency "github.com/spcent/plumego/x/data/idempotency"
//	)
//
//	store := dataidempotency.NewSQLStore(db, dataidempotency.DefaultSQLConfig()) // concrete implementation
//	var _ stableidempotency.Store = store    // satisfies this interface
package idempotency

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("idempotency: record not found")
	ErrInvalidKey = errors.New("idempotency: key is required")
	ErrExpired    = errors.New("idempotency: record expired")
)

// Status describes the lifecycle state of an idempotency record.
type Status string

const (
	// StatusInProgress marks a request that has claimed an idempotency key but
	// has not completed with a replayable response yet.
	StatusInProgress Status = "in_progress"

	// StatusCompleted marks a request that has completed and can replay its
	// stored response for matching duplicate requests.
	StatusCompleted Status = "completed"
)

// Record is the storage-agnostic representation of an idempotency entry.
//
// Store implementations own durable retention and expiry checks. Callers own
// request hashing policy: stores persist RequestHash but do not decide whether
// a duplicate request with a different hash is a conflict.
type Record struct {
	// Key is the caller-provided idempotency key.
	Key string

	// RequestHash identifies the request payload or semantic operation that
	// claimed the key. Concrete implementations define the hash format.
	RequestHash string

	// Status is the lifecycle state of the record.
	Status Status

	// Response is caller-owned. Implementations that retain it after Complete or
	// PutIfAbsent must make a defensive copy. Get returns implementation-owned
	// record values; callers that retain Response should copy it before mutation.
	Response []byte

	// CreatedAt records when the key was first claimed.
	CreatedAt time.Time

	// UpdatedAt records the last state transition time.
	UpdatedAt time.Time

	// ExpiresAt records when the entry should no longer be considered usable.
	ExpiresAt time.Time
}

// Store is the stable contract implemented by concrete idempotency backends.
//
// The stable layer tracks storage state only. It does not define HTTP retry
// policy, response replay policy, request-hash conflict handling, table schema,
// or backend duplicate-key mechanics.
type Store interface {
	// Get returns a record by key. The bool result reports whether a usable,
	// non-expired record was found without requiring callers to inspect
	// ErrNotFound. Missing and expired records should return found=false with a
	// nil error after any best-effort cleanup. Invalid keys return ErrInvalidKey.
	Get(ctx context.Context, key string) (Record, bool, error)

	// PutIfAbsent stores a new record only when the key is not already claimed by
	// a usable record. The bool result reports whether this call claimed the key.
	// If a usable in-progress or completed record already exists, implementations
	// return inserted=false with nil error; callers can then Get and compare
	// RequestHash if business policy needs conflict detection. A candidate record
	// that is already expired returns ErrExpired.
	PutIfAbsent(ctx context.Context, record Record) (bool, error)

	// Complete marks an existing usable record complete with a replayable
	// response. Missing records and records discovered to be expired during
	// completion return ErrNotFound after any best-effort cleanup.
	Complete(ctx context.Context, key string, response []byte) error

	// Delete removes a record by key. Delete is a storage cleanup operation; a
	// missing key returns ErrNotFound. Invalid keys return ErrInvalidKey.
	Delete(ctx context.Context, key string) error
}
