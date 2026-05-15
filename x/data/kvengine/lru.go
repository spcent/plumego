package kvengine

import "sync/atomic"

// LRU management for shard
func (kv *KVStore) moveToFront(shard *Shard, entry *Entry) {
	if shard.lruHead == entry {
		return
	}

	// Remove from current position
	if prev := shard.prev[entry]; prev != nil {
		shard.next[prev] = shard.next[entry]
	}
	if next := shard.next[entry]; next != nil {
		shard.prev[next] = shard.prev[entry]
	} else {
		shard.lruTail = shard.prev[entry]
	}

	// Add to front
	shard.prev[entry] = nil
	shard.next[entry] = shard.lruHead
	if shard.lruHead != nil {
		shard.prev[shard.lruHead] = entry
	}
	shard.lruHead = entry
	if shard.lruTail == nil {
		shard.lruTail = entry
	}
}

func (kv *KVStore) evictLRU(shard *Shard) {
	if shard.lruTail == nil {
		return
	}

	entry := shard.lruTail
	kv.deleteFromShard(shard, entry.Key, entry)
	// No need to log to WAL - LRU evictions are a cache management strategy

	atomic.AddInt64(&kv.evictions, 1)
}

func (kv *KVStore) evictOneLRUSkip(skipKey string) bool {
	for _, shard := range kv.shards {
		shard.mu.Lock()
		if shard.lruTail != nil && shard.lruTail.Key != skipKey {
			kv.evictLRU(shard)
			shard.mu.Unlock()
			return true
		}
		shard.mu.Unlock()
	}
	return false
}

func (kv *KVStore) deleteFromShard(shard *Shard, key string, entry *Entry) {
	// Remove from map
	delete(shard.data, key)

	// Remove from LRU
	if prev := shard.prev[entry]; prev != nil {
		shard.next[prev] = shard.next[entry]
	} else {
		shard.lruHead = shard.next[entry]
	}
	if next := shard.next[entry]; next != nil {
		shard.prev[next] = shard.prev[entry]
	} else {
		shard.lruTail = shard.prev[entry]
	}

	delete(shard.next, entry)
	delete(shard.prev, entry)

	// Update counters
	atomic.AddInt64(&kv.entries, -1)
	atomic.AddInt64(&kv.memoryUsage, -entry.Size)
}
