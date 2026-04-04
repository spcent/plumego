# Card 0505

Priority: P1

Goal:
- Reduce the three-layer error classification system (ErrorCategory / ErrorType
  / Code string) to two layers, and centralise the mapping logic so it cannot
  diverge.

Problem:
- Every error carries three representations of the same concept:
  - `ErrorCategory` (7 values) — coarse bucket
  - `ErrorType` (17 values) — fine-grained kind
  - `Code` string constant (15 values) — wire-format string
- The mapping between layers is scattered across `errors.go`:
  `CategoryForStatus()`, `HTTPStatusFromCategory()`, and inline switch
  statements inside convenience constructors.
- Not every `ErrorType` has a matching `Code` constant and vice-versa,
  making it easy to create partially populated errors.

Scope:
- Define a single lookup table (or method on `ErrorType`) that returns the
  canonical `(ErrorCategory, Code string, HTTPStatus int)` triple for each
  `ErrorType`.
- Replace the scattered `CategoryForStatus` / `HTTPStatusFromCategory` helpers
  with lookups into this table.
- Ensure every `ErrorType` maps to exactly one entry in the table.
- Remove `ErrorType` values that have no corresponding `Code` constant (or add
  the missing constants).

Non-goals:
- Do not change the `APIError` wire format (`code`, `category` fields stay).
- Do not merge `ErrorCategory` and `ErrorType` into one type — the coarse
  bucket is still useful for middleware routing decisions.

Files:
- `contract/errors.go`
- `contract/error_codes.go`
- `contract/error_utils.go`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- A single lookup table maps every `ErrorType` to its canonical Category, Code,
  and HTTP status.
- `CategoryForStatus` and `HTTPStatusFromCategory` delegate to this table.
- No `ErrorType` lacks a corresponding `Code` constant.
- All tests pass.
