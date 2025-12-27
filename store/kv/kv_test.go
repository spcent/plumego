package kvstore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestMain handles test setup and cleanup
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup test directories
	cleanupTestDirs()

	os.Exit(code)
}

func cleanupTestDirs() {
	testDirs, _ := filepath.Glob("testdata_*")
	for _, dir := range testDirs {
		os.RemoveAll(dir)
	}
}

// Test utilities
func createTestStore(t *testing.T) (*KVStore, func()) {
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())

	opts := Options{
		DataDir:       dataDir,
		MaxEntries:    100000, // Increased from 1000 to 100000 to reduce evict frequency
		MaxMemoryMB:   100,    // Increased from 10 to 100 to reduce evict frequency
		FlushInterval: 10 * time.Millisecond,
		CleanInterval: 100 * time.Millisecond,
		ShardCount:    4,
	}

	kv, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	cleanup := func() {
		kv.Close()
		os.RemoveAll(dataDir)
	}

	return kv, cleanup
}

// Basic functionality tests

func TestNewKVStore(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "valid options",
			opts: Options{
				DataDir:     "testdata",
				MaxEntries:  1000,
				MaxMemoryMB: 10,
				ShardCount:  4,
			},
			wantErr: false,
		},
		{
			name: "invalid shard count",
			opts: Options{
				DataDir:     "testdata",
				MaxEntries:  1000,
				MaxMemoryMB: 10,
				ShardCount:  3, // Not power of 2
			},
			wantErr: true,
		},
		{
			name: "zero max entries",
			opts: Options{
				DataDir:     "testdata",
				MaxEntries:  0,
				MaxMemoryMB: 10,
				ShardCount:  4,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv, err := NewKVStore(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKVStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if kv != nil {
				kv.Close()
				os.RemoveAll(tt.opts.DataDir)
			}
		})
	}
}

func TestBasicOperations(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Test Set and Get
	key := "test_key"
	value := []byte("test_value")

	err := kv.Set(key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := kv.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !bytes.Equal(value, retrieved) {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}

	// Test Exists
	if !kv.Exists(key) {
		t.Error("Key should exist")
	}

	// Test Delete
	err = kv.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Test Get after delete
	_, err = kv.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Test Exists after delete
	if kv.Exists(key) {
		t.Error("Key should not exist after delete")
	}
}

func TestTTL(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	key := "ttl_key"
	value := []byte("ttl_value")
	ttl := 100 * time.Millisecond

	// Set with TTL
	err := kv.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	// Should exist immediately
	if !kv.Exists(key) {
		t.Error("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(ttl + 50*time.Millisecond)

	// Should not exist after TTL
	if kv.Exists(key) {
		t.Error("Key should not exist after TTL")
	}

	// Get should return expired error
	_, err = kv.Get(key)
	if err != ErrKeyExpired && err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyExpired or ErrKeyNotFound, got %v", err)
	}
}

func TestConcurrentOperations(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	numWorkers := 10
	numOps := 100
	var wg sync.WaitGroup

	// Concurrent writes
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key_%d_%d", workerID, j)
				value := []byte(fmt.Sprintf("value_%d_%d", workerID, j))
				if err := kv.Set(key, value, 0); err != nil {
					t.Errorf("Set failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()

	// Verify all keys exist
	expectedKeys := numWorkers * numOps
	if kv.Size() != expectedKeys {
		t.Errorf("Expected %d keys, got %d", expectedKeys, kv.Size())
	}

	// Concurrent reads
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key_%d_%d", workerID, j)
				expectedValue := []byte(fmt.Sprintf("value_%d_%d", workerID, j))

				value, err := kv.Get(key)
				if err != nil {
					t.Errorf("Get failed for key %s: %v", key, err)
					continue
				}

				if !bytes.Equal(value, expectedValue) {
					t.Errorf("Value mismatch for key %s", key)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestLRUEviction(t *testing.T) {
	// Create store with very small limits
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())
	defer os.RemoveAll(dataDir)

	opts := Options{
		DataDir:     dataDir,
		MaxEntries:  5, // Very small limit
		MaxMemoryMB: 1,
		ShardCount:  2,
	}

	kv, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer kv.Close()

	// Add more entries than the limit
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := []byte(fmt.Sprintf("value_%d", i))
		kv.Set(key, value, 0)
	}

	// Should have evicted some entries
	if kv.Size() > 5 {
		t.Errorf("Expected max 5 entries due to eviction, got %d", kv.Size())
	}

	// Check that evictions were recorded
	stats := kv.GetStats()
	if stats.Evictions == 0 {
		t.Error("Expected some evictions to be recorded")
	}
}

func TestKeys(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Add some keys
	expectedKeys := []string{"apple", "banana", "cherry"}
	for _, key := range expectedKeys {
		kv.Set(key, []byte("value"), 0)
	}

	// Get all keys
	keys := kv.Keys()
	sort.Strings(keys)
	sort.Strings(expectedKeys)

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s, got %s", expectedKeys[i], key)
		}
	}
}

func TestStats(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Perform some operations
	kv.Set("key1", []byte("value1"), 0)
	kv.Set("key2", []byte("value2"), 0)
	kv.Get("key1") // Hit
	kv.Get("key3") // Miss

	stats := kv.GetStats()

	if stats.Entries != 2 {
		t.Errorf("Expected 2 entries, got %d", stats.Entries)
	}

	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	expectedHitRatio := 0.5
	if stats.HitRatio != expectedHitRatio {
		t.Errorf("Expected hit ratio %f, got %f", expectedHitRatio, stats.HitRatio)
	}
}

func TestPersistence(t *testing.T) {
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())
	defer os.RemoveAll(dataDir)

	opts := Options{
		DataDir:     dataDir,
		MaxEntries:  1000,
		MaxMemoryMB: 10,
		ShardCount:  4,
	}

	// Create store and add data
	kv1, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create first store: %v", err)
	}

	testData := map[string][]byte{
		"persistent1": []byte("value1"),
		"persistent2": []byte("value2"),
		"persistent3": []byte("value3"),
	}

	for k, v := range testData {
		if err = kv1.Set(k, v, 0); err != nil {
			t.Fatalf("Failed to set %s: %v", k, err)
		}
	}

	// Force WAL flush
	time.Sleep(50 * time.Millisecond)
	kv1.Close()

	// Create new store with same data directory
	kv2, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create second store: %v", err)
	}
	defer kv2.Close()

	// Verify data persisted
	for k, expectedV := range testData {
		v, err := kv2.Get(k)
		if err != nil {
			t.Errorf("Failed to get %s: %v", k, err)
			continue
		}

		if !bytes.Equal(v, expectedV) {
			t.Errorf("Value mismatch for %s: expected %s, got %s", k, expectedV, v)
		}
	}
}

func TestSnapshotLoadUsesDistinctEntries(t *testing.T) {
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())

	opts := Options{
		DataDir:       dataDir,
		MaxEntries:    100,
		MaxMemoryMB:   10,
		FlushInterval: 10 * time.Millisecond,
		CleanInterval: time.Second,
		ShardCount:    2,
	}

	kv, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		kv.Close()
		os.RemoveAll(dataDir)
	}()

	entries := map[string][]byte{
		"alpha": []byte("one"),
		"beta":  []byte("two"),
	}

	for k, v := range entries {
		if err = kv.Set(k, v, 0); err != nil {
			t.Fatalf("Set %s failed: %v", k, err)
		}
	}

	if err = kv.Snapshot(); err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	kv.Close()

	kvReloaded, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to reload store: %v", err)
	}
	defer kvReloaded.Close()

	for k, expected := range entries {
		value, err := kvReloaded.Get(k)
		if err != nil {
			t.Fatalf("Get %s after reload failed: %v", k, err)
		}

		if !bytes.Equal(value, expected) {
			t.Errorf("Value mismatch for %s: expected %s, got %s", k, expected, value)
		}
	}
}

func TestSnapshot(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Add some data
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("snap_key_%d", i)
		value := []byte(fmt.Sprintf("snap_value_%d", i))
		kv.Set(key, value, 0)
	}

	// Create snapshot
	err := kv.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Verify snapshot file exists
	snapshotPath := filepath.Join(kv.opts.DataDir, "snapshot.json")
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		t.Error("Snapshot file should exist")
	}

	// WAL should be reset after snapshot
	stats := kv.GetStats()
	if stats.WALSize > 0 {
		t.Error("WAL should be reset after snapshot")
	}
}

func TestCleanup(t *testing.T) {
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())
	defer os.RemoveAll(dataDir)

	opts := Options{
		DataDir:       dataDir,
		MaxEntries:    1000,
		MaxMemoryMB:   10,
		CleanInterval: 50 * time.Millisecond, // Fast cleanup for testing
		ShardCount:    4,
	}

	kv, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer kv.Close()

	// Add keys with short TTL
	ttl := 100 * time.Millisecond
	numKeys := 10

	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("cleanup_key_%d", i)
		value := []byte(fmt.Sprintf("cleanup_value_%d", i))
		kv.Set(key, value, ttl)
	}

	// Verify keys exist
	if kv.Size() != numKeys {
		t.Errorf("Expected %d keys, got %d", numKeys, kv.Size())
	}

	// Wait for TTL + cleanup cycle
	time.Sleep(ttl + 100*time.Millisecond)

	// Keys should be cleaned up
	if kv.Size() != 0 {
		t.Errorf("Expected 0 keys after cleanup, got %d", kv.Size())
	}
}

func TestGetExpiredIncrementsMisses(t *testing.T) {
	dataDir := fmt.Sprintf("testdata_%d", time.Now().UnixNano())
	defer os.RemoveAll(dataDir)

	opts := Options{
		DataDir:       dataDir,
		MaxEntries:    10,
		MaxMemoryMB:   1,
		FlushInterval: 10 * time.Millisecond,
		CleanInterval: time.Hour,
		ShardCount:    2,
	}

	kv, err := NewKVStore(opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer kv.Close()

	key := "ephemeral"
	if err = kv.Set(key, []byte("value"), 20*time.Millisecond); err != nil {
		t.Fatalf("Failed to set key with TTL: %v", err)
	}

	statsBefore := kv.GetStats()
	time.Sleep(40 * time.Millisecond)

	_, err = kv.Get(key)
	if err != ErrKeyExpired {
		t.Fatalf("Expected ErrKeyExpired, got %v", err)
	}

	statsAfter := kv.GetStats()
	if statsAfter.Misses != statsBefore.Misses+1 {
		t.Fatalf("Expected misses to increment after expired get: before=%d after=%d", statsBefore.Misses, statsAfter.Misses)
	}
}

func TestErrorConditions(t *testing.T) {
	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Test operations on non-existent key
	_, err := kv.Get("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	err = kv.Delete("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Test operations on closed store
	kv.Close()

	err = kv.Set("key", []byte("value"), 0)
	if err != ErrStoreClosed {
		t.Errorf("Expected ErrStoreClosed, got %v", err)
	}

	_, err = kv.Get("key")
	if err != ErrStoreClosed {
		t.Errorf("Expected ErrStoreClosed, got %v", err)
	}

	err = kv.Delete("key")
	if err != ErrStoreClosed {
		t.Errorf("Expected ErrStoreClosed, got %v", err)
	}
}

// Benchmark tests

func BenchmarkSet(b *testing.B) {
	kv, cleanup := createTestStore(&testing.T{})
	defer cleanup()

	value := []byte("benchmark_value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			kv.Set(key, value, 0)
			i++
		}
	})
}

func BenchmarkGet(b *testing.B) {
	kv, cleanup := createTestStore(&testing.T{})
	defer cleanup()

	// Pre-populate with data
	numKeys := 10000
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		kv.Set(key, value, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%numKeys)
			kv.Get(key)
			i++
		}
	})
}

func BenchmarkMixed(b *testing.B) {
	kv, cleanup := createTestStore(&testing.T{})
	defer cleanup()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := []byte(fmt.Sprintf("value_%d", i))
		kv.Set(key, value, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)

			// 70% reads, 30% writes
			if i%10 < 7 {
				kv.Get(key)
			} else {
				value := []byte(fmt.Sprintf("new_value_%d", i))
				kv.Set(key, value, 0)
			}
			i++
		}
	})
}

// Stress tests

func TestStressLargeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	kv, cleanup := createTestStore(t)
	defer cleanup()

	// Test with 1MB values
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	numKeys := 10
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("large_key_%d", i)
		if err := kv.Set(key, largeValue, 0); err != nil {
			t.Fatalf("Failed to set large value: %v", err)
		}
	}

	// Verify all values
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("large_key_%d", i)
		value, err := kv.Get(key)
		if err != nil {
			t.Fatalf("Failed to get large value: %v", err)
		}

		if !bytes.Equal(value, largeValue) {
			t.Errorf("Large value corrupted for key %s", key)
		}
	}
}

func TestStressHighConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	kv, cleanup := createTestStore(t)
	defer cleanup()

	numWorkers := 50
	opsPerWorker := 1000
	var wg sync.WaitGroup

	// Track operations
	var sets, gets, deletes int64

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("stress_%d_%d", workerID, j)
				value := []byte(fmt.Sprintf("value_%d_%d", workerID, j))

				op := j % 3
				switch op {
				case 0: // Set
					kv.Set(key, value, 0)
					atomic.AddInt64(&sets, 1)
				case 1: // Get
					kv.Get(key)
					atomic.AddInt64(&gets, 1)
				case 2: // Delete
					kv.Delete(key)
					atomic.AddInt64(&deletes, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Completed stress test: %d sets, %d gets, %d deletes",
		atomic.LoadInt64(&sets),
		atomic.LoadInt64(&gets),
		atomic.LoadInt64(&deletes))

	// Verify store is still functional
	testKey := "post_stress_test"
	testValue := []byte("post_stress_value")

	if err := kv.Set(testKey, testValue, 0); err != nil {
		t.Fatalf("Store not functional after stress test: %v", err)
	}

	retrievedValue, err := kv.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get value after stress test: %v", err)
	}

	if !bytes.Equal(testValue, retrievedValue) {
		t.Error("Value corrupted after stress test")
	}
}

// Memory usage test
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	kv, cleanup := createTestStore(t)
	defer cleanup()

	initialStats := kv.GetStats()
	initialMemory := initialStats.MemoryUsage

	// Add known amount of data
	numEntries := 1000
	valueSize := 1024
	value := make([]byte, valueSize)

	for i := 0; i < numEntries; i++ {
		key := fmt.Sprintf("mem_key_%d", i)
		kv.Set(key, value, 0)
	}

	finalStats := kv.GetStats()
	finalMemory := finalStats.MemoryUsage

	memoryIncrease := finalMemory - initialMemory
	t.Logf("Memory increase: %d bytes for %d entries", memoryIncrease, numEntries)

	// Memory usage should have increased
	if memoryIncrease <= 0 {
		t.Error("Memory usage should have increased")
	}

	// Rough check - should be somewhat proportional to data added
	expectedMinIncrease := int64(numEntries * valueSize / 2) // Allow for overhead
	if memoryIncrease < expectedMinIncrease {
		t.Errorf("Memory increase too small: got %d, expected at least %d",
			memoryIncrease, expectedMinIncrease)
	}
}

// Default store test
func TestDefault(t *testing.T) {
	kv, err := Default()
	if err != nil {
		t.Fatalf("Failed to create default store: %v", err)
	}
	defer func() {
		kv.Close()
		os.RemoveAll("data")
	}()

	// Basic functionality test
	err = kv.Set("test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("Default store Set failed: %v", err)
	}

	value, err := kv.Get("test")
	if err != nil {
		t.Fatalf("Default store Get failed: %v", err)
	}

	if string(value) != "value" {
		t.Errorf("Expected 'value', got '%s'", string(value))
	}
}
