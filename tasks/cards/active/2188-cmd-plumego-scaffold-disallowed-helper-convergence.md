# Card 2188: cmd/plumego Scaffold Disallowed Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold_test.go`
Depends On: none

Goal:
Consolidate scaffold disallowed-file assertions behind a local helper.

Problem:
The microservice legacy HTTP helper test repeats `slices.Contains` checks with
inline fatal messages. A named helper makes the generated-file exclusion rule
easier to extend and read.

Scope:
- Add a local helper for asserting a template file list omits a path.
- Use it in the microservice legacy helper test.

Non-goals:
- Do not change scaffold output.
- Do not change canonical HTTP contract pattern checks.
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
- Microservice disallowed-file checks use one local helper.
- The listed validation commands pass.

Outcome:
