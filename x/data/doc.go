// Package data provides topology-heavy data capabilities that do not belong in
// the stable store layer.
//
// While store/* owns thin interfaces and in-process primitives, x/data owns
// the implementations that require external dependencies, distributed topology,
// provider configuration, or durable storage beyond stdlib.
//
// Sub-packages:
//
//   - cache: Distributed cache (consistent hashing, replication, Redis adapter, leaderboard)
//   - file: Cloud file storage (S3, signed URLs, image processing)
//   - idempotency: SQL- and Redis-backed idempotency providers
//   - kvengine: WAL-backed durable KV engine
//   - migrate: Database schema migration runner
//   - pgx: PostgreSQL (pgx driver) helpers
//   - sqlx: SQL query builder and result mapper
//   - rw: Read-write splitting orchestration
//   - sharding: Sharding and topology routing
//
// Decision guide — store vs. x/data:
//
//	Need                                    → Use
//	In-process TTL cache                    → store/cache
//	Distributed / Redis / sharded cache     → x/data/cache
//	Local file storage                      → store/file
//	Cloud storage, S3, signed URLs          → x/data/file
//	Simple in-memory KV                     → store/kv
//	Persistent WAL-backed KV engine         → x/data/kvengine
//	In-memory idempotency guard             → store/idempotency
//	SQL / Redis-backed idempotency          → x/data/idempotency
//	Read-write splitting, sharding          → x/data/rw, x/data/sharding
package data
