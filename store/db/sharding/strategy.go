// Package sharding provides database sharding strategies and utilities for
// horizontal data partitioning across multiple database instances.
package sharding

import (
	"errors"
	"fmt"
)

// Common errors returned by sharding operations.
var (
	ErrInvalidShardKey   = errors.New("sharding: invalid shard key type")
	ErrNoMatchingRange   = errors.New("sharding: no matching range for shard key")
	ErrNoMatchingList    = errors.New("sharding: no matching list entry for shard key")
	ErrInvalidNumShards  = errors.New("sharding: invalid number of shards")
	ErrNilShardKey       = errors.New("sharding: shard key cannot be nil")
	ErrRangeNotFound     = errors.New("sharding: range definition not found")
	ErrInvalidRangeOrder = errors.New("sharding: range definitions must be in order")
	ErrOverlappingRanges = errors.New("sharding: overlapping range definitions")
)

// Strategy defines the interface for all sharding strategies.
// A sharding strategy determines how data is distributed across multiple
// database shards based on a shard key value.
type Strategy interface {
	// Shard calculates the target shard index for a given shard key.
	// The key parameter is the value used to determine the shard (e.g., user_id).
	// The numShards parameter specifies the total number of shards available.
	// Returns the shard index (0-based) or an error if the key is invalid.
	Shard(key any, numShards int) (int, error)

	// ShardRange calculates which shard indices contain data within the given range.
	// This is useful for range queries that may span multiple shards.
	// The start and end parameters define the inclusive range boundaries.
	// Returns a slice of shard indices or an error if the range is invalid.
	ShardRange(start, end any, numShards int) ([]int, error)

	// Name returns a human-readable name for this strategy (e.g., "hash", "mod", "range").
	Name() string
}

// validateShardKey checks if the shard key is valid (not nil).
func validateShardKey(key any) error {
	if key == nil {
		return ErrNilShardKey
	}
	return nil
}

// validateNumShards checks if the number of shards is valid (> 0).
func validateNumShards(numShards int) error {
	if numShards <= 0 {
		return fmt.Errorf("%w: got %d, must be > 0", ErrInvalidNumShards, numShards)
	}
	return nil
}

// normalizeShardIndex ensures the shard index is within valid bounds [0, numShards).
func normalizeShardIndex(idx, numShards int) int {
	if idx < 0 {
		idx = -idx
	}
	return idx % numShards
}
