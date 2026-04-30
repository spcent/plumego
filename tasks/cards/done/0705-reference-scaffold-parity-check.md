# Card 0705

Milestone: M-001
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, reference/standard-service/README.md, docs/getting-started.md, tasks/cards/done/0705-reference-scaffold-parity-check.md
Depends On: 0704-template-and-release-doc-truth

Goal:
Verify and document the relationship between scaffold output and `reference/standard-service` so the official example path and CLI-generated starter path do not drift.

Scope:
Compare scaffold output shape against the canonical reference app and record whether the relationship is structural parity, conceptual parity, or intentionally different.

Non-goals:
Do not refactor scaffold generation.
Do not change `reference/standard-service` behavior.
Do not introduce new generated templates.

Files:
cmd/plumego/internal/scaffold/scaffold.go
reference/standard-service/README.md
docs/getting-started.md
tasks/cards/done/0705-reference-scaffold-parity-check.md

Tests:
go run ./internal/checks/reference-layout
cd cmd/plumego && go test -timeout 20s ./...

Docs Sync:
Update docs only if the current relationship is undocumented or misleading.

Done Definition:
The scaffold/reference relationship is recorded in Outcome or user-facing docs.
`reference-layout` passes.
M-001 has a clear final verification path.

Outcome:
Completed.

Decision:

- The canonical scaffold has conceptual and runtime-structure parity with `reference/standard-service`.
- It is not intended to be a byte-for-byte copy: generated projects use `cmd/app/main.go`, include project-local files, and do not copy reference-only tests.

Changes:

- Documented the scaffold/reference relationship in `reference/standard-service/README.md`.
- Added the same generated-project caveat to `docs/getting-started.md`.

Validation:

- `go run ./internal/checks/reference-layout` passed.
- `cd cmd/plumego && go test -timeout 20s ./...` passed.

