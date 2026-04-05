# Card 0766

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `core/lifecycle.go`
- `core/app_test.go`
- `core/introspection_test.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
Depends On:

Goal:
- Remove the fake config-based shutdown timeout contract so `core` has one
  explicit shutdown model.

Problem:
- `core.AppConfig` still exposes `ShutdownTimeout`.
- `core.RuntimeSnapshot` and docs still present that value as if `core`
  enforces it.
- `(*App).Shutdown(ctx)` ignores `a.config.ShutdownTimeout` entirely and only
  obeys the caller-provided context.
- First-party apps defer `Shutdown(context.Background())`, so the config field
  currently adds API noise without affecting runtime behaviour.

Scope:
- Remove `ShutdownTimeout` from `core.AppConfig` and `core.RuntimeSnapshot`.
- Keep caller-owned shutdown deadlines as the single timeout source for
  `(*App).Shutdown(ctx)`.
- Update docs and tests so they no longer describe or assert a kernel-owned
  shutdown timeout.
- Remove first-party uses that surface the dead timeout in debug payloads.

Non-goals:
- Do not redesign connection draining behaviour in this card.
- Do not add a new convenience shutdown wrapper.
- Do not change the requirement that callers provide the shutdown context.

Files:
- `core/config.go`
- `core/introspection.go`
- `core/lifecycle.go`
- `core/app_test.go`
- `core/introspection_test.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./core/... ./x/devtools/...`

Docs Sync:
- Remove user-facing mentions of a core-managed shutdown timeout from root/core
  docs.

Done Definition:
- `core` exposes one shutdown timeout contract: the caller's context.
- `ShutdownTimeout` no longer appears in `core` config, snapshots, or docs.
- First-party debug/config surfaces no longer claim a timeout the kernel does
  not enforce.

Outcome:
