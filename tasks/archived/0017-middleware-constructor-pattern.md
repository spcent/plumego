# Card 0017

Priority: P2
State: done
Primary Module: middleware
Owned Files:
  - middleware/concurrencylimit/concurrency_limit.go
  - middleware/auth/contract.go

Depends On: —

Goal:
Constructor naming and parameter style across middleware sub-packages was highly inconsistent,
making it impossible for callers to form a unified mental model:

| Sub-package       | Constructor                                       | Parameter style         |
|-------------------|---------------------------------------------------|-------------------------|
| ratelimit         | `AbuseGuard(config AbuseGuardConfig)`             | Config struct           |
| auth              | `Authenticate(authenticator, opts...)`            | Interface + variadic    |
| bodylimit         | `BodyLimit(maxBytes int64, logger)`               | Bare parameters         |
| timeout           | `Timeout(cfg TimeoutConfig)`                      | Config struct           |
| coalesce          | `New(config) *Coalescer` + `Middleware(config)`   | Dual entry points       |
| cors              | `Middleware(opts CORSOptions)`                    | Options struct          |
| concurrencylimit  | `Middleware(maxConcurrent, queueDepth, queueTimeout, logger)` | Bare parameters |

Beyond naming chaos, two concrete problems existed:

1. **concurrencylimit accepted a logger but immediately discarded it** (`concurrency_limit.go:15:
   _ = logger`): the interface implied active logging but the implementation was a no-op,
   misleading callers.

2. **middleware/auth/contract.go filename was misleading**: all auth middleware implementation
   (Authenticate, Authorize, AuthorizeFunc, error handler, conversion functions) lived in
   `contract.go`, but by convention `contract.go` should contain only interface/type
   definitions, not implementations.

Scope:
- **concurrencylimit**: Remove the `logger log.StructuredLogger` parameter (breaking change —
  callers updated). If logging is genuinely needed in future, provide an optional `Logger` field
  in a Config struct and actually log on queue-timeout events. Grep callers before changing.
- **middleware/auth**: Rename `contract.go` to `auth.go`. If interface/implementation separation
  is needed later, create a new `types.go` for AuthErrorHandler/AuthOption types.
- **Document naming convention**: Add a note to `middleware/module.yaml` agent_hints:
  stateless middlewares use `Middleware(config T) middleware.Middleware`;
  stateful middlewares use `New(config T) *Type` with a `(t *Type) Middleware()` method;
  all Config types must provide `WithDefaults()` or `Default()`.

Non-goals:
- Do not unify constructors across all existing sub-packages (too large a change; only address
  the most egregious problems)
- Do not change the behavior logic of the auth middleware
- Do not rewrite bodylimit / cors and other sub-packages (address when touched)

Files:
  - middleware/concurrencylimit/concurrency_limit.go (logger parameter removed)
  - middleware/auth/contract.go → middleware/auth/auth.go (renamed)
  - middleware/auth/contract_test.go → middleware/auth/auth_test.go (renamed)
  - middleware/module.yaml (agent_hints naming convention added)
  - All files calling concurrencylimit.Middleware (updated after grep confirmation)

Tests:
  - go build ./middleware/...
  - go test ./middleware/...

Docs Sync:
  - middleware/module.yaml agent_hints augmented with naming convention

Done Definition:
- `grep -rn "_ = logger" middleware/concurrencylimit/` returns empty
- `middleware/auth/contract.go` does not exist; `middleware/auth/auth.go` exists
- `go build ./...` passes
- middleware/module.yaml contains constructor naming convention note

Outcome:
- Renamed `middleware/auth/contract.go` to `middleware/auth/auth.go` (and `contract_test.go`
  to `auth_test.go`).
- Removed `logger log.StructuredLogger` parameter from
  `middleware/concurrencylimit/concurrency_limit.Middleware` (was accepted but immediately
  discarded with `_ = logger`).
- Updated all callers of `concurrencylimit.Middleware`.
- Added constructor naming convention to `middleware/module.yaml` agent_hints:
  stateless = `Middleware(config T)`, stateful = `New(config T) *Type` with `.Middleware()`
  method; Config types need `WithDefaults()`/`Default()`.
