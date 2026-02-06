package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spcent/plumego/store/cache"
)

var errMiss = errors.New("cache miss")

type stubClient struct {
	data map[string][]byte
}

func (s *stubClient) Get(ctx context.Context, key string) ([]byte, error) {
	value, ok := s.data[key]
	if !ok {
		return nil, errMiss
	}
	return value, nil
}

func (s *stubClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s.data == nil {
		s.data = make(map[string][]byte)
	}
	s.data[key] = value
	return nil
}

func (s *stubClient) Del(ctx context.Context, keys ...string) (int64, error) {
	var removed int64
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			delete(s.data, key)
			removed++
		}
	}
	return removed, nil
}

func (s *stubClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	var count int64
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			count++
		}
	}
	return count, nil
}

type stubFlusher struct {
	stubClient
	flushed bool
}

func (s *stubFlusher) FlushDB(ctx context.Context) error {
	s.data = make(map[string][]byte)
	s.flushed = true
	return nil
}

func TestAdapterGetNotFound(t *testing.T) {
	adapter := NewAdapter(&stubClient{}, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	_, err := adapter.Get(context.Background(), "missing")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapterSetAndExists(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	if err := adapter.Set(context.Background(), "key", []byte("value"), time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, err := adapter.Exists(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("expected key to exist")
	}
}

func TestAdapterClear(t *testing.T) {
	client := &stubFlusher{stubClient: stubClient{data: map[string][]byte{"k": []byte("v")}}}
	adapter := NewAdapter(client, nil)

	if err := adapter.Clear(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.flushed {
		t.Fatalf("expected flush to be called")
	}
}

func TestAdapterClearUnsupported(t *testing.T) {
	adapter := NewAdapter(&stubClient{}, nil)

	if err := adapter.Clear(context.Background()); !errors.Is(err, ErrClearUnsupported) {
		t.Fatalf("expected ErrClearUnsupported, got %v", err)
	}
}

func TestAdapterKeyValidation(t *testing.T) {
	adapter := NewAdapter(&stubClient{data: make(map[string][]byte)}, nil)

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "valid key",
			key:     "valid:key",
			wantErr: false,
		},
		{
			name:    "key with newline",
			key:     "key\nwith\nnewline",
			wantErr: true,
		},
		{
			name:    "key with tab",
			key:     "key\twith\ttab",
			wantErr: true,
		},
		{
			name:    "key with null byte",
			key:     "key\x00null",
			wantErr: true,
		},
		{
			name:    "key with DEL character",
			key:     "key\x7Fdel",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := adapter.Set(context.Background(), tc.key, []byte("value"), 0)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Set() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestAdapterKeyTooLong(t *testing.T) {
	adapter := NewAdapter(&stubClient{data: make(map[string][]byte)}, nil)
	adapter.MaxKeyLength = 10

	longKey := "this_is_a_very_long_key"
	err := adapter.Set(context.Background(), longKey, []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for key too long")
	}
	if !errors.Is(err, cache.ErrKeyTooLong) {
		t.Fatalf("expected ErrKeyTooLong, got %v", err)
	}
}

func TestAdapterIncr(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test increment on non-existent key
	val1, err := adapter.Incr(context.Background(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != 5 {
		t.Fatalf("expected 5, got %d", val1)
	}

	// Test increment on existing key
	val2, err := adapter.Incr(context.Background(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != 8 {
		t.Fatalf("expected 8, got %d", val2)
	}

	// Test increment by negative number
	val3, err := adapter.Incr(context.Background(), "counter", -2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val3 != 6 {
		t.Fatalf("expected 6, got %d", val3)
	}
}

func TestAdapterDecr(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test decrement on non-existent key
	val1, err := adapter.Decr(context.Background(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != -5 {
		t.Fatalf("expected -5, got %d", val1)
	}

	// Test decrement on existing key
	val2, err := adapter.Decr(context.Background(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != -8 {
		t.Fatalf("expected -8, got %d", val2)
	}
}

func TestAdapterIncrNonInteger(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	// Set a non-integer value
	err := adapter.Set(context.Background(), "key", []byte("not an integer"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to increment
	_, err = adapter.Incr(context.Background(), "key", 1)
	if !errors.Is(err, cache.ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestAdapterAppend(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test append on non-existent key
	err := adapter.Append(context.Background(), "key", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := adapter.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}

	// Test append on existing key
	err = adapter.Append(context.Background(), "key", []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err = adapter.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val)
	}
}

func TestAdapterDelete(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	// Set a value
	err := adapter.Set(context.Background(), "key", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete it
	err = adapter.Delete(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's gone
	exists, err := adapter.Exists(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected key to be deleted")
	}
}
