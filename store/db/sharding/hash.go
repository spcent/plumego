package sharding

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

// HashStrategy implements consistent hash-based sharding using FNV-1a algorithm.
// This strategy provides good data distribution and is suitable for scenarios
// where you need even load distribution across shards.
//
// Supported key types: string, int, int8, int16, int32, int64,
// uint, uint8, uint16, uint32, uint64, []byte
type HashStrategy struct {
	// hashFunc is the hash function used to compute the hash value.
	// Defaults to FNV-1a if not specified.
	hashFunc func([]byte) uint64
}

// NewHashStrategy creates a new hash-based sharding strategy.
// Uses FNV-1a hash algorithm by default, which provides good distribution
// and performance characteristics for most use cases.
func NewHashStrategy() *HashStrategy {
	return &HashStrategy{
		hashFunc: fnv1aHash,
	}
}

// NewHashStrategyWithFunc creates a hash strategy with a custom hash function.
// This allows you to use alternative hash algorithms if needed.
func NewHashStrategyWithFunc(hashFunc func([]byte) uint64) *HashStrategy {
	return &HashStrategy{
		hashFunc: hashFunc,
	}
}

// Shard calculates the target shard index using hash-based sharding.
// The hash of the key is computed and then mapped to a shard using modulo.
//
// Example:
//
//	strategy := NewHashStrategy()
//	shardIdx, err := strategy.Shard("user123", 4)
//	// Returns 0-3 based on hash(user123) % 4
func (h *HashStrategy) Shard(key any, numShards int) (int, error) {
	if err := validateShardKey(key); err != nil {
		return 0, err
	}
	if err := validateNumShards(numShards); err != nil {
		return 0, err
	}

	bytes, err := keyToBytes(key)
	if err != nil {
		return 0, err
	}

	hash := h.hashFunc(bytes)
	return int(hash % uint64(numShards)), nil
}

// ShardRange calculates which shards might contain data in the given range.
// Since hash-based sharding doesn't preserve ordering, range queries typically
// need to query all shards. This method returns all shard indices.
//
// Example:
//
//	shards, err := strategy.ShardRange("user100", "user200", 4)
//	// Returns [0, 1, 2, 3] because hash doesn't preserve order
func (h *HashStrategy) ShardRange(start, end any, numShards int) ([]int, error) {
	if err := validateNumShards(numShards); err != nil {
		return nil, err
	}

	// Hash-based sharding doesn't preserve ordering, so range queries
	// need to check all shards
	shards := make([]int, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = i
	}
	return shards, nil
}

// Name returns the strategy name.
func (h *HashStrategy) Name() string {
	return "hash"
}

// fnv1aHash computes FNV-1a hash of the input bytes.
// FNV-1a is chosen for its simplicity, speed, and good distribution properties.
func fnv1aHash(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

// keyToBytes converts various key types to byte slices for hashing.
func keyToBytes(key any) ([]byte, error) {
	switch v := key.(type) {
	case string:
		return []byte(v), nil

	case []byte:
		return v, nil

	case int:
		return int64ToBytes(int64(v)), nil
	case int8:
		return int64ToBytes(int64(v)), nil
	case int16:
		return int64ToBytes(int64(v)), nil
	case int32:
		return int64ToBytes(int64(v)), nil
	case int64:
		return int64ToBytes(v), nil

	case uint:
		return uint64ToBytes(uint64(v)), nil
	case uint8:
		return uint64ToBytes(uint64(v)), nil
	case uint16:
		return uint64ToBytes(uint64(v)), nil
	case uint32:
		return uint64ToBytes(uint64(v)), nil
	case uint64:
		return uint64ToBytes(v), nil

	default:
		return nil, fmt.Errorf("%w: unsupported type %T", ErrInvalidShardKey, key)
	}
}

// int64ToBytes converts an int64 to bytes using little-endian encoding.
func int64ToBytes(n int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(n))
	return buf
}

// uint64ToBytes converts a uint64 to bytes using little-endian encoding.
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}
