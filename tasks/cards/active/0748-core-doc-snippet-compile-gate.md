# 0748 - core Doc Snippet Compile Gate

State: active
Priority: P2
Primary Module: core

## Goal

Add the missing documentation snippet compile gate for package-main core
examples.

## Scope

- Add `scripts/check-doc-snippets-compile.sh`.
- Compile package-main Go snippets from core-facing documentation through a
  temporary module that replaces `github.com/spcent/plumego` with the local
  checkout.
- Keep fragment snippets out of scope unless they are complete `package main`
  examples.

## Non-goals

- Do not rewrite all documentation examples.
- Do not add external dependencies.
- Do not make every small Go fragment independently compilable.

## Files

- `scripts/check-doc-snippets-compile.sh`
- `tasks/cards/done/0733-core-shutdown-doc-examples.md`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Not required unless snippets need small compile fixes.

## Done Definition

- The snippet compile script exists and passes.
- The previous card note about the missing script is updated with current
  status.
- Core tests and vet pass.
