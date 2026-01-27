# Router Contract

This document defines the stable, documented semantics for the Plumego router.
It freezes behavior so agents and users can safely rely on the contract.

## Scope
- Applies to `router.Router` request matching and dispatch.
- Preserves `net/http` handler semantics.
- Does not define business logic or middleware behavior beyond ordering.

## Route Matching Order
- Segment precedence is: static > param (`:id`) > wildcard (`*path`).
- Matching is left-to-right across path segments.
- Wildcard captures the remainder of the path and stops further matching.

## Method Resolution
- The router selects the tree for the request method.
- If no method tree exists, it falls back to the `ANY` tree.
- If no match is found in the method tree, it will try the `ANY` tree.
- If still no match is found, it returns `404 Not Found`.
- `405 Method Not Allowed` is disabled by default. Enable with `router.WithMethodNotAllowed(true)` to return `405` and set `Allow`.

## Path Normalization
- Request path uses `req.URL.Path` as-is (no URL decoding).
- Root path is always `/`.
- Leading and trailing slashes are trimmed for non-root paths, so `/a` and `/a/` are equivalent.
- Internal duplicate slashes are not normalized (e.g. `/a//b` is treated literally and will not match `/a/:id`).

## Parameter Extraction
- Param values come from `req.URL.Path` (the router does not perform additional decoding).
- `net/http` decodes percent-escapes in `URL.Path` (e.g. `%20` becomes space).
- Param keys are derived from route definitions in order.
- If duplicate param keys exist, later values overwrite earlier ones.
- Wildcard params include embedded slashes (e.g. `/files/*path` → `path = "a/b/c.txt"`).
- Percent-encoded slashes (`%2F`) are decoded by `net/http`, so a single-segment param will not match; wildcard params will capture the remainder (`a/b`).
- Params are injected into the request context:
  - `contract.ParamsContextKey` (map of params)
  - `contract.RequestContextKey` (RequestContext with Params)
- Router params override existing params stored in the request context.
- When a route matches, the router also sets:
  - `RequestContext.RoutePattern` (the matched route pattern)
  - `RequestContext.RouteName` (if provided)

## Cache Behavior
- The router caches match results using a normalized path key that strips extra slashes and trailing slash for non-root paths.
- Cache keys include request method and host.
- Cached handlers are invalidated whenever middleware registration changes (middleware version bump).
- Cache is an optimization only; it does not change matching semantics.

## Middleware Ordering (Router-Level)
- Global middlewares (`router.Use`) execute before group middlewares.
- Group middlewares execute from outer group to inner group, in the order added.
- Handler executes after all middleware layers.

## Conflict Rules (Registration-Time)
- Duplicate method+path registration fails.
- Different param names at the same path position are treated as conflicts.
- Only one wildcard is allowed per parent path segment.
