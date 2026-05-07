# middleware

## Purpose

`middleware` contains transport-only HTTP middleware.

## v1 Status

- `GA` in the Plumego v1 support matrix
- The stable package surface should stay small and expose one canonical
  constructor path per middleware.

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
- new configurable middleware should prefer `Middleware(Config)` with
  `Config.WithDefaults()` or `DefaultConfig()`; existing package-specific stable
  constructors such as `compression.Gzip(GzipConfig)`,
  `timeout.Timeout(TimeoutConfig)`, and
  `debug.DebugErrors(DebugErrorConfig)` remain the canonical public names for
  those packages
- if a stable package uses an exported `Config` or `Options` type without an exported default helper, document the exception and keep the defaulting local to the constructor; current intentional exceptions are `cors.CORSOptions`, `compression.GzipConfig`, and `timeout.TimeoutConfig`
- use `accesslog.Middleware(...)` as the canonical access-log constructor
- add ordering and error-path tests
- keep side effects explicit and local
- keep tenant-aware policy, resolution, and quota behavior in `x/tenant`
- keep auth and security-header transport adapters here, on top of `security/*` primitives
- keep stable rate limiting here as a thin `middleware/ratelimit` HTTP adapter over `security/abuse`, not as a catalog of limiter implementations
- rate limiting defaults must use the direct `RemoteAddr` peer IP; applications behind trusted proxies can opt into `X-Forwarded-For`/`X-Real-IP` by supplying an explicit `ratelimit.AbuseGuardConfig.KeyFunc`
- if a custom rate-limit `KeyFunc` returns an empty or all-whitespace key,
  `AbuseGuard` falls back to the direct `RemoteAddr` peer IP instead of sharing
  one global empty-key bucket
- production code that lets ratelimit create its own limiter should use
  `ratelimit.NewAbuseGuard(...)`, wire `guard.Middleware()`, and call
  `guard.Stop()` during application shutdown; callers that inject
  `AbuseGuardConfig.Limiter` own that limiter lifecycle themselves
- keep API version negotiation in `x/rest/versioning`
- keep protocol or payload adaptation in `x/gateway/*`
- keep request-id generation policy in middleware-owned packages; `contract` should only carry request-id context/header contracts
- customize request-id generation through `requestid.WithGenerator(...)`; `requestid.NewRequestID()` is a convenience default backed by package-local generator state
- treat request IDs as correlation identifiers only; the default generated ID is decodable and must not be used as a secret, token, nonce, or authorization value
- when `requestid.Middleware()` is absent, stable observability middleware may still stamp a fallback request id so access logs, tracing, and HTTP metrics share one correlation path

## Boundary with observability

- stable `middleware/*` owns transport-only observability primitives such as request IDs, tracing hooks, access logging, and HTTP metrics
- recommended production composition uses standalone `httpmetrics.Middleware(...)`
  and `tracing.Middleware(...)`, with `accesslog.Middleware(logger)` for
  access-log emission only
- the production reference service has a regression test that records one HTTP
  metric for one request, guarding against accidentally duplicating
  observability middleware for the same signal
- exporter catalogs, backend setup, sampling policy, and pipeline composition belong outside stable middleware
- `x/observability` owns broader adapter, export, and integration wiring
- do not turn stable `middleware` into an observability catch-all catalog
- middleware-owned instrumentation and logging callbacks are best-effort;
  callback panics are recovered so observability cannot replace a downstream
  handler panic or prevent canonical transport errors from being written
- tracing start failures are treated as missing tracing for that request; the
  downstream handler still runs
- accesslog fields use the shared middleware redaction policy before they are
  passed to the configured logger

## Production Security Profile

Plumego does not provide a hidden "production mode" bundle. Wire the stack
explicitly in application code so ordering and dependencies stay reviewable.

Recommended baseline order:

1. `requestid.Middleware(...)` for correlation.
2. `recovery.Recovery(app.Logger())` to convert panics from downstream middleware and handlers into structured errors.
3. `bodylimit.Middleware(bodylimit.Config{...})` for request body caps.
4. `timeout.Timeout(timeout.TimeoutConfig{...})` for bounded request runtime.
5. `middleware/security.SecurityHeaders(policy)` for response hardening; invalid custom header policies fail closed and do not call downstream handlers.
6. `ratelimit.NewAbuseGuard(...).Middleware()` for transport abuse limits when the limiter is middleware-owned.
7. `auth.Authenticate(...)` and `auth.Authorize(...)` only on protected route groups or handlers.
8. `httpmetrics.Middleware(...)` and `tracing.Middleware(...)` for transport telemetry when needed.
9. `accesslog.Middleware(app.Logger())` for logging-only access logs.

Keep `recovery.Recovery(...)` directly after `requestid.Middleware(...)` in
generated and reference stacks so request IDs are available and all later
transport middleware remains downstream of recovery.
Recovery logs only sanitized panic metadata such as the panic type; raw panic
values are not a stable logging surface because panic payloads can contain
secrets or request-specific data.

Keep tenant resolution, quota, and tenant policy in `x/tenant`. Keep exporter
and telemetry backend wiring in `x/observability`. The stable middleware layer
should remain a set of explicit transport wrappers, not a business policy
catalog.

### Request ID contract

`requestid.Middleware(...)` provides request correlation only. Generated request
IDs are safe to log as identifiers, but they are not secrets, tokens, nonces, or
authorization material. The default `requestid.NewRequestID()` format embeds a
timestamp component that can be decoded with `requestid.DecodeRequestID(...)`.
Applications that need a different shape or stricter generation policy should
wire `requestid.WithGenerator(...)`.

### Timeout contract

`timeout.Timeout(...)` creates a deadline-bound request context and returns a
structured `504` when the downstream handler does not finish before the
deadline. It does not forcibly stop downstream work. Handlers and services must
observe `r.Context().Done()` to stop side effects promptly after cancellation.

Timeout buffers successful responses so it can decide whether to return the
downstream response or the timeout error. Responses larger than
`TimeoutConfig.StreamingThreshold` are not streamed through; they become a
structured server error because the buffered response can no longer be replayed
safely. `StreamingThreshold` is the historical field name; treat it as the
maximum replayable response size, not as streaming support.

Writes attempted by a downstream handler after timeout observe the canceled
request context through the timeout response writer and return the context
deadline error. Timeout does not wait for that downstream goroutine after the
504 response is written. If the downstream handler panics before timeout, the
panic is re-thrown on the request goroutine so outer recovery middleware can
handle it. If it panics after the timeout response has already been emitted,
outer recovery can no longer rewrite the response; configure
`TimeoutConfig.OnPanic` to observe that late failure. The hook is best-effort
and should be non-blocking; if the hook itself panics, timeout recovers that
callback panic internally.

### Concurrency limit contract

`concurrencylimit.Middleware(...)` bounds active handlers and optional
queued waiters. Queued requests observe `r.Context().Done()` while waiting for a
worker slot; if the request is canceled before a slot is available, the
middleware returns without invoking the downstream handler or writing a
synthetic queue-timeout error.
Queue timeout responses include `queue_occupancy` and `queue_capacity` details;
`queue_occupancy` reports the internal queue channel occupancy, which includes
active and waiting requests represented in that channel.

### Gzip compression contract

`compression.Gzip(...)` compresses eligible non-error responses when the client
accepts gzip. It skips websocket upgrades, SSE, already-compressed responses,
and common binary content types.

`GzipConfig.MaxBufferBytes` controls only the pre-compression buffer. If a
response exceeds the limit before gzip output has started, the middleware
bypasses compression and writes the response as-is. Once gzip output has
started, later writes continue through the gzip writer; the middleware does not
switch a partially compressed response back to an uncompressed response.
If a handler calls `WriteHeader(200)` before setting `Content-Type` or writing a
body, gzip delays the compression decision until the first body write so it can
sniff the content type consistently.
If a handler calls `Flush()` before the first body write, gzip commits the
current headers and treats later writes as uncompressed pass-through data.

### Debug error contract

`debug.DebugErrors(...)` is a development-facing transport helper for replacing
empty or plain-text error responses with structured JSON. Do not wire it into a
production baseline unless the application owner explicitly accepts the risk of
debug metadata exposure.

Debug capture is bounded by `DebugErrorConfig.MaxBodyBytes`. If a response
exceeds the capture limit, the middleware stops debug replacement and passes the
original response through. It skips websocket upgrades, CONNECT requests, SSE
requests, and response content types that declare streaming. If a handler calls
`Flush` or `Hijack`, debug also switches to pass-through behavior and does not
attempt to replace the response.

### CORS contract

`cors.Middleware(...)` applies CORS headers only when the request origin and
preflight attributes match the configured allow lists. Disallowed origins,
methods, headers, and blank `Access-Control-Request-Headers` lists fall through
to the next handler without CORS headers; the middleware does not synthesize a
transport denial response for those cases.
Raw option lists are trimmed and blank entries are ignored before matching, so
common configuration whitespace does not silently disable an origin, method, or
header.

Wildcard `AllowedHeaders` preserves requested header names after trimming and
uses the same blank-list validation as explicit allowed-header checks.

The zero-value options default to `AllowedOrigins: ["*"]`. When credentials are
enabled with wildcard origins, the middleware echoes the request `Origin` value
instead of returning `*`, matching browser CORS requirements for credentialed
requests.

Production stacks should pass explicit origins or start from
`cors.StrictDefaultOptions("https://app.example")`. `StrictDefaultOptions`
trims and filters blank origins, keeps the standard method/header defaults, and
panics when no valid origin is supplied, so a missing origin list cannot
silently become wildcard access. It also rejects `"*"` because strict CORS
defaults require explicit origins. Use `cors.StrictDefaultOptionsE(...)` when
configuration should report an error instead of panicking.

### Body limit contract

`bodylimit.Middleware(...)` writes the structured `413` response from the request
body reader as soon as an overrun is detected and the response has not already
started. After that terminal error, downstream writes are suppressed; they
report the bytes as consumed to the handler but do not modify the body-limit
response.

### Rate limit contract

`ratelimit.NewAbuseGuard(...)` is the lifecycle entrypoint. Register
`guard.Middleware()` and call `guard.Stop()` during application shutdown when
the guard created the limiter. When `AbuseGuardConfig.Limiter` is supplied, the
caller owns that limiter and `guard.Stop()` is a no-op for it.

### Response writer compatibility

Buffering and capture middleware do not all preserve optional response writer
interfaces. Use this matrix when placing middleware around streaming, SSE, or
websocket handlers. The shared conformance suite covers panic propagation,
`Flush`, `Hijack`, post-timeout writes, and partial gzip panic finalization for
the high-risk wrappers represented here.
Each `Unwrap`, `Flush`, and `Hijack` claim in the matrix is also covered by a
shared positive or negative conformance case.

| Package | Buffers response body | `Unwrap` | `Flush` | `Hijack` | Streaming / websocket guidance |
|---|---:|---:|---:|---:|---|
| `accesslog` | no | yes | yes | yes | compatible with underlying writer support |
| `bodylimit` | no response buffering | yes | yes | yes | request-body limiting only; post-overrun writes are suppressed |
| `coalesce` | bounded capture for waiters | yes | yes | yes | use only for bounded safe responses; not for streaming/SSE/websocket |
| `compression` | pre-compression buffer | yes | yes | yes before compression starts | skips websocket/SSE; gzip output stays gzip once started |
| `debug` | bounded capture | yes | yes | yes | development-only; skips declared streaming and passes through on capture overflow, flush, or hijack |
| `httpmetrics` | no | yes | yes | yes | compatible with underlying writer support |
| `recovery` | no | yes | yes | yes | compatible with underlying writer support; cannot rewrite responses after headers/body are committed |
| `timeout` | full bounded replay buffer | no | no | no | not for streaming/SSE/websocket; large responses become structured errors |
| `tracing` | no | yes | yes | yes | compatible with underlying writer support |

### Coalesce capture contract

`coalesce.Middleware(...)` is a GA stable but high-risk transport response
coalescer, not a cache or a business freshness policy. It is appropriate only
for bounded, safe, non-streaming responses whose variants are fully represented
by the coalesce key. It forwards the leader request response to the leader client
while capturing a bounded copy for concurrent waiters with the same coalesce key.
The default `Config.MaxResponseBytes` is 10MB. If the leader response exceeds
the capture limit, the leader still receives the upstream response, but waiters
receive a structured upstream failure instead of replaying an unbounded
in-memory body. If the leader response uses `Flush` or `Hijack`, coalesce also
treats that leader response as unreplayable: the leader keeps the pass-through
transport behavior, while waiters receive the same structured upstream failure.

Only safe methods should be configured for coalescing. The default covers `GET`
and `HEAD`; configured methods are trimmed, uppercased, and blank entries are
ignored. Coalesced `HEAD` waiters receive the replayed status and headers
without a response body. Do not use coalesce for streaming, SSE, websocket,
long-polling, or response bodies whose correctness depends on per-request side
effects outside the key. When the leader explicitly commits headers with
`WriteHeader`, waiter replay uses that committed header snapshot rather than
later header mutations.

`Config.OnCoalesced` is called once for each waiter that successfully receives a
coalesced response. The callback `count` argument is a per-callback event count
and is currently always `1`; timed-out waiters are not counted. Aggregate in the
caller when total coalesced request counts are needed.
Waiters also observe their request context while waiting for the leader. If a
waiter request is canceled, coalesce returns without replaying the leader
response or writing a synthetic timeout response.
`Config.OnCoalesced` and `Config.OnError` are synchronous best-effort hooks;
panic from either hook is recovered internally so hooks cannot replace canonical
transport responses.

The default coalesce key is an FNV hash over method, host, URL, and common
variant headers. It is a transport deduplication key, not a security boundary;
handlers with additional response variants should provide an explicit `KeyFunc`.
A custom `KeyFunc` that returns an empty or all-whitespace key fails open:
coalesce forwards that request without creating or joining an in-flight slot.

## Internal transport primitives

`middleware/internal/transport` contains shared response-writing helpers used across middleware packages:

- `EnsureNoSniff(h http.Header)` — sets `X-Content-Type-Options: nosniff` unless already present
- `SafeWrite(w http.ResponseWriter, body []byte)` — writes body and sets the nosniff header; silently no-ops for nil writers
- `AddVary(h http.Header, values ...string)` — appends `Vary` tokens without duplicating existing comma-separated values
- `CopyHeaders(dst, src http.Header)` — overlays cloned source values onto destination headers and preserves destination keys absent from `src`
- `ReplaceHeaders(dst, src http.Header)` — replaces the complete destination header set with cloned source values; use for buffered response replay paths that own the full response header set
- `ClientIP(r *http.Request)` — extracts client IP from `X-Forwarded-For`, `X-Real-IP`, or `RemoteAddr` in that order
- `DirectClientIP(r *http.Request)` — extracts the direct peer IP from `RemoteAddr` only; use for security-sensitive defaults when trusted proxies are not configured
- `ResponseRecorder` — wraps an `http.ResponseWriter` to capture status code, body, and bytes written
- `BufferedResponse` — buffers the full response body with an optional max-bytes overflow guard; supports `WriteTo` for deferred flushing

These are internal; import them only from within the `middleware` module.

Do not confuse `middleware/internal/transport.CopyHeaders` with
`internal/httputil.CopyHeaders`. The middleware helper overlays destination
values; the lower-level `internal/httputil` helper appends values for
pass-through recorder use. Use `ReplaceHeaders` when replaying a fully buffered
response that should remove stale destination headers.
