package idempotency

import (
	"context"
	"errors"
	"testing"
	"time"
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
		{"ErrInvalidRecord", ErrInvalidRecord},
		{"ErrExpired", ErrExpired},
	}
	for _, s := range sentinels {
		if s.err == nil {
			t.Errorf("%s is nil", s.name)
		}
	}
	if errors.Is(ErrNotFound, ErrInvalidKey) || errors.Is(ErrNotFound, ErrInvalidRecord) ||
		errors.Is(ErrNotFound, ErrExpired) || errors.Is(ErrInvalidKey, ErrInvalidRecord) ||
		errors.Is(ErrInvalidKey, ErrExpired) || errors.Is(ErrInvalidRecord, ErrExpired) {
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

func TestStatusValid(t *testing.T) {
	if !StatusInProgress.Valid() {
		t.Fatal("StatusInProgress should be valid")
	}
	if !StatusCompleted.Valid() {
		t.Fatal("StatusCompleted should be valid")
	}
	if Status("failed").Valid() {
		t.Fatal("unknown status should be invalid")
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

func TestRecordCloneCopiesResponse(t *testing.T) {
	original := Record{
		Key:         "k",
		RequestHash: "sha256:request",
		Status:      StatusCompleted,
		Response:    []byte("body"),
	}

	clone := original.Clone()
	original.Response[0] = 'B'
	clone.Response[1] = 'A'

	if string(original.Response) != "Body" {
		t.Fatalf("original response = %q, want Body", original.Response)
	}
	if string(clone.Response) != "bAdy" {
		t.Fatalf("clone response = %q, want bAdy", clone.Response)
	}
	if (Record{}).Clone().Response != nil {
		t.Fatal("nil response should clone to nil")
	}
}

func TestValidateKey(t *testing.T) {
	if err := ValidateKey("request-123"); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
	for _, key := range []string{"", "   ", "bad\nkey"} {
		err := ValidateKey(key)
		if err == nil || !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("expected ErrInvalidKey for %q, got %v", key, err)
		}
	}
}

func TestValidateRecord(t *testing.T) {
	now := time.Now()
	valid := Record{
		Key:         "request-123",
		RequestHash: "sha256:request",
		Status:      StatusInProgress,
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Minute),
	}
	if err := ValidateRecord(valid); err != nil {
		t.Fatalf("valid record rejected: %v", err)
	}

	cases := []struct {
		name string
		mut  func(Record) Record
		want error
	}{
		{
			name: "invalid key",
			mut: func(record Record) Record {
				record.Key = ""
				return record
			},
			want: ErrInvalidKey,
		},
		{
			name: "missing request hash",
			mut: func(record Record) Record {
				record.RequestHash = ""
				return record
			},
			want: ErrInvalidRecord,
		},
		{
			name: "invalid status",
			mut: func(record Record) Record {
				record.Status = Status("failed")
				return record
			},
			want: ErrInvalidRecord,
		},
		{
			name: "expires before created",
			mut: func(record Record) Record {
				record.ExpiresAt = record.CreatedAt.Add(-time.Second)
				return record
			},
			want: ErrInvalidRecord,
		},
	}
	for _, tc := range cases {
		err := ValidateRecord(tc.mut(valid))
		if err == nil || !errors.Is(err, tc.want) {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, err)
		}
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
