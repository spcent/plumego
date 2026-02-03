// Package kvstore provides an embedded persistent key-value store with WAL.
//
// This package implements a high-performance, disk-backed key-value store featuring:
//   - Write-Ahead Logging (WAL) for durability
//   - LRU eviction with configurable memory limits
//   - TTL (time-to-live) support for automatic expiration
//   - Snapshot and restore capabilities
//   - Transaction support for atomic operations
//   - Compression (gzip) for reduced disk usage
//   - Metrics collection and monitoring
//
// The store is optimized for embedded use cases where you need persistence
// without the complexity of a separate database server. It's ideal for
// configuration storage, session management, and small-to-medium datasets.
//
// Example usage:
//
//	import kvstore "github.com/spcent/plumego/store/kv"
//
//	// Create or open a store
//	store, err := kvstore.Open(kvstore.Config{
//		Path:         "/data/mystore",
//		MaxMemoryMB:  512,
//		SyncWrites:   true,
//		EnableMetrics: true,
//	})
//	defer store.Close(context.Background())
//
//	// Set a key with 1-hour TTL
//	err = store.Set("session:abc", sessionData, 1*time.Hour)
//
//	// Get a key
//	val, err := store.Get("session:abc")
package kvstore

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/metrics"
)

var (
	ErrKeyExists          = errors.New("key already exists")
	ErrKeyNotFound        = errors.New("key not found")
	ErrKeyExpired         = errors.New("key expired")
	ErrStoreClosed        = errors.New("store is closed")
	ErrInvalidEntry       = errors.New("invalid WAL entry")
	ErrCloseTimeout       = errors.New("close operation timed out")
	ErrTransactionAborted = errors.New("transaction aborted")
)

const (
	opSet    byte = 1
	opDelete byte = 2

	defaultMaxEntries    = 100000
	defaultMaxMemoryMB   = 200
	defaultFlushInterval = 100 * time.Millisecond
	defaultCleanInterval = 30 * time.Second
	defaultShardCount    = 16
	defaultCloseTimeout  = 10 * time.Second

	magicNumber uint32 = 0x4B565354 // "KVST"
	version     uint32 = 1
)

// Entry represents a key-value pair with metadata
type Entry struct {
	Key      string    `json:"key"`
	Value    []byte    `json:"value"`
	ExpireAt time.Time `json:"expire_at"`
	Size     int64     `json:"size"`
	Version  int64     `json:"version"`
	// LRU chain pointers (not serialized)
	Prev *Entry `json:"-"`
	Next *Entry `json:"-"`
}

// WALEntry represents a Write-Ahead Log entry
type WALEntry struct {
	Op       byte      `json:"op"`
	Key      string    `json:"key"`
	Value    []byte    `json:"value,omitempty"`
	ExpireAt time.Time `json:"expire_at,omitempty"`
	Version  int64     `json:"version"`
	CRC      uint32    `json:"crc"`
}

// Options configures the KV store
type Options struct {
	DataDir           string              `json:"data_dir"`
	MaxEntries        int                 `json:"max_entries"`
	MaxMemoryMB       int                 `json:"max_memory_mb"`
	FlushInterval     time.Duration       `json:"flush_interval"`
	CleanInterval     time.Duration       `json:"clean_interval"`
	ShardCount        int                 `json:"shard_count"`
	EnableCompression bool                `json:"enable_compression"`
	ReadOnly          bool                `json:"read_only"`
	CloseTimeout      time.Duration       `json:"close_timeout"`
	SerializerFormat  SerializationFormat `json:"serializer_format"`  // Serialization format (binary/json)
	AutoDetectFormat  bool                `json:"auto_detect_format"` // Auto-detect format when loading
}

// Shard represents a single data shard with optimized locking
type Shard struct {
	mu      sync.RWMutex
	data    map[string]*Entry
	lruHead *Entry
	lruTail *Entry
	next    map[*Entry]*Entry // LRU next pointers
	prev    map[*Entry]*Entry // LRU prev pointers
}

// KVStore is a simplified, high-performance key-value store
type KVStore struct {
	// Configuration
	opts       Options
	serializer Serializer // Serialization strategy

	// Sharded data
	shards    []*Shard
	shardMask uint32

	// WAL
	walFile   *os.File
	walWriter *bufio.Writer
	walMutex  sync.Mutex

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Statistics (atomic)
	hits        int64
	misses      int64
	evictions   int64
	entries     int64
	memoryUsage int64
	walSize     int64
	version     int64

	// State
	closed int32

	// Unified metrics collector
	collector metrics.MetricsCollector
}

// Stats provides runtime statistics
type Stats struct {
	Entries     int64   `json:"entries"`
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	MemoryUsage int64   `json:"memory_usage"`
	WALSize     int64   `json:"wal_size"`
	HitRatio    float64 `json:"hit_ratio"`
}

// NewKVStore creates a new KV store
func NewKVStore(opts Options) (*KVStore, error) {
	if err := setDefaults(&opts); err != nil {
		return nil, err
	}

	if err := validateOptions(&opts); err != nil {
		return nil, err
	}

	// Create data directory
	if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize shards
	shards := make([]*Shard, opts.ShardCount)
	for i := 0; i < opts.ShardCount; i++ {
		shards[i] = &Shard{
			data: make(map[string]*Entry),
			next: make(map[*Entry]*Entry),
			prev: make(map[*Entry]*Entry),
		}
	}

	kv := &KVStore{
		opts:       opts,
		shards:     shards,
		shardMask:  uint32(opts.ShardCount - 1),
		serializer: GetSerializer(opts.SerializerFormat),
	}

	kv.ctx, kv.cancel = context.WithCancel(context.Background())

	// Initialize WAL
	if !opts.ReadOnly {
		if err := kv.initWAL(); err != nil {
			return nil, fmt.Errorf("failed to initialize WAL: %w", err)
		}

		// Start background workers
		kv.wg.Add(2)
		go kv.walFlusher()
		go kv.cleaner()
	}

	// Load existing data
	if err := kv.loadData(); err != nil {
		kv.Close()
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return kv, nil
}

func setDefaults(opts *Options) error {
	if opts.DataDir == "" {
		opts.DataDir = "data"
	}
	if opts.MaxEntries == 0 {
		opts.MaxEntries = defaultMaxEntries
	}
	if opts.MaxMemoryMB == 0 {
		opts.MaxMemoryMB = defaultMaxMemoryMB
	}
	if opts.FlushInterval == 0 {
		opts.FlushInterval = defaultFlushInterval
	}
	if opts.CleanInterval == 0 {
		opts.CleanInterval = defaultCleanInterval
	}
	if opts.ShardCount == 0 {
		opts.ShardCount = defaultShardCount
	}
	if opts.CloseTimeout == 0 {
		opts.CloseTimeout = defaultCloseTimeout
	}
	if opts.SerializerFormat == "" {
		// Default to binary for best performance
		opts.SerializerFormat = FormatBinary
	}
	// Auto-detect enabled by default for backward compatibility
	if !opts.AutoDetectFormat {
		opts.AutoDetectFormat = true
	}
	return nil
}

func validateOptions(opts *Options) error {
	if opts.MaxEntries <= 0 {
		return errors.New("max entries must be positive")
	}
	if opts.MaxMemoryMB <= 0 {
		return errors.New("max memory must be positive")
	}
	if opts.ShardCount <= 0 || (opts.ShardCount&(opts.ShardCount-1)) != 0 {
		return errors.New("shard count must be power of 2")
	}
	return nil
}

// getShard returns the shard for a given key
func (kv *KVStore) getShard(key string) *Shard {
	hash := crc32.ChecksumIEEE([]byte(key))
	return kv.shards[hash&kv.shardMask]
}

// initWAL initializes the Write-Ahead Log
func (kv *KVStore) initWAL() error {
	walPath := filepath.Join(kv.opts.DataDir, "store.wal")

	file, err := os.OpenFile(walPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL: %w", err)
	}

	kv.walFile = file
	kv.walWriter = bufio.NewWriter(file)
	return nil
}

// walFlusher handles periodic WAL flushing to disk
func (kv *KVStore) walFlusher() {
	defer kv.wg.Done()

	ticker := time.NewTicker(kv.opts.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kv.flushWAL()

		case <-kv.ctx.Done():
			kv.flushWAL()
			return
		}
	}
}

func (kv *KVStore) writeWALEntry(entry WALEntry) error {
	kv.walMutex.Lock()
	defer kv.walMutex.Unlock()

	data, err := kv.encodeWALEntry(entry)
	if err != nil {
		return err
	}

	n, err := kv.walWriter.Write(data)
	if err != nil {
		return err
	}

	atomic.AddInt64(&kv.walSize, int64(n))
	return nil
}

func (kv *KVStore) flushWAL() {
	kv.walMutex.Lock()
	defer kv.walMutex.Unlock()

	if kv.walWriter != nil {
		kv.walWriter.Flush()
		kv.walFile.Sync()
	}
}

// cleaner handles TTL cleanup
func (kv *KVStore) cleaner() {
	defer kv.wg.Done()

	ticker := time.NewTicker(kv.opts.CleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kv.cleanExpired()
		case <-kv.ctx.Done():
			return
		}
	}
}

func (kv *KVStore) cleanExpired() {
	now := time.Now()
	// Sample size per shard per cycle
	// We use a randomized sampling strategy similar to Redis to avoid O(N) scans
	const sampleSize = 20

	for _, shard := range kv.shards {
		// Loop until the percentage of expired keys is low
		for {
			var expired []string
			checkedCount := 0
			expiredCount := 0

			shard.mu.RLock()
			for key, entry := range shard.data {
				if !entry.ExpireAt.IsZero() {
					if now.After(entry.ExpireAt) {
						expired = append(expired, key)
						expiredCount++
					}
					checkedCount++
					if checkedCount >= sampleSize {
						break
					}
				}
			}
			shard.mu.RUnlock()

			if len(expired) > 0 {
				shard.mu.Lock()
				for _, key := range expired {
					if entry, exists := shard.data[key]; exists {
						// Double check expiration under write lock
						if !entry.ExpireAt.IsZero() && now.After(entry.ExpireAt) {
							kv.deleteFromShard(shard, key, entry)
							// No need to log to WAL - expired entries are handled during recovery
						}
					}
				}
				shard.mu.Unlock()
			}

			// If we didn't check enough keys (shard is small or has few items with TTL), stop
			if checkedCount < sampleSize {
				break
			}

			// If less than 25% of checked keys were expired, we assume the shard is mostly clean
			// and move to the next shard.
			if float64(expiredCount)/float64(checkedCount) < 0.25 {
				break
			}

			// Add a yield to prevent starving other goroutines during heavy cleanup
			// runtime.Gosched() could be used, or just relying on the lock release/acquire cycle
		}
	}
}

// LRU management for shard
func (kv *KVStore) moveToFront(shard *Shard, entry *Entry) {
	if shard.lruHead == entry {
		return
	}

	// Remove from current position
	if prev := shard.prev[entry]; prev != nil {
		shard.next[prev] = shard.next[entry]
	}
	if next := shard.next[entry]; next != nil {
		shard.prev[next] = shard.prev[entry]
	} else {
		shard.lruTail = shard.prev[entry]
	}

	// Add to front
	shard.prev[entry] = nil
	shard.next[entry] = shard.lruHead
	if shard.lruHead != nil {
		shard.prev[shard.lruHead] = entry
	}
	shard.lruHead = entry
	if shard.lruTail == nil {
		shard.lruTail = entry
	}
}

func (kv *KVStore) evictLRU(shard *Shard) {
	if shard.lruTail == nil {
		return
	}

	entry := shard.lruTail
	kv.deleteFromShard(shard, entry.Key, entry)
	// No need to log to WAL - LRU evictions are a cache management strategy

	atomic.AddInt64(&kv.evictions, 1)
}

func (kv *KVStore) deleteFromShard(shard *Shard, key string, entry *Entry) {
	// Remove from map
	delete(shard.data, key)

	// Remove from LRU
	if prev := shard.prev[entry]; prev != nil {
		shard.next[prev] = shard.next[entry]
	} else {
		shard.lruHead = shard.next[entry]
	}
	if next := shard.next[entry]; next != nil {
		shard.prev[next] = shard.prev[entry]
	} else {
		shard.lruTail = shard.prev[entry]
	}

	delete(shard.next, entry)
	delete(shard.prev, entry)

	// Update counters
	atomic.AddInt64(&kv.entries, -1)
	atomic.AddInt64(&kv.memoryUsage, -entry.Size)
}

// Public API

func (kv *KVStore) Set(key string, value []byte, ttl time.Duration) error {
	if kv.isClosed() {
		return ErrStoreClosed
	}
	if kv.opts.ReadOnly {
		return ErrStoreClosed
	}

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	size := int64(len(key) + len(value) + 64) // rough estimate
	version := atomic.AddInt64(&kv.version, 1)

	// Write to WAL first (before updating memory)
	walEntry := WALEntry{
		Op:       opSet,
		Key:      key,
		Value:    value,
		ExpireAt: expireAt,
		Version:  version,
	}
	walEntry.CRC = kv.calculateCRC(walEntry)

	// Synchronous WAL write - fail fast if WAL fails
	if err := kv.writeWALEntry(walEntry); err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	// WAL write succeeded, now update memory
	shard := kv.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Check limits and evict if necessary
	for atomic.LoadInt64(&kv.entries) >= int64(kv.opts.MaxEntries) {
		kv.evictLRU(shard)
	}

	maxMemory := int64(kv.opts.MaxMemoryMB) * 1024 * 1024
	for atomic.LoadInt64(&kv.memoryUsage)+size > maxMemory {
		kv.evictLRU(shard)
	}

	// Create or update entry
	entry := &Entry{
		Key:      key,
		Value:    append([]byte(nil), value...), // defensive copy
		ExpireAt: expireAt,
		Size:     size,
		Version:  version,
	}

	// Remove old entry if exists
	if oldEntry, exists := shard.data[key]; exists {
		kv.deleteFromShard(shard, key, oldEntry)
	}

	// Add new entry
	shard.data[key] = entry
	kv.moveToFront(shard, entry)

	atomic.AddInt64(&kv.entries, 1)
	atomic.AddInt64(&kv.memoryUsage, size)

	return nil
}

func (kv *KVStore) Get(key string) ([]byte, error) {
	if kv.isClosed() {
		return nil, ErrStoreClosed
	}

	shard := kv.getShard(key)

	// Use read lock for the entire operation to prevent race conditions
	shard.mu.RLock()
	entry, exists := shard.data[key]

	// Check if key exists
	if !exists {
		shard.mu.RUnlock()
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyNotFound
	}

	// Check expiration
	if !entry.ExpireAt.IsZero() && time.Now().After(entry.ExpireAt) {
		shard.mu.RUnlock()

		// Upgrade to write lock to delete expired entry
		shard.mu.Lock()
		// Double-check under write lock
		if e, exists := shard.data[key]; exists {
			if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
				kv.deleteFromShard(shard, key, e)
				// No need to log to WAL - expired entries are handled during recovery
				shard.mu.Unlock()
				atomic.AddInt64(&kv.misses, 1)
				return nil, ErrKeyExpired
			}
		}
		shard.mu.Unlock()

		// Entry was refreshed by another goroutine, retry the read
		return kv.Get(key)
	}

	// Create defensive copy while holding read lock
	valueCopy := append([]byte(nil), entry.Value...)
	shard.mu.RUnlock()

	// Update LRU with minimal lock time
	// Use TryLock to avoid blocking readers if contention is high.
	// If we can't get the lock immediately, we skip the LRU update.
	// This trades LRU accuracy for read throughput under high concurrency.
	if shard.mu.TryLock() {
		// Recheck existence and expiration under write lock
		if e, exists := shard.data[key]; exists {
			if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
				shard.mu.Unlock()
				atomic.AddInt64(&kv.misses, 1)
				return nil, ErrKeyExpired
			}
			// Move to front (LRU)
			kv.moveToFront(shard, e)
		}
		shard.mu.Unlock()
	}

	atomic.AddInt64(&kv.hits, 1)
	return valueCopy, nil
}

func (kv *KVStore) Delete(key string) error {
	if kv.isClosed() {
		return ErrStoreClosed
	}
	if kv.opts.ReadOnly {
		return ErrStoreClosed
	}

	// Write to WAL first (before updating memory)
	version := atomic.AddInt64(&kv.version, 1)
	walEntry := WALEntry{
		Op:      opDelete,
		Key:     key,
		Version: version,
	}
	walEntry.CRC = kv.calculateCRC(walEntry)

	// Synchronous WAL write - fail fast if WAL fails
	if err := kv.writeWALEntry(walEntry); err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	// WAL write succeeded, now update memory
	shard := kv.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	entry, exists := shard.data[key]
	if !exists {
		return ErrKeyNotFound
	}

	kv.deleteFromShard(shard, key, entry)

	return nil
}

func (kv *KVStore) Exists(key string) bool {
	if kv.isClosed() {
		return false
	}

	shard := kv.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	entry, exists := shard.data[key]
	if !exists {
		return false
	}

	// Check expiration
	if !entry.ExpireAt.IsZero() && time.Now().After(entry.ExpireAt) {
		return false
	}

	return true
}

func (kv *KVStore) Keys() []string {
	if kv.isClosed() {
		return nil
	}

	var keys []string
	now := time.Now()

	for _, shard := range kv.shards {
		shard.mu.RLock()
		for key, entry := range shard.data {
			if entry.ExpireAt.IsZero() || now.Before(entry.ExpireAt) {
				keys = append(keys, key)
			}
		}
		shard.mu.RUnlock()
	}

	sort.Strings(keys)
	return keys
}

func (kv *KVStore) Size() int {
	return int(atomic.LoadInt64(&kv.entries))
}

func (kv *KVStore) GetStats() Stats {
	hits := atomic.LoadInt64(&kv.hits)
	misses := atomic.LoadInt64(&kv.misses)

	var hitRatio float64
	if hits+misses > 0 {
		hitRatio = float64(hits) / float64(hits+misses)
	}

	return Stats{
		Entries:     atomic.LoadInt64(&kv.entries),
		Hits:        hits,
		Misses:      misses,
		Evictions:   atomic.LoadInt64(&kv.evictions),
		MemoryUsage: atomic.LoadInt64(&kv.memoryUsage),
		WALSize:     atomic.LoadInt64(&kv.walSize),
		HitRatio:    hitRatio,
	}
}

// Snapshot creates a point-in-time snapshot
func (kv *KVStore) Snapshot() error {
	if kv.isClosed() || kv.opts.ReadOnly {
		return ErrStoreClosed
	}

	// Use format-specific extension
	ext := ".bin"
	if kv.serializer.Format() == FormatJSON {
		ext = ".json"
	}
	snapshotPath := filepath.Join(kv.opts.DataDir, "snapshot"+ext)
	tempPath := snapshotPath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(tempPath)
	}

	var writer io.Writer = file
	var gzWriter *gzip.Writer
	if kv.opts.EnableCompression {
		gzWriter = gzip.NewWriter(file)
		writer = gzWriter
	}

	bufWriter := bufio.NewWriter(writer)

	// Write header using serializer
	if err := kv.serializer.WriteSnapshotHeader(bufWriter); err != nil {
		cleanup()
		return err
	}

	// Write all entries
	now := time.Now()
	for _, shard := range kv.shards {
		shard.mu.RLock()
		for _, entry := range shard.data {
			if entry.ExpireAt.IsZero() || now.Before(entry.ExpireAt) {
				data, err := kv.serializer.EncodeEntry(entry)
				if err != nil {
					shard.mu.RUnlock()
					cleanup()
					return err
				}
				if _, err := bufWriter.Write(data); err != nil {
					shard.mu.RUnlock()
					cleanup()
					return err
				}
			}
		}
		shard.mu.RUnlock()
	}

	if err := bufWriter.Flush(); err != nil {
		cleanup()
		return err
	}

	if gzWriter != nil {
		if err := gzWriter.Close(); err != nil {
			cleanup()
			return err
		}
	}

	if err := file.Sync(); err != nil {
		cleanup()
		return err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	// Atomic replace
	if err := os.Rename(tempPath, snapshotPath); err != nil {
		return fmt.Errorf("failed to rename snapshot: %w", err)
	}

	// Reset WAL
	kv.resetWAL()

	return nil
}

func (kv *KVStore) resetWAL() {
	kv.walMutex.Lock()
	defer kv.walMutex.Unlock()

	if kv.walFile != nil {
		kv.walFile.Close()
	}

	walPath := filepath.Join(kv.opts.DataDir, "store.wal")
	os.Remove(walPath)

	file, err := os.Create(walPath)
	if err != nil {
		return
	}

	kv.walFile = file
	kv.walWriter = bufio.NewWriter(file)
	atomic.StoreInt64(&kv.walSize, 0)
}

// loadData loads data from snapshot and replays WAL
func (kv *KVStore) loadData() error {
	// Load snapshot first
	if err := kv.loadSnapshot(); err != nil {
		return err
	}

	// Replay WAL
	return kv.replayWAL()
}

func (kv *KVStore) loadSnapshot() error {
	// Try both formats
	formats := []string{".bin", ".json"}
	var file *os.File
	var err error
	var detectedSerializer Serializer

	for _, ext := range formats {
		snapshotPath := filepath.Join(kv.opts.DataDir, "snapshot"+ext)
		file, err = os.Open(snapshotPath)
		if err == nil {
			// Auto-detect format if enabled
			if kv.opts.AutoDetectFormat {
				format, detectErr := DetectFormat(file)
				if detectErr == nil {
					detectedSerializer = GetSerializer(format)
				}
			}
			break
		}
	}

	if file == nil {
		return nil // No snapshot exists
	}
	defer file.Close()

	// Use detected serializer or configured one
	serializer := kv.serializer
	if detectedSerializer != nil {
		serializer = detectedSerializer
	}

	var reader io.Reader = file
	if kv.opts.EnableCompression {
		gzReader, err := gzip.NewReader(file)
		if err == nil {
			defer gzReader.Close()
			reader = gzReader
		}
	}

	bufReader := bufio.NewReader(reader)

	// Read header using serializer
	if err := serializer.ReadSnapshotHeader(bufReader); err != nil {
		return err
	}

	// Read entries
	for {
		entry, err := serializer.DecodeEntry(bufReader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		shard := kv.getShard(entry.Key)
		shard.mu.Lock()
		shard.data[entry.Key] = entry
		kv.moveToFront(shard, entry)
		shard.mu.Unlock()

		atomic.AddInt64(&kv.entries, 1)
		atomic.AddInt64(&kv.memoryUsage, entry.Size)
	}

	return nil
}

func (kv *KVStore) replayWAL() error {
	walPath := filepath.Join(kv.opts.DataDir, "store.wal")

	file, err := os.Open(walPath)
	if os.IsNotExist(err) {
		return nil // No WAL exists
	}
	if err != nil {
		return err
	}
	defer file.Close()

	// Auto-detect WAL format if enabled
	serializer := kv.serializer
	if kv.opts.AutoDetectFormat {
		format, detectErr := DetectWALFormat(file)
		if detectErr == nil {
			serializer = GetSerializer(format)
		}
	}

	reader := bufio.NewReader(file)

	for {
		entry, err := serializer.DecodeWALEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Stop on first error (corruption)
			break
		}

		if !kv.validateWALEntry(*entry) {
			continue // Skip corrupted entries
		}

		shard := kv.getShard(entry.Key)
		shard.mu.Lock()

		switch entry.Op {
		case opSet:
			size := int64(len(entry.Key) + len(entry.Value) + 64)
			e := &Entry{
				Key:      entry.Key,
				Value:    entry.Value,
				ExpireAt: entry.ExpireAt,
				Size:     size,
				Version:  entry.Version,
			}

			if oldEntry, exists := shard.data[entry.Key]; exists {
				kv.deleteFromShard(shard, entry.Key, oldEntry)
			}

			shard.data[entry.Key] = e
			kv.moveToFront(shard, e)
			atomic.AddInt64(&kv.entries, 1)
			atomic.AddInt64(&kv.memoryUsage, size)

		case opDelete:
			if entry, exists := shard.data[entry.Key]; exists {
				kv.deleteFromShard(shard, entry.Key, entry)
			}
		}

		shard.mu.Unlock()
	}

	return nil
}

// Utility methods

func (kv *KVStore) encodeWALEntry(entry WALEntry) ([]byte, error) {
	return kv.serializer.EncodeWALEntry(entry)
}

func (kv *KVStore) calculateCRC(entry WALEntry) uint32 {
	data := fmt.Sprintf("%d%s%s%d", entry.Op, entry.Key, entry.Value, entry.Version)
	return crc32.ChecksumIEEE([]byte(data))
}

func (kv *KVStore) validateWALEntry(entry WALEntry) bool {
	expected := kv.calculateCRC(entry)
	return entry.CRC == expected
}

func (kv *KVStore) isClosed() bool {
	return atomic.LoadInt32(&kv.closed) != 0
}

func (kv *KVStore) Close() error {
	if !atomic.CompareAndSwapInt32(&kv.closed, 0, 1) {
		return ErrStoreClosed
	}

	// Signal shutdown
	kv.cancel()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		kv.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Normal shutdown
	case <-time.After(kv.opts.CloseTimeout):
		return ErrCloseTimeout
	}

	// Close WAL
	if kv.walFile != nil {
		kv.flushWAL()
		kv.walFile.Close()
	}

	return nil
}

// Default creates a KV store with sensible defaults
func Default() (*KVStore, error) {
	return NewKVStore(Options{
		DataDir:           "data",
		MaxEntries:        100000,
		MaxMemoryMB:       200,
		EnableCompression: true,
	})
}

// SetMetricsCollector sets the unified metrics collector
func (kv *KVStore) SetMetricsCollector(collector metrics.MetricsCollector) {
	kv.collector = collector
}

// GetMetricsCollector returns the current metrics collector
func (kv *KVStore) GetMetricsCollector() metrics.MetricsCollector {
	return kv.collector
}

// recordMetrics records metrics using the unified collector
func (kv *KVStore) recordMetrics(operation, key string, duration time.Duration, err error, hit bool) {
	if kv.collector != nil {
		ctx := context.Background()
		kv.collector.ObserveKV(ctx, operation, key, duration, err, hit)
	}
}
