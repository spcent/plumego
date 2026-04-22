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

func TestLoadConfigParsesRuntimeSettings(t *testing.T) {
	cfg, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_KUBE_SYNC_ENABLED":         "true",
		"WORKERFLEET_STATUS_SWEEP_ENABLED":      "true",
		"WORKERFLEET_ALERT_EVALUATION_ENABLED":  "true",
		"WORKERFLEET_NOTIFICATION_ENABLED":      "true",
		"WORKERFLEET_KUBE_SYNC_INTERVAL":        "11s",
		"WORKERFLEET_STATUS_SWEEP_INTERVAL":     "12s",
		"WORKERFLEET_ALERT_EVALUATION_INTERVAL": "13s",
		"WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT": "14s",
		"WORKERFLEET_KUBE_API_HOST":             "https://kube.example",
		"WORKERFLEET_KUBE_BEARER_TOKEN":         "token",
		"WORKERFLEET_KUBE_NAMESPACE":            "sim",
		"WORKERFLEET_KUBE_LABEL_SELECTOR":       "app=worker",
		"WORKERFLEET_KUBE_WORKER_CONTAINER":     "worker",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Runtime.KubeSyncEnabled || !cfg.Runtime.StatusSweepEnabled || !cfg.Runtime.AlertEvaluationEnabled || !cfg.Runtime.NotificationEnabled {
		t.Fatalf("runtime flags not parsed: %#v", cfg.Runtime)
	}
	if cfg.Runtime.KubeSyncInterval != 11*time.Second {
		t.Fatalf("kube sync interval = %v, want 11s", cfg.Runtime.KubeSyncInterval)
	}
	if cfg.Runtime.StatusSweepInterval != 12*time.Second {
		t.Fatalf("status sweep interval = %v, want 12s", cfg.Runtime.StatusSweepInterval)
	}
	if cfg.Runtime.AlertEvaluationInterval != 13*time.Second {
		t.Fatalf("alert evaluation interval = %v, want 13s", cfg.Runtime.AlertEvaluationInterval)
	}
	if cfg.Runtime.NotifierDeliveryTimeout != 14*time.Second {
		t.Fatalf("notifier delivery timeout = %v, want 14s", cfg.Runtime.NotifierDeliveryTimeout)
	}
	if cfg.Kube.APIHost != "https://kube.example" || cfg.Kube.Namespace != "sim" || cfg.Kube.LabelSelector != "app=worker" || cfg.Kube.WorkerContainer != "worker" {
		t.Fatalf("kube settings not parsed: %#v", cfg.Kube)
	}
}

func TestLoadConfigRejectsInvalidRuntimeInterval(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_KUBE_SYNC_INTERVAL": "0s",
	}))
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_KUBE_SYNC_INTERVAL") {
		t.Fatalf("error = %v, want invalid kube sync interval", err)
	}
}

func TestLoadConfigRejectsInvalidRuntimeFlag(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_NOTIFICATION_ENABLED": "maybe",
	}))
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_NOTIFICATION_ENABLED") {
		t.Fatalf("error = %v, want invalid notification flag", err)
	}
}

func TestLoadConfigRejectsEnabledKubeSyncWithoutWorkerContainer(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_KUBE_SYNC_ENABLED":     "true",
		"WORKERFLEET_KUBE_WORKER_CONTAINER": " ",
	}))
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_KUBE_WORKER_CONTAINER") {
		t.Fatalf("error = %v, want missing worker container", err)
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
