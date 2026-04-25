# Card 2185: cmd/plumego Scaffold Parse Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Consolidate scaffold generated Go parseability checks behind a local helper.

Problem:
`scaffold_test.go` repeats parser setup and error formatting inside the
template parseability loop. A helper makes the validation intent clearer and
keeps error formatting consistent with other scaffold test helpers.

Scope:
- Add a local helper for parsing generated scaffold Go files.
- Use the helper in the parseability test.

Non-goals:
- Do not change scaffold output.
- Do not change package-name tests.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`

Tests:
- from `cmd/plumego`: `go test -race -timeout 60s ./internal/scaffold/...`
- from `cmd/plumego`: `go test -timeout 20s ./internal/scaffold/...`
- from `cmd/plumego`: `go vet ./internal/scaffold/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Scaffold parseability test uses one local parse helper.
- The listed validation commands pass.

Outcome:
