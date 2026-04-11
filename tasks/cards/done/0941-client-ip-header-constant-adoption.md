# Card 0941: Client IP Header Constant Adoption

Priority: P3
State: done
Primary Module: middleware

## Goal

Define canonical constants for client IP header names and use them consistently across transport helpers and tests.

## Problem

`middleware/internal/transport.ClientIP` and ratelimit tests embed `"X-Forwarded-For"` and `"X-Real-IP"` literals directly. This is a minor but repeated literal usage inside stable middleware.

## Scope

- Add header name constants in `middleware/internal/transport`.
- Use the constants in `ClientIP` and ratelimit tests.
- Keep precedence and parsing behavior unchanged.

## Non-Goals

- Do not change client IP resolution logic.
- Do not add new headers or change casing.

## Expected Files

- `middleware/internal/transport/http.go`
- `middleware/ratelimit/abuse_guard_test.go`

## Validation

```bash
go test -timeout 20s ./middleware/ratelimit ./middleware/internal/transport
go test -race -timeout 60s ./middleware/ratelimit
go vet ./middleware/ratelimit ./middleware/internal/transport
```

## Done Definition

- Client IP header literals are replaced by constants.
- Tests pass.

## Outcome

- Added client IP header constants in transport helper and updated ratelimit tests.
- Validation: `go test -timeout 20s ./middleware/ratelimit ./middleware/internal/transport`, `go test -race -timeout 60s ./middleware/ratelimit`, `go vet ./middleware/ratelimit ./middleware/internal/transport`.
