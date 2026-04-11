# Card 0943: Contract JSON Content-Type Constants

Priority: P3
State: active
Primary Module: contract

## Goal

Define canonical constants for the JSON response content type and header name used by `contract` response helpers.

## Problem

`contract.WriteJSON` and `contract.WriteError` embed `"Content-Type"` and `"application/json"` literals directly. This is minor but inconsistent with the recent move toward header constants in stable helpers.

## Scope

- Introduce constants in `contract` for the content type header name and JSON value.
- Update `WriteJSON`, `WriteError`, and relevant contract tests to use the constants.

## Non-Goals

- Do not change response shapes or content.
- Do not modify behavior in other modules that set content types.

## Expected Files

- `contract/response.go`
- `contract/errors.go`
- `contract/*_test.go`

## Validation

```bash
go test -timeout 20s ./contract
go test -race -timeout 60s ./contract
go vet ./contract
```

## Done Definition

- Contract JSON content type header/value are referenced through constants.
- Contract tests pass.
