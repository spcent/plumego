# Card 0755

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_response.go`
- `contract/response.go`
- `x/messaging/api.go`
- `x/ops/ops.go`
- `x/webhook/in.go`
- `x/webhook/out.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `docs/modules/contract/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

Goal:
- Collapse `contract` to one canonical success-response path for `Ctx`-based
  handlers so first-party code no longer mixes raw JSON writes and enveloped
  transport responses arbitrarily.

Problem:
- `Ctx.Response` delegates to `WriteResponse` and produces the standard
  `{data, meta, trace_id}` envelope.
- `Ctx.JSON` writes raw JSON directly with its own encoding path and no
  standardized envelope.
- First-party `Ctx` handlers in `x/messaging`, `x/ops`, `x/webhook`, and the
  dev dashboard use both styles interchangeably, so the package advertises two
  equally-valid success contracts at the same layer.
- This conflicts with the repository rule that each layer should expose one
  canonical success-response path.

Scope:
- Choose a single canonical success writer for `Ctx` handlers.
- Remove or quarantine the duplicate `Ctx` success helper that diverges from the
  canonical path.
- Migrate first-party `Ctx` handlers to the canonical writer.
- Keep the lower-level stdlib writer surface (`WriteJSON` / `WriteResponse`)
  explicit and documented instead of letting `Ctx` grow parallel convenience
  APIs.

Non-goals:
- Do not redesign the success envelope shape in this card.
- Do not change stdlib-only handlers that intentionally use raw payload writes
  outside the `Ctx` layer unless they are needed to keep docs/examples
  consistent.
- Do not add feature-specific response helpers.

Files:
- `contract/context_response.go`
- `contract/response.go`
- `x/messaging/api.go`
- `x/ops/ops.go`
- `x/webhook/in.go`
- `x/webhook/out.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `docs/modules/contract/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

Tests:
- Add or update coverage so canonical `Ctx` success writes surface the expected
  payload shape and trace behavior.
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./x/messaging/... ./x/ops/... ./x/webhook/...`
- `go test -timeout 20s ./internal/devserver/...` (run from `cmd/plumego/`)

Docs Sync:
- Update contract module docs and the style guide so `Ctx` handlers advertise
  only the canonical success writer.

Done Definition:
- `Ctx` exposes one canonical success-response path.
- First-party `Ctx` handlers no longer mix raw JSON and enveloped success writes
  at the same layer.
- Contract docs describe one preferred success path for `Ctx` handlers.

Outcome:
- Removed `Ctx.JSON` from `contract` so `Ctx` now exposes a single canonical
  success writer: `Ctx.Response`.
- Migrated all first-party `Ctx.JSON` callers in `x/messaging`, `x/rest`, and
  `cmd/plumego/internal/devserver` to `Ctx.Response` for success writes and
  existing structured error paths for failures.
- Updated contract docs and the canonical style guide to describe `Ctx.Response`
  / `WriteResponse` as the only preferred success path for `Ctx` handlers, with
  `WriteJSON` kept as an explicit low-level raw writer.
- Validation:
  - `go test -race -timeout 60s ./contract/...`
  - `go test -timeout 20s ./x/messaging/... ./x/ops/... ./x/webhook/... ./x/rest/...`
  - `go test -timeout 20s ./internal/devserver/...` (from `cmd/plumego/`)
  - `go vet ./contract/... ./x/messaging/... ./x/ops/... ./x/webhook/... ./x/rest/...`
  - `go vet ./internal/devserver/...` (from `cmd/plumego/`)
