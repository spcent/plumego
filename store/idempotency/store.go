// Package idempotency defines stable, storage-agnostic contracts for
// idempotent request processing.
//
// This package contains only interface definitions, sentinel errors, and shared
// types. Concrete implementations (SQL-backed, KV-backed) live in x/data/idempotency.
//
// Typical wiring:
//
//	import "github.com/spcent/plumego/x/data/idempotency"
//	store := idempotency.NewSQLStore(db)          // concrete implementation
//	_ idempotencycontract.Store = store            // satisfies this interface
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

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

type Record struct {
	Key         string
	RequestHash string
	Status      Status
	Response    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
}

type Store interface {
	Get(ctx context.Context, key string) (Record, bool, error)
	PutIfAbsent(ctx context.Context, record Record) (bool, error)
	Complete(ctx context.Context, key string, response []byte) error
	Delete(ctx context.Context, key string) error
}
