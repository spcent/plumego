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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrKeyNotFound   = errors.New("kv: key not found")
	ErrKeyExpired    = errors.New("kv: key expired")
	ErrInvalidKey    = errors.New("kv: key is required")
	ErrStoreClosed   = errors.New("kv: store is closed")
	ErrValueTooLarge = errors.New("kv: value too large")
)

const (
	defaultMaxEntries  = 100000
	defaultMaxMemoryMB = 200
	stateFileName      = "store.json"
)

// Options configures the stable embedded KV primitive.
type Options struct {
	// DataDir is the explicit directory where the state file is stored.
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
	MemoryUsage int64   `json:"memory_usage"`
	HitRatio    float64 `json:"hit_ratio"`
}

// NewKVStore creates a new embedded KV primitive backed by a single state file.
//
// DataDir must be explicitly configured so callers choose where filesystem
// state is created.
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
	store.evictIfNeededLocked()
	return store, nil
}

func setDefaults(opts *Options) {
	opts.DataDir = strings.TrimSpace(opts.DataDir)
	if opts.MaxEntries == 0 {
		opts.MaxEntries = defaultMaxEntries
	}
	if opts.MaxMemoryMB == 0 {
		opts.MaxMemoryMB = defaultMaxMemoryMB
	}
}

func validateOptions(opts Options) error {
	if strings.TrimSpace(opts.DataDir) == "" {
		return errors.New("kv: data dir is required")
	}
	if opts.MaxEntries <= 0 {
		return errors.New("kv: max entries must be positive")
	}
	if opts.MaxMemoryMB <= 0 {
		return errors.New("kv: max memory must be positive")
	}
	return nil
}

// Set stores a value with an optional TTL.
func (kv *KVStore) Set(key string, value []byte, ttl time.Duration) error {
	return kv.SetContext(context.Background(), key, value, ttl)
}

// SetContext stores a value with an optional TTL using the caller-provided context.
func (kv *KVStore) SetContext(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if kv == nil {
		return ErrStoreClosed
	}
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if err := contextErr(ctx); err != nil {
		return err
	}
	if kv.closed || kv.data == nil {
		return ErrStoreClosed
	}
	if err := validateKey(key); err != nil {
		return err
	}

	now := time.Now()
	var expireAt time.Time
	if ttl > 0 {
		expireAt = now.Add(ttl)
	}
	size := entrySize(key, value)
	if size > kv.maxMemoryBytes() {
		return fmt.Errorf("%w: size %d limit %d", ErrValueTooLarge, size, kv.maxMemoryBytes())
	}
	before := kv.cloneDataLocked()
	kv.data[key] = &entry{
		Value:     cloneBytes(value),
		ExpireAt:  expireAt,
		UpdatedAt: now,
		Size:      size,
	}
	kv.evictIfNeededLocked()
	if err := contextErr(ctx); err != nil {
		kv.data = before
		return err
	}
	if err := kv.persistLocked(); err != nil {
		kv.data = before
		return err
	}
	return nil
}

// Get returns a defensive copy of a value.
func (kv *KVStore) Get(key string) ([]byte, error) {
	return kv.GetContext(context.Background(), key)
}

// GetContext returns a defensive copy of a value using the caller-provided context.
func (kv *KVStore) GetContext(ctx context.Context, key string) ([]byte, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, ErrStoreClosed
	}
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if kv.closed || kv.data == nil {
		return nil, ErrStoreClosed
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}

	item, ok := kv.data[key]
	if !ok {
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyNotFound
	}
	if kv.isExpired(item, time.Now()) {
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyExpired
	}

	atomic.AddInt64(&kv.hits, 1)
	return cloneBytes(item.Value), nil
}

// Delete removes a key.
func (kv *KVStore) Delete(key string) error {
	return kv.DeleteContext(context.Background(), key)
}

// DeleteContext removes a key using the caller-provided context.
func (kv *KVStore) DeleteContext(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if kv == nil {
		return ErrStoreClosed
	}
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if err := contextErr(ctx); err != nil {
		return err
	}
	if kv.closed || kv.data == nil {
		return ErrStoreClosed
	}
	if err := validateKey(key); err != nil {
		return err
	}
	if _, ok := kv.data[key]; !ok {
		return ErrKeyNotFound
	}
	before := kv.cloneDataLocked()
	delete(kv.data, key)
	if err := contextErr(ctx); err != nil {
		kv.data = before
		return err
	}
	if err := kv.persistLocked(); err != nil {
		kv.data = before
		return err
	}
	return nil
}

// Exists reports whether a non-expired key exists. Errors are collapsed to
// false for compatibility; use ExistsContext when the caller needs errors.
func (kv *KVStore) Exists(key string) bool {
	exists, _ := kv.ExistsContext(context.Background(), key)
	return exists
}

// ExistsContext reports whether a non-expired key exists using the caller-provided context.
func (kv *KVStore) ExistsContext(ctx context.Context, key string) (bool, error) {
	if err := contextErr(ctx); err != nil {
		return false, err
	}
	if kv == nil {
		return false, ErrStoreClosed
	}
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if err := contextErr(ctx); err != nil {
		return false, err
	}
	if kv.closed || kv.data == nil {
		return false, ErrStoreClosed
	}
	if err := validateKey(key); err != nil {
		return false, err
	}
	item, ok := kv.data[key]
	if !ok {
		return false, nil
	}
	if kv.isExpired(item, time.Now()) {
		return false, nil
	}
	return true, nil
}

// Keys returns all non-expired keys in sorted order. Errors are collapsed to an
// empty slice for compatibility; use KeysContext when the caller needs errors.
func (kv *KVStore) Keys() []string {
	keys, _ := kv.KeysContext(context.Background())
	return keys
}

// KeysContext returns all non-expired keys in sorted order using the caller-provided context.
func (kv *KVStore) KeysContext(ctx context.Context) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, ErrStoreClosed
	}
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if kv.closed || kv.data == nil {
		return nil, ErrStoreClosed
	}

	now := time.Now()
	keys := make([]string, 0, len(kv.data))
	for key, item := range kv.data {
		if kv.isExpired(item, now) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys, nil
}

// Size returns the number of non-expired keys. Errors are collapsed to zero for
// compatibility; use SizeContext when the caller needs errors.
func (kv *KVStore) Size() int {
	size, _ := kv.SizeContext(context.Background())
	return size
}

// SizeContext returns the number of non-expired keys using the caller-provided context.
func (kv *KVStore) SizeContext(ctx context.Context) (int, error) {
	keys, err := kv.KeysContext(ctx)
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

// GetStats returns point-in-time statistics for the store.
func (kv *KVStore) GetStats() Stats {
	stats, _ := kv.GetStatsContext(context.Background())
	return stats
}

// GetStatsContext returns point-in-time statistics using the caller-provided context.
func (kv *KVStore) GetStatsContext(ctx context.Context) (Stats, error) {
	if err := contextErr(ctx); err != nil {
		return Stats{}, err
	}
	if kv == nil {
		return Stats{}, ErrStoreClosed
	}
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if err := contextErr(ctx); err != nil {
		return Stats{}, err
	}
	if kv.closed || kv.data == nil {
		return Stats{}, ErrStoreClosed
	}

	hits := atomic.LoadInt64(&kv.hits)
	misses := atomic.LoadInt64(&kv.misses)
	entries, memoryUsage := kv.currentUsageLocked(time.Now())

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
	}, nil
}

func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c < 0x20 || c == 0x7F {
			return fmt.Errorf("%w: control character at position %d", ErrInvalidKey, i)
		}
	}
	return nil
}

// Close closes the store.
func (kv *KVStore) Close() error {
	if kv == nil {
		return nil
	}
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.closed {
		return nil
	}
	kv.closed = true
	return nil
}

func (kv *KVStore) currentUsage() (int, int64) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	return kv.currentUsageLocked(time.Now())
}

func (kv *KVStore) currentUsageLocked(now time.Time) (int, int64) {
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
		if err := validateKey(key); err != nil {
			return fmt.Errorf("decode state key %q: %w", key, err)
		}
		itemCopy := item
		itemCopy.Value = cloneBytes(item.Value)
		itemCopy.Size = entrySize(key, item.Value)
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
			Value:     cloneBytes(item.Value),
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
	tmp, err := os.CreateTemp(kv.opts.DataDir, stateFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp state: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp state: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	if err := syncDir(kv.opts.DataDir); err != nil {
		return fmt.Errorf("sync state dir: %w", err)
	}
	committed = true
	return nil
}

func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Sync(); err != nil {
		if errors.Is(err, os.ErrInvalid) {
			return nil
		}
		return err
	}
	return nil
}

func (kv *KVStore) cloneDataLocked() map[string]*entry {
	cloned := make(map[string]*entry, len(kv.data))
	for key, item := range kv.data {
		itemCopy := *item
		itemCopy.Value = cloneBytes(item.Value)
		cloned[key] = &itemCopy
	}
	return cloned
}

func cloneBytes(in []byte) []byte {
	if in == nil {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
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

func (kv *KVStore) maxMemoryBytes() int64 {
	return int64(kv.opts.MaxMemoryMB) * 1024 * 1024
}

func entrySize(key string, value []byte) int64 {
	return int64(len(key) + len(value) + 64)
}

func (kv *KVStore) isExpired(item *entry, now time.Time) bool {
	return !item.ExpireAt.IsZero() && !item.ExpireAt.After(now)
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
