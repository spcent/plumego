// Package kvstore provides a small embedded persistent key-value primitive.
//
// The stable surface intentionally stays narrow:
//   - atomic file-backed persistence for small datasets
//   - TTL-aware get/set/delete operations
//   - key enumeration and basic runtime stats
//
// Durable-engine tuning such as WAL, snapshots, serializer selection,
// compression, and shard configuration lives in x/data/kvengine.
package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
	ErrStoreClosed = errors.New("store is closed")
)

const (
	defaultMaxEntries  = 100000
	defaultMaxMemoryMB = 200
	stateFileName      = "store.json"
)

// Options configures the stable embedded KV primitive.
type Options struct {
	DataDir     string `json:"data_dir"`
	MaxEntries  int    `json:"max_entries"`
	MaxMemoryMB int    `json:"max_memory_mb"`
}

type entry struct {
	Value     []byte    `json:"value"`
	ExpireAt  time.Time `json:"expire_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	Size      int64     `json:"size"`
}

type diskState struct {
	Entries map[string]entry `json:"entries"`
}

// KVStore is a small embedded persistent key-value store.
type KVStore struct {
	mu     sync.RWMutex
	opts   Options
	data   map[string]*entry
	closed bool

	hits   int64
	misses int64
}

// Stats provides runtime statistics.
type Stats struct {
	Entries     int64   `json:"entries"`
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	MemoryUsage int64   `json:"memory_usage"`
	WALSize     int64   `json:"wal_size"`
	HitRatio    float64 `json:"hit_ratio"`
}

// NewKVStore creates a new embedded KV primitive backed by a single state file.
func NewKVStore(opts Options) (*KVStore, error) {
	setDefaults(&opts)
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	store := &KVStore{
		opts: opts,
		data: make(map[string]*entry),
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	store.pruneExpiredLocked(time.Now())
	if err := store.persistLocked(); err != nil {
		return nil, err
	}
	return store, nil
}

func setDefaults(opts *Options) {
	if opts.DataDir == "" {
		opts.DataDir = "data"
	}
	if opts.MaxEntries == 0 {
		opts.MaxEntries = defaultMaxEntries
	}
	if opts.MaxMemoryMB == 0 {
		opts.MaxMemoryMB = defaultMaxMemoryMB
	}
}

func validateOptions(opts Options) error {
	if opts.MaxEntries <= 0 {
		return errors.New("max entries must be positive")
	}
	if opts.MaxMemoryMB <= 0 {
		return errors.New("max memory must be positive")
	}
	return nil
}

// Set stores a value with an optional TTL.
func (kv *KVStore) Set(key string, value []byte, ttl time.Duration) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return ErrStoreClosed
	}

	now := time.Now()
	var expireAt time.Time
	if ttl > 0 {
		expireAt = now.Add(ttl)
	}
	size := int64(len(key) + len(value) + 64)
	kv.data[key] = &entry{
		Value:     append([]byte(nil), value...),
		ExpireAt:  expireAt,
		UpdatedAt: now,
		Size:      size,
	}
	kv.evictIfNeededLocked()
	return kv.persistLocked()
}

// Get returns a defensive copy of a value.
func (kv *KVStore) Get(key string) ([]byte, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return nil, ErrStoreClosed
	}

	item, ok := kv.data[key]
	if !ok {
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyNotFound
	}
	if kv.isExpired(item, time.Now()) {
		delete(kv.data, key)
		atomic.AddInt64(&kv.misses, 1)
		if err := kv.persistLocked(); err != nil {
			return nil, err
		}
		return nil, ErrKeyExpired
	}

	atomic.AddInt64(&kv.hits, 1)
	return append([]byte(nil), item.Value...), nil
}

// Delete removes a key.
func (kv *KVStore) Delete(key string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return ErrStoreClosed
	}
	if _, ok := kv.data[key]; !ok {
		return ErrKeyNotFound
	}
	delete(kv.data, key)
	return kv.persistLocked()
}

// Exists reports whether a non-expired key exists.
func (kv *KVStore) Exists(key string) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return false
	}
	item, ok := kv.data[key]
	if !ok {
		return false
	}
	if kv.isExpired(item, time.Now()) {
		delete(kv.data, key)
		_ = kv.persistLocked()
		return false
	}
	return true
}

// Keys returns all non-expired keys in sorted order.
func (kv *KVStore) Keys() []string {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return nil
	}

	now := time.Now()
	keys := make([]string, 0, len(kv.data))
	dirty := false
	for key, item := range kv.data {
		if kv.isExpired(item, now) {
			delete(kv.data, key)
			dirty = true
			continue
		}
		keys = append(keys, key)
	}
	if dirty {
		_ = kv.persistLocked()
	}
	sort.Strings(keys)
	return keys
}

// Size returns the number of non-expired keys.
func (kv *KVStore) Size() int {
	return len(kv.Keys())
}

// GetStats returns point-in-time statistics for the store.
func (kv *KVStore) GetStats() Stats {
	hits := atomic.LoadInt64(&kv.hits)
	misses := atomic.LoadInt64(&kv.misses)
	entries, memoryUsage := kv.currentUsage()

	var hitRatio float64
	if hits+misses > 0 {
		hitRatio = float64(hits) / float64(hits+misses)
	}

	return Stats{
		Entries:     int64(entries),
		Hits:        hits,
		Misses:      misses,
		MemoryUsage: memoryUsage,
		HitRatio:    hitRatio,
	}
}

// Close closes the store.
func (kv *KVStore) Close() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return ErrStoreClosed
	}
	kv.closed = true
	return nil
}

func (kv *KVStore) currentUsage() (int, int64) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	now := time.Now()
	count := 0
	var memoryUsage int64
	for _, item := range kv.data {
		if kv.isExpired(item, now) {
			continue
		}
		count++
		memoryUsage += item.Size
	}
	return count, memoryUsage
}

func (kv *KVStore) load() error {
	path := filepath.Join(kv.opts.DataDir, stateFileName)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	var state diskState
	if err := json.Unmarshal(raw, &state); err != nil {
		return fmt.Errorf("decode state: %w", err)
	}
	for key, item := range state.Entries {
		itemCopy := item
		itemCopy.Value = append([]byte(nil), item.Value...)
		kv.data[key] = &itemCopy
	}
	return nil
}

func (kv *KVStore) persistLocked() error {
	state := diskState{
		Entries: make(map[string]entry, len(kv.data)),
	}
	for key, item := range kv.data {
		state.Entries[key] = entry{
			Value:     append([]byte(nil), item.Value...),
			ExpireAt:  item.ExpireAt,
			UpdatedAt: item.UpdatedAt,
			Size:      item.Size,
		}
	}

	raw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	path := filepath.Join(kv.opts.DataDir, stateFileName)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0644); err != nil {
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	return nil
}

func (kv *KVStore) pruneExpiredLocked(now time.Time) {
	for key, item := range kv.data {
		if kv.isExpired(item, now) {
			delete(kv.data, key)
		}
	}
}

func (kv *KVStore) evictIfNeededLocked() {
	kv.pruneExpiredLocked(time.Now())
	maxMemory := int64(kv.opts.MaxMemoryMB) * 1024 * 1024
	for len(kv.data) > kv.opts.MaxEntries || kv.memoryUsageLocked() > maxMemory {
		key := kv.oldestKeyLocked()
		if key == "" {
			return
		}
		delete(kv.data, key)
	}
}

func (kv *KVStore) oldestKeyLocked() string {
	var (
		oldestKey string
		oldestAt  time.Time
	)
	for key, item := range kv.data {
		if oldestKey == "" || item.UpdatedAt.Before(oldestAt) {
			oldestKey = key
			oldestAt = item.UpdatedAt
		}
	}
	return oldestKey
}

func (kv *KVStore) memoryUsageLocked() int64 {
	var total int64
	now := time.Now()
	for _, item := range kv.data {
		if kv.isExpired(item, now) {
			continue
		}
		total += item.Size
	}
	return total
}

func (kv *KVStore) isExpired(item *entry, now time.Time) bool {
	return !item.ExpireAt.IsZero() && !item.ExpireAt.After(now)
}
