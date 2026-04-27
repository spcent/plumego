# Card 0280: Contract Ctx Adapter Migration

Priority: P1
State: done
Primary Module: contract

## Goal

Remove the remaining non-canonical `contract.CtxHandlerFunc` and `contract.AdaptCtxHandler` surface by migrating extension and devserver handlers to `func(http.ResponseWriter, *http.Request)` or package-local adapters.

## Problem

Card 0925 prunes `contract.Ctx` response shortcuts, streaming helpers, and hidden request lifecycle behavior, but `AdaptCtxHandler` remains because deleting it in the same pass expands across multiple extension families:

- `x/rest`
- `x/messaging`
- `x/ops`
- `x/webhook`
- `x/devtools/pubsubdebug`
- `cmd/plumego/internal/devserver`

The stable style guide keeps canonical handlers as `func(http.ResponseWriter, *http.Request)`. Leaving the adapter in stable `contract` keeps a framework-style handler shape available as a stable entrypoint.

## Scope

- Enumerate all `contract.AdaptCtxHandler` and `contract.CtxHandlerFunc` references before editing.
- Migrate each owning module to explicit stdlib handlers or a package-local transition adapter.
- Remove `CtxHandlerFunc` and `AdaptCtxHandler` from `contract`.
- Keep `contract.Ctx` only if still needed for narrow binding helpers after migration; otherwise remove it in the same pass.
- Update tests and docs to avoid advertising Ctx handlers.

## Non-Goals

- Do not change response envelope shape.
- Do not reintroduce response shortcuts on `Ctx`.
- Do not move business DTO logic into middleware.
- Do not add new stable helper aliases for route registration.

## Expected Files

- `contract/context_core.go`
- `x/rest/*`
- `x/messaging/*`
- `x/ops/*`
- `x/webhook/*`
- `x/devtools/pubsubdebug/*`
- `cmd/plumego/internal/devserver/*`
- `docs/modules/contract/README.md`
- `contract/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./contract ./x/rest ./x/messaging ./x/ops ./x/webhook ./x/devtools/...
go test -race -timeout 60s ./contract
go vet ./contract ./x/rest ./x/messaging ./x/ops ./x/webhook ./x/devtools/...
```

Also run the CLI submodule gate:

```bash
(cd cmd/plumego && go test -timeout 20s ./...)
```

Then run the required repo-wide gates before committing.

## Done Definition

- `contract.AdaptCtxHandler` and `contract.CtxHandlerFunc` have zero residual Go references.
- Stable `contract` no longer advertises a non-canonical handler shape.
- Affected extension handlers remain explicit and pass package tests.
- Focused gates and repo-wide gates pass.

## Outcome

- Removed `CtxHandlerFunc` and `AdaptCtxHandler` from `contract`.
- Added package-local adapters in `x/rest`, `x/messaging`, `x/ops`, `x/webhook`, and `x/devtools/pubsubdebug`.
- Updated devserver routes to build `contract.Ctx` locally without stable adapters.
