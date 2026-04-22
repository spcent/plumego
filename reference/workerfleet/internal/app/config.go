package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"workerfleet/internal/platform/store"
)

const (
	StoreBackendMemory = "memory"
	StoreBackendMongo  = "mongo"
)

type Config struct {
	StoreBackend string
	Mongo        MongoConfig
	Retention    time.Duration
	Runtime      RuntimeConfig
	Kube         KubeConfig
}

type MongoConfig struct {
	URI              string
	Database         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
}

type RuntimeConfig struct {
	KubeSyncEnabled         bool
	StatusSweepEnabled      bool
	AlertEvaluationEnabled  bool
	NotificationEnabled     bool
	KubeSyncInterval        time.Duration
	StatusSweepInterval     time.Duration
	AlertEvaluationInterval time.Duration
	NotifierDeliveryTimeout time.Duration
}

type KubeConfig struct {
	APIHost         string
	BearerToken     string
	Namespace       string
	LabelSelector   string
	WorkerContainer string
}

func DefaultConfig() Config {
	return Config{
		StoreBackend: StoreBackendMemory,
		Mongo: MongoConfig{
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 10 * time.Second,
		},
		Retention: store.DefaultRetention,
		Runtime: RuntimeConfig{
			KubeSyncInterval:        30 * time.Second,
			StatusSweepInterval:     30 * time.Second,
			AlertEvaluationInterval: 30 * time.Second,
			NotifierDeliveryTimeout: 5 * time.Second,
		},
		Kube: KubeConfig{
			WorkerContainer: "worker",
		},
	}
}

func LoadConfigFromEnv() (Config, error) {
	return LoadConfig(os.LookupEnv)
}

func LoadConfig(lookup func(string) (string, bool)) (Config, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	cfg := DefaultConfig()
	if value, ok := lookup("WORKERFLEET_STORE_BACKEND"); ok {
		cfg.StoreBackend = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := lookup("WORKERFLEET_MONGO_URI"); ok {
		cfg.Mongo.URI = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_MONGO_DATABASE"); ok {
		cfg.Mongo.Database = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_MONGO_CONNECT_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_MONGO_CONNECT_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.ConnectTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_MONGO_OPERATION_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_MONGO_OPERATION_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.OperationTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_MONGO_MAX_POOL_SIZE"); ok {
		poolSize, err := parseUintEnv("WORKERFLEET_MONGO_MAX_POOL_SIZE", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Mongo.MaxPoolSize = poolSize
	}
	if value, ok := lookup("WORKERFLEET_RETENTION_DAYS"); ok {
		days, err := parseUintEnv("WORKERFLEET_RETENTION_DAYS", value)
		if err != nil {
			return Config{}, err
		}
		if days == 0 {
			return Config{}, errors.New("WORKERFLEET_RETENTION_DAYS must be greater than zero")
		}
		cfg.Retention = time.Duration(days) * 24 * time.Hour
	}
	if value, ok := lookup("WORKERFLEET_KUBE_SYNC_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_KUBE_SYNC_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.KubeSyncEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_STATUS_SWEEP_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_STATUS_SWEEP_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.StatusSweepEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_ALERT_EVALUATION_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_ALERT_EVALUATION_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.AlertEvaluationEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_NOTIFICATION_ENABLED"); ok {
		enabled, err := parseBoolEnv("WORKERFLEET_NOTIFICATION_ENABLED", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.NotificationEnabled = enabled
	}
	if value, ok := lookup("WORKERFLEET_KUBE_SYNC_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_KUBE_SYNC_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.KubeSyncInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_STATUS_SWEEP_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_STATUS_SWEEP_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.StatusSweepInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_ALERT_EVALUATION_INTERVAL"); ok {
		interval, err := parseDurationEnv("WORKERFLEET_ALERT_EVALUATION_INTERVAL", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.AlertEvaluationInterval = interval
	}
	if value, ok := lookup("WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT"); ok {
		timeout, err := parseDurationEnv("WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT", value)
		if err != nil {
			return Config{}, err
		}
		cfg.Runtime.NotifierDeliveryTimeout = timeout
	}
	if value, ok := lookup("WORKERFLEET_KUBE_API_HOST"); ok {
		cfg.Kube.APIHost = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_BEARER_TOKEN"); ok {
		cfg.Kube.BearerToken = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_NAMESPACE"); ok {
		cfg.Kube.Namespace = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_LABEL_SELECTOR"); ok {
		cfg.Kube.LabelSelector = strings.TrimSpace(value)
	}
	if value, ok := lookup("WORKERFLEET_KUBE_WORKER_CONTAINER"); ok {
		cfg.Kube.WorkerContainer = strings.TrimSpace(value)
	}

	return cfg, ValidateConfig(cfg)
}

func ValidateConfig(cfg Config) error {
	switch cfg.StoreBackend {
	case "", StoreBackendMemory:
	case StoreBackendMongo:
		if strings.TrimSpace(cfg.Mongo.URI) == "" {
			return errors.New("WORKERFLEET_MONGO_URI is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
		if strings.TrimSpace(cfg.Mongo.Database) == "" {
			return errors.New("WORKERFLEET_MONGO_DATABASE is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
	default:
		return fmt.Errorf("unsupported WORKERFLEET_STORE_BACKEND %q", cfg.StoreBackend)
	}
	if cfg.Runtime.KubeSyncEnabled && strings.TrimSpace(cfg.Kube.WorkerContainer) == "" {
		return errors.New("WORKERFLEET_KUBE_WORKER_CONTAINER is required when WORKERFLEET_KUBE_SYNC_ENABLED=true")
	}
	return nil
}

func parseDurationEnv(name string, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return duration, nil
}

func parseUintEnv(name string, value string) (uint64, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}

func parseBoolEnv(name string, value string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}
