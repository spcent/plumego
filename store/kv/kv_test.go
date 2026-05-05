package kvstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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

func TestKVStoreNilAndEmptyValueRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}

	if err := store.Set("nil-value", nil, 0); err != nil {
		t.Fatalf("Set nil value: %v", err)
	}
	if err := store.Set("empty-value", []byte{}, 0); err != nil {
		t.Fatalf("Set empty value: %v", err)
	}

	nilValue, err := store.Get("nil-value")
	if err != nil {
		t.Fatalf("Get nil value: %v", err)
	}
	if nilValue != nil {
		t.Fatalf("nil value round-trip = %#v, want nil", nilValue)
	}
	emptyValue, err := store.Get("empty-value")
	if err != nil {
		t.Fatalf("Get empty value: %v", err)
	}
	if emptyValue == nil || len(emptyValue) != 0 {
		t.Fatalf("empty value round-trip = %#v, want non-nil empty slice", emptyValue)
	}
	if !store.Exists("nil-value") || !store.Exists("empty-value") {
		t.Fatal("nil and empty values should be existing keys")
	}
	if _, err := store.Get("missing-value"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("missing value error = %v, want ErrKeyNotFound", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	emptyValue, err = reopened.Get("empty-value")
	if err != nil {
		t.Fatalf("Get reopened empty value: %v", err)
	}
	if emptyValue == nil || len(emptyValue) != 0 {
		t.Fatalf("reopened empty value = %#v, want non-nil empty slice", emptyValue)
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

func TestKVStoreZeroValueFailsClosed(t *testing.T) {
	var store KVStore
	ctx := t.Context()

	if err := store.SetContext(ctx, "key", []byte("value"), 0); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value SetContext error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.GetContext(ctx, "key"); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value GetContext error = %v, want ErrStoreClosed", err)
	}
	if err := store.DeleteContext(ctx, "key"); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value DeleteContext error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.ExistsContext(ctx, "key"); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value ExistsContext error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.KeysContext(ctx); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value KeysContext error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.SizeContext(ctx); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value SizeContext error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.GetStatsContext(ctx); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("zero-value GetStatsContext error = %v, want ErrStoreClosed", err)
	}

	if store.Exists("key") {
		t.Fatal("zero-value compat Exists should collapse closed error to false")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("zero-value compat Keys = %v, want empty", keys)
	}
	if size := store.Size(); size != 0 {
		t.Fatalf("zero-value compat Size = %d, want zero", size)
	}
	if stats := store.GetStats(); stats != (Stats{}) {
		t.Fatalf("zero-value compat GetStats = %+v, want zero", stats)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("zero-value Close error = %v, want nil", err)
	}

	var nilStore *KVStore
	if err := nilStore.Close(); err != nil {
		t.Fatalf("nil Close error = %v, want nil", err)
	}
	if err := nilStore.SetContext(ctx, "key", []byte("value"), 0); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("nil SetContext error = %v, want ErrStoreClosed", err)
	}
}

func TestSetDefaultsTrimsWhitespaceDataDirWithoutDefaulting(t *testing.T) {
	opts := Options{DataDir: " \t ", MaxEntries: 1, MaxMemoryMB: 1}

	setDefaults(&opts)

	if opts.DataDir != "" {
		t.Fatalf("expected whitespace-only DataDir to stay empty after trimming, got %q", opts.DataDir)
	}
}

func TestNewKVStoreRequiresExplicitDataDir(t *testing.T) {
	cwd := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	for _, opts := range []Options{
		{},
		{DataDir: " \t "},
	} {
		if _, err := NewKVStore(opts); err == nil || !strings.Contains(err.Error(), "data dir is required") {
			t.Fatalf("NewKVStore(%+v) error = %v, want data dir required", opts, err)
		}
	}
	if _, err := os.Stat(filepath.Join(cwd, "data")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Options{} should not create implicit data directory, stat error = %v", err)
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
		{name: "value too large", err: ErrValueTooLarge, want: "kv: value too large"},
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
			name: "data dir",
			opts: Options{MaxEntries: 1, MaxMemoryMB: 1},
			want: "kv: data dir is required",
		},
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
	if _, err := store.ExistsContext(t.Context(), "bad\nkey"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("ExistsContext control key error = %v, want ErrInvalidKey", err)
	}
	if err := store.Set("bad\nkey", []byte("value"), 0); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("Set control key error = %v, want ErrInvalidKey", err)
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

func TestKVStoreLoadRejectsCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, stateFileName), []byte(`{"entries":`), 0644); err != nil {
		t.Fatalf("write corrupt state: %v", err)
	}

	if _, err := NewKVStore(Options{DataDir: dir}); !errors.Is(err, ErrCorruptState) {
		t.Fatalf("NewKVStore corrupt JSON error = %v, want ErrCorruptState", err)
	} else if errors.Is(err, ErrInvalidKey) {
		t.Fatalf("corrupt JSON should not be classified as ErrInvalidKey: %v", err)
	}
}

func TestKVStoreLoadRejectsInvalidKeys(t *testing.T) {
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

	if _, err := NewKVStore(Options{DataDir: dir}); !errors.Is(err, ErrCorruptState) {
		t.Fatalf("NewKVStore invalid persisted key error = %v, want ErrCorruptState", err)
	} else if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("NewKVStore invalid persisted key error = %v, want ErrInvalidKey in chain", err)
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

func TestKVStoreOpenDoesNotPersistPrunedStartupState(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeState(t, dir, diskState{Entries: map[string]entry{
		"expired": {
			Value:     []byte("old"),
			ExpireAt:  now.Add(-time.Minute),
			UpdatedAt: now.Add(-time.Hour),
			Size:      entrySize("expired", []byte("old")),
		},
		"current": {
			Value:     []byte("new"),
			UpdatedAt: now,
			Size:      entrySize("current", []byte("new")),
		},
	}})

	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if keys := store.Keys(); len(keys) != 1 || keys[0] != "current" {
		t.Fatalf("loaded keys = %v, want [current]", keys)
	}

	var persisted diskState
	readState(t, dir, &persisted)
	if _, ok := persisted.Entries["expired"]; !ok {
		t.Fatalf("startup pruning should not rewrite state file: %+v", persisted.Entries)
	}
}

func TestKVStoreRejectsOversizedValue(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir(), MaxMemoryMB: 1})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	oversized := bytes.Repeat([]byte("x"), 1024*1024)
	err = store.Set("too-large", oversized, 0)
	if err == nil {
		t.Fatal("expected oversized value to be rejected")
	}
	if !errors.Is(err, ErrValueTooLarge) {
		t.Fatalf("expected ErrValueTooLarge, got %v", err)
	}
	if !strings.Contains(err.Error(), "size ") || !strings.Contains(err.Error(), " limit ") {
		t.Fatalf("expected size and limit details, got %v", err)
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

func TestKVStoreExpiredGetDoesNotPersistCleanup(t *testing.T) {
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

	if _, err := store.Get("ttl"); !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("Get expired error = %v, want ErrKeyExpired", err)
	}

	store.mu.RLock()
	item, stillLoaded := store.data["ttl"]
	store.mu.RUnlock()
	if !stillLoaded {
		t.Fatal("expired key should remain loaded after read-only Get")
	}
	if !bytes.Equal(item.Value, []byte("value")) {
		t.Fatalf("loaded value = %q, want %q", item.Value, []byte("value"))
	}
}

func TestKVStoreContextMethodsRejectCanceledContext(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := store.SetContext(ctx, "alpha", []byte("one"), 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("SetContext canceled error = %v, want context.Canceled", err)
	}
	if _, err := store.GetContext(ctx, "alpha"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetContext canceled error = %v, want context.Canceled", err)
	}
	if err := store.DeleteContext(ctx, "alpha"); !errors.Is(err, context.Canceled) {
		t.Fatalf("DeleteContext canceled error = %v, want context.Canceled", err)
	}
	if _, err := store.ExistsContext(ctx, "alpha"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ExistsContext canceled error = %v, want context.Canceled", err)
	}
	if _, err := store.KeysContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("KeysContext canceled error = %v, want context.Canceled", err)
	}
	if _, err := store.SizeContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("SizeContext canceled error = %v, want context.Canceled", err)
	}
	if _, err := store.GetStatsContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetStatsContext canceled error = %v, want context.Canceled", err)
	}
}

func TestKVStoreContextMethodsReturnClosedErrors(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := store.ExistsContext(t.Context(), "alpha"); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("ExistsContext closed error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.KeysContext(t.Context()); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("KeysContext closed error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.SizeContext(t.Context()); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("SizeContext closed error = %v, want ErrStoreClosed", err)
	}
	if _, err := store.GetStatsContext(t.Context()); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("GetStatsContext closed error = %v, want ErrStoreClosed", err)
	}

	if store.Exists("alpha") {
		t.Fatal("compat Exists should collapse closed error to false")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("compat Keys should collapse closed error to empty slice, got %v", keys)
	}
	if size := store.Size(); size != 0 {
		t.Fatalf("compat Size should collapse closed error to zero, got %d", size)
	}
	if stats := store.GetStats(); stats != (Stats{}) {
		t.Fatalf("compat GetStats should collapse closed error to zero stats, got %+v", stats)
	}
}

func TestKVStoreSetContextRejectsCanceledContextAfterWaitingForLock(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	store.mu.Lock()
	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		errCh <- store.SetContext(ctx, "blocked", []byte("value"), 0)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	store.mu.Unlock()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("SetContext error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("SetContext did not return after lock was released")
	}
	if store.Exists("blocked") {
		t.Fatal("canceled SetContext should not store value")
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

func TestUnsupportedDirSyncErrorClassification(t *testing.T) {
	if !isUnsupportedDirSync(os.ErrInvalid) {
		t.Fatal("os.ErrInvalid should be unsupported directory sync")
	}
	if !isUnsupportedDirSync(&os.PathError{Op: "sync", Path: "dir", Err: syscall.EINVAL}) {
		t.Fatal("PathError wrapping EINVAL should be unsupported directory sync")
	}
	if isUnsupportedDirSync(&os.PathError{Op: "sync", Path: "dir", Err: syscall.EIO}) {
		t.Fatal("EIO should remain fatal")
	}
}

func blockStatePath(t *testing.T, dir string) {
	t.Helper()

	statePath := filepath.Join(dir, stateFileName)
	if err := os.Remove(statePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("remove state file: %v", err)
		}
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
