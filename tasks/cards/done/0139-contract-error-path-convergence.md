# Card 0139

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/context_response.go`
- `contract/error_utils.go`
- `contract/errors_test.go`
- `contract/error_handling_test.go`
- `x/messaging/api.go`
- `docs/modules/contract/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

Goal:
- Keep one explicit error-construction and error-write path in `contract`, and
  remove the extra convenience surfaces that currently behave like a mini error
  framework.

Problem:
- `WriteError` + `NewErrorBuilder` is the intended structured error path.
- `Ctx.ErrorJSON` constructs API errors through a separate ad hoc helper based
  on status/code/message/details.
- `HandleError`, `HandlePanic`, `SafeExecute`, and `SafeExecuteWithResult` add
  wrapping, logging, panic recovery, and response writing convenience inside
  `contract`, even though they are not part of the transport contract itself.
- `ErrorLogger` and `logErrorWithContext` also create two logging shapes for the
  same error data.
- Repo usage already prefers `WriteError` directly, while the extra helpers are
  mostly unused outside tests and a narrow set of first-party callers.

Scope:
- Define one canonical `contract` error path centered on `APIError`,
  `NewErrorBuilder`, and `WriteError`.
- Remove duplicate `Ctx`/utility helpers that hide or duplicate that path.
- Migrate first-party callers still using the duplicate helpers.
- Keep bind-error adaptation only if it remains a thin adapter into the
  canonical error path rather than a parallel framework.

Non-goals:
- Do not redesign the JSON error envelope in this card.
- Do not move domain-specific error mapping into `contract`.
- Do not add new logging abstractions to replace the removed helpers.

Files:
- `contract/errors.go`
- `contract/context_response.go`
- `contract/error_utils.go`
- `contract/errors_test.go`
- `contract/error_handling_test.go`
- `x/messaging/api.go`
- `docs/modules/contract/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

Tests:
- Add or update tests to cover the canonical builder + writer path after the
  duplicate helpers are removed.
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./x/messaging/...`

Docs Sync:
- Update contract docs and style guidance so error construction/writing is
  documented through one path only.

Done Definition:
- `contract` exposes one explicit error-construction and write path.
- Duplicate error convenience helpers are removed.
- First-party callers no longer depend on the removed helpers.

Outcome:
- Removed duplicate `contract` error helpers: `Ctx.ErrorJSON`, `HandleError`,
  `HandlePanic`, `SafeExecute`, `SafeExecuteWithResult`, `ErrorLogger`, and the
  old `error_utils.go` mini-framework.
- Standardized first-party callers on `NewErrorBuilder().Build()` plus
  `WriteError`, including `x/messaging`, `x/rest`, and
  `cmd/plumego/internal/devserver`.
- Tightened `NewErrorBuilder` so category is derived during `Build()` when the
  caller sets status without an explicit category, which makes the canonical
  builder path usable without a parallel helper.
- Updated tests and docs to describe one error path centered on
  `NewErrorBuilder` + `WriteError`.
- Validation:
  - `go test -race -timeout 60s ./contract/...`
  - `go test -timeout 20s ./x/messaging/... ./x/rest/... ./x/webhook/... ./x/devtools/... ./x/scheduler/... ./x/ai/streaming/... ./x/websocket/... ./x/pubsub/... ./core/...`
  - `go test -timeout 20s ./internal/devserver/...` (from `cmd/plumego/`)
  - `go vet ./contract/... ./x/messaging/... ./x/rest/... ./x/webhook/... ./x/devtools/... ./x/scheduler/... ./x/ai/streaming/... ./x/websocket/... ./x/pubsub/... ./core/...`
  - `go vet ./internal/devserver/...` (from `cmd/plumego/`)
  - `go test -timeout 20s ./...`
  - `go vet ./...`
  - `go test -timeout 20s ./...` (from `cmd/plumego/`)
  - `go vet ./...` (from `cmd/plumego/`)
