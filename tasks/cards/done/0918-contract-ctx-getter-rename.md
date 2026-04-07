# Card 0918

Priority: P3
State: active
Primary Module: contract
Owned Files:
- `contract/context_core.go`
Depends On: —

Goal:
- Rename `GetBodySize()` and `GetRequestDuration()` to drop the non-idiomatic `Get`
  prefix so they align with every other accessor method on `Ctx`.

---

## Problem

**File:** `contract/context_core.go` lines 382–388

```go
func (c *Ctx) GetRequestDuration() time.Duration {
    return time.Since(c.startedAt)
}

func (c *Ctx) GetBodySize() int64 {
    return c.bodySize.Load()
}
```

All other `Ctx` accessor methods use no `Get` prefix:

| Method | Pattern |
|--------|---------|
| `RequestID()` | no prefix |
| `IsAborted()` | `Is` prefix (boolean) |
| `IsCompressed()` | `Is` prefix (boolean) |
| `RequestHeaders()` | descriptive name |
| `CollectedErrors()` | descriptive name |
| `GetBodySize()` | **`Get` prefix — inconsistent** |
| `GetRequestDuration()` | **`Get` prefix — inconsistent** |

Go's standard library and the rest of the codebase use `Get` only when a method
explicitly corresponds to a map/store lookup (e.g., `ctx.Get(key)`). Standalone
accessors do not carry the prefix.

---

## Fix

```go
// Before                                 // After
func (c *Ctx) GetBodySize() int64          func (c *Ctx) BodySize() int64
func (c *Ctx) GetRequestDuration() time.Duration  func (c *Ctx) RequestDuration() time.Duration
```

**Migration:**
1. `grep -rn 'GetBodySize\|GetRequestDuration' . --include='*.go'` — collect all callers.
2. Rename the methods.
3. Update every caller in the same PR.

---

Scope:
- Rename the two methods and migrate all callers.
- No behavior change.

Non-goals:
- Do not rename `Get(key string)` or `MustGet(key string)` — those are map-lookup methods
  where the `Get` prefix correctly describes the operation.
- Do not add new `Ctx` methods.

Files:
- `contract/context_core.go`
- All files identified by the grep above (likely test files and middleware).

Tests:
- `grep -rn 'GetBodySize\|GetRequestDuration' . --include='*.go'` returns empty after migration.
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `BodySize()` and `RequestDuration()` exist; `GetBodySize` and `GetRequestDuration` do not.
- `grep -rn 'GetBodySize\|GetRequestDuration' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
