package sharding

import (
	"fmt"
)

// ModStrategy implements simple modulo-based sharding.
// This is the simplest sharding strategy, calculating shard index as: key % numShards
// It only supports integer keys and provides deterministic routing with minimal overhead.
//
// Advantages:
//   - Extremely fast (single modulo operation)
//   - Simple to understand and debug
//   - Deterministic routing
//
// Disadvantages:
//   - Only supports integer keys
//   - Difficult to scale (adding shards requires data migration)
//   - Data distribution depends on key distribution
//
// Supported key types: int, int8, int16, int32, int64,
// uint, uint8, uint16, uint32, uint64
type ModStrategy struct {
	// No configuration needed for basic modulo sharding
}

// NewModStrategy creates a new modulo-based sharding strategy.
func NewModStrategy() *ModStrategy {
	return &ModStrategy{}
}

// Shard calculates the target shard index using modulo operation.
// The shard index is computed as: abs(key) % numShards
//
// Example:
//
//	strategy := NewModStrategy()
//	shardIdx, err := strategy.Shard(12345, 4)
//	// Returns 1 because 12345 % 4 = 1
func (m *ModStrategy) Shard(key any, numShards int) (int, error) {
	if err := validateShardKey(key); err != nil {
		return 0, err
	}
	if err := validateNumShards(numShards); err != nil {
		return 0, err
	}

	id, err := keyToInt64(key)
	if err != nil {
		return 0, err
	}

	return normalizeShardIndex(int(id), numShards), nil
}

// ShardRange calculates which shards contain data in the given range.
// For modulo sharding, we need to check which shards the range boundaries
// map to and all shards in between.
//
// Example:
//
//	shards, err := strategy.ShardRange(10, 25, 4)
//	// For range [10, 25] with 4 shards:
//	// 10 % 4 = 2, 25 % 4 = 1
//	// Returns [0, 1, 2, 3] because the range wraps around
func (m *ModStrategy) ShardRange(start, end any, numShards int) ([]int, error) {
	if err := validateNumShards(numShards); err != nil {
		return nil, err
	}

	startID, err := keyToInt64(start)
	if err != nil {
		return nil, fmt.Errorf("invalid start key: %w", err)
	}

	endID, err := keyToInt64(end)
	if err != nil {
		return nil, fmt.Errorf("invalid end key: %w", err)
	}

	// Ensure start <= end
	if startID > endID {
		startID, endID = endID, startID
	}

	// If the range spans more than numShards values, all shards are affected
	if endID-startID >= int64(numShards) {
		shards := make([]int, numShards)
		for i := 0; i < numShards; i++ {
			shards[i] = i
		}
		return shards, nil
	}

	// Calculate which shards are affected
	shardSet := make(map[int]struct{})
	for id := startID; id <= endID; id++ {
		shardSet[normalizeShardIndex(int(id), numShards)] = struct{}{}
	}

	// Convert set to sorted slice
	shards := make([]int, 0, len(shardSet))
	for i := 0; i < numShards; i++ {
		if _, exists := shardSet[i]; exists {
			shards = append(shards, i)
		}
	}

	return shards, nil
}

// Name returns the strategy name.
func (m *ModStrategy) Name() string {
	return "mod"
}

// keyToInt64 converts various integer types to int64.
func keyToInt64(key any) (int64, error) {
	switch v := key.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		// Check for overflow
		if v > uint64(1<<63-1) {
			return 0, fmt.Errorf("%w: uint64 value %d too large for int64", ErrInvalidShardKey, v)
		}
		return int64(v), nil
	default:
		return 0, fmt.Errorf("%w: unsupported type %T (mod strategy requires integer keys)", ErrInvalidShardKey, key)
	}
}
