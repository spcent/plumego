# Card 0757

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- `contract/validation.go`
- `contract/context_test.go`
- `contract/context_extended_test.go`
- `contract/active_cards_regression_test.go`
- `x/messaging/api.go`
- `docs/modules/contract/README.md`

Goal:
- Reduce the bind/validate API in `contract` to one clear binding model per
  request source, without silent option dropping or duplicated “bind + validate”
  convenience families.

Problem:
- JSON binding currently exposes `BindJSON`, `BindJSONWithOptions`,
  `BindAndValidateJSON`, and `BindAndValidateJSONWithOptions`.
- Query binding exposes `BindQuery`, `BindAndValidateQuery`, and
  `BindAndValidateQueryWithOptions`, but some `BindOptions` fields are silently
  ignored for query sources.
- Validation is duplicated across these helpers through repeated
  “bind then validate then log” flows.
- This makes binding behavior harder to predict and leaves no single obvious API
  surface for callers.

Scope:
- Converge on one canonical bind entry per source (`json`, `query`) with
  explicit option semantics.
- Remove or fold duplicated “bind + validate” helpers into a smaller surface.
- Ensure option handling is explicit; unsupported options must not be silently
  ignored.
- Keep field-error extraction and bind-error-to-API-error mapping aligned with
  the converged binding surface.

Non-goals:
- Do not redesign validation rule syntax in this card.
- Do not add mixed-source binding or framework-style auto-binding.
- Do not move validation into middleware.

Files:
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- `contract/validation.go`
- `contract/context_test.go`
- `contract/context_extended_test.go`
- `contract/active_cards_regression_test.go`
- `x/messaging/api.go`
- `docs/modules/contract/README.md`

Tests:
- Add or update tests to cover the reduced binding surface, explicit option
  behavior, and validation error mapping.
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./x/messaging/...`

Docs Sync:
- Update contract module docs so JSON/query binding has one documented entry
  pattern per source.

Done Definition:
- `contract` exposes one clear binding model per source.
- Silent option dropping is removed.
- First-party callers use the converged binding surface.

Outcome:
- Removed duplicated bind/validate helpers from `contract`: `BindJSONWithOptions`,
  `BindAndValidateJSON`, `BindAndValidateJSONWithOptions`,
  `BindAndValidateQuery`, and `BindAndValidateQueryWithOptions`.
- Converged on one explicit bind entry per source:
  - `BindJSON(dst, BindOptions{...})` for JSON
  - `BindQuery(dst)` for query
  - `ValidateStruct(dst)` as the explicit second-step validator
- Reduced `BindOptions` to JSON-only fields so query binding no longer accepts
  silently-ignored validation options.
- Migrated first-party callers in `x/messaging` and `x/ops` to the explicit
  `Bind*` then `ValidateStruct` flow, using `WriteBindError` for failures.
- Removed hidden bind-layer logging side effects and updated tests/docs to
  assert the explicit two-step model.
- Validation:
  - `go test -race -timeout 60s ./contract/...`
  - `go test -timeout 20s ./x/messaging/... ./x/ops/...`
  - `go vet ./contract/... ./x/messaging/... ./x/ops/...`
  - `go test -timeout 20s ./...`
  - `go vet ./...`
  - `go test -timeout 20s ./...` (from `cmd/plumego/`)
  - `go vet ./...` (from `cmd/plumego/`)
