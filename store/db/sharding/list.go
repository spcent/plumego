package sharding

import (
	"fmt"
)

// ListStrategy implements list-based (explicit mapping) sharding.
// Each possible key value is explicitly mapped to a specific shard.
//
// This strategy is ideal for:
//   - Categorical data with known values (country codes, tenant IDs)
//   - Enum-like fields with limited distinct values
//   - Scenarios requiring manual control over data placement
//
// Advantages:
//   - Complete control over data placement
//   - Easy to balance load manually
//   - Supports any comparable key type
//
// Disadvantages:
//   - Requires maintaining the mapping table
//   - Not suitable for high-cardinality keys
//   - Adding new values requires configuration updates
type ListStrategy struct {
	// mapping stores the explicit key -> shard mappings
	mapping map[any]int

	// defaultShard is used when a key is not in the mapping (-1 means error)
	defaultShard int
}

// NewListStrategy creates a new list-based sharding strategy.
// The mapping parameter defines explicit key-to-shard mappings.
// Keys not in the mapping will cause an error unless a default shard is set.
//
// Example:
//
//	strategy := NewListStrategy(map[any]int{
//	    "US": 0,
//	    "EU": 1,
//	    "CN": 2,
//	    "JP": 3,
//	})
func NewListStrategy(mapping map[any]int) *ListStrategy {
	return &ListStrategy{
		mapping:      mapping,
		defaultShard: -1, // No default shard
	}
}

// NewListStrategyWithDefault creates a list strategy with a default shard
// for keys that are not explicitly mapped.
//
// Example:
//
//	strategy := NewListStrategyWithDefault(map[any]int{
//	    "US": 0,
//	    "EU": 1,
//	}, 2)  // All other countries go to shard 2
func NewListStrategyWithDefault(mapping map[any]int, defaultShard int) *ListStrategy {
	return &ListStrategy{
		mapping:      mapping,
		defaultShard: defaultShard,
	}
}

// Shard looks up the shard index for the given key in the mapping table.
//
// Example:
//
//	shardIdx, err := strategy.Shard("EU", 4)
//	// Returns 1 (as mapped)
func (l *ListStrategy) Shard(key any, numShards int) (int, error) {
	if err := validateShardKey(key); err != nil {
		return 0, err
	}
	if err := validateNumShards(numShards); err != nil {
		return 0, err
	}

	// Look up the key in the mapping
	shardIdx, exists := l.mapping[key]
	if !exists {
		// Key not found - use default shard if configured
		if l.defaultShard >= 0 {
			shardIdx = l.defaultShard
		} else {
			return 0, fmt.Errorf("%w: key %v", ErrNoMatchingList, key)
		}
	}

	// Validate shard index
	if shardIdx < 0 || shardIdx >= numShards {
		return 0, fmt.Errorf("mapped shard index %d out of range [0, %d)", shardIdx, numShards)
	}

	return shardIdx, nil
}

// ShardRange returns all shards that contain keys in the given list.
// For list-based sharding, we cannot efficiently determine which shards
// contain values in a range, so this returns all shards.
//
// If you need range query support, consider using RangeStrategy instead.
func (l *ListStrategy) ShardRange(start, end any, numShards int) ([]int, error) {
	if err := validateNumShards(numShards); err != nil {
		return nil, err
	}

	// For list-based sharding, we cannot determine which shards contain
	// values in a range without knowing all possible keys.
	// Return all shards to be safe.
	shards := make([]int, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = i
	}
	return shards, nil
}

// Name returns the strategy name.
func (l *ListStrategy) Name() string {
	return "list"
}

// AddMapping adds or updates a key-to-shard mapping.
// This is useful for dynamically updating the strategy configuration.
func (l *ListStrategy) AddMapping(key any, shard int) {
	l.mapping[key] = shard
}

// RemoveMapping removes a key from the mapping table.
// After removal, the key will use the default shard if configured,
// or return an error.
func (l *ListStrategy) RemoveMapping(key any) {
	delete(l.mapping, key)
}

// GetMapping returns the current key-to-shard mappings.
// Returns a copy to prevent external modification.
func (l *ListStrategy) GetMapping() map[any]int {
	result := make(map[any]int, len(l.mapping))
	for k, v := range l.mapping {
		result[k] = v
	}
	return result
}

// SetDefaultShard sets the default shard for unmapped keys.
// Set to -1 to disable default shard (unmapped keys will error).
func (l *ListStrategy) SetDefaultShard(shard int) {
	l.defaultShard = shard
}

// GetDefaultShard returns the current default shard setting.
// Returns -1 if no default shard is configured.
func (l *ListStrategy) GetDefaultShard() int {
	return l.defaultShard
}
