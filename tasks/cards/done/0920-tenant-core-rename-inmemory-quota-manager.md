# Card 0920

Priority: P2
State: active
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota.go`
Depends On: 0915 (QuotaConfig legacy fields removal, not a hard dependency)

Goal:
- Rename `InMemoryQuotaManager` to `FixedWindowQuotaManager` so the name signals the
  single-window constraint that currently requires reading the comment to discover.

---

## Problem

**File:** `x/tenant/core/quota.go` line 54

```go
// InMemoryQuotaManager is a single-window in-memory quota manager.
// It respects QuotaConfig.Limits by using the first valid limit entry.
// For multi-window enforcement (hour + day + month simultaneously),
// use WindowQuotaManager with an InMemoryQuotaStore instead.
type InMemoryQuotaManager struct { ... }
```

The package has three quota manager implementations:

| Name | Algorithm | Windows | Storage |
|------|-----------|---------|---------|
| `InMemoryQuotaManager` | Fixed window | **First limit entry only** | In-memory |
| `SlidingWindowQuotaManager` | Sliding window | Minute only | In-memory (sync.Map) |
| `WindowQuotaManager` | Fixed window | All limits simultaneously | `QuotaStore` |

The name `InMemoryQuotaManager` describes the *storage medium* but says nothing about
the *window semantics*. Callers who see it alongside `WindowQuotaManager` and
`SlidingWindowQuotaManager` have no hint from the name alone that:
1. It enforces only the first entry in `Limits`.
2. Callers with hour + day + month limits must use `WindowQuotaManager` instead.

This creates a trap: a developer who sets
```go
QuotaConfig{
    Limits: []QuotaLimit{
        {Window: QuotaWindowMinute, Requests: 100},
        {Window: QuotaWindowDay, Requests: 5000},
    },
}
```
and uses `InMemoryQuotaManager` will only get minute enforcement. The daily limit is
silently ignored. The comment documents this, but names should not require reading
comments to understand fundamental constraints.

Renaming to `FixedWindowQuotaManager` makes the constraint explicit:
- "Fixed" contrasts with "Sliding" and signals the window-boundary reset model.
- "Window" (singular) indicates one active window, complementing `WindowQuotaManager`
  which enforces multiple.
- The storage medium ("InMemory") is less important than the algorithm contract.

---

## Fix

Rename the type and its constructor:

```go
// Before                                After
InMemoryQuotaManager                    FixedWindowQuotaManager
NewInMemoryQuotaManager(provider)       NewFixedWindowQuotaManager(provider)
```

**Migration:**
1. `grep -rn 'InMemoryQuotaManager\|NewInMemoryQuotaManager' . --include='*.go'`
2. Rename type and constructor in `quota.go`.
3. Update all callers, tests, and any documentation.

---

Scope:
- Rename type and constructor only. No behavior change.
- Update all references in the same PR.

Non-goals:
- Do not change the single-window behavior.
- Do not add multi-window support to this type.
- Do not rename `InMemoryQuotaStore` (storage, not a manager — the name is appropriate).
- Do not rename `WindowQuotaManager` or `SlidingWindowQuotaManager`.

Files:
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_test.go`
- All caller files identified by grep.

Tests:
- `grep -rn 'InMemoryQuotaManager\|NewInMemoryQuotaManager' . --include='*.go'` returns empty after migration.
- `go build ./...`
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync:
- Update the QuotaConfig struct comment if it references `InMemoryQuotaManager` by name.

Done Definition:
- `FixedWindowQuotaManager` and `NewFixedWindowQuotaManager` exist; `InMemoryQuotaManager` does not.
- `grep -rn 'InMemoryQuotaManager' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
