# Card 2191: cmd/plumego Scaffold Template Subtest Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Use named subtests for scaffold template-level loops.

Problem:
Several scaffold tests iterate across templates without `t.Run`, so failures
are less localized than the helper-based assertions now used in the file.

Scope:
- Add template-name subtests to basic scaffold template loops.
- Preserve existing assertions and generated output.

Non-goals:
- Do not change scaffold behavior or templates.
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
- Template-level scaffold loops use named subtests where useful.
- The listed validation commands pass.

Outcome:
