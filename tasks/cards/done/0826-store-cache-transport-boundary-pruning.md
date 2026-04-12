# Card 0826

Priority: P1
State: done
Primary Module: store
Owned Files:
- `store/cache/cache.go`
- `store/cache/http_helpers.go`
- `store/cache/leaderboard.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
Depends On:

Goal:
- Restore `store/cache` to a transport-agnostic primitive by removing HTTP middleware ownership and feature-specific leaderboard behavior from the stable store root.
- Reset `store/cache` to one clear primitive surface, without leaving HTTP or leaderboard compatibility APIs behind.

Problem:
- `store/cache/cache.go` currently embeds HTTP cache-middleware behavior (`Cached`, `CachedWithConfig`, request-derived cache keys, cached response serialization) directly into the stable store primitive.
- `store/cache/http_helpers.go` duplicates response-writing helpers that already exist under middleware internals, which is direct transport leakage into `store`.
- `store/cache/leaderboard.go` adds a leaderboard-specific sorted-set surface on top of the stable cache primitive, and the package comments still advertise distributed and leaderboard use cases as if they were core stable responsibilities.

Scope:
- Remove or relocate HTTP response caching helpers and request-derived cache middleware out of stable `store/cache`.
- Remove or relocate leaderboard-specific sorted-set behavior out of the stable cache primitive surface.
- Tighten the store manifest and module primer so they describe only the stable cache responsibilities left after the pruning.
- Delete the removed stable entrypoints instead of preserving forwarding wrappers in `store/cache`.

Non-goals:
- Do not add new distributed cache implementations to the stable root.
- Do not remove the basic cache contract or the in-memory primitive without a replacement stable home.
- Do not introduce feature-specific HTTP caching conventions in `middleware`.
- Do not keep HTTP or leaderboard APIs in stable `store/cache` for backward compatibility.

Files:
- `store/cache/cache.go`
- `store/cache/http_helpers.go`
- `store/cache/leaderboard.go`
- `store/module.yaml`
- `docs/modules/store/README.md`

Tests:
- `go test -timeout 20s ./store/... ./x/cache/...`
- `go test -race -timeout 60s ./store/... ./x/cache/...`
- `go vet ./store/... ./x/cache/...`

Docs Sync:
- Keep the store manifest and primer aligned on the rule that stable `store/cache` owns only transport-agnostic cache primitives.

Done Definition:
- Stable `store/cache` no longer owns HTTP response caching adapters or leaderboard-specific APIs.
- Any relocated cache behavior lands in an owning extension package with explicit docs.
- The store primer and manifest no longer describe stable cache as a feature bucket.
- Removed transport and leaderboard APIs have zero residual references under `store/cache`.

Outcome:
- Completed.
- Removed HTTP response caching, request-derived cache helpers, and ranked-data behavior from stable `store/cache`.
- Moved transport- and feature-owned cache behavior to extension layers so stable store cache ownership stays primitive-only.
