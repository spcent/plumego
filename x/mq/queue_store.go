package mq

import (
	"context"
	"time"
)

type TaskStore interface {
	Insert(ctx context.Context, task Task) error
	Reserve(ctx context.Context, opts ReserveOptions) ([]Task, error)
	Ack(ctx context.Context, taskID, consumerID string, now time.Time) error
	Release(ctx context.Context, taskID, consumerID string, opts ReleaseOptions) error
	MoveToDLQ(ctx context.Context, taskID, consumerID, reason string, now time.Time) error
	ExtendLease(ctx context.Context, taskID, consumerID string, lease time.Duration, now time.Time) error
	Stats(ctx context.Context) (Stats, error)
}
