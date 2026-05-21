# Card 2041

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Context Package: implementation
Priority: P2
State: done
Primary Module: docs
Owned Files:
- docs/why-plumego.md
- docs/migration/from-gin.md
- docs/migration/from-echo.md
- docs/migration/from-chi.md
- cmd/plumego/README.md
Depends On: 2040

## Goal

Remove obvious current-doc drift around canonical lifecycle and reference layout without rewriting historical release records.

## Scope

Review current docs/examples for stale `Start()` lifecycle, old handler/domain layout, or stale canonical scaffold descriptions, and patch only current guidance.

## Non-goals

- Do not edit historical release notes to match current command output.
- Do not edit website files that already have unrelated local changes.
- Do not change code behavior.

## Files

- docs/why-plumego.md
- docs/migration/from-gin.md
- docs/migration/from-echo.md
- docs/migration/from-chi.md
- cmd/plumego/README.md

## Acceptance Tests

## Tests

- Documentation grep for stale canonical `Start()` examples in current guidance.

## Docs Sync

- Same as Files.

## Validation

- go run ./internal/tools/doc-snippets
- rg -n "app\\.Start\\(\\)|func \\(a \\*App\\) Start\\(\\)" docs cmd/plumego/README.md
- gofmt -l .

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

- Updated current guidance in why, migration, and CLI docs to reference `main.run -> app.Start(ctx)` and `internal/domain/<name>` layout.
- Left historical release output unchanged.
- Validation:
  - `go run ./internal/tools/doc-snippets`
  - `rg -n "app\\.Start\\(\\)|func \\(a \\*App\\) Start\\(\\)" docs cmd/plumego/README.md` returned no matches.
  - `gofmt -l .`
