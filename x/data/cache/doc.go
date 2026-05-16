// Package cache provides extension-layer cache adapters and topology-heavy
// cache implementations. It builds on the stable store/cache abstractions.
//
// Sub-packages:
//
//   - distributed: consistent-hashing distributed cache with replication and failover
//   - leaderboard: ranked-data cache built on store/cache primitives
//   - redis: Redis client adapter implementing store/cache.Cache
package cache
