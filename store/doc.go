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
//   - kv: KV store interface and file-backed in-process KV implementation
//   - idempotency: Idempotency store interface and in-memory implementation
//
// For the store vs. x/data decision guide, see x/data package documentation.
package store
