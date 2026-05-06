# 0770 - Core Package Docs

State: active
Priority: P2
Primary module: core

## Goal

Improve pkg.go.dev readability for stable core without widening the API.

## Scope

- Add package-level Go documentation for `core`.
- Sharpen exported config type and field comments for lifecycle, TLS, HTTP/2, and router policy.
- Keep comments consistent with module docs.

## Non-goals

- Do not add examples that require broader docs restructuring.
- Do not add public symbols.
- Do not change runtime behavior.

## Files

- `core/doc.go`
- `core/config.go`
- `core/app.go`
- `tasks/cards/active/0770-core-package-docs.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

No markdown docs required unless Go comments reveal drift.

## Done Definition

- Package doc explains core ownership and lifecycle usage.
- Exported config comments are precise enough for stable package documentation.
- Core tests and vet pass.

