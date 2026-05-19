# Card 1499

Milestone: M-020
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- `security/authn/*.go`
- `security/jwt/*.go`
- `store/cache/*.go`
- `store/kv/*.go`
- `health/*.go`
- `log/*.go`
- `metrics/*.go`
Depends On: 1498

## Goal

Add or strengthen package-level doc comments for the remaining five stable roots (`security`, `store`, `health`, `log`, `metrics`) so that all nine stable roots render clear, useful summaries on `pkg.go.dev`.

## Scope

- Package-level doc comments in `security/authn`, `security/jwt`, `security/password`, `security/headers`, `security/input`, `security/abuse`.
- Package-level doc comments in `store/cache`, `store/kv`, `store/db`, `store/file`, `store/idempotency`.
- Package-level doc comments in `health`, `log`, `metrics`.
- One `Example` function each for: `security/jwt` (verify), `store/cache` (get/set), `log` (NewLogger).

## Non-goals

- Do not change any exported function signatures, types, or behavior.
- Do not add `x/*` doc comments (extension packages are addressed separately).
- Do not add benchmark functions.
- Do not change `core`, `router`, `contract`, `middleware` (covered in Card 1498).

## Files

- `security/authn/doc.go` or `security/authn/authn.go`
- `security/jwt/doc.go` or `security/jwt/jwt.go`
- `security/jwt/example_test.go`
- `store/cache/doc.go` or `store/cache/cache.go`
- `store/cache/example_test.go`
- `health/doc.go` or `health/health.go`
- `log/doc.go` or `log/log.go`
- `log/example_test.go`
- `metrics/doc.go` or `metrics/metrics.go`

## Acceptance Tests

```
security/jwt/example_test.go: ExampleVerify
store/cache/example_test.go: ExampleMemoryCache_Get
log/example_test.go: ExampleNewLogger
```

## Tests

```bash
go test -run=Example ./security/...
go test -run=Example ./store/...
go test -run=Example ./health/...
go test -run=Example ./log/...
go test -run=Example ./metrics/...
go vet ./security/... ./store/... ./health/... ./log/... ./metrics/...
```

## Docs Sync

- No external docs targets; doc comments are the deliverable.

## Validation

```bash
go test -run=Example ./security/... ./store/... ./health/... ./log/... ./metrics/...
go vet ./security/... ./store/... ./health/... ./log/... ./metrics/...
gofmt -l ./security/ ./store/ ./health/ ./log/ ./metrics/
go run ./internal/checks/public-entrypoints-sync
go run ./internal/checks/module-manifests
```

## Done Definition

- [ ] All five stable root families (`security`, `store`, `health`, `log`, `metrics`) have non-empty package-level doc comments in all sub-packages.
- [ ] `ExampleVerify`, `ExampleMemoryCache_Get`, `ExampleNewLogger` compile and pass.
- [ ] All Acceptance Tests pass (`go test -run=Example`).
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l` produces no output for affected paths.
- [ ] Together with Card 1498: all nine stable roots are documented on `pkg.go.dev`.

## Outcome

<!-- Agent fills after completion -->
