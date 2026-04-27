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
	if StatusInProgress != "in_progress" {
		t.Errorf("StatusInProgress = %q, want %q", StatusInProgress, "in_progress")
	}
	if StatusCompleted != "completed" {
		t.Errorf("StatusCompleted = %q, want %q", StatusCompleted, "completed")
	}
	if StatusInProgress == StatusCompleted {
		t.Error("StatusInProgress and StatusCompleted must differ")
	}
}

func TestRecordZeroValue(t *testing.T) {
	var r Record
	if r.Key != "" || r.RequestHash != "" || r.Status != "" || r.Response != nil {
		t.Fatalf("unexpected zero-value record: %+v", r)
	}
	if !r.CreatedAt.IsZero() || !r.UpdatedAt.IsZero() || !r.ExpiresAt.IsZero() {
		t.Fatalf("zero-value record should have zero timestamps: %+v", r)
	}
}

func TestRecordFields(t *testing.T) {
	response := []byte("body")
	r := Record{
		Key:         "k",
		RequestHash: "sha256:request",
		Status:      StatusInProgress,
		Response:    response,
	}
	if r.Key != "k" {
		t.Errorf("Key = %q, want %q", r.Key, "k")
	}
	if r.RequestHash != "sha256:request" {
		t.Errorf("RequestHash = %q, want %q", r.RequestHash, "sha256:request")
	}
	if r.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", r.Status, StatusInProgress)
	}
	if string(r.Response) != "body" {
		t.Errorf("Response = %q, want %q", r.Response, "body")
	}

	response[0] = 'B'
	if string(r.Response) != "Body" {
		t.Errorf("Record value should expose caller-owned response bytes, got %q", r.Response)
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
