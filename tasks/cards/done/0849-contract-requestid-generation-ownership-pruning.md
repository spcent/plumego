# Card 0849

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/request_id_generation.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `middleware/requestid`
- `middleware/internal/observability`

Goal:
- Remove request-id generation ownership and hidden global generator state from stable `contract`.
- Keep `contract` limited to request-id carriage (`WithRequestID`, `RequestIDFromContext`, header constant) while moving generation policy to middleware ownership.

Problem:
- Stable `contract` still exports `RequestIDGenerator`, `NewRequestIDGenerator`, and `NewRequestID()`, even though its manifest now describes transport carriers rather than observability/generation policy.
- `contract/request_id_generation.go` maintains a package-global generator singleton, which conflicts with the repo rule against hidden globals in stable roots.
- Repository usage is effectively middleware-owned, with generation called from `middleware/internal/observability`, not from transport contracts generally.

Scope:
- Move request-id generation and generator state out of stable `contract` into middleware ownership.
- Keep `contract` limited to request-id carriage and transport constants.
- Update middleware callers and tests in the same change.
- Sync contract and middleware docs/manifests to the reduced ownership boundary.

Non-goals:
- Do not change the canonical request-id header or context carrier.
- Do not collapse request correlation into tracing.
- Do not introduce compatibility wrappers in `contract`.

Files:
- `contract/request_id_generation.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `middleware/requestid`
- `middleware/internal/observability`

Tests:
- `go test -timeout 20s ./contract/... ./middleware/requestid ./middleware/internal/observability`
- `go test -race -timeout 60s ./contract/... ./middleware/requestid ./middleware/internal/observability`
- `go vet ./contract/... ./middleware/requestid ./middleware/internal/observability`

Docs Sync:
- Keep contract docs aligned on the rule that `contract` owns only request-id carriers, while generation policy belongs to middleware.

Done Definition:
- Stable `contract` no longer exports request-id generation APIs or hidden generator state.
- Middleware owns request-id generation policy.
- No residual callers reference the removed `contract` generation helpers.

Outcome:
- Completed.
- Moved request-id generation and decode logic out of stable `contract` into `middleware/internal/observability`.
- Stable `contract` now keeps only request-id carriage/header contracts, while middleware-owned code provides generation policy.
- Updated `middleware/requestid` and shared observability helpers to use the middleware-owned generator with no residual `contract.NewRequestID*` APIs.
