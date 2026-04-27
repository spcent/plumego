# Card 0147

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/config.go`
- `core/options.go`
- `core/options_test.go`
- `core/introspection_test.go`
- `core/lifecycle_test.go`
- `reference/standard-service/internal/app/app.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Replace the fragmented `core` config input surface with one clear typed
  configuration model.

Problem:
- `New` inlines defaults directly in `core/app.go`.
- `core/options.go` mixes per-field setters, grouped setters, and convenience
  wrappers: `WithServerTimeouts`, `WithMaxHeaderBytes`, `WithHTTP2`,
  `WithTLS`, `WithTLSConfig`, and enable-only `WithDebug`.
- Some runtime config fields such as `DrainInterval` have no dedicated option at
  all, so tests mutate `app.config` directly to cover them.
- This leaves `core` configuration split across defaults embedded in `New`,
  overlapping option helpers, and direct struct mutation in tests.

Scope:
- Define one canonical typed config input for `core.New(...)`.
- Migrate first-party callers and tests to the chosen config path.
- Remove overlapping option helpers once their last first-party caller is gone.
- Keep default values unchanged unless a card explicitly says otherwise.

Non-goals:
- Do not redesign route or middleware wiring in this card.
- Do not change actual timeout/TLS default values.
- Do not add environment loading behaviour to `core`.

Files:
- `core/app.go`
- `core/config.go`
- `core/options.go`
- `core/options_test.go`
- `core/introspection_test.go`
- `core/lifecycle_test.go`
- `reference/standard-service/internal/app/app.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `docs/modules/core/README.md`

Tests:
- Remove direct `app.config` mutation from tests that should instead exercise
  the canonical config input path.
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./reference/...`
- `go test -timeout 20s ./cmd/plumego/internal/scaffold/...`
- `go vet ./core/... ./reference/...`

Docs Sync:
- Update the core primer and scaffold output to show the typed config input as
  the only preferred construction path.

Done Definition:
- `core` construction uses one typed config input model.
- Overlapping per-field/grouped option helpers are removed after migration.
- Tests and first-party callers no longer need direct `app.config` mutation to
  exercise supported configuration.

Outcome:
- Added `core.DefaultConfig()` and changed `core.New(...)` to take an explicit
  typed `AppConfig`, making config the single canonical construction input.
- Removed the fragmented config option helpers (`WithAddr`, `WithEnvPath`,
  `WithShutdownTimeout`, `WithServerTimeouts`, `WithMaxHeaderBytes`,
  `WithHTTP2`, `WithTLS`, `WithTLSConfig`, `WithDebug`) after migrating all
  first-party callers and examples to the typed config path.
- Updated core tests, reference apps/config defaults, scaffold output,
  dashboard wiring, and root/core docs so they all show the same
  `DefaultConfig()` -> mutate fields -> `core.New(cfg, ...)` model.
