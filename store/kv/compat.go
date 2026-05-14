package kvstore

import "context"

// Exists reports whether a non-expired key exists. Errors are collapsed to
// false for backward compatibility; use ExistsContext when the caller needs
// errors.
func (kv *KVStore) Exists(key string) bool {
	exists, _ := kv.ExistsContext(context.Background(), key)
	return exists
}

// Keys returns all non-expired keys in sorted order. Errors are collapsed to an
// empty slice for backward compatibility; use KeysContext when the caller needs
// errors.
func (kv *KVStore) Keys() []string {
	keys, _ := kv.KeysContext(context.Background())
	return keys
}

// Size returns the number of non-expired keys. Errors are collapsed to zero for
// backward compatibility; use SizeContext when the caller needs errors.
func (kv *KVStore) Size() int {
	size, _ := kv.SizeContext(context.Background())
	return size
}

// GetStats returns point-in-time statistics for the store.
func (kv *KVStore) GetStats() Stats {
	stats, _ := kv.GetStatsContext(context.Background())
	return stats
}
