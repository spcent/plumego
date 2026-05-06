# 0766 - Core Open Connection Terminology

State: active
Priority: P1
Primary module: core docs

## Goal

Align public docs and source comments with the actual open connection count semantics.

## Scope

- Replace stale `active connection` wording where the tracker counts open HTTP connections.
- Keep runtime behavior and log field names unchanged.
- Keep wording consistent between module docs and top-level docs.

## Non-goals

- Do not change `ConnState` counting behavior.
- Do not add request-level inflight tracking.
- Do not change public APIs.

## Files

- `core/app.go`
- `core/config.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update English and Chinese docs only where wording currently describes open connection tracking incorrectly.

## Done Definition

- No stale active-connection wording remains for core connection tracking.
- Core tests and doc snippet compile checks pass.

