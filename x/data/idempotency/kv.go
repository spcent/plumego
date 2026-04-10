package idempotency

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

type KVConfig struct {
	Prefix string
	Now    func() time.Time
}

func DefaultKVConfig() KVConfig {
	return KVConfig{
		Prefix: "idem:",
		Now:    time.Now,
	}
}

type KVStore struct {
	store  *kvstore.KVStore
	prefix string
	now    func() time.Time
}

func NewKVStore(store *kvstore.KVStore, cfg KVConfig) *KVStore {
	if cfg.Prefix == "" {
		cfg.Prefix = "idem:"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &KVStore{
		store:  store,
		prefix: cfg.Prefix,
		now:    cfg.Now,
	}
}

func (s *KVStore) Get(_ context.Context, key string) (Record, bool, error) {
	if s == nil || s.store == nil {
		return Record{}, false, ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return Record{}, false, ErrInvalidKey
	}

	data, err := s.store.Get(s.key(key))
	if err != nil {
		if err == kvstore.ErrKeyNotFound || err == kvstore.ErrKeyExpired {
			return Record{}, false, nil
		}
		return Record{}, false, err
	}

	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, false, err
	}

	if s.isExpired(record) {
		_ = s.store.Delete(s.key(key))
		return Record{}, false, nil
	}

	return record, true, nil
}

func (s *KVStore) PutIfAbsent(ctx context.Context, record Record) (bool, error) {
	if s == nil || s.store == nil {
		return false, ErrNotFound
	}
	if strings.TrimSpace(record.Key) == "" {
		return false, ErrInvalidKey
	}

	if s.isExpired(record) {
		return false, ErrExpired
	}

	if existing, found, err := s.Get(ctx, record.Key); err != nil {
		return false, err
	} else if found {
		_ = existing
		return false, nil
	}

	now := s.now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	if record.Status == "" {
		record.Status = StatusInProgress
	}

	data, err := json.Marshal(record)
	if err != nil {
		return false, err
	}

	return true, s.store.Set(s.key(record.Key), data, s.ttl(record))
}

func (s *KVStore) Complete(ctx context.Context, key string, response []byte) error {
	if s == nil || s.store == nil {
		return ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}

	record, found, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}

	record.Status = StatusCompleted
	record.Response = response
	record.UpdatedAt = s.now()

	if s.isExpired(record) {
		_ = s.store.Delete(s.key(key))
		return ErrExpired
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return s.store.Set(s.key(key), data, s.ttl(record))
}

func (s *KVStore) Delete(_ context.Context, key string) error {
	if s == nil || s.store == nil {
		return ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	return s.store.Delete(s.key(key))
}

func (s *KVStore) key(key string) string {
	return s.prefix + key
}

func (s *KVStore) isExpired(record Record) bool {
	if record.ExpiresAt.IsZero() {
		return false
	}
	return !record.ExpiresAt.After(s.now())
}

func (s *KVStore) ttl(record Record) time.Duration {
	if record.ExpiresAt.IsZero() {
		return 0
	}
	ttl := record.ExpiresAt.Sub(s.now())
	if ttl < 0 {
		return 0
	}
	return ttl
}
