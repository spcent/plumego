# Card 0316: Core Route Metadata Surface Canonicalization

Priority: P1
State: done
Primary Module: core

## Goal

Remove the named-route shorthand helpers from `core` and restore one canonical route-metadata path: `App.AddRoute(..., router.WithRouteName(...))`.

## Problem

- `core/routing.go` currently exposes `GetWithName`, `PostWithName`, `PutWithName`, `DeleteWithName`, `PatchWithName`, and `AnyWithName`.
- This directly conflicts with the current control plane:
  - `docs/CANONICAL_STYLE_GUIDE.md` says named routes must use `app.AddRoute(..., router.WithRouteName(...))` and explicitly forbids named-route helper aliases.
  - `docs/modules/core/README.md` and `docs/modules/router/README.md` describe `AddRoute(..., WithRouteName(...))` as the canonical metadata path.
- The result is a stable public API that now advertises two ways to express the same metadata, even though the repository style contract says there should be only one.
- `rg -n 'GetWithName|PostWithName|PutWithName|DeleteWithName|PatchWithName|AnyWithName' .` currently shows definitions in `core/routing.go` and no live production callers, which makes this cleanup low-risk but still a symbol-removal task under `AGENTS.md §7.1`.

## Scope

- Remove the six `*WithName` helpers from `core/routing.go`.
- Update `core` tests to assert the canonical named-route path through `AddRoute(..., router.WithRouteName(...))`.
- Run the symbol-removal completeness protocol:
  - enumerate every `*WithName` reference before editing
  - migrate or delete each site in the same change
  - re-run the search and confirm no residual references remain

## Non-Goals

- Do not remove the method shorthands `Get`, `Post`, `Put`, `Delete`, `Patch`, or `Any`.
- Do not change `router.WithRouteName(...)` behavior.
- Do not redesign route registration beyond restoring the single canonical metadata path.

## Files

- `core/routing.go`
- `core/routing_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
- `go build ./...`

## Docs Sync

- None expected if code is brought back into alignment with the already-updated style and module docs.

## Done Definition

- `core` no longer exports named-route shorthand aliases.
- The only supported named-route path in stable `core` is `AddRoute(..., router.WithRouteName(...))`.
- The pre/post removal grep for all six symbols is clean.
- `core` tests and repo build pass.

## Outcome

- Removed `GetWithName`, `PostWithName`, `PutWithName`, `DeleteWithName`, `PatchWithName`, and `AnyWithName` from `core/routing.go`.
- Kept named-route coverage on the canonical `App.AddRoute(..., router.WithRouteName(...))` path.
- Extended the canonical named-route test to cover the `ANY` registration path without reintroducing alias helpers.

## Validation Run

```bash
gofmt -w core/routing.go core/routing_test.go
rg -n 'GetWithName|PostWithName|PutWithName|DeleteWithName|PatchWithName|AnyWithName' . --glob '*.go'
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go build ./...
```
