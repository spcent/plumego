# Router Architecture

> **Package**: `github.com/spcent/plumego/router`
> **Audience**: Maintainers and contributors

This document defines the internal layering of Plumego's router and the
invariants that must be preserved when changing it.

The router has five internal layers:

1. `state`
2. `registration`
3. `dispatch`
4. `metadata`
5. `cache`

These layers are intentionally separate. New features should fit into one of
them instead of introducing parallel state or duplicate matching logic.

---

## 1. State

Source of truth:

- [`router.go`](/Users/bingrong.yan/projects/go/plumego/router/router.go)

Key type:

- `routerState`

Responsibilities:

- Own all shared mutable routing state
- Back both the root router and all route groups
- Hold the route trees, registered route list, named routes, metadata,
  validations, runtime cache, and router-wide config flags
- Provide the single lock boundary for shared routing data

Rules:

- `routerState` is the only shared mutable state for routing
- Route groups may add prefix and middleware layering, but they must not create
  their own copies of trees, metadata, validations, or cache
- If a feature needs shared routing data, add it to `routerState`
- Do not reintroduce per-group copies of shared maps guarded by different locks

Design intent:

- Root and groups must observe the same route graph
- A route registered through a group is not a separate router; it is a scoped
  view over the same router state

---

## 2. Registration

Primary files:

- [`router_registration.go`](/Users/bingrong.yan/projects/go/plumego/router/router_registration.go)
- [`router_api.go`](/Users/bingrong.yan/projects/go/plumego/router/router_api.go)

Responsibilities:

- Build and mutate the trie
- Normalize prefixes and full route patterns
- Detect duplicate and conflicting route registrations
- Attach route-level middleware snapshots
- Bind validation already known at registration time onto the final trie node
- Expose method helpers such as `Get`, `Post`, `Any`, `Handle`, `Resource`

Rules:

- Registration is the only place that may mutate trie structure
- Path normalization for stored route patterns must be consistent with metadata
  and validation storage
- Conflict detection must remain centralized here
- Registration may attach derived data to trie nodes, but it must not create a
  second routing index with its own matching semantics

Design intent:

- Matching behavior is defined by the trie built during registration
- Any later runtime behavior must derive from that trie, not bypass it

---

## 3. Dispatch

Primary file:

- [`router_dispatch.go`](/Users/bingrong.yan/projects/go/plumego/router/router_dispatch.go)

Responsibilities:

- Implement `http.Handler`
- Normalize incoming request paths
- Resolve route matches from trie or cache
- Apply 404 / 405 behavior
- Build parameter maps and request context
- Run validation for the matched route
- Apply middleware and invoke the final handler

Rules:

- Dispatch must not contain alternative route matching rules
- Dispatch may read metadata and validation associated with the matched route,
  but must not search for them using independent pattern matching
- Dispatch must treat the matched route as the source of truth for:
  - `RoutePattern`
  - `RouteMethod`
  - `Validation`
  - route-scoped middleware

Design intent:

- Requests should follow a single pipeline:
  `normalize path -> match route -> validate params -> apply middleware -> serve`

---

## 4. Metadata

Primary file:

- [`router_metadata.go`](/Users/bingrong.yan/projects/go/plumego/router/router_metadata.go)

Responsibilities:

- Store route metadata by `method -> pattern`
- Manage named routes and reverse URL generation
- Expose route inspection APIs such as `Routes()` and `Print()`

Rules:

- Metadata is keyed by normalized route pattern, not by request path
- Metadata lookups must never infer matches by reparsing request URLs
- Reverse routing must be derived from named route definitions, not from trie
  traversal

Design intent:

- Metadata is descriptive, not authoritative for request matching
- Matching determines the route; metadata enriches the matched route

---

## 5. Cache

Primary file:

- [`cache.go`](/Users/bingrong.yan/projects/go/plumego/router/cache.go)

Responsibilities:

- Cache match results for static and parameterized routes
- Preserve trie matching semantics while accelerating repeated lookups
- Track cache statistics

Rules:

- Cache is an optimization, never the source of truth
- Cache entries must be derived from trie matches only
- Cache lookup ordering must mirror trie specificity ordering
- Cache must not change route semantics
- Cache must remain concurrency-safe under parallel request load

Design intent:

- If cache is cleared, behavior must remain identical
- Cache may reduce work, but it must never introduce alternate routing results

---

## Validation Principles

Validation previously tended to drift because it was easy to maintain a
separate index of `method + pattern` rules and then perform another round of
runtime pattern matching.

That pattern is explicitly disallowed now.

Current rule:

- Validation is attached to the matched trie node and carried through
  `MatchResult.Validation`

Implications:

- Validation follows the same semantics as route matching
- Group prefixes, `ANY` routes, and cached parameterized routes all use the
  same validation decision
- There is no second runtime pattern matcher for validation

Do:

- Bind validation to normalized route patterns during registration or by
  updating the registered node
- Read validation from the final matched route during dispatch

Do not:

- Build a second validation trie
- Scan a list of patterns at request time
- Reconstruct validation decisions from `req.URL.Path` after matching

---

## Cache Principles

The router uses two cache forms:

- exact-path cache
- parameterized-pattern cache

The correctness rule is simple:

- cache lookup must produce the same route that a trie lookup would produce

Implications:

- Pattern cache ordering must preserve trie specificity
- A cache hit must still run route validation and middleware
- Cache entries must not embed request-specific parameter maps
- Middleware cache invalidation must remain tied to middleware versioning

Do:

- Cache `MatchResult`
- Extract fresh parameter values for parameterized requests
- Keep cache statistics scoped to one logical lookup

Do not:

- Cache validation outcomes per request path
- Cache request context objects
- Add host-, header-, or query-sensitive routing into cache keys unless the
  trie matching semantics also depend on them

---

## Anti-Patterns To Avoid

Do not reintroduce any of the following:

- Shared route maps guarded by different locks
- Group-local copies of route trees or route metadata
- Runtime validation pattern scanning
- Cache logic that can match a different route than the trie
- Metadata lookups based on raw request path instead of normalized route pattern
- Feature-specific side indexes that duplicate route ownership without a single
  source of truth

---

## Change Checklist

Before merging router changes, verify:

1. Which layer owns this change: `state`, `registration`, `dispatch`,
   `metadata`, or `cache`?
2. Does it add a second source of truth for route identity?
3. Does it preserve trie-first semantics?
4. Does it keep root and groups on the same shared state?
5. Does `go test -race ./router` still pass?
6. Does cache-cleared behavior match cache-enabled behavior?

If any answer is unclear, the design is probably too implicit.
