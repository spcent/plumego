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
	if cfg.Profile != ProfileDevelopment {
		t.Fatalf("profile = %q, want %q", cfg.Profile, ProfileDevelopment)
	}
	if cfg.Retention != 7*24*time.Hour {
		t.Fatalf("retention = %v, want 7d", cfg.Retention)
	}
	if !cfg.Metrics.ExperimentalSeriesEnabled {
		t.Fatalf("experimental metrics should default enabled in dev profile")
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
		"WORKERFLEET_KUBE_SYNC_ENABLED":                "true",
		"WORKERFLEET_STATUS_SWEEP_ENABLED":             "true",
		"WORKERFLEET_ALERT_EVALUATION_ENABLED":         "true",
		"WORKERFLEET_NOTIFICATION_ENABLED":             "true",
		"WORKERFLEET_KUBE_SYNC_INTERVAL":               "11s",
		"WORKERFLEET_STATUS_SWEEP_INTERVAL":            "12s",
		"WORKERFLEET_ALERT_EVALUATION_INTERVAL":        "13s",
		"WORKERFLEET_NOTIFICATION_DELIVERY_INTERVAL":   "14s",
		"WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT":        "15s",
		"WORKERFLEET_LOOP_LEASE_TTL":            "45s",
		"WORKERFLEET_LOOP_LEASE_OWNER":          "workerfleet-test",
		"WORKERFLEET_KUBE_API_HOST":             "https://kube.example",
		"WORKERFLEET_KUBE_BEARER_TOKEN":         "token",
		"WORKERFLEET_KUBE_NAMESPACE":            "sim",
		"WORKERFLEET_KUBE_LABEL_SELECTOR":       "app=worker",
		"WORKERFLEET_KUBE_WORKER_CONTAINER":     "worker",
		"WORKERFLEET_FEISHU_WEBHOOK_URL":        "https://feishu.example/hook",
		"WORKERFLEET_WEBHOOK_URL":               "https://webhook.example/hook",
		"WORKERFLEET_WEBHOOK_HEADERS":           "X-Test=one,X-Token=two",
		"WORKERFLEET_WORKER_AUTH_TOKEN":         "worker-secret",
		"WORKERFLEET_ADMIN_AUTH_TOKEN":          "admin-secret",
		"WORKERFLEET_QUERY_AUTH_REQUIRED":       "true",
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
	if cfg.Runtime.NotificationDeliveryInterval != 14*time.Second {
		t.Fatalf("notification delivery interval = %v, want 14s", cfg.Runtime.NotificationDeliveryInterval)
	}
	if cfg.Runtime.NotifierDeliveryTimeout != 15*time.Second {
		t.Fatalf("notifier delivery timeout = %v, want 15s", cfg.Runtime.NotifierDeliveryTimeout)
	}
	if cfg.Runtime.LoopLeaseTTL != 45*time.Second || cfg.Runtime.LoopLeaseOwner != "workerfleet-test" {
		t.Fatalf("loop lease settings not parsed: ttl=%v owner=%q", cfg.Runtime.LoopLeaseTTL, cfg.Runtime.LoopLeaseOwner)
	}
	if cfg.Kube.APIHost != "https://kube.example" || cfg.Kube.Namespace != "sim" || cfg.Kube.LabelSelector != "app=worker" || cfg.Kube.WorkerContainer != "worker" {
		t.Fatalf("kube settings not parsed: %#v", cfg.Kube)
	}
	if cfg.Notifier.FeishuWebhookURL == "" || cfg.Notifier.WebhookURL == "" || cfg.Notifier.WebhookHeaders["X-Test"] != "one" || cfg.Notifier.WebhookHeaders["X-Token"] != "two" {
		t.Fatalf("notifier settings not parsed: %#v", cfg.Notifier)
	}
	if cfg.WorkerAuth.Token != "worker-secret" {
		t.Fatalf("worker auth token was not parsed")
	}
	if cfg.AdminAuth.Token != "admin-secret" || !cfg.AdminAuth.Required {
		t.Fatalf("admin auth not parsed: %#v", cfg.AdminAuth)
	}
}

func TestLoadConfigParsesStatusAndAlertPolicies(t *testing.T) {
	cfg, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_PROFILE":                        "prod",
		"WORKERFLEET_STATUS_STALE_AFTER":             "70s",
		"WORKERFLEET_STATUS_OFFLINE_AFTER":           "140s",
		"WORKERFLEET_STATUS_STAGE_STUCK_AFTER":       "20m",
		"WORKERFLEET_STATUS_RESTART_BURST_THRESHOLD": "6",
		"WORKERFLEET_ALERT_STAGE_STUCK_AFTER":        "25m",
		"WORKERFLEET_ALERT_RESTART_BURST_THRESHOLD":  "8",
		"WORKERFLEET_WORKER_AUTH_TOKEN":              "worker-secret",
		"WORKERFLEET_ADMIN_AUTH_TOKEN":               "admin-secret",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Profile != ProfileProduction {
		t.Fatalf("profile = %q, want %q", cfg.Profile, ProfileProduction)
	}
	if cfg.Policy.Status.StaleAfter != 70*time.Second {
		t.Fatalf("status stale after = %v, want 70s", cfg.Policy.Status.StaleAfter)
	}
	if cfg.Policy.Status.OfflineAfter != 140*time.Second {
		t.Fatalf("status offline after = %v, want 140s", cfg.Policy.Status.OfflineAfter)
	}
	if cfg.Policy.Status.StageStuckAfter != 20*time.Minute {
		t.Fatalf("status stage stuck after = %v, want 20m", cfg.Policy.Status.StageStuckAfter)
	}
	if cfg.Policy.Status.RestartBurstThreshold != 6 {
		t.Fatalf("status restart threshold = %d, want 6", cfg.Policy.Status.RestartBurstThreshold)
	}
	if cfg.Policy.Alert.StageStuckAfter != 25*time.Minute {
		t.Fatalf("alert stage stuck after = %v, want 25m", cfg.Policy.Alert.StageStuckAfter)
	}
	if cfg.Policy.Alert.RestartBurstThreshold != 8 {
		t.Fatalf("alert restart threshold = %d, want 8", cfg.Policy.Alert.RestartBurstThreshold)
	}
}

func TestLoadConfigParsesExperimentalMetricFlags(t *testing.T) {
	cfg, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_PROFILE":                      "prod",
		"WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED": "true",
		"WORKERFLEET_WORKER_AUTH_TOKEN":            "worker-secret",
		"WORKERFLEET_ADMIN_AUTH_TOKEN":             "admin-secret",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Metrics.ExperimentalSeriesEnabled {
		t.Fatalf("experimental metrics should be enabled by explicit flag")
	}

	cfg, err = LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_PROFILE":           "prod",
		"WORKERFLEET_WORKER_AUTH_TOKEN": "worker-secret",
		"WORKERFLEET_ADMIN_AUTH_TOKEN":  "admin-secret",
	}))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Metrics.ExperimentalSeriesEnabled {
		t.Fatalf("experimental metrics should default disabled in prod profile")
	}
}

func TestLoadConfigRejectsInvalidRuntimeInterval(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_KUBE_SYNC_INTERVAL": "0s",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_KUBE_SYNC_INTERVAL")
}

func TestLoadConfigRejectsRetentionOverflow(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_RETENTION_DAYS": "9223372036854775807",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_RETENTION_DAYS")
}

func TestLoadConfigRejectsInvalidRuntimeFlag(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_NOTIFICATION_ENABLED": "maybe",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_NOTIFICATION_ENABLED")
}

func TestLoadConfigRejectsUnsafeLoopLeaseTTL(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_LOOP_LEASE_TTL": "5s",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_LOOP_LEASE_TTL")
}

func TestLoadConfigRejectsEmptyLoopLeaseOwnerWhenSet(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_LOOP_LEASE_OWNER": " ",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_LOOP_LEASE_OWNER")
}

func TestLoadConfigRejectsEnabledKubeSyncWithoutWorkerContainer(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_KUBE_SYNC_ENABLED":     "true",
		"WORKERFLEET_KUBE_WORKER_CONTAINER": " ",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_KUBE_WORKER_CONTAINER")
}

func TestLoadConfigRejectsInvalidWebhookHeaders(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_WEBHOOK_HEADERS": "bad-entry",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_WEBHOOK_HEADERS")
}

func TestLoadConfigRejectsEmptyWorkerAuthTokenWhenSet(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_WORKER_AUTH_TOKEN": " ",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_WORKER_AUTH_TOKEN")
}

func TestLoadConfigRejectsEmptyAdminAuthTokenWhenSet(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_ADMIN_AUTH_TOKEN": " ",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_ADMIN_AUTH_TOKEN")
}

func TestLoadConfigRejectsProductionWithoutWorkerAuthToken(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_PROFILE":          "prod",
		"WORKERFLEET_ADMIN_AUTH_TOKEN": "admin-secret",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_WORKER_AUTH_TOKEN")
}

func TestLoadConfigRejectsProductionWithoutAdminAuthToken(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_PROFILE":           "prod",
		"WORKERFLEET_WORKER_AUTH_TOKEN": "worker-secret",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_ADMIN_AUTH_TOKEN")
}

func TestLoadConfigRejectsRequiredQueryAuthWithoutAdminAuthToken(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_QUERY_AUTH_REQUIRED": "true",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_ADMIN_AUTH_TOKEN")
}

func TestLoadConfigRejectsEnabledNotificationsWithoutSink(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_NOTIFICATION_ENABLED": "true",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_FEISHU_WEBHOOK_URL or WORKERFLEET_WEBHOOK_URL")
}

func TestLoadConfigRejectsMongoWithoutURI(t *testing.T) {
	_, err := LoadConfig(testLookup(map[string]string{
		"WORKERFLEET_STORE_BACKEND":  "mongo",
		"WORKERFLEET_MONGO_DATABASE": "workerfleet",
	}))
	assertConfigErrorMentions(t, err, "WORKERFLEET_MONGO_URI")
}

func TestValidateConfigRejectsUnsafeThresholds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Policy.Status.StaleAfter = 4 * time.Second
	err := ValidateConfig(cfg)
	assertConfigErrorMentions(t, err, "WORKERFLEET_STATUS_STALE_AFTER")

	cfg = DefaultConfig()
	cfg.Policy.Alert.StageStuckAfter = 20 * time.Second
	err = ValidateConfig(cfg)
	assertConfigErrorMentions(t, err, "WORKERFLEET_ALERT_STAGE_STUCK_AFTER")
}

func assertConfigErrorMentions(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want mention of %s", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want mention of %s", err, want)
	}
}

func testLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
