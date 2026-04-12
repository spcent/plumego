# Card 0959: Reference And Diagnostic Handler Response Convergence

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: reference
Depends On: —

## Goal

Bring the reference applications and small diagnostic/export handlers back to
the repository's canonical single-response-path guidance, so examples and
operator-facing endpoints stop teaching mixed `contract`/`http.Error` patterns.

## Problem

- `reference/standard-service/internal/handler/api.go` and
  `reference/standard-service/internal/handler/health.go` call
  `contract.WriteResponse` and then fall back to `http.Error("encoding error")`,
  which creates a second success/error write path inside the canonical example
  app.
- The same fallback pattern appears in demo route handlers under
  `reference/with-gateway`, `reference/with-messaging`, `reference/with-webhook`,
  and `reference/with-websocket`, so example code teaches two incompatible
  response styles.
- `x/ops/healthhttp/history.go` and `x/ai/metrics/prometheus.go` still emit
  plain-text `http.Error` responses on otherwise structured or documented
  endpoints.
- Because these files are documentation-by-example surfaces, inconsistencies
  here have outsized impact on future code shape.

## Scope

- Choose one explicit write-failure policy for reference handlers and apply it
  consistently across the reference app and demo handlers.
- Preserve successful CSV and Prometheus text responses where those content
  types are intentional; only failure-path behavior needs convergence.
- Remove `"encoding error"` fallback teaching from the canonical reference app.
- Add or update focused tests where reference or diagnostic handler behavior is
  asserted.

## Non-Goals

- Do not redesign reference app bootstrap or route layout.
- Do not change successful CSV or Prometheus payload formats.
- Do not widen `contract` just to support example-only behavior.

## Expected Files

- `reference/standard-service/internal/handler/api.go`
- `reference/standard-service/internal/handler/health.go`
- `reference/with-*/internal/app/routes.go`
- `reference/with-messaging/internal/handler/messaging.go`
- `x/ops/healthhttp/history.go`
- `x/ai/metrics/prometheus.go`

## Validation

```bash
rg -n 'http\.Error\(w, "encoding error"|http\.Error\(w, "Failed to generate metrics"|failed to write CSV' reference x/ops/healthhttp x/ai/metrics -g '!**/*_test.go'
go test -timeout 20s ./reference/... ./x/ops/... ./x/ai/...
go vet ./reference/... ./x/ops/... ./x/ai/...
```

## Docs Sync

- reference READMEs only if the example output or handler guidance changes

## Done Definition

- Reference handlers no longer teach `contract.WriteResponse` followed by
  `http.Error("encoding error")`.
- Diagnostic/export handlers keep their intentional success content types but no
  longer mix in ad hoc plain-text failure handling where a repo-standard path is
  expected.
- The canonical reference app demonstrates one response path per handler layer.

## Outcome

- Removed `encoding error` fallback teaching from the reference application and demo handlers so examples now show a single response path.
- Kept CSV and Prometheus success payloads as intentional text outputs, but rewrote them to buffer before writing instead of mixing in ad hoc `http.Error` fallbacks.
- Added focused assertions for invalid health-history export format and AI metrics handler response behavior.

## Validation Run

```bash
rg -n 'http\.Error\(w, "encoding error"|http\.Error\(w, "Failed to generate metrics"|failed to write CSV' reference x/ops/healthhttp x/ai/metrics -g '!**/*_test.go'
go test -timeout 20s ./reference/... ./x/ops/... ./x/ai/...
go vet ./reference/... ./x/ops/... ./x/ai/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
