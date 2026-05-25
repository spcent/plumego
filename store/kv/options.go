package kvstore

import (
	"errors"
	"strings"
)

const (
	defaultMaxEntries  = 100000
	defaultMaxMemoryMB = 200
)

// Options configures the stable embedded KV primitive.
type Options struct {
	// DataDir is the explicit directory where the state file is stored.
	DataDir     string `json:"data_dir"`
	MaxEntries  int    `json:"max_entries"`
	MaxMemoryMB int    `json:"max_memory_mb"`
}

// DefaultOptions returns the default stable KV options for the caller-selected
// data directory. DataDir remains explicit so callers choose where filesystem
// state is created.
func DefaultOptions(dataDir string) Options {
	opts := Options{DataDir: dataDir}
	setDefaults(&opts)
	return opts
}

func setDefaults(opts *Options) {
	opts.DataDir = strings.TrimSpace(opts.DataDir)
	if opts.MaxEntries == 0 {
		opts.MaxEntries = defaultMaxEntries
	}
	if opts.MaxMemoryMB == 0 {
		opts.MaxMemoryMB = defaultMaxMemoryMB
	}
}

func validateOptions(opts Options) error {
	if strings.TrimSpace(opts.DataDir) == "" {
		return errors.New("kv: data dir is required")
	}
	if opts.MaxEntries <= 0 {
		return errors.New("kv: max entries must be positive")
	}
	if opts.MaxMemoryMB <= 0 {
		return errors.New("kv: max memory must be positive")
	}
	return nil
}
