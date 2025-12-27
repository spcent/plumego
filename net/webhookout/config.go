package webhookout

import (
	"time"

	"github.com/spcent/plumego/config"
)

type DropPolicy string

const (
	DropNewest     DropPolicy = "drop_newest"
	BlockWithLimit DropPolicy = "block_timeout"
	FailFast       DropPolicy = "fail_fast"
)

type Config struct {
	Enabled bool

	QueueSize  int
	Workers    int
	DrainMax   time.Duration
	DropPolicy DropPolicy
	BlockWait  time.Duration

	DefaultTimeout    time.Duration
	DefaultMaxRetries int
	BackoffBase       time.Duration
	BackoffMax        time.Duration
	RetryOn429        bool

	// For safety (recommended)
	AllowPrivateNetwork bool
}

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

	if cfg.QueueSize < 1 {
		cfg.QueueSize = 1
	}
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}
	if cfg.BlockWait < 0 {
		cfg.BlockWait = 0
	}
	switch cfg.DropPolicy {
	case DropNewest, BlockWithLimit, FailFast:
	default:
		cfg.DropPolicy = BlockWithLimit
	}
	return cfg
}
