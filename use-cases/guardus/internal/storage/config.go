package storage

import "errors"

const (
	DefaultMaximumNumberOfResults = 100
	DefaultMaximumNumberOfEvents  = 50
)

var (
	ErrSQLStorageRequiresPath          = errors.New("sql storage requires a non-empty path to be defined")
	ErrMemoryStorageDoesNotSupportPath = errors.New("memory storage does not support persistence, use sqlite if you want persistence on file")
)

// Config controls how a Store is built.
type Config struct {
	// Path used by the store to achieve persistence. Required for TypeSQLite.
	Path string `json:"path"`

	// Type of store. Empty string defaults to TypeMemory.
	Type Type `json:"type"`

	// Caching enables a write-through cache for read latency. Only meaningful
	// for TypeSQLite.
	Caching bool `json:"caching,omitempty"`

	// MaximumNumberOfResults retained per endpoint.
	MaximumNumberOfResults int `json:"maximum-number-of-results,omitempty"`

	// MaximumNumberOfEvents retained per endpoint.
	MaximumNumberOfEvents int `json:"maximum-number-of-events,omitempty"`
}

// ValidateAndSetDefaults applies defaults and validates the configuration.
func (c *Config) ValidateAndSetDefaults() error {
	if c.Type == "" {
		c.Type = TypeMemory
	}
	if c.Type == TypeSQLite && len(c.Path) == 0 {
		return ErrSQLStorageRequiresPath
	}
	if c.Type == TypeMemory && len(c.Path) > 0 {
		return ErrMemoryStorageDoesNotSupportPath
	}
	if c.MaximumNumberOfResults <= 0 {
		c.MaximumNumberOfResults = DefaultMaximumNumberOfResults
	}
	if c.MaximumNumberOfEvents <= 0 {
		c.MaximumNumberOfEvents = DefaultMaximumNumberOfEvents
	}
	return nil
}
