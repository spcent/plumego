package store

import (
	"context"
	"database/sql"
	"time"
)

type Dialect string

const (
	DialectPostgres Dialect = "postgres"
	DialectMySQL    Dialect = "mysql"
)

type SQLConfig struct {
	Dialect          Dialect
	Table            string
	DLQTable         string
	AttemptsTable    string
	EnableAttemptLog bool
	AttemptLogHook   AttemptLogHook
	Now              func() time.Time
}

type AttemptLogError struct {
	Op      string
	TaskID  string
	Attempt int
	Err     error
}

type AttemptLogHook func(ctx context.Context, info AttemptLogError)

func DefaultSQLConfig() SQLConfig {
	return SQLConfig{
		Dialect:          DialectPostgres,
		Table:            "mq_tasks",
		DLQTable:         "mq_task_dlq",
		AttemptsTable:    "mq_task_attempts",
		EnableAttemptLog: false,
		Now:              time.Now,
	}
}

type SQLStore struct {
	db      *sql.DB
	cfg     SQLConfig
	nowFunc func() time.Time
}

// ReplayOptions configures DLQ replay behavior.
type ReplayOptions struct {
	Max           int
	Now           time.Time
	AvailableAt   time.Time
	ResetAttempts bool
}

// ReplayResult reports DLQ replay results.
type ReplayResult struct {
	Replayed  int
	Remaining int
}

// DLQReplayer exposes DLQ replay operations when supported.
type DLQReplayer interface {
	ReplayDLQ(ctx context.Context, opts ReplayOptions) (ReplayResult, error)
}

type MemConfig struct {
	Now func() time.Time
}

func DefaultMemConfig() MemConfig {
	return MemConfig{Now: time.Now}
}
