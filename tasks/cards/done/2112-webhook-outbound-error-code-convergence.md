# Card 2112: Webhook Outbound Error Code Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: done
Primary Module: x/webhook
Owned Files:
- x/webhook/out.go
- x/webhook/outbound_test.go
- x/webhook/webhook_component_test.go
- docs/modules/x-webhook/README.md
Depends On: none

Goal:
- Finish outbound webhook HTTP error-code convergence.
- Remove repeated one-off `NewErrorBuilder().Type(...).Message(...)` branches that lack explicit stable codes.

Scope:
- Audit outbound route handlers in `x/webhook/out.go`.
- Add local, canonical helpers for required, not-found, forbidden, unauthorized, validation, and internal response errors where that reduces duplication.
- Keep fail-closed auth behavior and timing-safe token comparison intact.
- Add or update tests for missing endpoint IDs, missing requests, forbidden/unauthorized tokens, validation failures, and internal service failures.

Non-goals:
- Do not change inbound verification behavior.
- Do not change outbound delivery retry, queue, signer, or store semantics.
- Do not promote `x/webhook` into the app-facing messaging family root.

Files:
- `x/webhook/out.go`: consolidate outbound handler error construction and codes.
- `x/webhook/outbound_test.go`: cover outbound handler error branches.
- `x/webhook/webhook_component_test.go`: update component-level response assertions if needed.
- `docs/modules/x-webhook/README.md`: document outbound error codes if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/webhook/...`
- `go test -timeout 20s ./x/webhook/...`
- `go vet ./x/webhook/...`

Docs Sync:
- Required if public response codes/messages or route behavior change.

Done Definition:
- Outbound webhook route errors have explicit stable codes.
- Repeated builder chains are reduced to one canonical local construction path.
- Auth and verification remain fail-closed.
- The three listed validation commands pass.

Outcome:
- Added local outbound webhook error helpers for required, not-found, forbidden, unauthorized, invalid JSON, validation, and internal failures.
- Migrated outbound route handler branches to those helpers while preserving fail-closed trigger-token behavior and timing-safe comparison.
- Added tests for missing IDs, missing targets, unauthorized trigger tokens, invalid JSON, and validation failures.
- Documented outbound route error-code expectations in `docs/modules/x-webhook/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./x/webhook/...`
  - `go test -timeout 20s ./x/webhook/...`
  - `go vet ./x/webhook/...`
