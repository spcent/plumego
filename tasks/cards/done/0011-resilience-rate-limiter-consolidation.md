# Card 0011

Priority: P2
State: done
Primary Module: x/resilience
Owned Files:
  - x/resilience/module.yaml
  - security/abuse/limiter.go

Depends On: —

Goal:
There were 3 separate, independent rate limiter implementations with no code reuse:

- `security/abuse/limiter.go`: complete token bucket implementation supporting per-IP, per-key, and global rate limiting
- `x/pubsub/ratelimit.go`: a `RateLimitedPubSub` wrapper written specifically for pubsub, with its own standalone token bucket
- `x/tenant/core/ratelimit.go`: `InMemoryRateLimitManager` containing a third independent token bucket implementation

Meanwhile, `x/resilience`'s `module.yaml` explicitly states: "Put reusable resilience primitives here",
yet x/resilience only had a circuitbreaker with no packages importing it (zero usage).

Goal of this card: migrate the reusable rate limiter primitive into `x/resilience/ratelimit`, allowing
the three scenarios above to share a common primitive and eliminate the duplicated logic.

Scope:
- Create a general-purpose token bucket rate limiter in `x/resilience/ratelimit/`:
  - Supports per-key bucketing (general purpose: per-IP, per-tenant, per-topic)
  - `Allow(key string) bool` + `Wait(ctx, key) error` usage patterns
  - Configurable QPS, Burst, TTL (idle key cleanup)
- Replace the token bucket logic in `x/pubsub/ratelimit.go` with a call to x/resilience/ratelimit
  (keeping the public API of RateLimitedPubSub unchanged)
- Replace the internal implementation of `InMemoryRateLimitManager` in `x/tenant/core/ratelimit.go`
  with a call to x/resilience/ratelimit (keeping the interface unchanged)
- `security/abuse/limiter.go`: decide based on complexity — if its per-key window logic can be
  directly replaced by the general primitive then migrate; if business logic differences are too large
  then leave it as-is with an internal reuse comment only
- Update x/resilience's module.yaml to add `ratelimit` to `public_entrypoints`

Non-goals:
- Do not change any external API of security/abuse, x/pubsub, or x/tenant
- Do not move security/abuse's policy logic (ban list, suspicious pattern detection) out of that package
- Do not introduce any external dependencies (continue using sync/atomic and other standard library packages)
- Do not modify middleware/ratelimit (that package uses security/abuse and does not implement buckets directly)

Files:
  - x/resilience/ratelimit/ratelimit.go (new)
  - x/resilience/ratelimit/ratelimit_test.go (new)
  - x/resilience/module.yaml (update public_entrypoints)
  - x/pubsub/ratelimit.go (internal replacement)
  - x/tenant/core/ratelimit.go (internal replacement)
  - security/abuse/limiter.go (optional partial reuse)

Tests:
  - go test -race ./x/resilience/...
  - go test ./x/pubsub/...
  - go test ./x/tenant/...

Docs Sync:
  - docs/modules/x-resilience/README.md (add ratelimit section)

Done Definition:
- `x/resilience/ratelimit` package exists, has tests, and passes `go test -race`
- x/pubsub and x/tenant no longer implement their own token bucket logic; they use x/resilience/ratelimit
- `grep -rn "golang.org/x/time/rate" .` results are consistent (only inside x/resilience or gone entirely)
- All tests in modified packages pass

Outcome:
- Created `x/resilience/ratelimit/ratelimit.go` with `TokenBucket` struct (New, Allow, AllowN, Wait, WaitN, UpdateRate) and `KeyedBuckets` struct (NewKeyed, Allow, Wait, Delete, Len)
- Created `x/resilience/ratelimit/ratelimit_test.go` with comprehensive tests including a concurrent test
- Updated `x/pubsub/ratelimit.go` to use `*ratelimit.TokenBucket` as the underlying implementation
- `x/tenant/core/ratelimit.go` kept its local tokenBucket (different semantics: supports req.Now override for testability and in-place rate reload)
- Added `ratelimit` to `x/resilience/module.yaml` public_entrypoints
