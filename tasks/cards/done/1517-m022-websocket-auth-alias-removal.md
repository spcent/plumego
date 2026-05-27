# Card 1517

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/auth.go`
- `x/websocket/auth_test.go`
- `docs/modules/x/websocket/README.md`
- `specs/deprecation-inventory.yaml`
Depends On: none

Goal:
- Remove the dead `x/websocket` auth compatibility aliases once their last repo
  callers are migrated, instead of keeping them under an open-ended
  compatibility inventory entry.

Scope:
- Delete `NewHS256TokenAuth` and `(*SimpleHS256TokenAuth).AuthenticateToken`.
- Migrate local tests to `NewSimpleHS256TokenAuth` and `VerifyJWT`.
- Shrink the websocket compatibility inventory entry to the aliases that still
  have live callers.

Non-goals:
- Do not remove `NewHubE` or `ReadMessageStream` in this card.
- Do not change websocket auth behavior beyond alias removal.

Files:
- `x/websocket/auth.go`
- `x/websocket/auth_test.go`
- `docs/modules/x/websocket/README.md`
- `specs/deprecation-inventory.yaml`

Acceptance Tests:
- `go test -timeout 20s ./x/websocket/...`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- `docs/modules/x/websocket/README.md`

Validation:
- `go test -timeout 20s ./x/websocket/...`
- `go run ./internal/checks/deprecation-inventory -strict`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Removed the dead websocket auth aliases `NewHS256TokenAuth` and
  `(*SimpleHS256TokenAuth).AuthenticateToken` after migrating their remaining
  repo callers to `NewSimpleHS256TokenAuth` and `VerifyJWT`.
- Updated the websocket docs to point callers at the canonical auth names and
  shrank the compatibility inventory entry so it now tracks only the aliases
  that still have live compatibility callers (`NewHubE` and
  `ReadMessageStream`).
- Validation:
  - `go test -timeout 20s ./x/websocket/...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`
