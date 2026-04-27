# Card 0479: cmd/plumego Scaffold Pattern Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Make scaffold template required/disallowed pattern assertions use shared local
helpers.

Problem:
`scaffold_test.go` still has inline `strings.Contains` loops for canonical HTTP
contract and response DTO checks, while the same file already uses helpers for
NoTODO, parseability, and omitted files.

Scope:
- Add local include/exclude pattern assertion helpers.
- Use them in canonical HTTP contract, route param, and local response DTO tests.

Non-goals:
- Do not change scaffold output or pattern lists.
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
- Scaffold template pattern assertions use shared helpers.
- The listed validation commands pass.

Outcome:
