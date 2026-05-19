# Card 0008

Priority: P2
State: done
Primary Module: middleware
Owned Files:
  - middleware/error_registry.go

Depends On: —

Goal:
`WriteTransportError` and three error code constants
(CodeServerBusy, CodeServerQueueTimeout, CodeUpstreamFailed) defined in
`middleware/error_registry.go` live in the parent `middleware` package,
forcing all sub-packages (concurrencylimit, ratelimit, recovery, timeout, bodylimit, coalesce)
to import the parent package, creating an asymmetric "sub-package depends on parent package" coupling.
There are 15+ call sites with a consistent pattern, but the dependency chain is not clean.

Move `WriteTransportError` and the related error codes to the `middleware/internal/transport`
sub-package so that sub-packages import the internal package instead, and the parent package
re-exports from that internal package (keeping the public API unchanged).

Scope:
- Create `middleware/internal/transport/transport.go`, migrating in:
  - The `WriteTransportError` function
  - The `CodeServerBusy`, `CodeServerQueueTimeout`, `CodeUpstreamFailed` constants
- Change `middleware/error_registry.go` to re-export from `middleware/internal/transport`,
  keeping the external API unchanged (existing callers using `mw.WriteTransportError`, `mw.CodeServerBusy`, etc. require no modifications)
- Sub-packages may optionally import directly from `middleware/internal/transport` (if re-export is kept, sub-package changes are not required)
- The focus is eliminating sub-packages' implicit dependency on the parent, without introducing new external breakage

Non-goals:
- Do not change the signature or behavior of `WriteTransportError`
- Do not modify callers of `mw.WriteTransportError` in x/gateway or x/rest
- Do not unify the middleware and handler-layer error-writing paths (that is a separate discussion)
- Do not migrate middleware functionality unrelated to error codes

Files:
  - middleware/error_registry.go (retained as re-export shim or deleted, depending on decision)
  - middleware/internal/transport/transport.go (new file)
  - middleware/concurrencylimit/concurrency_limit.go (update import)
  - middleware/ratelimit/abuse_guard.go (update import)
  - middleware/recovery/recover.go (update import)
  - middleware/timeout/timeout.go (update import)
  - middleware/bodylimit/body_limit.go (update import)
  - middleware/coalesce/coalesce.go (update import)

Tests:
  - go build ./middleware/...
  - go test ./middleware/...

Docs Sync: —

Done Definition:
- `middleware/internal/transport/transport.go` exists, containing the function and constants
- `middleware/error_registry.go` contains only re-exports (no implementation logic), or is deleted
- All sub-packages compile without circular dependencies
- `go test ./middleware/...` passes

Outcome:
- Created `middleware/internal/transport/errors.go` containing the implementation
- Changed `middleware/error_registry.go` to re-export via type aliases and var declarations: `CodeServerBusy = transport.CodeServerBusy`, etc.
- All callers continue to use the same API surface without any modifications required
