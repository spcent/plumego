# Card 0783

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/app.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `core/module.yaml`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
Depends On:
- `0781-core-runtime-state-contract-collapse.md`

Goal:
- Remove logger lifecycle ownership from `core` so the kernel treats the logger
  as a passive dependency instead of secretly starting and stopping it during
  app preparation and shutdown.

Problem:
- `Prepare()` currently calls `log.Lifecycle.Start(context.Background())`.
- `Shutdown(ctx)` conditionally calls `log.Lifecycle.Stop(ctx)`.
- This creates hidden side effects, asymmetric context ownership, and a fake
  kernel startup responsibility that no longer matches `core`'s reduced scope.
- `core` therefore still owns a non-obvious runtime hook even after serving and
  readiness ownership were removed.

Scope:
- Remove `log.Lifecycle` start/stop handling from `core`.
- Make `core` treat its logger dependency as already-constructed and already-
  owned by the caller.
- Update tests and docs to match the stricter dependency boundary.

Non-goals:
- Do not add a replacement startup hook API in `core`.
- Do not keep legacy lifecycle calls behind hidden wrappers.
- Do not redesign logging APIs outside what is required to remove kernel
  ownership.

Files:
- `core/app.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `core/module.yaml`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Remove any suggestion that `Prepare()` starts logger subsystems or that
  `Shutdown(ctx)` flushes logger lifecycle hooks on behalf of the caller.

Done Definition:
- `core` no longer starts or stops logger lifecycle hooks.
- The kernel logger is treated as a passive injected dependency.
- Tests and docs no longer describe hidden logger lifecycle ownership.
