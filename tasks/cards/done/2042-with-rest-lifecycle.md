# Card 2042

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: reference/with-rest
Owned Files:
- reference/with-rest/main.go
- reference/with-rest/internal/app/app.go
- reference/with-rest/README.md
- reference/with-rest/ARCHITECTURE.md
- reference/with-rest/AGENT_TASKS.md
Depends On: 2041

## Goal

Align `reference/with-rest` with the canonical startup lifecycle.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, and update current local docs that describe the startup shape.

## Non-goals

- Do not change REST resource behavior.
- Do not change `x/rest`.
- Do not change route contracts or response shapes.

## Files

- reference/with-rest/main.go
- reference/with-rest/internal/app/app.go
- reference/with-rest/README.md
- reference/with-rest/ARCHITECTURE.md
- reference/with-rest/AGENT_TASKS.md

## Acceptance Tests

- Documentation grep has no current `App.Start()` lifecycle examples for `reference/with-rest`.

## Tests

- Existing package tests.

## Docs Sync

- reference/with-rest/README.md
- reference/with-rest/ARCHITECTURE.md
- reference/with-rest/AGENT_TASKS.md

## Validation

- cd reference/with-rest && go test -timeout 20s ./...
- cd reference/with-rest && go vet ./...
- gofmt -l reference/with-rest

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l reference/with-rest produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

- Moved process signal ownership to `main.run` and changed `App.Start(ctx)` to use caller-owned cancellation.
- Updated local scenario docs to describe canonical lifecycle ownership.
- Validation:
  - `cd reference/with-rest && go test -timeout 20s ./...`
  - `cd reference/with-rest && go vet ./...`
  - `gofmt -l reference/with-rest`
  - `rg -n "func \\(a \\*App\\) Start\\(\\)|a\\.Start\\(\\)|signal.NotifyContext" reference/with-rest`
