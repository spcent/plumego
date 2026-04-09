# Card 0855

Priority: P1
State: active
Primary Module: middleware
Owned Files:
- `middleware/ratelimit`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `security/abuse`

Goal:
- Converge stable rate limiting onto one canonical transport adapter path.
- Keep stable `middleware/ratelimit` as a thin transport layer over stable abuse primitives instead of a catalog of parallel algorithms.

Problem:
- `middleware/ratelimit` currently exposes three overlapping surfaces: `AbuseGuard(...)`, `TokenBucket(...)`, and `NewRateLimiter(...).Middleware()`, plus a separate key-function/store helper family.
- `AbuseGuard(...)` already wraps the stable `security/abuse.Limiter`, while `TokenBucket(...)` duplicates bucket logic and `RateLimiter` adds queue/backpressure lifecycle in the same stable package.
- This violates the repo rule that a middleware package should expose one canonical constructor path and leaves rate-limit policy ownership split across transport and security layers.

Scope:
- Keep one canonical stable middleware entrypoint for rate limiting.
- Remove duplicate algorithm/object surfaces from stable `middleware/ratelimit` in the same change; do not leave compatibility wrappers.
- Route the surviving primitive ownership through `security/abuse`, and move non-canonical transport/topology variants out of the stable middleware surface.
- Update tests, docs, and manifests to match the reduced package shape.

Non-goals:
- Do not redesign rate-limit response headers or canonical transport error shape.
- Do not redesign `security/abuse.Limiter` semantics unless required by the convergence.
- Do not add a second stable rate-limit package.

Files:
- `middleware/ratelimit`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `security/abuse`

Tests:
- `go test -timeout 20s ./middleware/ratelimit ./security/abuse`
- `go test -race -timeout 60s ./middleware/ratelimit ./security/abuse`
- `go vet ./middleware/ratelimit ./security/abuse`

Docs Sync:
- Keep middleware docs aligned on the rule that each middleware package should expose one canonical constructor path.
- Keep security docs aligned on the rule that abuse/rate-limit primitives live in `security/abuse`, while middleware only adapts them to HTTP.

Done Definition:
- Stable `middleware/ratelimit` exposes one canonical constructor path only.
- Duplicate token-bucket/object-lifecycle surfaces are removed from the stable middleware package.
- Security and middleware docs/manifests describe the same ownership split implemented in code.

Outcome:
- Completed.
- Reduced stable `middleware/ratelimit` to the single canonical `AbuseGuard(...)` HTTP adapter over `security/abuse.Limiter`.
- Removed the duplicate token-bucket implementation, in-memory limiter store helpers, and queue/backpressure `RateLimiter` lifecycle surface from the stable middleware package.
- Updated middleware docs and manifest guidance so rate limiting is described as a thin transport adapter over stable abuse primitives, not a catalog of limiter implementations.
