package idempotency

import stable "github.com/spcent/plumego/store/idempotency"

var (
	ErrNotFound   = stable.ErrNotFound
	ErrInvalidKey = stable.ErrInvalidKey
	ErrExpired    = stable.ErrExpired
)

type Status = stable.Status

const (
	StatusInProgress = stable.StatusInProgress
	StatusCompleted  = stable.StatusCompleted
)

type Record = stable.Record

type Store = stable.Store
