# Card 0014

Priority: P2
State: done
Primary Module: router
Owned Files:
  - router/types.go
  - router/cache.go

Depends On: —

Goal:
The `router` package exported several symbols that are only used internally, polluting the public API surface:

1. **`MatchResult`** (router/types.go:8): `matchRoute` returns `*MatchResult`, but that method is
   package-private (lowercase). `MatchResult` itself is never directly referenced by code outside the
   package. Exporting a type that callers can never reach is meaningless.

2. **`DefaultPatternCacheSize = 50`** (router/cache.go:13): this constant is used internally in the
   `cache.go` constructor; no external callers exist, yet it was exported, misleading users into
   thinking it was configurable.

3. Additionally, `RouterOption` (router.go:86) and `RouteOption` (registration.go) coexist in the
   router package with inconsistent naming style — the former is a middleware injection option, the
   latter is a route registration option; their purpose and layer are different, yet the similar
   naming causes confusion.

Scope:
- Rename `MatchResult` to `matchResult` (unexported), update all intra-package references
  (router/cache.go, router/dispatch.go, router/matcher.go)
- Rename `DefaultPatternCacheSize` to `defaultPatternCacheSize`, update intra-package references
- If `DefaultPoolSliceCap` (pool.go) is likewise only used internally, make it unexported too
- Add a comment at the top of router/types.go explaining the difference in purpose between
  RouterOption and RouteOption, eliminating naming confusion (if renaming is warranted, handle it
  as a separate future card)
- Before executing: `grep -rn "router\.MatchResult\|router\.DefaultPatternCacheSize" . --include="*.go"`
  to confirm no external references exist

Non-goals:
- Do not change the route matching algorithm or performance characteristics
- Do not rename RouterOption / RouteOption (only add clarifying comments)
- Do not modify the router package's public API (Handle, Group, etc.)

Files:
  - router/types.go (MatchResult → matchResult)
  - router/cache.go (DefaultPatternCacheSize → defaultPatternCacheSize, update references)
  - router/dispatch.go (update matchResult references)
  - router/matcher.go (update matchResult references)
  - router/pool.go (check DefaultPoolSliceCap)

Tests:
  - go build ./router/...
  - go test ./router/...

Docs Sync: —

Done Definition:
- `grep -rn "router\.MatchResult" . --include="*.go"` returns empty
- `grep -rn "router\.DefaultPatternCacheSize" . --include="*.go"` returns empty
- `go build ./...` passes (confirms no external dependencies broken)
- `go test ./router/...` passes

Outcome:
- All three were already unexported: `matchResult` (lowercase), `defaultPatternCacheSize` (lowercase), `defaultPoolSliceCap` (lowercase)
- All done conditions were already met at verification time
