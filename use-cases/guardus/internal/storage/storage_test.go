package storage_test

import (
	"path/filepath"
	"testing"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
	"guardus/internal/storage"
	"guardus/internal/storage/common/paging"
)

func newStores(t *testing.T) map[string]storage.Store {
	t.Helper()
	logger := plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
	mem, err := storage.New(&storage.Config{Type: storage.TypeMemory}, logger)
	if err != nil {
		t.Fatalf("memory store: %v", err)
	}
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "guardus.db")
	sqliteCfg := &storage.Config{Type: storage.TypeSQLite, Path: dbPath}
	sqlite, err := storage.New(sqliteCfg, logger)
	if err != nil {
		t.Fatalf("sqlite store: %v", err)
	}
	t.Cleanup(func() {
		mem.Close()
		sqlite.Close()
	})
	return map[string]storage.Store{"memory": mem, "sqlite": sqlite}
}

func TestStoreContract(t *testing.T) {
	for name, store := range newStores(t) {
		store := store
		t.Run(name, func(t *testing.T) {
			ep := &endpoint.Endpoint{Name: "site", Group: "core", URL: "https://example.com"}
			r1 := &endpoint.Result{Success: true, Duration: 100 * time.Millisecond, Timestamp: time.Now()}
			r2 := &endpoint.Result{Success: false, Duration: 200 * time.Millisecond, Timestamp: time.Now().Add(time.Second)}
			if err := store.InsertEndpointResult(ep, r1); err != nil {
				t.Fatalf("insert r1: %v", err)
			}
			if err := store.InsertEndpointResult(ep, r2); err != nil {
				t.Fatalf("insert r2: %v", err)
			}
			params := paging.NewEndpointStatusParams().WithResults(1, 20).WithEvents(1, 20)
			status, err := store.GetEndpointStatusByKey(ep.Key(), params)
			if err != nil {
				t.Fatalf("get by key: %v", err)
			}
			if got, want := len(status.Results), 2; got != want {
				t.Errorf("results: got %d want %d", got, want)
			}
			if len(status.Events) < 2 {
				t.Errorf("expected at least 2 events on flip, got %d", len(status.Events))
			}
			all, err := store.GetAllEndpointStatuses(params)
			if err != nil {
				t.Fatalf("get all: %v", err)
			}
			if len(all) != 1 {
				t.Errorf("all: got %d want 1", len(all))
			}
			from := time.Now().Add(-2 * time.Hour)
			to := time.Now().Add(2 * time.Hour)
			if _, err := store.GetUptimeByKey(ep.Key(), from, to); err != nil {
				t.Errorf("uptime: %v", err)
			}
			if _, err := store.GetAverageResponseTimeByKey(ep.Key(), from, to); err != nil {
				t.Errorf("avg: %v", err)
			}
			if _, err := store.GetHourlyAverageResponseTimeByKey(ep.Key(), from, to); err != nil {
				t.Errorf("hourly avg: %v", err)
			}
			if newer, err := store.HasEndpointStatusNewerThan(ep.Key(), time.Now().Add(-time.Hour)); err != nil || !newer {
				t.Errorf("newer: %v %v", newer, err)
			}
			a := &alert.Alert{Type: alert.TypeSlack}
			if err := store.UpsertTriggeredEndpointAlert(ep, a); err != nil {
				t.Errorf("upsert alert: %v", err)
			}
			if err := store.DeleteTriggeredEndpointAlert(ep, a); err != nil {
				t.Errorf("delete alert: %v", err)
			}
			if removed := store.DeleteAllEndpointStatusesNotInKeys([]string{"keep-me"}); removed < 1 {
				t.Errorf("delete not in keys: removed=%d", removed)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	c := &storage.Config{Type: storage.TypeSQLite}
	if err := c.ValidateAndSetDefaults(); err == nil {
		t.Errorf("expected error: sqlite without path")
	}
	c = &storage.Config{Type: storage.TypeMemory, Path: "x"}
	if err := c.ValidateAndSetDefaults(); err == nil {
		t.Errorf("expected error: memory with path")
	}
	c = &storage.Config{}
	if err := c.ValidateAndSetDefaults(); err != nil {
		t.Errorf("default: %v", err)
	}
	if c.Type != storage.TypeMemory {
		t.Errorf("default type want memory got %s", c.Type)
	}
	if c.MaximumNumberOfResults != 100 {
		t.Errorf("default maximum-results want 100 got %d", c.MaximumNumberOfResults)
	}
}
