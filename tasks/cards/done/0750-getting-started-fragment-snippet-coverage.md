# 0750 - getting-started Fragment Snippet Coverage

State: done
Priority: P1
Primary Module: docs

## Goal

Compile the core-facing routing fragments in `docs/getting-started.md` so small
examples cannot drift from the stable route registration API.

## Scope

- Extend `scripts/check-doc-snippets-compile.sh` to compile selected
  non-`package main` Go fragments from `docs/getting-started.md`.
- Keep wrappers minimal and local to the script.
- Make the getting-started routing fragments show explicit route registration
  error handling.

## Non-goals

- Do not require every prose fragment in all docs to compile.
- Do not turn the getting-started page into another full application example.
- Do not change route registration APIs.

## Files

- `scripts/check-doc-snippets-compile.sh`
- `docs/getting-started.md`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Required for `docs/getting-started.md` because the fragments should document the
current explicit error-handling style.

## Done Definition

- The getting-started routing fragments compile under the snippet gate.
- The fragments show explicit handling of route registration errors.
- Core tests and vet pass.

## Outcome

- Extended `scripts/check-doc-snippets-compile.sh` to wrap and compile
  non-`package main` Go fragments from `docs/getting-started.md`.
- Updated getting-started routing fragments to handle route registration errors
  explicitly.
- Verified with `bash scripts/check-doc-snippets-compile.sh`,
  `go test -timeout 20s ./core/...`, and `go vet ./core/...`.
