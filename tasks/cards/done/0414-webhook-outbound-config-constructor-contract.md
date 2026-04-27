# Card 0414: Webhook Outbound Config Constructor Contract

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/webhook
Owned Files:
- x/webhook/outbound_config.go
- x/webhook/outbound_test.go
- docs/modules/x-webhook/README.md
Depends On: none

Goal:
- Add an explicit error-returning construction path for outbound webhook config loaded from a reader.
- Preserve current panic compatibility while avoiding panic-only behavior for nil configuration readers.

Scope:
- Audit `ConfigFromReader` nil-reader behavior.
- Add a safe `ConfigFromReaderE`-style variant or equivalent local pattern that returns a typed/sentinel error.
- Keep environment parsing, defaults, clamping, drop-policy fallback, and outbound delivery behavior unchanged.
- Add focused tests for nil reader error behavior and existing panic compatibility.

Non-goals:
- Do not change inbound verification, trigger auth, outbound queue, retry, or delivery semantics.
- Do not promote `x/webhook` into the app-facing messaging family root.
- Do not add new configuration dependencies.

Files:
- `x/webhook/outbound_config.go`: add explicit error-returning config reader path and preserve panic wrapper.
- `x/webhook/outbound_test.go`: cover nil reader error path and panic compatibility.
- `docs/modules/x-webhook/README.md`: document the safe config reader path if public guidance changes.

Tests:
- `go test -race -timeout 60s ./x/webhook/...`
- `go test -timeout 20s ./x/webhook/...`
- `go vet ./x/webhook/...`

Docs Sync:
- Required if public constructor guidance or examples change.

Done Definition:
- Nil outbound config readers can be handled without panic through an explicit error-returning API.
- Existing `ConfigFromReader` behavior remains source-compatible for current callers.
- Focused tests cover both safe and compatibility paths.
- The three listed validation commands pass.

Outcome:
- Added `ConfigFromReaderE` and comparable `ErrNilValueReader` for explicit nil-reader handling.
- Preserved `ConfigFromReader` as the panic-compatible wrapper around the error-returning path.
- Kept environment parsing and config default/clamping behavior unchanged.
- Documented the safe config reader constructor in `docs/modules/x-webhook/README.md`.
- Added tests for nil-reader error behavior, panic compatibility, and valid reader defaults.
- Validation passed:
  - `go test -race -timeout 60s ./x/webhook/...`
  - `go test -timeout 20s ./x/webhook/...`
  - `go vet ./x/webhook/...`
