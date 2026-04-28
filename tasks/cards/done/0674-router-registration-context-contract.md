# 0674 - router registration and context contract

State: done
Priority: P0
Primary module: `router`

## Goal

Make route registration fail predictably for invalid handlers and prevent
non-parameterized route dispatch from inheriting stale path params from an
incoming request context.

## Scope

- Reject nil handlers in `Router.AddRoute`.
- Keep registration errors explicit and non-panic oriented.
- Ensure dispatch replaces route params with the current match state, including
  clearing params for routes without path parameters.
- Add focused tests for nil handler rejection and stale-param cleanup.

## Non-goals

- Do not change exported router symbols.
- Do not add new registration helper families.
- Do not alter route matching precedence.

## Files

- `router/registration.go`
- `router/dispatch.go`
- `router/router_contract_test.go`

## Tests

- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`

## Docs Sync

No docs sync expected; this hardens existing registration and context behavior.

## Done Definition

- `AddRoute` returns an error for nil handlers before mutating router state.
- Exact/static routes do not expose stale route params already present on the
  request context.
- Router validation commands pass.

## Outcome

- Rejected nil handlers before route state is mutated.
- Flattened registration error construction for frozen, duplicate, and
  parameter-conflict paths.
- Replaced route params on every matched request so no-param routes clear stale
  params from an incoming request context.
- Validation:
  - `go test -timeout 20s ./router/...`
  - `go test -race -timeout 60s ./router/...`
  - `go vet ./router/...`
