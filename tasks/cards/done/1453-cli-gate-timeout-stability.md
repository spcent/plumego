# Card 1453

Milestone: M-003
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/commands`
- `cmd/plumego/internal`
- `Makefile`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`
Depends On: tasks/cards/done/1452-post-v1-maturity-roadmap.md

Goal:
- Remove the mismatch between real CLI test runtime and the release gate timeout so `cmd/plumego` does not fail near the `20s` boundary.

Scope:
- Identify whether `cmd/plumego/commands` has slow test cleanup or whether the gate timeout is too tight for the accepted generated-project smoke layer.
- Keep the CLI behavior unchanged.
- Align `Makefile` and release checklist timeout only if the test layer is intentionally slow.

Non-goals:
- Do not change CLI command semantics.
- Do not weaken root module gates.
- Do not change extension maturity evidence.

Files:
- `cmd/plumego/commands/*_test.go`
- `cmd/plumego/internal/**`
- `Makefile`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`

Tests:
- `cd cmd/plumego && go test -timeout 60s ./...`
- `cd cmd/plumego && go test -race -timeout 60s ./...`
- `go test -timeout 20s ./...`

Docs Sync:
- Required only if gate timeout policy changes.

Done Definition:
- CLI tests pass consistently under the documented release gate timeout, or the documented timeout is updated to match the intentional slow smoke layer and passes locally.

Outcome:
- Full CLI confidence keeps the generated-project smoke layer and now uses a
  `60s` timeout in the release gate, matching the existing CLI race timeout and
  avoiding cold-cache failures near the old `20s` boundary.
