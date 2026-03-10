# Plumego v1 Core API Freeze List

Status: Active (v1.x)  
Last updated: 2026-03-10  
Scope: `core`, `router`, `middleware`, `contract`, top-level `plumego` convenience exports

This document defines the v1 API freeze baseline and compatibility tiers.

## v1 GA Scope Decision (2026-03-10)

Decision: adopt **Option A** for v1.0 GA production guarantees.

- GA production guarantee includes canonical core runtime modules:
  - `core`, `router`, `middleware`, `contract`, `security`, `store`
- `tenant/*` and `net/mq/*` remain `experimental` in v1.0 and are excluded from
  GA stability guarantees.
- Experimental modules may be used in production by advanced adopters, but API
  and behavior compatibility is not guaranteed at v1.0 level.
- Stabilization target for these experimental modules is tracked as post-v1.0
  work (v1.1+ stream).

## Stability Tiers

- `canonical`: Primary v1 API. No breaking behavior/signature changes in v1.x.
- `compatibility`: Supported in v1.x for compatibility or convenience, but not the primary style for new code.
- `experimental`: Not part of the v1 stable contract. Behavior/signatures may change before promotion.

## Canonical v1 APIs (Frozen)

### `core` package

- `core.New(options ...Option) *App`
- `(*App).Use(middlewares ...middleware.Middleware) error`
- `(*App).AddRoute(method, path string, handler http.Handler) error`
- `(*App).AddRouteWithName(method, path, name string, handler http.Handler) error`
- `(*App).Get(path string, handler http.HandlerFunc)`
- `(*App).Post(path string, handler http.HandlerFunc)`
- `(*App).Put(path string, handler http.HandlerFunc)`
- `(*App).Delete(path string, handler http.HandlerFunc)`
- `(*App).Patch(path string, handler http.HandlerFunc)`
- `(*App).Boot() error`
- `(*App).ServeHTTP(w http.ResponseWriter, r *http.Request)`
- `(*App).Router() *router.Router`
- `(*App).Logger() log.StructuredLogger`
- `(*App).Register(runner Runner) error`
- `(*App).OnShutdown(hook ShutdownHook) error`

### `core` options

- `WithRouter`
- `WithAddr`
- `WithEnvPath`
- `WithShutdownTimeout`
- `WithServerTimeouts`
- `WithMaxHeaderBytes`
- `WithHTTP2`
- `WithTLS`
- `WithTLSConfig`
- `WithDebug`
- `WithLogger`
- `WithComponent`
- `WithComponents`
- `WithMethodNotAllowed`
- `WithShutdownHook`
- `WithShutdownHooks`
- `WithRunner`
- `WithRunners`
- `WithMetricsCollector`
- `WithTracer`
- `WithHealthManager`

### `router` package (core route composition)

- `router.NewRouter`
- `(*Router).Use`
- `(*Router).Group`
- `(*Router).GroupFunc`
- `(*Router).Get/Post/Put/Delete/Patch/Any`
- `(*Router).GetNamed/PostNamed/PutNamed/DeleteNamed/PatchNamed/AnyNamed`
- `(*Router).AddRoute`
- `(*Router).AddRouteWithOptions`
- `(*Router).AddRouteWithName`
- `(*Router).SetMethodNotAllowed`
- `(*Router).MethodNotAllowedEnabled`
- `(*Router).Freeze`
- `(*Router).URL`
- `(*Router).URLMust`
- `(*Router).HasRoute`
- `(*Router).NamedRoutes`

### `middleware` + `contract` canonical style contracts

- `type middleware.Middleware func(http.Handler) http.Handler`
- `middleware.Apply`
- `contract.WriteResponse`
- `contract.WriteError`
- `contract.Param`
- `contract.RequestContextFrom`

## Compatibility APIs (Supported, Non-Canonical)

These remain supported in v1.x but are not the primary style in canonical examples.

### `core.App` compatibility/convenience methods

- `(*App).Handle(pattern string, handler http.Handler)`
- `(*App).HandleFunc(pattern string, handler http.HandlerFunc)`
- `(*App).Any(path string, handler http.HandlerFunc)`
- `(*App).GetNamed(name, path string, handler http.HandlerFunc)`
- `(*App).PostNamed(name, path string, handler http.HandlerFunc)`
- `(*App).PutNamed(name, path string, handler http.HandlerFunc)`
- `(*App).DeleteNamed(name, path string, handler http.HandlerFunc)`
- `(*App).PatchNamed(name, path string, handler http.HandlerFunc)`
- `(*App).AnyNamed(name, path string, handler http.HandlerFunc)`

### Top-level `plumego` package exports

The top-level package re-exports many `core` and optional module types/functions for convenience.
In v1 this is supported as `compatibility`/`convenience`, while canonical documentation prefers explicit package ownership (`core`, `router`, `middleware`, `contract`).

## Experimental (Not Frozen for v1 Stable Contract)

- `tenant/*` (feature set is still marked experimental in project docs)
- Modules explicitly marked experimental in docs/reports (for example parts of `net/mq` ecosystem and optional advanced extensions)

## Freeze Rules

1. Canonical APIs must not break in v1.x.
2. Compatibility APIs must remain callable in v1.x unless a critical security issue requires emergency change.
3. Experimental APIs may evolve before promotion; changes must be documented.
4. New APIs must not introduce a second canonical style for handlers/routes/errors.
