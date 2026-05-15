package cache

import (
	"fmt"
	"time"
)

const (
	// DefaultMaxKeyLength is the default maximum key length.
	DefaultMaxKeyLength = 256

	// DefaultMaxMemoryUsage is the default maximum tracked payload bytes (0 = no limit).
	DefaultMaxMemoryUsage = 0

	// DefaultCleanupInterval is the default cleanup interval for expired items.
	DefaultCleanupInterval = 5 * time.Minute
)

// Config defines the configuration for cache backends.
type Config struct {
	// MaxKeyLength is the maximum allowed key length (0 = no limit).
	MaxKeyLength int

	// MaxMemoryUsage is the maximum tracked payload bytes (0 = no limit).
	MaxMemoryUsage uint64

	// CleanupInterval is the interval for cleaning up expired items.
	CleanupInterval time.Duration

	// DefaultTTL is the default time-to-live for items without explicit TTL.
	DefaultTTL time.Duration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		MaxKeyLength:    DefaultMaxKeyLength,
		MaxMemoryUsage:  DefaultMaxMemoryUsage,
		CleanupInterval: DefaultCleanupInterval,
		DefaultTTL:      10 * time.Minute,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.MaxKeyLength < 0 {
		return fmt.Errorf("%w: MaxKeyLength cannot be negative", ErrInvalidConfig)
	}
	if c.CleanupInterval < 0 {
		return fmt.Errorf("%w: CleanupInterval cannot be negative", ErrInvalidConfig)
	}
	if c.DefaultTTL < 0 {
		return fmt.Errorf("%w: DefaultTTL cannot be negative", ErrInvalidConfig)
	}
	return nil
}
