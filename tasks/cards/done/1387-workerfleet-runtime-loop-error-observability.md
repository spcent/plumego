# Card 1387

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- reference/workerfleet/internal/app/runtime_loops.go
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/bootstrap.go
- reference/workerfleet/internal/app/runtime_loops_test.go
- reference/workerfleet/internal/app/alert_loop_test.go
Depends On:

Goal:
- Stop silently dropping runtime-loop, alert-evaluation, and notification errors.

Scope:
- Add an explicit app-local runtime error observer or logger dependency.
- Report errors from Kubernetes sync, status sweep, alert evaluation, and notifier dispatch.
- Keep loop retries bounded to the existing interval model; do not make transient loop errors crash the HTTP server.
- Distinguish startup/configuration errors from per-tick runtime errors.
- Add tests proving loop callbacks report errors instead of discarding them.

Non-goals:
- Do not add non-stdlib dependencies.
- Do not change HTTP route contracts.
- Do not redesign alert delivery durability or add a queue in this card.

Files:
- reference/workerfleet/internal/app/runtime_loops.go
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/bootstrap.go
- reference/workerfleet/internal/app/runtime_loops_test.go
- reference/workerfleet/internal/app/alert_loop_test.go

Tests:
- cd reference/workerfleet && go test -timeout 20s ./internal/app/...
- cd reference/workerfleet && go test -timeout 20s ./internal/platform/metrics/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Required only if new runtime error behavior, config, or observability fields are user-visible.

Done Definition:
- No runtime loop uses `_ =` to discard actionable errors.
- Runtime errors are visible through explicit app-local instrumentation.
- Startup/config errors still fail before handlers are exposed.
- Target checks pass.

Outcome:
- Implemented app-local runtime error observation through `RuntimeErrorObserver`.
- Runtime loop, alert evaluation, and notifier dispatch errors now report low-cardinality runtime error metrics instead of being silently discarded.
- Added focused app and metrics tests for status-sweep errors, notifier delivery failure reporting, and `workerfleet_runtime_errors_total`.
- Validation run:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/platform/metrics/...`
  - `cd reference/workerfleet && go test -timeout 20s ./...`
