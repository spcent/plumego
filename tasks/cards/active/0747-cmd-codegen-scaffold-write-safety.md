# Card 0747

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego codegen and scaffold
Owned Files: cmd/plumego/internal/codegen/codegen.go, cmd/plumego/internal/codegen/codegen_test.go, cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go
Depends On: 0746

Goal:
Harden code generation and scaffold writes before stable release.

Scope:
- Validate every generated output path before writing, including optional test files.
- Prevent test file overwrite unless `Force` is explicitly set.
- Make unknown scaffold templates fail closed at the internal API boundary.
- Surface requested `git init` failures instead of silently reporting success.

Non-goals:
- Do not split scaffold templates into separate files in this card.
- Do not change generated project layout.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/codegen/codegen_test.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`

Tests:
- `go test ./internal/codegen ./internal/scaffold`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless CLI-visible scaffold behavior changes.

Done Definition:
- Codegen never overwrites any generated file without `Force`.
- Scaffold rejects unknown templates internally.
- `plumego new --git` reports git initialization failures.

Outcome:
