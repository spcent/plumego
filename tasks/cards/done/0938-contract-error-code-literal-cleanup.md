# Card 0938: Contract Error Code Literal Cleanup

Priority: P3
State: done
Primary Module: contract

## Goal

Use `contract` error code constants in tests instead of raw string literals so that the canonical vocabulary is referenced consistently.

## Problem

Some tests still assert on string literals like `"VALIDATION_ERROR"` or `"http_request"` rather than using the exported constants. This makes the tests brittle and less aligned with the canonical code registry.

## Scope

- Replace string literal error codes in `contract` tests with their constant equivalents.
- Keep test intent unchanged.

## Non-Goals

- Do not change error code values.
- Do not modify error envelope shape or response semantics.

## Expected Files

- `contract/*_test.go`

## Validation

```bash
go test -timeout 20s ./contract
go test -race -timeout 60s ./contract
go vet ./contract
```

## Done Definition

- No raw error-code string literals remain in contract tests.
- Contract tests pass.

## Outcome

- Replaced error-code string literals in contract tests with canonical constants.
- Validation: `go test -timeout 20s ./contract`, `go test -race -timeout 60s ./contract`, `go vet ./contract`.
