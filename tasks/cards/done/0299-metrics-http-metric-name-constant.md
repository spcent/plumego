# Card 0299: Metrics HTTP Name Constant Adoption

Priority: P2
State: done
Primary Module: metrics

## Goal

Use the canonical `MetricHTTPRequest` constant everywhere the HTTP request metric name is referenced.

## Problem

`metrics/collector.go` defines `MetricHTTPRequest = "http_request"`, but `ObserveHTTP` and tests still embed the string literal `"http_request"`. This is minor but creates inconsistent constants usage inside a stable root.

## Scope

- Replace literal `"http_request"` with `MetricHTTPRequest` in metrics implementation and tests.
- Ensure the tests assert on the constant rather than raw string.

## Non-Goals

- Do not rename the metric name value.
- Do not change labels or observation semantics.

## Expected Files

- `metrics/collector.go`
- `metrics/*_test.go`

## Validation

```bash
go test -timeout 20s ./metrics/...
go test -race -timeout 60s ./metrics/...
go vet ./metrics/...
```

## Done Definition

- No `"http_request"` string literal remains in metrics package code/tests.
- Metrics tests pass.

## Outcome

- Replaced HTTP metric name literals with `MetricHTTPRequest` in implementation and tests.
- Validation: `go test -timeout 20s ./metrics/...`, `go test -race -timeout 60s ./metrics/...`, `go vet ./metrics/...`.
