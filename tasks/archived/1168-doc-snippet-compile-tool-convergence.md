# 1168 - Doc Snippet Compile Tool Convergence

State: done
Priority: P2
Primary module: docs tooling

## Goal

Make documentation snippet compilation easier to maintain by replacing duplicated shell/awk parsing with one small standard-library Go tool.

## Scope

- Add a small internal tool that parses configured Markdown files for Go code fences.
- Compile complete `package main` snippets.
- Continue compiling `docs/getting-started.md` fragment snippets through an explicit wrapper.
- Keep `scripts/check-doc-snippets-compile.sh` as the stable entrypoint and make it delegate to the tool.
- Preserve current behavior and success output shape.

## Non-goals

- Do not broaden snippet coverage beyond the current configured docs.
- Do not introduce third-party dependencies.
- Do not change documentation examples except where required by the tool contract.

## Files

- `scripts/check-doc-snippets-compile.sh`
- `internal/tools/doc-snippets/main.go`

## Tests

- `go test -timeout 20s ./internal/tools/...`
- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`

## Docs Sync

No docs update required unless the tool needs explicit snippet markers in existing docs.

## Done Definition

- The shell script no longer contains duplicated Markdown parsing logic.
- The Go tool compiles the same documentation snippets currently covered by the script.
- Tool tests, snippet compilation, and core tests pass.

## Validation

- `go test -timeout 20s ./internal/tools/...`
- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
