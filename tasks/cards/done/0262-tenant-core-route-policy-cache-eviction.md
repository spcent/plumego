# Card 0262

Priority: P2
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/route_policy_cache.go`
Depends On: —

Goal:
- Fix `InMemoryRoutePolicyCache.Set` to evict expired entries first when the cache is
  full instead of deleting an arbitrary random entry.

---

## Problem

**File:** `x/tenant/core/route_policy_cache.go` lines 74–79

```go
func (c *InMemoryRoutePolicyCache) Set(ctx context.Context, tenantID string, policy RoutePolicy) error {
    ...
    if len(c.entries) >= c.maxSize {
        for key := range c.entries {
            delete(c.entries, key)  // ← deletes one random entry (map iteration order is undefined)
            break
        }
    }
    ...
```

When the cache reaches capacity, Go's map iteration provides no ordering guarantees,
so this evicts an arbitrary entry. The evicted entry may be fresh (just written a moment
ago) while genuinely stale entries remain in the cache. Over time this leads to:

1. **Wasted work**: recently-resolved policies are evicted before they expire, forcing
   redundant provider calls.
2. **Missed cleanup**: expired entries continue to occupy cache slots even though they
   would be rejected at read time (`Get` checks `expiresAt`).

The `Get` path already performs lazy expiry deletion (lines 59–61), so the cache already
has an `expiresAt` per entry. Eviction at `Set` should prefer entries that have already
expired, then fall back to any entry only if nothing is expired.

---

## Fix

Replace the single-random-delete with an evict-expired-first pass:

```go
if len(c.entries) >= c.maxSize {
    now := time.Now().UTC()
    evicted := false
    for key, entry := range c.entries {
        if now.After(entry.expiresAt) {
            delete(c.entries, key)
            evicted = true
            break
        }
    }
    if !evicted {
        // No expired entry found; evict an arbitrary entry to stay within maxSize.
        for key := range c.entries {
            delete(c.entries, key)
            break
        }
    }
}
```

This is a minimal, correct fix that does not require introducing a heap or ordered
structure. A single pass through the map to find the first expired entry is O(n) in the
worst case but typical cache usage will hit expired entries early.

---

Scope:
- Fix `Set` eviction logic only.
- No change to `Get`, `Delete`, TTL handling, or constructor defaults.

Non-goals:
- Do not implement LRU eviction or any ordered eviction strategy.
- Do not add background eviction goroutines.
- Do not change the `RoutePolicyCache` interface.

Files:
- `x/tenant/core/route_policy_cache.go`
- `x/tenant/core/route_policy_test.go`

Tests:
- Add a test: fill cache to `maxSize` with expired entries, call `Set`, verify an
  expired entry was removed and the new entry is present.
- Add a test: fill cache to `maxSize` with fresh entries, call `Set`, verify the new
  entry is present and total size does not exceed `maxSize`.
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync: —

Done Definition:
- `Set` evicts an expired entry when one exists before falling back to random eviction.
- New tests verify expired-first and fresh-fallback behavior.
- All tests pass.

Outcome:
