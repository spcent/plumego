# Card 0282: Contract Doc And Spec Drift Convergence

Priority: P1
State: done
Primary Module: contract

## Goal

Align canonical docs and machine-readable specs with the current `contract` surface so agents no longer receive conflicting guidance around `Ctx` response helpers or non-canonical handler shapes.

## Problem

Recent contract pruning removed the `Ctx` success-response path and narrowed `Ctx` toward legacy binding and route-param access, but the canonical guidance still contains stale instructions:

- `docs/CANONICAL_STYLE_GUIDE.md` still presents `ctx.Response(...)` as a canonical success path.
- `specs/dependency-rules.yaml` still lists `ctx.Response(...)` in the `response_write` rule.
- `docs/modules/contract/README.md` correctly narrows `Ctx`, but should explicitly say response writing is on stdlib-shaped handlers via `WriteResponse`.
- Card 0928 covers the remaining `CtxHandlerFunc` / `AdaptCtxHandler` migration; this card should keep docs consistent with that direction without implementing 0928.

This creates boundary drift: future agents may reintroduce `Ctx` response helpers because the specs still advertise them.

## Scope

- Grep all active canonical guidance for stale `Ctx` success-response and handler-shape references before editing.
- Update `docs/CANONICAL_STYLE_GUIDE.md` so success responses use only `contract.WriteResponse(w, r, status, data, meta)` on stdlib handlers.
- Update `specs/dependency-rules.yaml` to remove `ctx.Response(...)` from the canonical response-write rule.
- Update `docs/modules/contract/README.md` to state that `Ctx` is not a response-writing surface.
- Preserve historical references under `tasks/cards/done/*`; do not rewrite completed task history.
- Do not remove `CtxHandlerFunc` or `AdaptCtxHandler`; that belongs to Card 0928.

## Non-Goals

- Do not change `contract` runtime code in this card unless a doc-only check requires a manifest wording adjustment.
- Do not change response envelope behavior.
- Do not migrate extension handlers; execute Card 0928 separately.
- Do not add new contract helper APIs.

## Expected Files

- `docs/CANONICAL_STYLE_GUIDE.md`
- `specs/dependency-rules.yaml`
- `docs/modules/contract/README.md`
- `contract/module.yaml` only if the manifest wording still conflicts after docs sync

## Validation

Run focused gates first:

```bash
go run ./internal/checks/agent-workflow
go run ./internal/checks/dependency-rules
go test -timeout 20s ./contract/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- Canonical docs/specs no longer recommend `ctx.Response(...)`.
- `contract` docs describe `WriteResponse` as the only stable success-response path.
- Card 0928 remains the implementation task for deleting the remaining adapter surface.
- Focused gates and repo-wide gates pass.

## Outcome

- Removed `ctx.Response(...)` guidance from `docs/CANONICAL_STYLE_GUIDE.md`.
- Updated `specs/dependency-rules.yaml` to list only `WriteResponse(...)` as the success response path.
- Clarified in `docs/modules/contract/README.md` that `Ctx` does not provide response-writing helpers.
