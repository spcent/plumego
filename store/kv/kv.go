package kvstore

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
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
	DataDir           string        `json:"data_dir"`
	MaxEntries        int           `json:"max_entries"`
	MaxMemoryMB       int           `json:"max_memory_mb"`
	FlushInterval     time.Duration `json:"flush_interval"`
	CleanInterval     time.Duration `json:"clean_interval"`
	ShardCount        int           `json:"shard_count"`
	EnableCompression bool          `json:"enable_compression"`
	ReadOnly          bool          `json:"read_only"`
	CloseTimeout      time.Duration `json:"close_timeout"`
}

// Shard represents a single data shard
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
	opts Options

	// Sharded data
	shards    []*Shard
	shardMask uint32

	// WAL
	walFile   *os.File
	walWriter *bufio.Writer
	walMutex  sync.Mutex
	logChan   chan WALEntry

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
		opts:      opts,
		shards:    shards,
		shardMask: uint32(opts.ShardCount - 1),
		logChan:   make(chan WALEntry, 1000),
	}

	kv.ctx, kv.cancel = context.WithCancel(context.Background())

	// Initialize WAL
	if !opts.ReadOnly {
		if err := kv.initWAL(); err != nil {
			return nil, fmt.Errorf("failed to initialize WAL: %w", err)
		}

		// Start background workers
		kv.wg.Add(2)
		go kv.writeWal()
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

// walWriter handles background WAL writing
func (kv *KVStore) writeWal() {
	defer kv.wg.Done()

	ticker := time.NewTicker(kv.opts.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-kv.logChan:
			if err := kv.writeWALEntry(entry); err != nil {
				// In production, you'd want proper logging here
				fmt.Printf("WAL write error: %v\n", err)
			}

		case <-ticker.C:
			kv.flushWAL()

		case <-kv.ctx.Done():
			// Drain remaining entries
			for {
				select {
				case entry := <-kv.logChan:
					kv.writeWALEntry(entry)
				default:
					kv.flushWAL()
					return
				}
			}
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

	for _, shard := range kv.shards {
		var expired []string

		shard.mu.RLock()
		for key, entry := range shard.data {
			if !entry.ExpireAt.IsZero() && now.After(entry.ExpireAt) {
				expired = append(expired, key)
			}
		}
		shard.mu.RUnlock()

		if len(expired) > 0 {
			shard.mu.Lock()
			for _, key := range expired {
				if entry, exists := shard.data[key]; exists {
					if !entry.ExpireAt.IsZero() && now.After(entry.ExpireAt) {
						kv.deleteFromShard(shard, key, entry)
						if !kv.opts.ReadOnly {
							kv.logDelete(key)
						}
					}
				}
			}
			shard.mu.Unlock()
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

	if !kv.opts.ReadOnly {
		kv.logDelete(entry.Key)
	}

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

	shard := kv.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	size := int64(len(key) + len(value) + 64) // rough estimate
	version := atomic.AddInt64(&kv.version, 1)

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

	// Log to WAL
	if !kv.opts.ReadOnly {
		kv.logSet(key, value, expireAt, version)
	}

	return nil
}

func (kv *KVStore) Get(key string) ([]byte, error) {
	if kv.isClosed() {
		return nil, ErrStoreClosed
	}

	shard := kv.getShard(key)

	// First, try with read lock for non-modifying operations
	shard.mu.RLock()
	entry, exists := shard.data[key]

	// Check if key exists and is not expired
	if !exists {
		shard.mu.RUnlock()
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyNotFound
	}

	// Check expiration (read-only check first)
	isExpired := !entry.ExpireAt.IsZero() && time.Now().After(entry.ExpireAt)
	if isExpired {
		shard.mu.RUnlock()
		// Need write lock to delete expired entry
		shard.mu.Lock()
		// Recheck existence and expiration under write lock
		if e, exists := shard.data[key]; exists {
			if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
				kv.deleteFromShard(shard, key, e)
				if !kv.opts.ReadOnly {
					kv.logDelete(key)
				}
			}
		}
		shard.mu.Unlock()
		atomic.AddInt64(&kv.misses, 1)
		return nil, ErrKeyExpired
	}

	// For LRU update, we need write lock
	// Create defensive copy first under read lock
	valueCopy := append([]byte(nil), entry.Value...)
	shard.mu.RUnlock()

	// Acquire write lock for LRU update
	shard.mu.Lock()
	// Recheck existence (could have been deleted by another goroutine)
	if e, exists := shard.data[key]; exists {
		// Recheck expiration
		if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
			shard.mu.Unlock()
			atomic.AddInt64(&kv.misses, 1)
			return nil, ErrKeyExpired
		}
		// Move to front (LRU)
		kv.moveToFront(shard, e)
	}
	shard.mu.Unlock()

	atomic.AddInt64(&kv.hits, 1)
	return valueCopy, nil
}

func (kv *KVStore) Delete(key string) error {
	if kv.isClosed() {
		return ErrStoreClosed
	}

	shard := kv.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	entry, exists := shard.data[key]
	if !exists {
		return ErrKeyNotFound
	}

	kv.deleteFromShard(shard, key, entry)

	if !kv.opts.ReadOnly {
		kv.logDelete(key)
	}

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

	snapshotPath := filepath.Join(kv.opts.DataDir, "snapshot.json")
	tempPath := snapshotPath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	if kv.opts.EnableCompression {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
	}

	encoder := json.NewEncoder(writer)

	// Write header
	header := map[string]interface{}{
		"magic":   magicNumber,
		"version": version,
		"created": time.Now(),
	}
	if err := encoder.Encode(header); err != nil {
		return err
	}

	// Write all entries
	now := time.Now()
	for _, shard := range kv.shards {
		shard.mu.RLock()
		for _, entry := range shard.data {
			if entry.ExpireAt.IsZero() || now.Before(entry.ExpireAt) {
				if err := encoder.Encode(entry); err != nil {
					shard.mu.RUnlock()
					return err
				}
			}
		}
		shard.mu.RUnlock()
	}

	file.Sync()

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
	snapshotPath := filepath.Join(kv.opts.DataDir, "snapshot.json")

	file, err := os.Open(snapshotPath)
	if os.IsNotExist(err) {
		return nil // No snapshot exists
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file
	if kv.opts.EnableCompression {
		gzReader, err := gzip.NewReader(file)
		if err == nil {
			defer gzReader.Close()
			reader = gzReader
		}
	}

	decoder := json.NewDecoder(reader)

	// Read header
	var header map[string]interface{}
	if err := decoder.Decode(&header); err != nil {
		return err
	}

	// Read entries
	for decoder.More() {
		var entry Entry
		if err := decoder.Decode(&entry); err != nil {
			return err
		}

		shard := kv.getShard(entry.Key)
		shard.mu.Lock()
		entryCopy := entry
		shard.data[entry.Key] = &entryCopy
		kv.moveToFront(shard, &entryCopy)
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

	decoder := json.NewDecoder(file)

	for decoder.More() {
		var entry WALEntry
		if err := decoder.Decode(&entry); err != nil {
			// Stop on first error (corruption)
			break
		}

		if !kv.validateWALEntry(entry) {
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

func (kv *KVStore) logSet(key string, value []byte, expireAt time.Time, version int64) {
	entry := WALEntry{
		Op:       opSet,
		Key:      key,
		Value:    value,
		ExpireAt: expireAt,
		Version:  version,
	}
	entry.CRC = kv.calculateCRC(entry)

	select {
	case kv.logChan <- entry:
	default:
		// Channel full, drop entry (or implement backpressure)
	}
}

func (kv *KVStore) logDelete(key string) {
	entry := WALEntry{
		Op:      opDelete,
		Key:     key,
		Version: atomic.AddInt64(&kv.version, 1),
	}
	entry.CRC = kv.calculateCRC(entry)

	select {
	case kv.logChan <- entry:
	default:
		// Channel full, drop entry
	}
}

func (kv *KVStore) encodeWALEntry(entry WALEntry) ([]byte, error) {
	return json.Marshal(entry)
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
