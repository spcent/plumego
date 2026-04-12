# Card 0858

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/error_wrap.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `core/app_helpers.go`
- `core/routing.go`
- `router/registration.go`

Goal:
- Keep stable `contract` focused on API error payloads and write helpers, not repo-wide wrapped-error utilities or response-parsing convenience.
- Remove generic error utility ownership from stable transport contracts.

Problem:
- Stable `contract` still exports `WrapError`, `WrapErrorf`, `GetErrorDetails`, `FormatError`, `PanicToError`, `ParseErrorFromResponse`, `IsClientError`, `IsServerError`, and `IsAPIErrorRetryable`.
- `core` and `router` currently depend on `contract.WrapError(...)` for internal registration and setup errors, which is not a transport primitive.
- This leaves `contract` acting as both the HTTP error model and a repo-wide generic error utility package, which conflicts with the rule that `contract` should contain transport primitives only.

Scope:
- Remove or relocate generic wrapped-error and response-parsing/classification utilities out of stable `contract`.
- Update `core` and `router` callers in the same change.
- Keep `APIError`, `NewErrorBuilder`, `WriteError`, `WriteResponse`, `WriteJSON`, and bind-error helpers intact.
- Sync docs and manifests to the reduced transport-only contract boundary.

Non-goals:
- Do not change the on-wire error schema.
- Do not redesign `NewErrorBuilder` or `WriteError`.
- Do not add deprecated forwarding wrappers for removed utility helpers.

Files:
- `contract/errors.go`
- `contract/error_wrap.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `core/app_helpers.go`
- `core/routing.go`
- `router/registration.go`

Tests:
- `go test -timeout 20s ./contract/... ./core ./router`
- `go test -race -timeout 60s ./contract/... ./core ./router`
- `go vet ./contract/... ./core ./router`

Docs Sync:
- Keep contract docs aligned on the rule that stable `contract` owns transport error contracts only.

Done Definition:
- Stable `contract` no longer exports repo-wide wrapped-error or response-parsing utility helpers outside the transport contract boundary.
- `core` and `router` no longer depend on `contract` for generic internal error wrapping.
- Contract docs/manifests describe the same transport-only boundary implemented in code.

Outcome:
- Completed.
- Removed generic wrapped-error, retry classification, panic-recovery, and HTTP response-parsing utilities from stable `contract`, leaving the transport error model and write helpers intact.
- Switched `core` and `router` back to local `fmt.Errorf`-based registration/setup errors instead of depending on `contract` for generic internal wrapping.
- Updated `contract` docs and manifest language to make repo-wide error utility ownership explicitly out of scope for the stable transport contract layer.
