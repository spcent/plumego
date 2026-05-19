# Card 0928

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/scaffold
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go, cmd/plumego/README.md
Depends On: 0734

Goal:
Remove advertised legacy scaffold runtime shapes that conflict with the canonical stable service style.

Scope:
- Make `fullstack` and `microservice` templates use the same canonical bootstrap and route style as other supported templates.
- Remove generated raw `http.ListenAndServe` and generated app-level `log.Fatal` paths from advertised templates.
- Update tests so all supported templates parse and preserve canonical handler/contract style.
- Update README template descriptions if file sets change.

Non-goals:
- Do not remove template names.
- Do not add frontend assets or container orchestration behavior.
- Do not modify reference/standard-service.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/scaffold ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Advertised templates no longer emit conflicting runtime styles.
- Scaffold tests cover stable canonical style for all supported templates.
- README no longer implies legacy runtime behavior.

Outcome:
- Mapped `minimal`, `fullstack`, and `microservice` template names to the canonical bootstrap file set.
- Removed generated `log.Fatal*` usage from canonical main output.
- Added tests for canonical file-set aliases and disallowed legacy runtime shapes.

Validation:
- `go test ./internal/scaffold ./commands`
- `go build .`
