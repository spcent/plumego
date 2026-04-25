# Card 2187: cmd/plumego Scaffold Package Rule Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Make scaffold package-name validation use an explicit helper for generated cmd
paths.

Problem:
The scaffold package-name test keeps the rule for cmd paths inline in the loop.
That rule is part of the generated layout contract and is clearer as a small
named helper.

Scope:
- Add a local helper that identifies generated Go files that must be package
  `main`.
- Use it in `TestTemplateContent_CorrectPackageNames`.

Non-goals:
- Do not change scaffold output.
- Do not change parseability or NoTODO tests.
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
- The package-name test uses a named helper for cmd/main package detection.
- The listed validation commands pass.

Outcome:
