package idempotency

import (
	"context"
	"errors"
	"testing"
)

// compile-time: noopStore must satisfy Store.
var _ Store = noopStore{}

func TestSentinelErrorsAreNonNilAndDistinct(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrInvalidKey", ErrInvalidKey},
		{"ErrExpired", ErrExpired},
	}
	for _, s := range sentinels {
		if s.err == nil {
			t.Errorf("%s is nil", s.name)
		}
	}
	if errors.Is(ErrNotFound, ErrInvalidKey) || errors.Is(ErrNotFound, ErrExpired) ||
		errors.Is(ErrInvalidKey, ErrExpired) {
		t.Error("sentinel errors must not wrap each other")
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusInProgress == "" {
		t.Error("StatusInProgress must not be empty")
	}
	if StatusCompleted == "" {
		t.Error("StatusCompleted must not be empty")
	}
	if StatusInProgress == StatusCompleted {
		t.Error("StatusInProgress and StatusCompleted must differ")
	}
}

func TestRecordFields(t *testing.T) {
	r := Record{Key: "k", Status: StatusInProgress, Response: []byte("body")}
	if r.Key != "k" {
		t.Errorf("Key = %q, want %q", r.Key, "k")
	}
	if r.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", r.Status, StatusInProgress)
	}
	if string(r.Response) != "body" {
		t.Errorf("Response = %q, want %q", r.Response, "body")
	}
}

// noopStore is a compile-time-only implementation used to verify the interface.
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
