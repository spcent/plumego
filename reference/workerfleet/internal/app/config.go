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
}

type MongoConfig struct {
	URI              string
	Database         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
}

func DefaultConfig() Config {
	return Config{
		StoreBackend: StoreBackendMemory,
		Mongo: MongoConfig{
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 10 * time.Second,
		},
		Retention: store.DefaultRetention,
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

	return cfg, ValidateConfig(cfg)
}

func ValidateConfig(cfg Config) error {
	switch cfg.StoreBackend {
	case "", StoreBackendMemory:
		return nil
	case StoreBackendMongo:
		if strings.TrimSpace(cfg.Mongo.URI) == "" {
			return errors.New("WORKERFLEET_MONGO_URI is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
		if strings.TrimSpace(cfg.Mongo.Database) == "" {
			return errors.New("WORKERFLEET_MONGO_DATABASE is required when WORKERFLEET_STORE_BACKEND=mongo")
		}
		return nil
	default:
		return fmt.Errorf("unsupported WORKERFLEET_STORE_BACKEND %q", cfg.StoreBackend)
	}
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
