# Card 2190: cmd/plumego Codegen Subtest Naming Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/codegen/codegen_test.go`
Depends On: none

Goal:
Use named subtests for codegen table-driven checks that currently rely on
inline error labels.

Problem:
Several codegen tests loop over HTTP methods or method sets without `t.Run`,
which makes failures less targeted than the rest of the test suite.

Scope:
- Add `t.Run` around service-method, handler-call, and handler-test injection
  table cases.

Non-goals:
- Do not change generated code or assertions.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/codegen/codegen_test.go`

Tests:
- from `cmd/plumego`: `go test -race -timeout 60s ./internal/codegen/...`
- from `cmd/plumego`: `go test -timeout 20s ./internal/codegen/...`
- from `cmd/plumego`: `go vet ./internal/codegen/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Targeted codegen table loops use named subtests.
- The listed validation commands pass.

Outcome:
