package kvstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestKVStoreBasicOperations(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("alpha", []byte("one"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get("alpha")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, []byte("one")) {
		t.Fatalf("unexpected value %q", got)
	}

	if !store.Exists("alpha") {
		t.Fatalf("expected alpha to exist")
	}

	if err := store.Delete("alpha"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("alpha"); err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestKVStoreTTL(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("ttl", []byte("value"), 20*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if store.Exists("ttl") {
		t.Fatalf("expected expired key to be gone")
	}
	if _, err := store.Get("ttl"); err != ErrKeyNotFound && err != ErrKeyExpired {
		t.Fatalf("expected expiration error, got %v", err)
	}
}

func TestKVStorePersistenceAcrossReopen(t *testing.T) {
	dir := t.TempDir()

	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Set("persist", []byte("value"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	got, err := reopened.Get("persist")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if !bytes.Equal(got, []byte("value")) {
		t.Fatalf("unexpected reopened value %q", got)
	}
}

func TestKVStoreKeysAndStats(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	_ = store.Set("b", []byte("2"), 0)
	_ = store.Set("a", []byte("1"), 0)
	_, _ = store.Get("a")
	_, _ = store.Get("missing")

	keys := store.Keys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("unexpected keys: %#v", keys)
	}

	stats := store.GetStats()
	if stats.Entries != 2 {
		t.Fatalf("expected 2 entries, got %d", stats.Entries)
	}
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestKVStoreClose(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := store.Set("closed", []byte("x"), 0); err != ErrStoreClosed {
		t.Fatalf("expected ErrStoreClosed, got %v", err)
	}
}

func TestKVStoreCloseIdempotent(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("expected Close to be idempotent, got %v", err)
	}
}

func TestSetDefaultsTrimsWhitespaceDataDir(t *testing.T) {
	opts := Options{DataDir: " \t ", MaxEntries: 1, MaxMemoryMB: 1}

	setDefaults(&opts)

	if opts.DataDir != "data" {
		t.Fatalf("expected whitespace-only DataDir to use default, got %q", opts.DataDir)
	}
}

func TestSentinelErrorMessagesAreNamespaced(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{name: "not found", err: ErrKeyNotFound, want: "kv: key not found"},
		{name: "expired", err: ErrKeyExpired, want: "kv: key expired"},
		{name: "invalid key", err: ErrInvalidKey, want: "kv: key is required"},
		{name: "closed", err: ErrStoreClosed, want: "kv: store is closed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("error string = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateOptionsErrorsAreNamespaced(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want string
	}{
		{
			name: "max entries",
			opts: Options{DataDir: t.TempDir(), MaxEntries: -1, MaxMemoryMB: 1},
			want: "kv: max entries must be positive",
		},
		{
			name: "max memory",
			opts: Options{DataDir: t.TempDir(), MaxEntries: 1, MaxMemoryMB: -1},
			want: "kv: max memory must be positive",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOptions(tc.opts)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if got := err.Error(); got != tc.want {
				t.Fatalf("error string = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKVStoreRejectsEmptyKeys(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("", []byte("value"), 0); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("Set empty key error = %v, want ErrInvalidKey", err)
	}
	if _, err := store.Get(""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("Get empty key error = %v, want ErrInvalidKey", err)
	}
	if err := store.Delete(""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("Delete empty key error = %v, want ErrInvalidKey", err)
	}
	if store.Exists("") {
		t.Fatal("Exists should return false for empty key")
	}

	stats := store.GetStats()
	if stats.Misses != 0 {
		t.Fatalf("invalid keys should not count as misses, got %+v", stats)
	}
}

func TestKVStoreLoadRecomputesEntrySize(t *testing.T) {
	dir := t.TempDir()
	writeState(t, dir, diskState{Entries: map[string]entry{
		"alpha": {
			Value:     []byte("one"),
			UpdatedAt: time.Now(),
			Size:      0,
		},
	}})

	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	stats := store.GetStats()
	if stats.MemoryUsage != entrySize("alpha", []byte("one")) {
		t.Fatalf("MemoryUsage = %d, want %d", stats.MemoryUsage, entrySize("alpha", []byte("one")))
	}
}

func TestKVStoreLoadSkipsInvalidKeys(t *testing.T) {
	dir := t.TempDir()
	writeState(t, dir, diskState{Entries: map[string]entry{
		"": {
			Value:     []byte("invalid"),
			UpdatedAt: time.Now(),
			Size:      entrySize("", []byte("invalid")),
		},
		"alpha": {
			Value:     []byte("one"),
			UpdatedAt: time.Now(),
			Size:      entrySize("alpha", []byte("one")),
		},
	}})

	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	keys := store.Keys()
	if len(keys) != 1 || keys[0] != "alpha" {
		t.Fatalf("expected only valid key after load, got %v", keys)
	}
	if store.Exists("") {
		t.Fatal("invalid empty key should not be visible after load")
	}

	var persisted diskState
	readState(t, dir, &persisted)
	if _, ok := persisted.Entries[""]; ok {
		t.Fatalf("invalid empty key should not be persisted after normalization: %+v", persisted.Entries)
	}
}

func TestKVStoreLoadAppliesMaxEntries(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeState(t, dir, diskState{Entries: map[string]entry{
		"old": {
			Value:     []byte("old"),
			UpdatedAt: now.Add(-time.Minute),
			Size:      entrySize("old", []byte("old")),
		},
		"new": {
			Value:     []byte("new"),
			UpdatedAt: now,
			Size:      entrySize("new", []byte("new")),
		},
	}})

	store, err := NewKVStore(Options{DataDir: dir, MaxEntries: 1})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	keys := store.Keys()
	if len(keys) != 1 || keys[0] != "new" {
		t.Fatalf("expected only newest key after load eviction, got %v", keys)
	}
}

func TestKVStoreRejectsOversizedValue(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir(), MaxMemoryMB: 1})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	oversized := bytes.Repeat([]byte("x"), 1024*1024)
	if err := store.Set("too-large", oversized, 0); err == nil {
		t.Fatal("expected oversized value to be rejected")
	}
	if store.Exists("too-large") {
		t.Fatal("oversized value should not be stored in memory")
	}
}

func TestKVStorePersistFailureCleansTempFile(t *testing.T) {
	dir := t.TempDir()
	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	blockStatePath(t, dir)

	if err := store.Set("alpha", []byte("one"), 0); err == nil {
		t.Fatal("expected persist failure")
	}

	matches, err := filepath.Glob(filepath.Join(dir, stateFileName+".*.tmp"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected temp files to be cleaned up, got %v", matches)
	}
}

func TestKVStoreSetRollbackOnPersistFailure(t *testing.T) {
	dir := t.TempDir()
	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("alpha", []byte("one"), 0); err != nil {
		t.Fatalf("initial Set: %v", err)
	}
	blockStatePath(t, dir)

	if err := store.Set("alpha", []byte("two"), 0); err == nil {
		t.Fatal("expected persist failure")
	}
	got, err := store.Get("alpha")
	if err != nil {
		t.Fatalf("Get after failed Set: %v", err)
	}
	if !bytes.Equal(got, []byte("one")) {
		t.Fatalf("expected rollback to old value, got %q", got)
	}
}

func TestKVStoreDeleteRollbackOnPersistFailure(t *testing.T) {
	dir := t.TempDir()
	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("alpha", []byte("one"), 0); err != nil {
		t.Fatalf("initial Set: %v", err)
	}
	blockStatePath(t, dir)

	if err := store.Delete("alpha"); err == nil {
		t.Fatal("expected persist failure")
	}
	got, err := store.Get("alpha")
	if err != nil {
		t.Fatalf("Get after failed Delete: %v", err)
	}
	if !bytes.Equal(got, []byte("one")) {
		t.Fatalf("expected deleted value to be restored, got %q", got)
	}
}

func TestKVStoreExpiredGetRollbackOnPersistFailure(t *testing.T) {
	dir := t.TempDir()
	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("ttl", []byte("value"), 10*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	blockStatePath(t, dir)

	if _, err := store.Get("ttl"); err == nil {
		t.Fatal("expected persist failure")
	} else if errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected persist failure to be returned, got %v", err)
	}

	store.mu.RLock()
	item, stillLoaded := store.data["ttl"]
	store.mu.RUnlock()
	if !stillLoaded {
		t.Fatal("expired key should be restored when cleanup cannot persist")
	}
	if !bytes.Equal(item.Value, []byte("value")) {
		t.Fatalf("restored value = %q, want %q", item.Value, []byte("value"))
	}
}

func TestKVStoreReadOnlyExpiredChecksDoNotMutate(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("ttl", []byte("value"), 10*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	if store.Exists("ttl") {
		t.Fatal("expected expired key to not exist")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("expected no non-expired keys, got %v", keys)
	}

	store.mu.RLock()
	_, stillLoaded := store.data["ttl"]
	store.mu.RUnlock()
	if !stillLoaded {
		t.Fatal("read-only checks should not mutate expired entries")
	}
}

func TestKVStoreConcurrentReadOnlyInspection(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("alpha", []byte("one"), 0); err != nil {
		t.Fatalf("Set alpha: %v", err)
	}
	if err := store.Set("beta", []byte("two"), 0); err != nil {
		t.Fatalf("Set beta: %v", err)
	}

	var readers sync.WaitGroup
	for i := 0; i < 50; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			if !store.Exists("alpha") {
				t.Error("expected alpha to exist")
			}
			if keys := store.Keys(); len(keys) != 2 {
				t.Errorf("expected two keys, got %v", keys)
			}
			if size := store.Size(); size != 2 {
				t.Errorf("expected size 2, got %d", size)
			}
		}()
	}
	readers.Wait()
}

func blockStatePath(t *testing.T, dir string) {
	t.Helper()

	statePath := filepath.Join(dir, stateFileName)
	if err := os.Remove(statePath); err != nil {
		t.Fatalf("remove state file: %v", err)
	}
	if err := os.Mkdir(statePath, 0755); err != nil {
		t.Fatalf("create blocking state dir: %v", err)
	}
}

func writeState(t *testing.T, dir string, state diskState) {
	t.Helper()

	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, stateFileName), raw, 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}
}

func readState(t *testing.T, dir string, state *diskState) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(dir, stateFileName))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if err := json.Unmarshal(raw, state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
}
