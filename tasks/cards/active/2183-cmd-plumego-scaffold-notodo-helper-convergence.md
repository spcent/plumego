# Card 2183: cmd/plumego Scaffold NoTODO Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Consolidate repeated scaffold `// TODO` assertions behind one local helper.

Problem:
`scaffold_test.go` repeats the same bare `// TODO` string check for template
content and default file content. The duplication is small but makes future
template checks easier to drift.

Scope:
- Add a local helper for asserting generated scaffold content has no bare
  `// TODO`.
- Replace repeated inline checks with the helper.

Non-goals:
- Do not change scaffold output.
- Do not change parseability or package-name tests.
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
- Scaffold NoTODO tests use one local helper.
- The listed validation commands pass.

Outcome:
