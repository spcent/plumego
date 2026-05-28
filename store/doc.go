// Package store provides stable, transport-agnostic storage primitives.
//
// Each sub-package owns a thin interface and an in-process reference
// implementation backed only by the standard library. There are no external
// dependencies.
//
// Sub-packages:
//
//   - cache: Cache interface and MemoryCache (in-process, TTL-based)
//   - db: SQL helpers for *sql.DB wiring and query helpers
//   - file: FileStore interface and local filesystem implementation
//   - kv: KV interface and in-memory KV implementation
//   - idempotency: Idempotency store interface and in-memory implementation
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
package store
