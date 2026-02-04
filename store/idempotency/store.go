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
