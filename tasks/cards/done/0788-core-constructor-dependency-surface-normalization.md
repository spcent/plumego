# Card 0788

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `reference/standard-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`
- `reference/with-websocket/internal/app/app.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`
- `docs/modules/core/README.md`
- `core/module.yaml`
Depends On:
- `0786-core-internal-preparation-state-unification.md`

Goal:
- Replace the leftover one-off `Option` pattern with a clearer constructor
  dependency contract for `core`.

Problem:
- `core` is already config-first, but constructor dependencies still flow
  through `Option`.
- Only `WithLogger(...)` remains, which means `Option` is no longer a real
  extension surface; it is a wrapper around a single dependency and still
  panics on nil.
- This leaves `core.New(...)` split across typed config plus an ad hoc,
  panic-driven option contract.

Scope:
- Remove the generic `Option` constructor pattern from `core`.
- Introduce one explicit typed dependency input for constructor-owned
  dependencies.
- Migrate all first-party callers and docs to the normalized constructor.

Non-goals:
- Do not move logger into `AppConfig`.
- Do not keep `WithLogger(...)` as a dead compatibility wrapper.
- Do not add a new generic plugin/decorator constructor surface.

Files:
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `reference/standard-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`
- `reference/with-websocket/internal/app/app.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`
- `docs/modules/core/README.md`
- `core/module.yaml`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./reference/...`
- `go test -timeout 20s ./internal/scaffold/... ./internal/devserver/...`
- `go vet ./...`

Docs Sync:
- Update all constructor examples and module metadata to the normalized
  dependency contract.

Done Definition:
- `core.New(...)` uses one explicit typed dependency contract instead of
  `Option`.
- `WithLogger(...)` and the exported `Option` surface are removed.
- First-party callers and docs use the normalized constructor.
