# 0752 - doc Snippet Script Path Hardening

State: done
Priority: P2
Primary Module: docs

## Goal

Make the documentation snippet compile script robust when checkout or temporary
paths contain spaces.

## Scope

- Quote the local replacement path written into the temporary `go.mod`.
- Avoid unquoted shell paths inside awk-generated commands.
- Keep the script portable for the repository's current shell/tooling baseline.

## Non-goals

- Do not add external parser dependencies.
- Do not replace the script with a Go utility.
- Do not widen snippet coverage beyond the current documented scope.

## Files

- `scripts/check-doc-snippets-compile.sh`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Not required.

## Done Definition

- The script quotes module replacement paths and generated directory creation
  paths safely.
- The snippet check, core tests, and vet pass.

## Outcome

- Quoted the local module replacement path written to the temporary `go.mod`.
- Added shell quoting for awk-created snippet directories.
- Verified with `bash scripts/check-doc-snippets-compile.sh`,
  `go test -timeout 20s ./core/...`, and `go vet ./core/...`.
