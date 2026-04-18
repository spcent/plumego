# Card 0013

Priority: P2
State: done
Primary Module: store
Owned Files:
  - store/file/file.go
  - store/idempotency/store.go

Depends On: —

Goal:
`store/file` and `store/idempotency` are pure interface packages in the stable root (defining only
interfaces and types), with concrete implementations distributed under `x/data/`. However, neither
package has any comments or documentation explaining this design decision, nor any pointers to the
implementations, which leads to:

1. Developers who find the `store/file.Storage` interface don't know where to find the implementation
2. `store/idempotency/store_test.go` is only 14 lines, using only a `noopStore` to verify interface
   compilation — it has no coverage of the semantic correctness of `ErrNotFound`, `ErrInvalidKey`,
   `ErrExpired`, and other error sentinel constants
3. `store/file`'s test file (`coverage_test.go`) is an empty placeholder

Additionally, `store/cache/cache.go` contains a full implementation (MemoryCache), which is a
completely different style from the two pure-interface packages above, but there is no documentation
explaining why.

Scope:
- In `store/file/doc.go` (or top-of-file comment) explicitly state:
  - This package defines stable interfaces; implementations are in `x/data/file` (local, S3, etc.)
  - Reference `x/fileapi` as the HTTP handler layer
- In `store/idempotency/doc.go` or file comment, similarly state:
  - Implementations are in `x/data/idempotency` (sql, kv)
- Expand `store/idempotency/store_test.go` to test the following behaviors:
  - `StatusInProgress` and `StatusComplete` constants are reasonable (non-zero values)
  - `ErrNotFound`, `ErrExpired`, `ErrInvalidKey` sentinel errors are all non-nil and mutually distinct
  - `Record` struct field completeness (Key, Status, Response fields are assignable)
- Delete the `store/file/coverage_test.go` placeholder (or replace with meaningful compilation tests)
- In `store/cache/doc.go` or top-of-file comment, explain: this package contains an in-memory cache
  implementation (different positioning from the pure-interface packages store/file and store/idempotency)

Non-goals:
- Do not write new concrete implementations for store/file or store/idempotency
- Do not change any logic in x/data/file or x/data/idempotency
- Do not modify interface signatures

Files:
  - store/file/file.go (add package doc at top)
  - store/file/coverage_test.go (delete or replace)
  - store/idempotency/store.go (add package doc at top)
  - store/idempotency/store_test.go (add meaningful tests)
  - store/cache/cache.go (add positioning note at top)

Tests:
  - go test ./store/...

Docs Sync: —

Done Definition:
- `store/file` and `store/idempotency` package docs contain explicit pointers to the implementation packages
- `store/idempotency/store_test.go` verifies all sentinel errors are non-nil and mutually distinct
- `store/file` has no empty placeholder tests
- `go test ./store/...` passes

Outcome:
- `store/file/file.go` already has a package doc explaining implementations are in x/data/file and HTTP handlers in x/fileapi
- `store/idempotency/store.go` already has a package doc pointing to x/data/idempotency
- `store/idempotency/store_test.go` already has TestSentinelErrorsAreNonNilAndDistinct, TestStatusConstants, TestRecordFields
- `store/file/coverage_test.go` already has meaningful tests (TestError_Error_WithPath, TestFileStat_Zero, etc.)
- All done conditions were already met at verification time
