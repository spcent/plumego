# Card 0558

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/README.md
- reference/standard-service/README.md
Depends On: 2267

Goal:
Keep `plumego new` scaffold output aligned with `reference/standard-service` as the canonical app layout.

Scope:
- Compare scaffold templates with `reference/standard-service` wiring and config shape.
- Remove scaffold drift where generated projects teach a different bootstrap path.
- Add or update scaffold tests that assert generated files avoid stale TODOs and use canonical route/middleware patterns.
- Update CLI docs only for implemented scaffold behavior.

Non-goals:
- Do not add new scaffold templates in this card.
- Do not change stable root APIs.
- Do not introduce third-party dependencies.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`
- `reference/standard-service/README.md`

Tests:
- `go test -timeout 20s ./cmd/plumego/internal/scaffold/...`
- `go test -timeout 20s ./cmd/plumego/commands/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Required for CLI scaffold behavior.

Done Definition:
- Generated starter projects teach the same app structure as `reference/standard-service`.
- Scaffold tests lock the canonical bootstrap and no-bare-TODO expectations.

Outcome:
Completed. Updated the canonical scaffold route and handler templates to match
the current `reference/standard-service` surface more closely, including
explicit `http.HandlerFunc` route binding, `/api/status`, `/api/v1/greet`, and
local response DTOs. Added scaffold tests to lock the canonical route and
handler surface, and documented the `canonical` template in CLI/reference docs.
