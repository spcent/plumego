# Card 0345: X Webhook Transport Surface Cleanup

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/webhook
Depends On: —

## Goal

Clean up `x/webhook`'s HTTP transport surface so route registration stays
explicit and error handling no longer depends on a pile of per-scenario helper
constructors.

## Problem

- `x/webhook/in.go` and `x/webhook/out.go` both hide route registration behind
  `routesOnce sync.Once`, which turns `RegisterRoutes` into stateful behavior
  instead of explicit router wiring.
- The same files also define local `err*` helper families such as
  `errBadRequest`, `errInvalidJSON`, `errNotFound`, `errInMissingSecret`, and
  `errInPublishFailed`, even though the handlers already depend on
  `contract.WriteError` and `contract.NewErrorBuilder`.
- That combination leaves `x/webhook` with hidden registration state plus a
  second local error vocabulary that drifts from the repository rule of one
  canonical error-construction path per layer.

## Scope

- Remove stateful `routesOnce` registration guards from inbound and outbound
  webhook route registration.
- Converge inbound/outbound webhook handlers on one local error-construction
  pattern built directly on `contract.NewErrorBuilder`, not per-scenario helper
  wrappers.
- Update focused tests for repeat registration behavior, verification failures,
  and webhook CRUD/trigger deny paths.

## Non-Goals

- Do not redesign signature verification, deduplication, delivery retry, or
  store semantics.
- Do not widen this card into broader `x/messaging` family API redesign.
- Do not weaken fail-closed behavior on missing secrets, invalid signatures, or
  trigger token checks.

## Files

- `x/webhook/in.go`
- `x/webhook/out.go`
- `x/webhook/*_test.go`
- `docs/modules/x-webhook/README.md` only if route registration or error policy
  wording changes

## Tests

```bash
rg -n 'routesOnce sync\\.Once|errBadRequest\\(|errInvalidJSON\\(|errNotFound\\(|errStoreError\\(|errForbidden\\(|errUnauthorized\\(|errInMissingSecret\\(|errInInvalidSignature\\(|errInPublishFailed\\(' x/webhook -g '*.go'
go test -timeout 20s ./x/webhook/...
go vet ./x/webhook/...
```

## Docs Sync

- `docs/modules/x-webhook/README.md` only if it still implies stateful route
  registration or documents the removed helper pattern

## Done Definition

- `x/webhook` no longer uses `sync.Once` to hide HTTP route registration state.
- Inbound and outbound handlers share one direct `contract.NewErrorBuilder`
  error-construction pattern instead of per-scenario `err*` helpers.
- Focused webhook tests cover the cleaned-up route registration and deny-path
  behavior.

## Outcome

- Removed `routesOnce` from inbound and outbound webhook components so route
  registration is explicit and duplicate registration now surfaces as a router
  error instead of being silently ignored.
- Replaced inbound and outbound `err*` helper families with direct
  `contract.NewErrorBuilder` + `contract.WriteError` calls at the transport
  site.
- Added focused tests covering duplicate route registration, inbound missing
  secret fail-closed behavior, and outbound trigger-token forbidden responses.

## Validation Run

```bash
rg -n 'routesOnce sync\.Once|errBadRequest\(|errInvalidJSON\(|errNotFound\(|errStoreError\(|errForbidden\(|errUnauthorized\(|errInMissingSecret\(|errInInvalidSignature\(|errInPublishFailed\(' x/webhook -g '*.go'
go test -timeout 20s ./x/webhook/...
go vet ./x/webhook/...
```
