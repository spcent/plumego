# Card 0122

Milestone:
Recipe: specs/change-recipes/stable-api.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
  - middleware/timeout/timeout.go
  - middleware/timeout/timeout_test.go
  - docs/modules/middleware/README.md
Depends On: 0121

Goal:
Make timeout middleware stable by documenting and testing its actual
cancellation and buffering contract: it cancels request context and returns a
timeout response, but cannot forcibly stop downstream side effects.

Scope:
- Clarify Go doc and module docs for timeout behavior.
- Add tests covering context cancellation and the buffered-response large-body path.
- Remove misleading comments that imply true streaming bypass.
- Keep public API unchanged.

Non-goals:
- Do not rename `Timeout`.
- Do not add goroutine-killing behavior.
- Do not introduce a new timeout abstraction.

Files:
- `middleware/timeout/timeout.go`
- `middleware/timeout/timeout_test.go`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/timeout`
- `go test -race -timeout 60s ./middleware/timeout`
- `go vet ./middleware/...`

Docs Sync:
- `docs/modules/middleware/README.md`

Done Definition:
- Timeout docs match implementation.
- Tests assert cancellation and large response behavior.
- Targeted tests pass.

Outcome:

