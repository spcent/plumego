# Card 1391

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- reference/workerfleet/main.go
- reference/workerfleet/main_test.go
- reference/workerfleet/internal/app/bootstrap.go
- reference/workerfleet/internal/app/routes.go
- reference/workerfleet/internal/app/bootstrap_test.go
Depends On:
- 1390

Goal:
- Move workerfleet toward the `reference/standard-service` thin-entrypoint shape without changing runtime behavior.

Scope:
- Keep `main.go` focused on loading config, constructing the app, and running it.
- Move Plumego core app construction, middleware wiring, route registration, runtime loop start/stop, and graceful shutdown orchestration into `internal/app`.
- Keep dependency injection constructor-based and visible.
- Preserve current environment variables and default values.
- Preserve the existing HTTP address, metrics route, health routes, and shutdown semantics.

Non-goals:
- Do not introduce a framework wrapper or route auto-discovery.
- Do not move workerfleet code into Plumego stable roots.
- Do not change workerfleet module path.

Files:
- reference/workerfleet/main.go
- reference/workerfleet/main_test.go
- reference/workerfleet/internal/app/bootstrap.go
- reference/workerfleet/internal/app/routes.go
- reference/workerfleet/internal/app/bootstrap_test.go

Tests:
- cd reference/workerfleet && go test -timeout 20s .
- cd reference/workerfleet && go test -timeout 20s ./internal/app/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Update README if run lifecycle wording changes.

Done Definition:
- `main.go` is thin and no longer owns detailed middleware, route, or loop wiring.
- Runtime behavior is preserved by focused tests.
- No hidden globals or init-time registration are introduced.
- Target checks pass.

Outcome:
- Added `internal/app.App` as the workerfleet composition/runtime owner.
- Moved server config parsing, core app construction, middleware wiring, route registrar invocation, runtime loop startup, alert loop startup, HTTP serving, graceful shutdown, and runtime close orchestration out of `main.go`.
- Kept `main.go` thin: load app config, load server config, construct app with explicit route registrar DI, then run.
- Preserved the 1390 app/handler boundary by injecting `handler.RegisterServiceRoutes` instead of importing handler from `internal/app`.
- Validation:
  - `cd reference/workerfleet && go test -timeout 20s .`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
  - `cd reference/workerfleet && go test -timeout 20s ./...`
  - `rg -n 'workerfleet/internal/handler|handler\.' reference/workerfleet/internal/app` returned no matches.
  - `git diff --check`
