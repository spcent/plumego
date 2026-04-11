# Card 0942: NoSniff Header Constant Adoption

Priority: P3
State: done
Primary Module: middleware

## Goal

Use constants for the `X-Content-Type-Options` header name and its `nosniff` value in transport helpers.

## Problem

`middleware/internal/transport.EnsureNoSniff` embeds the header name and value as string literals, which is inconsistent with other header constants added nearby.

## Scope

- Introduce constants for the header name and value in `middleware/internal/transport`.
- Update `EnsureNoSniff` to use the constants.

## Non-Goals

- Do not change behavior or header values.

## Expected Files

- `middleware/internal/transport/http.go`

## Validation

```bash
go test -timeout 20s ./middleware/internal/transport
go vet ./middleware/internal/transport
```

## Done Definition

- NoSniff header name/value are referenced through constants.

## Outcome

- Introduced NoSniff header constants and updated transport helper to use them.
- Validation: `go test -timeout 20s ./middleware/internal/transport`, `go vet ./middleware/internal/transport`.
