package mq

import (
	"context"
	"time"
)

// TaskDeduper prevents duplicate processing of tasks.
// Implementations should return true from IsCompleted when the key has already been processed.
type TaskDeduper interface {
	IsCompleted(ctx context.Context, key string) (bool, error)
	MarkCompleted(ctx context.Context, key string, ttl time.Duration) error
}
