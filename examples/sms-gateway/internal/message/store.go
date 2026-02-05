package message

import (
	"context"
	"errors"

	"github.com/spcent/plumego/net/mq"
)

var (
	ErrMessageNotFound      = errors.New("message not found")
	ErrMessageStateConflict = errors.New("message state conflict")
	ErrMessageExists        = errors.New("message already exists")
)

// Repository is the minimal persistence interface needed by handler examples.
type Repository interface {
	Insert(ctx context.Context, msg Message) error
	UpdateStatus(ctx context.Context, id string, from Status, to Status, reason Reason) error
}

// Store extends Repository with message lookups.
type Store interface {
	Repository
	Get(ctx context.Context, id string) (Message, bool, error)
}

// ProviderUpdater optionally persists provider selection.
type ProviderUpdater interface {
	UpdateProvider(ctx context.Context, id string, provider string) error
}

// TaskEnqueuer enqueues message tasks into the queue.
type TaskEnqueuer interface {
	Enqueue(ctx context.Context, task mq.Task, opts mq.EnqueueOptions) error
}

// AttemptRecorder records attempt counts for a message.
type AttemptRecorder interface {
	RecordAttempt(ctx context.Context, id string, attempts int) error
}

// DLQRecorder records DLQ reason for a message.
type DLQRecorder interface {
	RecordDLQ(ctx context.Context, id string, reason Reason) error
}
