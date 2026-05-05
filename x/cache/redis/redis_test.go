package redis

import (
	"context"
	"errors"
	"math"
	"strconv"
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

type stubAtomicClient struct {
	stubClient
}

func (s *stubAtomicClient) IncrBy(ctx context.Context, key string, delta int64) (int64, error) {
	var current int64
	if data, err := s.Get(ctx, key); err == nil && len(data) > 0 {
		num, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return 0, cache.ErrNotInteger
		}
		current = num
	} else if err != nil && !errors.Is(err, errMiss) {
		return 0, err
	}

	next, err := addTestInt64(current, delta)
	if err != nil {
		return 0, err
	}
	if err := s.Set(ctx, key, []byte(strconv.FormatInt(next, 10)), 0); err != nil {
		return 0, err
	}
	return next, nil
}

func (s *stubAtomicClient) Append(ctx context.Context, key string, data []byte) (int64, error) {
	existing, err := s.Get(ctx, key)
	if err != nil && !errors.Is(err, errMiss) {
		return 0, err
	}
	value := append(append([]byte(nil), existing...), data...)
	if err := s.Set(ctx, key, value, 0); err != nil {
		return 0, err
	}
	return int64(len(value)), nil
}

func addTestInt64(value, delta int64) (int64, error) {
	if delta > 0 && value > math.MaxInt64-delta {
		return 0, cache.ErrNotInteger
	}
	if delta < 0 && value < math.MinInt64-delta {
		return 0, cache.ErrNotInteger
	}
	return value + delta, nil
}

func TestAdapterGetNotFound(t *testing.T) {
	adapter := NewAdapter(&stubClient{}, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	_, err := adapter.Get(t.Context(), "missing")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapterSetAndExists(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	if err := adapter.Set(t.Context(), "key", []byte("value"), time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, err := adapter.Exists(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("expected key to exist")
	}
}

func TestAdapterCopiesSetAndGetValues(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	value := []byte("value")
	if err := adapter.Set(t.Context(), "key", value, time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	value[0] = 'X'

	got, err := adapter.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "value" {
		t.Fatalf("Get after mutating input = %q, want value", string(got))
	}

	got[0] = 'Y'
	got, err = adapter.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if string(got) != "value" {
		t.Fatalf("Get after mutating returned slice = %q, want value", string(got))
	}
}

func TestAdapterClear(t *testing.T) {
	client := &stubFlusher{stubClient: stubClient{data: map[string][]byte{"k": []byte("v")}}}
	adapter := NewAdapter(client, nil)

	if err := adapter.Clear(t.Context()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.flushed {
		t.Fatalf("expected flush to be called")
	}
}

func TestAdapterClearUnsupported(t *testing.T) {
	adapter := NewAdapter(&stubClient{}, nil)

	if err := adapter.Clear(t.Context()); !errors.Is(err, ErrClearUnsupported) {
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
			err := adapter.Set(t.Context(), tc.key, []byte("value"), 0)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Set() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, cache.ErrInvalidKey) {
				t.Fatalf("Set() error = %v, want ErrInvalidKey", err)
			}
		})
	}
}

func TestAdapterKeyTooLong(t *testing.T) {
	adapter := NewAdapter(&stubClient{data: make(map[string][]byte)}, nil)
	adapter.MaxKeyLength = 10

	longKey := "this_is_a_very_long_key"
	err := adapter.Set(t.Context(), longKey, []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for key too long")
	}
	if !errors.Is(err, cache.ErrKeyTooLong) {
		t.Fatalf("expected ErrKeyTooLong, got %v", err)
	}
}

func TestAdapterIncr(t *testing.T) {
	client := &stubAtomicClient{stubClient: stubClient{data: make(map[string][]byte)}}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test increment on non-existent key
	val1, err := adapter.Incr(t.Context(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != 5 {
		t.Fatalf("expected 5, got %d", val1)
	}
	if got := string(client.data["counter"]); got != "5" {
		t.Fatalf("stored counter = %q, want decimal text 5", got)
	}

	// Test increment on existing key
	val2, err := adapter.Incr(t.Context(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != 8 {
		t.Fatalf("expected 8, got %d", val2)
	}

	// Test increment by negative number
	val3, err := adapter.Incr(t.Context(), "counter", -2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val3 != 6 {
		t.Fatalf("expected 6, got %d", val3)
	}
}

func TestAdapterIncrOverflow(t *testing.T) {
	client := &stubAtomicClient{stubClient: stubClient{data: make(map[string][]byte)}}
	adapter := NewAdapter(client, nil)

	if err := adapter.Set(t.Context(), "counter", []byte("9223372036854775807"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := adapter.Incr(t.Context(), "counter", 1); !errors.Is(err, cache.ErrNotInteger) {
		t.Fatalf("Incr overflow error = %v, want ErrNotInteger", err)
	}
}

func TestAdapterDecr(t *testing.T) {
	client := &stubAtomicClient{stubClient: stubClient{data: make(map[string][]byte)}}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test decrement on non-existent key
	val1, err := adapter.Decr(t.Context(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != -5 {
		t.Fatalf("expected -5, got %d", val1)
	}

	// Test decrement on existing key
	val2, err := adapter.Decr(t.Context(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != -8 {
		t.Fatalf("expected -8, got %d", val2)
	}
}

func TestAdapterDecrRejectsMinDelta(t *testing.T) {
	adapter := NewAdapter(&stubClient{data: make(map[string][]byte)}, nil)

	if _, err := adapter.Decr(t.Context(), "counter", math.MinInt64); !errors.Is(err, cache.ErrNotInteger) {
		t.Fatalf("Decr min delta error = %v, want ErrNotInteger", err)
	}
}

func TestAdapterIncrNonInteger(t *testing.T) {
	client := &stubAtomicClient{stubClient: stubClient{data: make(map[string][]byte)}}
	adapter := NewAdapter(client, nil)

	// Set a non-integer value
	err := adapter.Set(t.Context(), "key", []byte("not an integer"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to increment
	_, err = adapter.Incr(t.Context(), "key", 1)
	if !errors.Is(err, cache.ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestAdapterAppend(t *testing.T) {
	client := &stubAtomicClient{stubClient: stubClient{data: make(map[string][]byte)}}
	adapter := NewAdapter(client, func(err error) bool {
		return errors.Is(err, errMiss)
	})

	// Test append on non-existent key
	err := adapter.Append(t.Context(), "key", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := adapter.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}

	// Test append on existing key
	err = adapter.Append(t.Context(), "key", []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err = adapter.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val)
	}
}

func TestAdapterAtomicCapabilitiesUnsupported(t *testing.T) {
	adapter := NewAdapter(&stubClient{data: make(map[string][]byte)}, nil)

	if _, err := adapter.Incr(t.Context(), "counter", 1); !errors.Is(err, cache.ErrCapabilityUnsupported) {
		t.Fatalf("Incr error = %v, want ErrCapabilityUnsupported", err)
	}
	if _, err := adapter.Decr(t.Context(), "counter", 1); !errors.Is(err, cache.ErrCapabilityUnsupported) {
		t.Fatalf("Decr error = %v, want ErrCapabilityUnsupported", err)
	}
	if err := adapter.Append(t.Context(), "key", []byte("value")); !errors.Is(err, cache.ErrCapabilityUnsupported) {
		t.Fatalf("Append error = %v, want ErrCapabilityUnsupported", err)
	}
}

func TestAdapterDelete(t *testing.T) {
	client := &stubClient{data: make(map[string][]byte)}
	adapter := NewAdapter(client, nil)

	// Set a value
	err := adapter.Set(t.Context(), "key", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete it
	err = adapter.Delete(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's gone
	exists, err := adapter.Exists(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected key to be deleted")
	}
}
