# Card 0235

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_core.go`
- `contract/context_abort_test.go`
- `contract/context_extended_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`

Goal:
- Remove framework-style runtime state and service-locator behavior from stable `contract.Ctx`.
- Keep `Ctx` focused on transport helpers: params, binding, validation, responses, cookies, and streaming.

Problem:
- `contract.Ctx` still exports `Set/Get/MustGet`, `Abort/AbortWithStatus`, `Error/CollectedErrors`, `RequestDuration`, `BodySize`, and `IsCompressed`.
- These helpers act like a mutable framework request bag and runtime recorder rather than a transport contract, which conflicts with the repo rule against context service-locator patterns.
- Repository grep currently shows these helpers are exercised only by `contract` tests, not by stable or extension runtime call sites, so the public surface is wider than the real canonical usage.

Scope:
- Remove the string-key request store and the non-canonical runtime bookkeeping helpers from `contract.Ctx`.
- Keep `NewCtx`, bind helpers, param helpers, response helpers, and streaming helpers intact.
- Update `contract` tests, docs, and manifest language in the same change.

Non-goals:
- Do not redesign `BindJSON`, `BindQuery`, `ValidateStruct`, `Response`, or `Stream`.
- Do not change request-id or trace-context carriage.
- Do not introduce replacement wrapper helpers with a different name.

Files:
- `contract/context_core.go`
- `contract/context_abort_test.go`
- `contract/context_extended_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`

Tests:
- `go test -timeout 20s ./contract/...`
- `go test -race -timeout 60s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- Keep contract docs aligned on the rule that `Ctx` is a transport helper, not a framework-local request state container.

Done Definition:
- Stable `contract.Ctx` no longer exposes string-key storage or runtime bookkeeping helpers with no canonical runtime callers.
- The remaining `Ctx` surface is transport-focused and consistent with the contract docs.
- There are zero residual runtime references to the removed `Ctx` helpers.

Outcome:
- Completed.
- Removed framework-style request-bag and runtime-bookkeeping helpers from `contract.Ctx`, including string-key storage, abort state, error collection, request timing, body-size inspection, and compression tracking.
- Kept `Ctx` focused on transport helpers such as params, binding, validation, response writing, cookies, and streaming, with internal cleanup reduced to an unexported release path.
- Updated `contract` docs and manifest language to explicitly forbid turning `Ctx` into a mutable runtime state container.
