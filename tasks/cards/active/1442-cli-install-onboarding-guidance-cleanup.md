# Card 1442

Milestone: M-006
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- `docs/release/v1.0.0.md`
- CLI/onboarding docs discovered during execution
Depends On:
- 1441

Goal:
- Align CLI install and onboarding guidance with the nested `cmd/plumego` module
  fact recorded during v1 release evidence.

Scope:
- Search README, getting-started, and CLI docs for tagged `go install` claims.
- Update only statements that imply unsupported tagged CLI install behavior.
- Keep source-checkout CLI validation documented.

Non-goals:
- Do not restructure `cmd/plumego`.
- Do not change CLI commands or scaffold behavior.
- Do not change runtime APIs.

Files:
- `docs/release/v1.0.0.md`
- README/getting-started/CLI docs as discovered

Tests:
- `cd cmd/plumego && go test -timeout 20s ./...`
- `bash scripts/check-doc-snippets-compile.sh`
- `git diff --check`

Docs Sync:
- Required if install or onboarding text changes.

Done Definition:
- CLI install/onboarding docs no longer overclaim tagged install support.
- CLI tests and docs snippets pass.
