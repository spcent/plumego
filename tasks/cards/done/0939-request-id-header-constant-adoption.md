# Card 0939: Request ID Header Constant Adoption

Priority: P3
State: done
Primary Module: contract

## Goal

Use `contract.RequestIDHeader` everywhere request-id headers are referenced in stable tests to avoid string drift.

## Problem

Several tests still use the string literal `"X-Request-ID"` instead of the canonical constant, which makes the stable test suite less resilient to header naming changes.

## Scope

- Replace `"X-Request-ID"` literals in stable tests with `contract.RequestIDHeader`.
- Keep test behavior unchanged.

## Non-Goals

- Do not change request-id propagation behavior.
- Do not change header naming values.

## Expected Files

- `core/options_test.go`
- `middleware/accesslog/accesslog_test.go`
- `middleware/requestid/request_id_test.go`

## Validation

```bash
go test -timeout 20s ./core ./middleware/... ./contract
```

## Done Definition

- No `"X-Request-ID"` literals remain in stable tests.
- Tests pass.

## Outcome

- Replaced request-id header literals in stable tests with `contract.RequestIDHeader`.
- Validation: `go test -timeout 20s ./core ./middleware/... ./contract`.
