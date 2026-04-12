# Card 0944: Authn Header Constant Adoption

Priority: P3
State: done
Primary Module: security

## Goal

Define canonical constants for auth header names used by `security/authn` and its tests.

## Problem

`security/authn` uses `"Authorization"` and `"X-Token"` as literals in both code and tests. This is minor but inconsistent with the recent header-constant cleanup across stable middleware.

## Scope

- Introduce constants for `Authorization` and `X-Token` header names in `security/authn`.
- Use those constants in the authn token extractor and tests.

## Non-Goals

- Do not change header semantics or parsing logic.
- Do not add or remove supported auth schemes.

## Expected Files

- `security/authn/token.go`
- `security/authn/token_test.go`

## Validation

```bash
go test -timeout 20s ./security/authn
go test -race -timeout 60s ./security/authn
go vet ./security/authn
```

## Done Definition

- Authn header names are referenced through constants.
- Authn tests pass.
