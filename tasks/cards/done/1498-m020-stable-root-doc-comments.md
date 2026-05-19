# Card 1498

Milestone: M-020
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/*.go`
- `router/*.go`
- `contract/*.go`
- `middleware/chain.go`
Depends On: 1497

## Goal

Add or strengthen package-level doc comments and `Example` functions for `core`, `router`, `contract`, and `middleware/chain` so that every package renders a clear, useful summary on `pkg.go.dev` and `godoc`.

## Scope

- Package-level doc comments in `core`, `router`, `contract`, `middleware` (chain and at least 3 sub-packages).
- `Example` functions in `core/example_test.go`, `router/example_test.go`, `contract/example_test.go`.
- Examples must compile and pass `go test -run=Example`.

## Non-goals

- Do not change any exported function signatures, types, or behavior.
- Do not add doc comments to `security`, `store`, `health`, `log`, `metrics` (those are Card 1499).
- Do not add `x/*` doc comments in this card.
- Do not add benchmark functions.

## Files

- `core/doc.go` (create if needed) or `core/core.go`
- `core/example_test.go`
- `router/doc.go` (create if needed) or `router/router.go`
- `router/example_test.go`
- `contract/doc.go` (create if needed) or `contract/errors.go`
- `contract/example_test.go`
- `middleware/chain.go` (package doc comment)

## Acceptance Tests

```
core/example_test.go: ExampleNew
router/example_test.go: ExampleRouter_Group
contract/example_test.go: ExampleWriteError
contract/example_test.go: ExampleWriteResponse
```

## Tests

```bash
go test -run=Example ./core/...
go test -run=Example ./router/...
go test -run=Example ./contract/...
go vet ./core/... ./router/... ./contract/... ./middleware/...
```

## Docs Sync

- No external docs targets; doc comments are the deliverable.

## Validation

```bash
go test -run=Example ./core/... ./router/... ./contract/... ./middleware/...
go vet ./core/... ./router/... ./contract/... ./middleware/...
gofmt -l ./core/ ./router/ ./contract/ ./middleware/
go run ./internal/checks/public-entrypoints-sync
```

## Done Definition

- [ ] `core`, `router`, `contract`, `middleware` each have a non-empty package-level doc comment (first sentence is a complete summary of what the package provides).
- [ ] `ExampleNew` in `core/example_test.go` shows `core.New` → `app.Use` → `app.Get` → `app.Prepare`.
- [ ] `ExampleRouter_Group` in `router/example_test.go` shows a route group with prefix and middleware.
- [ ] `ExampleWriteError` and `ExampleWriteResponse` in `contract/example_test.go` show handler usage.
- [ ] All Acceptance Tests pass (`go test -run=Example`).
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l` produces no output for affected paths.

## Outcome

Completed in commit `f1bae9a3`.

Added package-level godoc and examples for `core`, `router`, `contract`, and
selected `middleware` subpackages.

Validation run:

- `go test -run=Example ./core/...`
- `go test -run=Example ./router/...`
- `go test -run=Example ./contract/...`
- `go test -run=Example ./middleware/...`
- `go vet ./core/... ./router/... ./contract/... ./middleware/...`
- `gofmt -l ./core/ ./router/ ./contract/ ./middleware/`
- `go run ./internal/checks/public-entrypoints-sync`
