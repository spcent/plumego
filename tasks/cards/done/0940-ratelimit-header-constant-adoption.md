# Card 0940: Ratelimit Header Constant Adoption

Priority: P3
State: done
Primary Module: middleware

## Goal

Define canonical constants for rate-limit headers and use them consistently across middleware and tests.

## Problem

`middleware/ratelimit` embeds header names like `"X-RateLimit-Limit"` and `"Retry-After"` as string literals in both implementation and tests. This duplicates literals and makes future changes error-prone.

## Scope

- Introduce header constants in `middleware/ratelimit` (or the nearest stable helper).
- Replace header string literals in `abuse_guard.go` and tests with the constants.
- Keep header values and semantics unchanged.

## Non-Goals

- Do not change rate-limit policy behavior.
- Do not change header values or casing.

## Expected Files

- `middleware/ratelimit/abuse_guard.go`
- `middleware/ratelimit/abuse_guard_test.go`

## Validation

```bash
go test -timeout 20s ./middleware/ratelimit
go test -race -timeout 60s ./middleware/ratelimit
go vet ./middleware/ratelimit
```

## Done Definition

- No rate-limit header literals remain in ratelimit middleware/tests.
- Tests pass.

## Outcome

- Added rate-limit header constants and replaced literals in ratelimit middleware/tests.
- Validation: `go test -timeout 20s ./middleware/ratelimit ./middleware/internal/transport`, `go test -race -timeout 60s ./middleware/ratelimit`, `go vet ./middleware/ratelimit ./middleware/internal/transport`.
