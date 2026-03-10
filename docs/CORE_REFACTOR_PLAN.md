# Core Package Refactoring Plan

> Historical document (superseded).  
> This refactor plan is archived and is **not** the canonical v1 guidance as of March 10, 2026.  
> Use `docs/other/V1_CORE_API_FREEZE.md` and `docs/CANONICAL_STYLE_GUIDE.md` as the source of truth.

> Authority: `docs/CANONICAL_STYLE_GUIDE.md`
> Scope: `core/` package and its sub-packages
> Compatibility: Not required — clean break, follow best practices

---

## Executive Summary

The `core` package has accumulated concerns that violate the canonical style guide's core
principles: **stdlib first**, **one obvious way**, **explicit over implicit**, and
**small-step refactorability**. The package has become a feature catalog instead of a
focused app-construction entry point. This plan identifies every violation and provides a
concrete fix for each, organized into ten sequential phases.

---

## Violation Catalogue

### V1 — `core/di`: Reflection-Based Service Locator (Critical)

**Files:** `core/di/di.go`, `core/di/di_test.go`, `core/di/di_race_test.go`, `core/di/di_scoped_test.go`

The `DIContainer` uses `reflect.Type`-based lookup, `inject` struct tags, and named
string resolution. This is the exact anti-pattern the style guide forbids (§12):

```go
// ❌ Anti-pattern — struct tag injection obscures dependency shape
type Handler struct {
    DB *Database `inject:""`
}
container.Inject(&handler)

// ✅ Canonical — explicit constructor injection
type Handler struct {
    DB *Database
}
func NewHandler(db *Database) Handler { return Handler{DB: db} }
```

Style guide references: §12 ("Not canonical: `svc := MustServiceFromContext(...)`"),
§14 ("Do not introduce: New service locator patterns").

**Fix:** Delete the entire `core/di` package. Zero callers outside the package exist in
the current codebase. Any future dependency wiring must use constructor injection.

---

### V2 — App Struct Is a Feature Catalog (Critical)

**File:** `core/app.go`

The `App` struct holds 25+ fields spanning lifecycle state, extension capabilities,
observability, and HTTP internals. The style guide (§2) states: "Must not become a
feature catalog."

Specific offenders:

| Field | Problem |
|---|---|
| `loggingEnabled bool` | Middleware state tracked as flag instead of registry truth |
| `requestIDEnabled bool` | Same — flag is a proxy for "was middleware added" |
| `recoveryEnabled bool` | Same |
| `corsEnabled bool` | Same |
| `corsOptions *cors.CORSOptions` | CORS config embedded in app lifecycle struct |
| `metricsCollector metrics.MetricsCollector` | Observability concern in app core |
| `tracer observability.Tracer` | Observability concern in app core |
| `pub pubsub.PubSub` | Extension capability reference in app core |
| `connTracker *connectionTracker` | HTTP server detail mixed into App |

**Fix:** Remove all boolean middleware-state flags. The middleware registry is the source
of truth — if a middleware is registered, it is active. Remove `corsOptions`,
`metricsCollector`, `tracer`, and `pub` as direct App fields. Pass them explicitly to
their respective factory functions at call sites.

**Resulting App struct (target):**

```go
type App struct {
    config        *AppConfig
    router        *router.Router
    middlewareReg *middleware.Registry
    logger        log.StructuredLogger

    mu           sync.RWMutex
    started      bool
    configFrozen bool

    httpServer  *http.Server
    handler     http.Handler
    handlerOnce sync.Once

    components        []Component
    startedComponents []Component
    componentStopOnce sync.Once

    runners        []Runner
    startedRunners []Runner
    runnerStopOnce sync.Once

    shutdownHooks []ShutdownHook
    shutdownOnce  sync.Once
}
```

---

### V3 — Multiple Middleware Registration Styles (Critical)

**Files:** `core/options.go`, `core/middleware.go`

The style guide (§4, §18) requires "one bootstrap" and forbids "Teaching multiple equally
valid bootstraps." Currently there are three families for the same operation:

> Legacy snippet (historical): shows pre-freeze API variants for analysis context only.

```go
// Style A — direct Use (canonical)
app.Use(recovery.RecoveryWithLogger(logger))

// Style B — Option functions (redundant alternative)
app := core.New(core.WithRecovery(), core.WithLogging(), core.WithRequestID())

// Style C — method shortcuts (third alternative)
app.EnableAuth("token")
app.EnableRateLimit(100, 200)
```

Redundant Option functions that duplicate `app.Use()`:

- `WithRecovery()` → already expressible as `app.Use(recovery.RecoveryWithLogger(logger))`
- `WithLogging()` → already expressible as `app.Use(observability.Logging(...))`
- `WithRequestID()` → already expressible as `app.Use(observability.RequestID())`
- `WithRecommendedMiddleware()` → bundles three middlewares, hiding order
- `WithCORS()` / `WithCORSOptions()` → already expressible as `app.Use(cors.CORS(...))`

App methods that should be `app.Use()` calls:

- `EnableAuth(token string)` — shortcut that hides the middleware being used
- `EnableRateLimit(max, queue int64)` — shortcut that obscures configuration shape

**Fix:**

Delete from `options.go`: `WithRecovery`, `WithLogging`, `WithRequestID`,
`WithRecommendedMiddleware`, `WithCORS`, `WithCORSOptions`.

Delete from `middleware.go`: `EnableAuth`, `EnableRateLimit`, `enableLogging`,
`enableRecovery`, `enableCORS`, `corsMiddleware`, `loggingMiddleware`, `metricsAdapter`.

Delete corresponding private methods: `enableLogging()`, `enableRecovery()`,
`enableRequestID()`, `enableCORS()`.

The single canonical registration path:

```go
app := core.New(core.WithAddr(":8080"))

app.Use(observability.RequestID())
app.Use(recovery.RecoveryWithLogger(logger))
app.Use(observability.Logging(logger, metricsCollector, tracer))
app.Use(cors.CORS())

registerRoutes(app)
```

---

### V4 — Implicit Guardrail Middleware (Critical)

**File:** `core/middleware.go` (`applyGuardrails`), `core/config.go`, `core/options.go`

`applyGuardrails()` silently prepends four middleware layers before request dispatch:
security headers, abuse guard, body limit, and concurrency limit. The user never sees
these in the bootstrap. The style guide (§4) says "Global middleware attached explicitly
near startup" and (§1) "explicit over implicit."

The Option functions that configure implicit behaviour:

- `WithSecurityHeadersEnabled(bool)`
- `WithSecurityHeadersPolicy(*headers.Policy)`
- `WithAbuseGuardEnabled(bool)`
- `WithAbuseGuardConfig(ratelimit.AbuseGuardConfig)`
- `WithMaxBodyBytes(int64)` — activates a hidden middleware
- `WithConcurrencyLimits(int, int, time.Duration)` — activates hidden middleware

**Fix:**

Remove `applyGuardrails()` entirely. Remove from `AppConfig`: `EnableSecurityHeaders`,
`SecurityHeadersPolicy`, `EnableAbuseGuard`, `AbuseGuardConfig`, `MaxBodyBytes`,
`MaxConcurrency`, `QueueDepth`, `QueueTimeout`. Remove corresponding Option functions.

Move the constants `DefaultMaxConcurrency` and `DefaultQueueDepth` to the
`middleware/limits` package where they are used.

Users add protection explicitly in their bootstrap:

```go
app.Use(security.SecurityHeaders(nil))
app.Use(ratelimit.AbuseGuard(ratelimit.DefaultAbuseGuardConfig()))
app.Use(limits.BodyLimit(10<<20, logger))
app.Use(limits.ConcurrencyLimit(256, 512, 250*time.Millisecond, logger))
```

---

### V5 — Extension Configs Embedded in AppConfig

**Files:** `core/config.go`, `core/config_aliases.go`, `core/options.go`,
`core/components_default.go`

`AppConfig` contains three extension-capability configs:

```go
PubSub     PubSubConfig
WebhookOut WebhookOutConfig
WebhookIn  WebhookInConfig
```

`config_aliases.go` exists only because these leaked into the core config struct:

```go
type PubSubConfig = pubsubdebug.PubSubConfig
type WebhookOutConfig = webhook.WebhookOutConfig
type WebhookInConfig = webhook.WebhookInConfig
```

Style guide §2: "Extension packages … Must not define primary coding style" and core
"Must not become a feature catalog."

**Fix:**

Remove `PubSub`, `WebhookOut`, `WebhookIn` from `AppConfig`.
Delete `core/config_aliases.go`.
Remove from `options.go`: `WithPubSub`, `WithPubSubDebug`, `WithWebhookOut`, `WithWebhookIn`.

Extension components are mounted explicitly with their config:

```go
app := core.New()
app.Add(pubsubdebug.New(pubsubdebug.Config{Enabled: true, Path: "/_debug/pubsub"}, ps))
app.Add(webhook.NewOutbound(webhook.OutConfig{Enabled: true, BasePath: "/webhooks"}))
app.Add(webhook.NewInbound(webhook.InConfig{GitHubPath: "/webhooks/github"}, ps, logger))
```

`builtInComponents()` is simplified to only provide devtools when `Debug` is true.

---

### V6 — Tenant Concerns in Core Options (Critical)

**Files:** `core/options.go`, `core/tenant_alias.go`

`TenantMiddlewareOptions` (50+ lines) and `WithTenantMiddleware` live in
`core/options.go`. The tenant domain is an extension capability (§2: "tenant — capability
layers, not the core learning path"). The style guide (§9) says middleware must be
"Transport-layer cross-cutting only."

> Legacy snippet (historical): this "Current" sample is intentionally non-canonical.

```go
// ❌ Current — business domain orchestration inside core
app := core.New(core.WithTenantMiddleware(core.TenantMiddlewareOptions{
    QuotaManager:    quotaMgr,
    PolicyEvaluator: policyEval,
}))

// ✅ Target — explicit middleware registration from the tenant package
app.Use(tenant.Resolver(tenant.ResolverOptions{HeaderName: "X-Tenant-ID"}))
app.Use(tenant.RateLimit(tenant.RateLimitOptions{Limiter: rateLimiter}))
app.Use(tenant.Quota(tenant.QuotaOptions{Manager: quotaMgr}))
app.Use(tenant.Policy(tenant.PolicyOptions{Evaluator: policyEval}))
```

**Fix:**

Remove `TenantMiddlewareOptions`, `WithTenantMiddleware`, `WithTenantConfigManager` from
`core/options.go`. Delete `core/tenant_alias.go`. Move any convenience chain builder
(if desired) to the `tenant` package itself, e.g. `tenant.FullChain(opts)` returning
`[]middleware.Middleware`.

---

### V7 — `Component.Dependencies() []reflect.Type`

**File:** `core/component.go`

The `Component` interface declares:

```go
Dependencies() []reflect.Type
```

This implies a reflection-based dependency-ordering system. The style guide (§12) forbids
reflection-based dependency resolution. Ordering is a static, explicit concern — it
should be determined by the order in which `WithComponent()` calls appear at the
bootstrap site.

**Fix:**

Remove `Dependencies() []reflect.Type` from the `Component` interface.
Remove its implementation from `BaseComponent`.
Remove it from all concrete component types under `core/components/`.
Remove the `reflect` import from `core/component.go`.

Component ordering is controlled entirely by registration order:

```go
// Components start in registration order, stop in reverse order.
app := core.New(
    core.WithComponent(dbComponent),    // starts first
    core.WithComponent(cacheComponent), // starts second
    core.WithComponent(apiComponent),   // starts third
)
```

---

### V8 — Observability Wrapper Aliases in Core

**File:** `core/observability_wrapper.go`

```go
type MetricsConfig = observability.MetricsConfig
type TracingConfig = observability.TracingConfig
type ObservabilityConfig = observability.ObservabilityConfig

func DefaultMetricsConfig() MetricsConfig     { ... }
func DefaultTracingConfig() TracingConfig      { ... }
func DefaultObservabilityConfig() ObservabilityConfig { ... }
```

Re-exporting sub-package types and constructors through core creates a redundant API
surface and blurs package boundaries. The style guide (§14): "Do not introduce: New
package roles that blur core boundaries."

**Fix:**

Delete `core/observability_wrapper.go`. The `ConfigureObservability` method can remain
on `App` (it is app-scoped configuration), but callers import
`core/components/observability` directly for the config types:

```go
import obscomp "github.com/spcent/plumego/core/components/observability"

cfg := obscomp.DefaultObservabilityConfig()
cfg.Metrics.Enabled = true
app.ConfigureObservability(cfg)
```

---

### V9 — Devtools Path Constants Re-Exported Through Core

**File:** `core/devtools_wrapper.go`

The file re-exports 11 path constants from `core/components/devtools` into the `core`
package namespace and defines a `formatType` utility function. These constants are only
needed inside the devtools component and in the `applyGuardrails` debug-mode branch (which
is itself being removed in V4). The `formatType` helper is pure utility and should be
local to the devtools package.

**Fix:**

Delete all constant re-exports from `core/devtools_wrapper.go`. Keep only
`newDevToolsComponent`, `attachDevMetrics`, `devtoolsMiddlewareList`, and
`devtoolsConfigSnapshot` (the functions used by `components_default.go`). Move
`formatType` into `core/components/devtools` where it is used. Inline the constants
directly in the one location inside `lifecycle.go` that logs the debug-mode paths, or
remove that logging once the devtools component self-registers its own routes.

---

### V10 — `ResettableOnce`: Testing Concern in Production Code

**Files:** `core/resettable_once.go`, `core/app.go` (4 uses), `core/lifecycle.go`,
`core/app_test_resettable.go`

`ResettableOnce` and `ResetForTesting()` exist to let test code mutate production
lifecycle state. The style guide (§13): "No hidden global state between tests." The
correct fix is to construct a new `App` per test, not to reset a shared one.

```go
// ❌ Current — resets singleton state
app.ResetForTesting()
app.Get("/users", handler)
app.ServeHTTP(rec, req)

// ✅ Canonical — new App per test
func TestCreateUser(t *testing.T) {
    app := core.New()
    app.Get("/users", handler)
    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)
    ...
}
```

**Fix:**

Delete `core/resettable_once.go`.
Replace all four `ResettableOnce` fields in `App` with `sync.Once`.
Delete `ResetForTesting()` from the public API.
Delete `core/app_test_resettable.go`.
Rewrite all tests that called `ResetForTesting()` to construct a fresh `App` per case.

---

## Phase Execution Plan

Phases are ordered by dependency and blast radius. Complete each phase fully (tests
passing) before starting the next.

### Phase 1 — Delete `core/di`

**Goal:** Remove the service locator package.

**Steps:**
1. Delete `core/di/di.go`, `core/di/di_test.go`, `core/di/di_race_test.go`, `core/di/di_scoped_test.go`.
2. Delete the `core/di/` directory.
3. Verify no import outside the package: `grep -r '"github.com/spcent/plumego/core/di"' .`.
4. Run `go build ./...` — must pass.

**Acceptance:** `core/di` does not exist. `go test ./...` passes.

---

### Phase 2 — Single Middleware Registration Path (V3)

**Goal:** Remove all middleware-registration shortcuts. One path: `app.Use(middleware)`.

**Steps:**
1. Delete from `core/options.go`:
   - `WithRecovery`, `WithLogging`, `WithRequestID`, `WithRecommendedMiddleware`
   - `WithCORS`, `WithCORSOptions`
2. Delete from `core/middleware.go`:
   - `EnableAuth`, `EnableRateLimit` (public methods on `*App`)
   - `enableLogging`, `enableRecovery`, `enableCORS` (private helpers)
   - `loggingMiddleware`, `corsMiddleware` (lazy factories)
   - `metricsAdapter` struct and its `Observe` method
3. Remove from `App` struct: `loggingEnabled`, `requestIDEnabled`, `recoveryEnabled`,
   `corsEnabled`, `corsOptions` fields.
4. Remove the `enableRequestID()` private method and the `requestIDEnabled` guard in
   `applyGuardrails` (the whole guardrails function is removed in Phase 4).
5. Update all call sites in tests and examples to use explicit `app.Use()`.
6. Run `go test ./...`.

**Acceptance:** `app.EnableAuth`, `app.EnableRateLimit`, `WithLogging`, `WithRecovery`,
`WithRequestID`, `WithRecommendedMiddleware`, `WithCORS`, `WithCORSOptions` do not exist.

---

### Phase 3 — Remove Tenant Concerns from Core (V6)

**Goal:** Tenant middleware wiring lives in the `tenant` package.

**Steps:**
1. Delete from `core/options.go`: `TenantMiddlewareOptions`, `WithTenantMiddleware`,
   `WithTenantConfigManager`.
2. Delete `core/tenant_alias.go`.
3. Remove the `middleware/tenant` import from `core/options.go`.
4. Move any convenience `FullChain` builder (if desired) to `tenant/chain.go` in the
   `tenant` package, returning `[]middleware.Middleware`.
5. Rewrite any code or example that used `WithTenantMiddleware` to call `app.Use()`
   with individual `tenant.*` middleware.
6. Run `go test ./...`.

**Acceptance:** `core` has no import path to `middleware/tenant` or `tenant` packages.

---

### Phase 4 — Make Guardrails Explicit (V4)

**Goal:** No hidden middleware. Users register hardening middleware explicitly.

**Steps:**
1. Delete `applyGuardrails()` from `core/middleware.go`.
2. Remove from `core/config.go` (`AppConfig`):
   - `EnableSecurityHeaders`, `SecurityHeadersPolicy`
   - `EnableAbuseGuard`, `AbuseGuardConfig`
   - `MaxBodyBytes`, `MaxConcurrency`, `QueueDepth`, `QueueTimeout`
3. Remove from `core/options.go`:
   - `WithSecurityHeadersEnabled`, `WithSecurityHeadersPolicy`
   - `WithAbuseGuardEnabled`, `WithAbuseGuardConfig`
   - `WithMaxBodyBytes`, `WithConcurrencyLimits`
4. Move `DefaultMaxConcurrency` and `DefaultQueueDepth` constants to `middleware/limits`.
5. Remove the `applyGuardrails` call from `ensureHandler()` in `core/http_handler.go`.
6. Remove no-longer-needed imports from `core/middleware.go`:
   `middleware/limits`, `middleware/ratelimit`, `middleware/security`, `middleware/debug`.
7. Update the canonical bootstrap example in `README.md` to show explicit middleware.
8. Run `go test ./...`.

**Acceptance:** `applyGuardrails` does not exist. `AppConfig` has no security or
concurrency fields.

---

### Phase 5 — Remove Extension Configs from AppConfig (V5)

**Goal:** `AppConfig` contains only server lifecycle configuration.

**Steps:**
1. Remove `PubSub PubSubConfig`, `WebhookOut WebhookOutConfig`, `WebhookIn WebhookInConfig`
   from `AppConfig` in `core/config.go`.
2. Delete `core/config_aliases.go`.
3. Remove from `core/options.go`: `WithPubSub`, `WithPubSubDebug`, `WithWebhookOut`,
   `WithWebhookIn`.
4. Refactor `core/components_default.go` — `builtInComponents()` no longer reads
   extension configs from `AppConfig`. Devtools is the only built-in component.
   PubSub debug, webhook in/out become explicit `WithComponent(...)` calls.
5. Update component constructors to accept config directly:
   - `pubsubdebug.New(cfg, fallbackPS)` — creates the component with its own config
   - `webhook.NewInbound(cfg, ps, logger)`
   - `webhook.NewOutbound(cfg)`
6. Run `go test ./...`.

**Acceptance:** `AppConfig` does not reference any extension package types.

---

### Phase 6 — Remove `Dependencies()` from Component Interface (V7)

**Goal:** Component ordering is static, determined by registration order.

**Steps:**
1. Remove `Dependencies() []reflect.Type` from the `Component` interface in
   `core/component.go`.
2. Remove the `reflect` import from `core/component.go`.
3. Remove `Dependencies() []reflect.Type` from `BaseComponent`.
4. Remove `Dependencies()` from all concrete component implementations:
   - `core/components/builtin/`
   - `core/components/devtools/`
   - `core/components/observability/`
   - `core/components/ops/`
   - `core/components/pubsubdebug/`
   - `core/components/tenant/`
   - `core/components/webhook/`
   - `core/components/websocket/`
5. Remove any dependency-sort logic that consumed `Dependencies()`.
6. Run `go test ./...`.

**Acceptance:** `Component` interface has four methods: `RegisterRoutes`, `RegisterMiddleware`,
`Start`, `Stop`. No `reflect` import in `core/component.go`.

---

### Phase 7 — Remove `ResettableOnce` (V10)

**Goal:** Tests construct fresh `App` instances. No mutable singleton test state.

**Steps:**
1. Delete `core/resettable_once.go`.
2. In `core/app.go`, replace the four `ResettableOnce` fields with `sync.Once`:
   - `handlerOnce sync.Once`
   - `componentStopOnce sync.Once`
   - `runnerStopOnce sync.Once`
   - `shutdownOnce sync.Once`
3. Delete `ResetForTesting()` from `core/lifecycle.go`.
4. Delete `core/app_test_resettable.go`.
5. Rewrite all tests that called `ResetForTesting()`:
   - Each test case constructs a fresh `core.New()` instance.
   - Table-driven tests pass `app` as a local variable, not a package-level shared one.
6. Run `go test -race ./...`.

**Acceptance:** `ResettableOnce` does not exist. `ResetForTesting` does not exist.
`go test -race ./...` passes.

---

### Phase 8 — Remove Observability Wrapper (V8)

**Goal:** Remove type aliases and proxy functions that blur the `core` / `core/components/observability` boundary.

**Steps:**
1. Delete `core/observability_wrapper.go` (type aliases `MetricsConfig`, `TracingConfig`,
   `ObservabilityConfig` and `DefaultXxx()` functions).
2. Update all call sites to import `core/components/observability` directly.
3. Keep `App.ConfigureObservability(cfg)` — it is legitimately app-scoped configuration.
   Update its signature to use the concrete type directly (not via an alias):
   ```go
   import obscomp "github.com/spcent/plumego/core/components/observability"
   func (a *App) ConfigureObservability(cfg obscomp.ObservabilityConfig) error
   ```
4. Remove `metricsCollector` and `tracer` from the `App` struct (already removed in
   Phase 2). `ConfigureObservability` stores state on the observability component
   registered via `WithComponent`.
5. Run `go test ./...`.

**Acceptance:** `core` package does not re-export types from `core/components/observability`.

---

### Phase 9 — Clean Up Devtools Wrapper (V9)

**Goal:** Remove constant re-exports and misplaced utility from `core/devtools_wrapper.go`.

**Steps:**
1. Move `formatType` to `core/components/devtools/devtools.go` (package-private).
2. Remove all 11 constant declarations (`devToolsRoutesPath`, etc.) from
   `core/devtools_wrapper.go`.
3. Any remaining reference to those constants in `core/lifecycle.go` (debug-mode startup
   logging) should either:
   - Import `core/components/devtools` directly for the constant values, or
   - Remove the startup log lines (the devtools component self-registers its own routes).
4. The file `core/devtools_wrapper.go` shrinks to only `newDevToolsComponent`,
   `attachDevMetrics`, `devtoolsMiddlewareList`, `devtoolsConfigSnapshot`. Rename to
   `core/devtools.go` to drop the misleading `_wrapper` suffix.
5. Run `go test ./...`.

**Acceptance:** No `const` declarations in `core/devtools.go`. `formatType` lives in
the devtools sub-package.

---

### Phase 10 — Slim `App` Struct and Final Cleanup (V2 completion)

**Goal:** Verify `App` matches the target struct defined in V2. Remove all remaining
field and method cruft accumulated by earlier phases.

**Steps:**
1. Audit `core/app.go` against the target struct (see V2). Remove any remaining
   fields that were not cleaned up in earlier phases.
2. Remove `guardsApplied bool` (no more guardrails).
3. Remove `envLoaded bool` guard or keep if `.env` loading is still part of `Boot()`.
4. Verify `core/options.go` contains only these Option functions:
   - `WithRouter`, `WithAddr`, `WithEnvPath`
   - `WithShutdownTimeout`, `WithServerTimeouts`, `WithMaxHeaderBytes`
   - `WithHTTP2`, `WithTLS`, `WithTLSConfig`
   - `WithDebug`, `WithLogger`
   - `WithComponent`, `WithComponents`
   - `WithRunner`, `WithRunners`
   - `WithShutdownHook`, `WithShutdownHooks`
   - `WithMetricsCollector`, `WithTracer` (if retained for explicit observability)
   - `WithMethodNotAllowed`
5. Run `go vet ./...`, `gofmt -w .`, `go test -race ./...`.
6. Run `go test -timeout 20s ./...` — all must pass.

**Acceptance:** App struct matches the target. No TODO/FIXME comments remain. All
style guide checklist items (§16) pass review.

---

## Target AppConfig (after all phases)

```go
type AppConfig struct {
    Addr            string
    EnvFile         string
    TLS             TLSConfig
    Debug           bool
    ShutdownTimeout time.Duration

    // HTTP server hardening
    ReadTimeout       time.Duration
    ReadHeaderTimeout time.Duration
    WriteTimeout      time.Duration
    IdleTimeout       time.Duration
    MaxHeaderBytes    int
    EnableHTTP2       bool
    DrainInterval     time.Duration
}
```

---

## Canonical Bootstrap After Refactoring

```go
func main() {
    logger := log.NewGLogger()

    app := core.New(
        core.WithAddr(":8080"),
        core.WithLogger(logger),
        core.WithShutdownTimeout(5*time.Second),
        core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
        core.WithDebug(), // dev only
    )

    // Hardening — explicit, ordered, visible
    app.Use(security.SecurityHeaders(nil))
    app.Use(ratelimit.AbuseGuard(ratelimit.DefaultAbuseGuardConfig()))
    app.Use(limits.BodyLimit(10<<20, logger))
    app.Use(limits.ConcurrencyLimit(256, 512, 250*time.Millisecond, logger))

    // Observability
    app.Use(observability.RequestID())
    app.Use(recovery.RecoveryWithLogger(logger))
    app.Use(observability.Logging(logger, metricsCollector, tracer))

    // Domain routes
    registerRoutes(app)

    if err := app.Boot(); err != nil {
        logger.Error("server failed", log.Fields{"error": err})
        os.Exit(1)
    }
}

func registerRoutes(app *core.App) {
    userHandler := handlers.UserHandler{Service: userService}
    app.Get("/healthz", healthHandler)
    app.Post("/users", userHandler.Create)
    app.Get("/users/:id", userHandler.Get)
}
```

---

## File Change Summary

| File | Action |
|---|---|
| `core/di/` (entire package) | **Delete** |
| `core/resettable_once.go` | **Delete** |
| `core/app_test_resettable.go` | **Delete** |
| `core/config_aliases.go` | **Delete** |
| `core/tenant_alias.go` | **Delete** |
| `core/observability_wrapper.go` | **Delete** |
| `core/devtools_wrapper.go` | **Rename → `core/devtools.go`**, remove constants |
| `core/app.go` | Remove 9 fields (see V2 target struct) |
| `core/config.go` | Remove 9 fields from `AppConfig` (see Phase 4, 5) |
| `core/middleware.go` | Remove `applyGuardrails`, `enableLogging`, `enableRecovery`, `enableCORS`, `metricsAdapter`, `EnableAuth`, `EnableRateLimit` |
| `core/options.go` | Remove ~15 Option functions and `TenantMiddlewareOptions` |
| `core/component.go` | Remove `Dependencies() []reflect.Type` from interface and `BaseComponent` |
| `core/components_default.go` | Simplify to devtools-only built-in |
| `core/lifecycle.go` | Remove `ResetForTesting`, remove `guardsApplied` references |
| `core/http_handler.go` | Remove `applyGuardrails()` call |
| All `core/components/*/` | Remove `Dependencies()` implementations |
| Tests in `core/` | Rewrite to construct fresh `App` per test |

---

## Review Checklist (§16) After Refactoring

- [x] Request flow obvious from route to response?
- [x] Code stays close to `net/http` semantics?
- [x] Single visible source for each input?
- [x] JSON decoding explicit or via one transparent helper?
- [x] Middleware transport-only?
- [x] Dependencies explicit in structs or constructors?
- [x] Response and error shapes consistent with framework conventions?
- [x] A new contributor can trace control flow in a few minutes?
