# Card 0150

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `core/introspection_test.go`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
- `cmd/plumego/internal/devserver/analyzer.go`
- `cmd/plumego/internal/devserver/analyzer_test.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `reference/standard-service/internal/config/config.go`
- `reference/with-gateway/internal/config/config.go`
- `reference/with-messaging/internal/config/config.go`
- `reference/with-webhook/internal/config/config.go`
- `reference/with-websocket/internal/config/config.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`
Depends On:
- `0149-core-shutdown-timeout-contract-removal.md`

Goal:
- Remove app-local dev/debug metadata from the `core` kernel contract.

Problem:
- `core.AppConfig` still carries `Debug` and `EnvFile`.
- `core` itself does not branch on either field; they only flow through config
  defaults, snapshots, and first-party tooling.
- `x/devtools`, scaffolded apps, and the dev dashboard treat `core` as a
  transport for devtools/debug metadata rather than a pure HTTP kernel.
- This blurs the kernel boundary and keeps `core.RuntimeSnapshot` wider than
  the runtime state `core` actually owns.

Scope:
- Remove `Debug` and `EnvFile` from `core.AppConfig` and `core.RuntimeSnapshot`.
- Move debug/env-file ownership to app-local config and `x/devtools` options /
  hooks instead of routing it through `core`.
- Update first-party config loaders, scaffold output, and dashboard tooling to
  use the new explicit devtools/debug metadata path.

Non-goals:
- Do not remove devtools functionality.
- Do not change how reference apps load environment files.
- Do not redesign `x/devtools` endpoints in this card.

Files:
- `core/config.go`
- `core/introspection.go`
- `core/introspection_test.go`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
- `cmd/plumego/internal/devserver/analyzer.go`
- `cmd/plumego/internal/devserver/analyzer_test.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `reference/standard-service/internal/config/config.go`
- `reference/with-gateway/internal/config/config.go`
- `reference/with-messaging/internal/config/config.go`
- `reference/with-webhook/internal/config/config.go`
- `reference/with-websocket/internal/config/config.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/devtools/...`
- `go test -timeout 20s ./reference/...`
- `go test -timeout 20s ./internal/scaffold/... ./internal/devserver/...`
- `go vet ./core/... ./x/devtools/...`

Docs Sync:
- Remove `Debug` / `EnvFile` from the documented `core` config contract and
  describe them as app-local or devtools-local inputs instead.

Done Definition:
- `core` no longer owns `Debug` or `EnvFile`.
- First-party devtools/debug workflows use explicit non-core metadata inputs.
- `RuntimeSnapshot` only carries state the kernel actually owns.

Outcome:
- Removed `Debug` and `EnvFile` from `core.AppConfig` and `core.RuntimeSnapshot`,
  so the kernel now exposes only HTTP runtime state it actually owns.
- Switched first-party reference and scaffolded config loaders to an app-local
  `cfg.App` section that owns `EnvFile` and `Debug`.
- Made `x/devtools` own the merged debug config payload through
  `devtools.ConfigSnapshot`, with `core.RuntimeSnapshot` flowing in only as the
  kernel runtime subset.
- Updated the devserver analyzer/dashboard path to consume the devtools-owned
  config snapshot instead of treating `/_debug/config` as a raw `core`
  contract.
