# middleware

## Purpose

`middleware` contains transport-only HTTP middleware.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- adding request/response wrappers
- enforcing ordering-sensitive transport behavior
- instrumenting requests without owning business policy

## Do not use this module for

- business validation
- tenant policy catalogs
- tenant resolution or tenant quota enforcement
- service construction
- API version negotiation
- resource or protocol transformation

## First files to read

- `middleware/module.yaml`
- `docs/CANONICAL_STYLE_GUIDE.md`
- the target package under `middleware/*`

## Canonical change shape

- implement `func(http.Handler) http.Handler`
- keep one constructor path per middleware package; delete parallel wrapper families
- keep stable middleware packages single-purpose; split unrelated transport behaviors into separate packages instead of umbrella buckets
- new configurable middleware should prefer `Middleware(Config)` with `Config.WithDefaults()` or `DefaultConfig()`; existing package-specific stable constructors such as `compression.Gzip(GzipConfig)`, `timeout.Timeout(TimeoutConfig)`, `ratelimit.AbuseGuard(AbuseGuardConfig)`, and `debug.DebugErrors(DebugErrorConfig)` remain the canonical public names for those packages
- if a stable package uses an exported `Config` or `Options` type without an exported default helper, document the exception and keep the defaulting local to the constructor; current intentional exceptions are `cors.CORSOptions`, `compression.GzipConfig`, and `timeout.TimeoutConfig`
- use `accesslog.Middleware(...)` as the canonical access-log constructor
- for middleware that must preserve a panicking legacy constructor, expose an `E` variant such as `accesslog.MiddlewareE(...)` or `recovery.RecoveryE(...)` so new callers can handle invalid dependencies without panic
- add ordering and error-path tests
- keep side effects explicit and local
- keep tenant-aware policy, resolution, and quota behavior in `x/tenant`
- keep auth and security-header transport adapters here, on top of `security/*` primitives
- keep stable rate limiting here as a thin `middleware/ratelimit.AbuseGuard(...)` adapter over `security/abuse`, not as a catalog of limiter implementations
- keep API version negotiation in `x/rest/versioning`
- keep protocol or payload adaptation in `x/gateway/*`
- keep request-id generation policy in middleware-owned packages; `contract` should only carry request-id context/header contracts
- customize request-id generation through `requestid.WithGenerator(...)`; do not rely on mutable package-level generator state
- when `requestid.Middleware()` is absent, stable observability middleware may still stamp a fallback request id so access logs, tracing, and HTTP metrics share one correlation path

## Boundary with observability

- stable `middleware/*` owns transport-only observability primitives such as request IDs, tracing hooks, access logging, and HTTP metrics
- `x/observability` owns broader adapter, export, and integration wiring
- do not turn stable `middleware` into an observability catch-all catalog

## Production Security Profile

Plumego does not provide a hidden "production mode" bundle. Wire the stack
explicitly in application code so ordering and dependencies stay reviewable.

Recommended baseline order:

1. `requestid.Middleware(...)` for correlation.
2. `recovery.Recovery(app.Logger())` to convert panics into structured errors.
3. `bodylimit.BodyLimit(maxBytes, app.Logger())` for request body caps.
4. `timeout.Timeout(timeout.TimeoutConfig{...})` for bounded request runtime.
5. `middleware/security.SecurityHeaders(policy)` for response hardening.
6. `ratelimit.AbuseGuard(ratelimit.AbuseGuardConfig{...})` for transport abuse limits.
7. `auth.Authenticate(...)` and `auth.Authorize(...)` only on protected route groups or handlers.
8. `httpmetrics.Middleware(...)`, `tracing.Middleware(...)`, and `accesslog.Middleware(...)` for transport observability.

Keep tenant resolution, quota, and tenant policy in `x/tenant`. Keep exporter
and telemetry backend wiring in `x/observability`. The stable middleware layer
should remain a set of explicit transport wrappers, not a business policy
catalog.

## Internal transport primitives

`middleware/internal/transport` contains shared response-writing helpers used across middleware packages:

- `EnsureNoSniff(h http.Header)` — sets `X-Content-Type-Options: nosniff` unless already present
- `SafeWrite(w http.ResponseWriter, body []byte)` — writes body and sets the nosniff header; silently no-ops for nil writers
- `ClientIP(r *http.Request)` — extracts client IP from `X-Forwarded-For`, `X-Real-IP`, or `RemoteAddr` in that order
- `ResponseRecorder` — wraps an `http.ResponseWriter` to capture status code, body, and bytes written
- `BufferedResponse` — buffers the full response body with an optional max-bytes overflow guard; supports `WriteTo` for deferred flushing

These are internal; import them only from within the `middleware` module.
