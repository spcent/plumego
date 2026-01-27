# Middleware Contract

This document defines the stable, documented semantics for Plumego middleware.
It freezes behavior so agents and users can safely rely on the contract.

## Scope
- Applies to `middleware.Middleware`, `middleware.Apply`, and `middleware.Chain`.
- Preserves `net/http` compatibility (`http.Handler` and `http.HandlerFunc`).

## Ordering Semantics
- Middleware execution order follows registration order.
- Example: `Apply(h, A, B)` executes as `A -> B -> h`.
- `Chain.Use(A).Use(B)` executes as `A -> B -> h`.

## Short-Circuiting
- A middleware may short-circuit by not calling `next.ServeHTTP`.
- Short-circuiting stops all downstream handlers.
- Middleware must be explicit when returning early (e.g., auth failure).

## Context Propagation
- Middleware must preserve `context.Context` and may attach request-scoped values.
- Standardized keys live in `contract` (e.g. `ContextWithPrincipal`).
- Middleware must not mutate shared global state unless documented.
- Middleware registration is expected to happen during startup before serving traffic; the registry is not goroutine-safe.

## Response Behavior
- Middleware should not write headers/body after calling `next` unless it owns the response.
- Middleware should avoid double-writes or conflicting status codes.

## Compatibility
- All middleware must accept and return `http.Handler`.
- Middleware should not assume a custom response writer type unless documented.
