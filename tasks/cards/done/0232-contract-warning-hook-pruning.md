# Card 0232

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/trace.go`
- `contract/errors_test.go`
- `contract/trace_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`

Goal:
- Remove package-global warning hooks from stable `contract`.
- Make transport helpers deterministic and side-effect-free instead of relying on mutable global diagnostics.

Problem:
- `contract.WarnFunc` is a mutable package-global hook used by `WriteError(...)` and `WithSpanIDString(...)`, which violates the repo rule against hidden globals in stable roots.
- The current warning path preserves compatibility for partially-populated `APIError` values and invalid span ids by emitting side-channel diagnostics, but the user requested full convergence rather than compatibility retention.
- The stable transport layer should not depend on mutable global callbacks for correctness or observability.

Scope:
- Remove `contract.WarnFunc` entirely.
- Make `WriteError(...)` normalize incomplete `APIError` values without global side effects.
- Make `WithSpanIDString(...)` deterministic without warning callbacks; keep only the canonical valid trace carrier behavior.
- Update tests and docs in the same change so no code still depends on the removed hook.

Non-goals:
- Do not redesign the structured error schema.
- Do not move request-id or trace carriers out of `contract`.
- Do not introduce replacement logging hooks or app-level diagnostics globals.

Files:
- `contract/errors.go`
- `contract/trace.go`
- `contract/errors_test.go`
- `contract/trace_test.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`

Tests:
- `go test -timeout 20s ./contract/...`
- `go test -race -timeout 60s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- Keep the contract docs aligned on explicit transport helpers with no hidden global diagnostics path.

Done Definition:
- `contract.WarnFunc` is removed with zero residual references.
- `WriteError(...)` and trace helpers behave deterministically without package-global callbacks.
- Contract docs/manifests no longer imply or mention mutable warning hooks.

Outcome:
- Completed.
- Removed the package-global `contract.WarnFunc` hook and the remaining warning-path behavior from stable `contract`.
- `WriteError(...)` now normalizes incomplete `APIError` values without side effects, and `WithSpanIDString(...)` ignores invalid span ids instead of storing malformed carrier values.
- Updated contract tests and primer text to match the deterministic transport-only behavior.
