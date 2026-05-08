package idempotency

import (
	"errors"
	"strings"

	stable "github.com/spcent/plumego/store/idempotency"
)

var (
	ErrNotFound      = stable.ErrNotFound
	ErrInvalidKey    = stable.ErrInvalidKey
	ErrInvalidRecord = stable.ErrInvalidRecord
	ErrExpired       = stable.ErrExpired
	ErrInvalidConfig = errors.New("idempotency: invalid config")
)

type Status = stable.Status

const (
	StatusInProgress = stable.StatusInProgress
	StatusCompleted  = stable.StatusCompleted
)

type Record = stable.Record

type Store = stable.Store

func ValidateKey(key string) error {
	return stable.ValidateKey(key)
}

func ValidateRecord(record Record) error {
	return stable.ValidateRecord(record)
}

func normalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if err := ValidateKey(key); err != nil {
		return "", err
	}
	return key, nil
}
