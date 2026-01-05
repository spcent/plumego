package webhookout

import (
	"errors"
	"time"

	"github.com/spcent/plumego/config"
)

// DropPolicy defines the queue overflow behavior.
type DropPolicy string

const (
	// DropNewest drops the newest task when queue is full
	DropNewest DropPolicy = "drop_newest"
	// BlockWithLimit blocks on enqueue with timeout when queue is full
	BlockWithLimit DropPolicy = "block_timeout"
	// FailFast immediately returns error when queue is full
	FailFast DropPolicy = "fail_fast"
)

// Config holds webhook delivery service configuration.
type Config struct {
	// Service control
	Enabled bool

	// Queue and workers
	QueueSize  int
	Workers    int
	DrainMax   time.Duration
	DropPolicy DropPolicy
	BlockWait  time.Duration

	// Retry and timeout
	DefaultTimeout    time.Duration
	DefaultMaxRetries int
	BackoffBase       time.Duration
	BackoffMax        time.Duration
	RetryOn429        bool

	// Security
	AllowPrivateNetwork bool
}

// DefaultConfig returns production-ready defaults.
func DefaultConfig() Config {
	return Config{
		Enabled:             true,
		QueueSize:           2048,
		Workers:             8,
		DrainMax:            5 * time.Second,
		DropPolicy:          BlockWithLimit,
		BlockWait:           50 * time.Millisecond,
		DefaultTimeout:      5 * time.Second,
		DefaultMaxRetries:   6,
		BackoffBase:         500 * time.Millisecond,
		BackoffMax:          30 * time.Second,
		RetryOn429:          true,
		AllowPrivateNetwork: false,
	}
}

// ConfigFromEnv creates config from environment variables.
func ConfigFromEnv() Config {
	cfg := Config{
		Enabled:    config.GetBool("WEBHOOK_ENABLED", true),
		QueueSize:  config.GetInt("WEBHOOK_QUEUE_SIZE", 2048),
		Workers:    config.GetInt("WEBHOOK_WORKERS", 8),
		DrainMax:   config.GetDurationMs("WEBHOOK_DRAIN_MAX_MS", 5000),
		DropPolicy: DropPolicy(config.GetString("WEBHOOK_DROP_POLICY", string(BlockWithLimit))),
		BlockWait:  config.GetDurationMs("WEBHOOK_BLOCK_WAIT_MS", 50),

		DefaultTimeout:    config.GetDurationMs("WEBHOOK_DEFAULT_TIMEOUT_MS", 5000),
		DefaultMaxRetries: config.GetInt("WEBHOOK_DEFAULT_MAX_RETRIES", 6),
		BackoffBase:       config.GetDurationMs("WEBHOOK_BACKOFF_BASE_MS", 500),
		BackoffMax:        config.GetDurationMs("WEBHOOK_BACKOFF_MAX_MS", 30000),
		RetryOn429:        config.GetBool("WEBHOOK_RETRY_ON_429", true),

		AllowPrivateNetwork: config.GetBool("WEBHOOK_ALLOW_PRIVATE_NET", false),
	}

	// Validate and apply defaults
	if cfg.QueueSize < 1 {
		cfg.QueueSize = 1
	}
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}
	if cfg.BlockWait < 0 {
		cfg.BlockWait = 0
	}
	if cfg.BackoffBase <= 0 {
		cfg.BackoffBase = 500 * time.Millisecond
	}
	if cfg.BackoffMax <= 0 {
		cfg.BackoffMax = 30 * time.Second
	}
	if cfg.DrainMax <= 0 {
		cfg.DrainMax = 5 * time.Second
	}

	// Validate drop policy
	switch cfg.DropPolicy {
	case DropNewest, BlockWithLimit, FailFast:
	default:
		cfg.DropPolicy = BlockWithLimit
	}

	return cfg
}

// Validate checks if configuration is valid.
func (c Config) Validate() error {
	if c.QueueSize < 1 {
		return errors.New("queue_size must be at least 1")
	}
	if c.Workers < 1 {
		return errors.New("workers must be at least 1")
	}
	if c.DefaultTimeout <= 0 {
		return errors.New("default_timeout must be positive")
	}
	if c.DefaultMaxRetries < 0 {
		return errors.New("default_max_retries cannot be negative")
	}
	if c.BackoffBase <= 0 {
		return errors.New("backoff_base must be positive")
	}
	if c.BackoffMax <= 0 {
		return errors.New("backoff_max must be positive")
	}
	if c.BackoffBase > c.BackoffMax {
		return errors.New("backoff_base cannot be greater than backoff_max")
	}
	return nil
}
