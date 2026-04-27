# Card 0295: Contract Validation Code Normalization

Priority: P1
State: done
Primary Module: contract

## Goal

Use the canonical `contract` error codes for field-level validation errors so that validation output does not invent a second code vocabulary.

## Problem

`contract.ValidateStruct` emits `FieldError.Code` values like `"required"`, `"min"`, `"max"`, `"email"`, and `"unknown_rule"`. Meanwhile the stable module already defines canonical error codes (`CodeRequired`, `CodeInvalidFormat`, `CodeOutOfRange`, `CodeDuplicate`, etc.). This creates two incompatible code vocabularies inside the same stable root.

## Scope

- Enumerate all `FieldError.Code` values emitted by `contract` validation.
- Map each rule to canonical `contract` error codes:
  - `required` → `CodeRequired`
  - `email` → `CodeInvalidFormat`
  - `min` / `max` → `CodeOutOfRange`
  - `unknown_rule` / validator config errors → `CodeInvalidFormat`
  - `max_depth_exceeded` → `CodeOutOfRange` (or the closest canonical code; document the decision)
- Update tests asserting on field-level codes.
- Update docs if they list the old lower-case codes.

## Non-Goals

- Do not change error envelope shape or `BindErrorToAPIError` mapping.
- Do not add new error codes unless an existing canonical code is insufficient.
- Do not change validation rule semantics beyond code normalization.

## Expected Files

- `contract/validation.go`
- `contract/*_test.go`
- `docs/modules/contract/README.md` (if field codes are documented)

## Validation

```bash
go test -timeout 20s ./contract
go test -race -timeout 60s ./contract
go vet ./contract
```

## Done Definition

- All field-level validation errors use canonical `contract` code constants.
- No lower-case ad hoc validation codes remain in stable `contract`.
- Tests and docs match the new canonical codes.

## Outcome

- Normalized validation rule codes to canonical `contract` error codes and updated regression assertions.
- Validation: `go test -timeout 20s ./contract`, `go test -race -timeout 60s ./contract`, `go vet ./contract`.
