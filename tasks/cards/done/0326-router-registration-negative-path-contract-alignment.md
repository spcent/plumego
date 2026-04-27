# Card 0326: Router Registration Negative-Path Contract Alignment

Priority: P1
State: done
Primary Module: router

## Goal

Make the `router` registration failure contract internally consistent so
duplicate, conflict, and frozen-router failures are modeled and tested through
the same non-panic `error` path across the stable surface.

## Problem

- `router/registration.go` models registration failures through `AddRoute(... ) error`.
- `router/router_conflict_test.go` and `core/routing_test.go` already assert
  duplicate/conflict/frozen cases through returned errors.
- `router/negative_test.go` still documents and asserts panic behavior for the
  same failure class by routing through `mustAddRoute(...)`, which panics on any
  registration error.
- That leaves the stable router contract split between the real exported API
  (`error`) and a stale panic-oriented negative test shape.

This is small on code size but high on control-plane clarity: route
registration is one of the main stable surfaces and should have one obvious
failure model.

## Scope

- Normalize router negative-path tests to assert returned errors instead of panic
  recovery for duplicate/frozen registration failures.
- Tighten router-facing docs and test comments so they describe registration
  failure as an `error` contract, not a panic contract.
- Keep helper usage (`mustAddRoute`) limited to success-path test setup only.

## Non-Goals

- Do not change successful route registration semantics.
- Do not add new public router APIs.
- Do not change route conflict rules or dispatch behavior in this card.

## Files

- `router/negative_test.go`
- `router/test_helpers_test.go`
- `router/router_conflict_test.go`
- `docs/modules/router/README.md`
- any other `router/*_test.go` files that still describe panic-oriented
  registration behavior

## Tests

- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`

## Docs Sync

- Keep `docs/modules/router/README.md` aligned if the negative-path contract is
  made more explicit.

## Done Definition

- Router registration failure tests no longer assert panic behavior for
  duplicate or frozen registration.
- `mustAddRoute(...)` is not used to model expected failure cases.
- Router docs and test comments describe one canonical failure path:
  `AddRoute(... )` returns an error.

## Outcome

- Rewrote `router/negative_test.go` to assert duplicate/frozen registration
  failures through returned errors instead of panic recovery.
- Left `mustAddRoute(...)` as a success-path test helper only.
- Tightened `docs/modules/router/README.md` so router registration failure is
  described as an `error` contract.

## Validation Run

```bash
go test -timeout 20s ./router/...
go test -race -timeout 60s ./router/...
go vet ./router/...
```
