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
