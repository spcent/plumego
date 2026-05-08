# Card 1201

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego scaffold
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go, cmd/plumego/README.md
Depends On: 0757

Goal:
Remove unreachable legacy template paths and make generated app startup style clearer.

Scope:
- Remove unreachable legacy default template content that is no longer emitted by stable templates.
- Convert generated stable app entrypoints to `run() error` plus a single `main` exit/log path where practical.
- Keep current stable template file sets unchanged.
- Add tests proving legacy `internal/httpapp` shapes are not reachable.

Non-goals:
- Do not add new templates.
- Do not change scenario profile route behavior.
- Do not change generated dependency set.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/scaffold`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md` only if generated startup guidance changes.

Done Definition:
- Unreachable legacy scaffold content is removed or made impossible to invoke.
- Stable generated app entrypoint style is deterministic and tested.

Outcome:
- Removed the legacy scaffold fallback content that could synthesize
  `internal/httpapp`, domain-user, frontend, Docker, or ad hoc package files
  outside the declared stable template file sets.
- `cmd/app/main.go` content now always uses the canonical `main` plus
  `run() error` entrypoint shape.
- Added regression coverage proving legacy fallback paths do not emit content.

Validation:
- `go test ./internal/scaffold`
- `go test ./...`
- `go vet ./...`
