package idempotency

import (
	"context"
	"testing"
)

func TestStorePrimitiveTypes(t *testing.T) {
	record := Record{
		Key:    "key",
		Status: StatusInProgress,
	}
	if record.Status != StatusInProgress {
		t.Fatalf("status = %q, want %q", record.Status, StatusInProgress)
	}

	var _ Store = noopStore{}
}

type noopStore struct{}

func (noopStore) Get(context.Context, string) (Record, bool, error) {
	return Record{}, false, ErrNotFound
}

func (noopStore) PutIfAbsent(context.Context, Record) (bool, error) {
	return false, ErrInvalidKey
}

func (noopStore) Complete(context.Context, string, []byte) error {
	return ErrExpired
}

func (noopStore) Delete(context.Context, string) error {
	return nil
}
