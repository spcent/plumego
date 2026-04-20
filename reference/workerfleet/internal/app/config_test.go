package app

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaultsToMemory(t *testing.T) {
	cfg, err := LoadConfig(testLookup(nil))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.StoreBackend != StoreBackendMemory {
		t.Fatalf("store backend = %q, want memory", cfg.StoreBackend)
	}
	if cfg.Retention != 7*24*time.Hour {
		t.Fatalf("retention = %v, want 7d", cfg.Retention)
	}
}

func TestLoadConfigParsesMongoSettings(t *testing.T) {
	cfg, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_STORE_BACKEND":           "mongo",
		"WORKERFLEET_MONGO_URI":               "mongodb://127.0.0.1:27017",
		"WORKERFLEET_MONGO_DATABASE":          "workerfleet",
		"WORKERFLEET_MONGO_CONNECT_TIMEOUT":   "3s",
		"WORKERFLEET_MONGO_OPERATION_TIMEOUT": "4s",
		"WORKERFLEET_MONGO_MAX_POOL_SIZE":     "20",
		"WORKERFLEET_RETENTION_DAYS":          "3",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.StoreBackend != StoreBackendMongo {
		t.Fatalf("store backend = %q, want mongo", cfg.StoreBackend)
	}
	if cfg.Mongo.ConnectTimeout != 3*time.Second {
		t.Fatalf("connect timeout = %v, want 3s", cfg.Mongo.ConnectTimeout)
	}
	if cfg.Mongo.OperationTimeout != 4*time.Second {
		t.Fatalf("operation timeout = %v, want 4s", cfg.Mongo.OperationTimeout)
	}
	if cfg.Mongo.MaxPoolSize != 20 {
		t.Fatalf("max pool size = %d, want 20", cfg.Mongo.MaxPoolSize)
	}
	if cfg.Retention != 3*24*time.Hour {
		t.Fatalf("retention = %v, want 3d", cfg.Retention)
	}
}

func TestLoadConfigRejectsMongoWithoutURI(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_STORE_BACKEND":  "mongo",
		"WORKERFLEET_MONGO_DATABASE": "workerfleet",
	}))
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_MONGO_URI") {
		t.Fatalf("error = %v, want missing mongo uri", err)
	}
}

func testLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
