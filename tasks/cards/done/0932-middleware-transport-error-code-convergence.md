# Card 0932: Middleware Transport Error Code Convergence

Priority: P1
State: done
Primary Module: middleware

## Goal

Use one canonical set of error codes for transport-layer middleware failures. Prefer the existing `contract` error codes where they already describe the same failure class, and avoid duplicating them in `middleware`.

## Problem

Stable `contract` already defines error codes such as `CodeRateLimited`, `CodeUnauthorized`, `CodeInternalError`, and request-bind error codes. Meanwhile `middleware/error_registry.go` defines overlapping lower-case codes (`rate_limited`, `auth_unauthenticated`, `internal_error`, `request_body_too_large`, `request_timeout`), and middleware adapters use the `middleware` versions in error envelopes. This results in two incompatible error-code vocabularies for the same failures in stable code paths and violates the “one canonical error construction path per layer” rule.

## Scope

- Enumerate all usages of `middleware.Code*` constants that overlap `contract.Code*`.
- Decide a single canonical vocabulary: prefer `contract.Code*` for common transport errors (auth, rate-limit, timeout, internal error, request bind failures).
- Remove or replace duplicate `middleware.Code*` constants that map to `contract` equivalents.
- Update middleware adapters (auth, ratelimit, timeout, recovery, coalesce, conformance helpers) to use `contract.Code*` or other canonical codes.
- Keep middleware-specific codes that have no `contract` equivalent (tenant/queue/server busy) but validate naming is consistent.
- Update tests asserting on error codes (`middleware/conformance`, `middleware/*_test.go`).
- Update module docs if they list the legacy lower-case codes.

## Non-Goals

- Do not change the error envelope shape (`contract.WriteError` output).
- Do not introduce new error codes unless there is a clear gap.
- Do not change HTTP status semantics for existing middleware failures.

## Expected Files

- `middleware/error_registry.go`
- `middleware/*/*.go`
- `middleware/*/*_test.go`
- `contract/error_codes.go` (read-only reference)
- `docs/modules/middleware/README.md` (if code lists exist)

## Validation

```bash
go test -timeout 20s ./middleware/... ./contract
go test -race -timeout 60s ./middleware
go vet ./middleware ./contract
```

## Done Definition

- No overlapping middleware error codes remain where `contract` already defines the canonical code.
- All middleware transport errors emit the same code vocabulary.
- All conformance and middleware tests pass.
- If any exported symbols are removed, full symbol-change protocol is followed.

## Outcome

- Removed overlapping middleware error-code constants and standardized middleware error envelopes on `contract` codes.
- Updated middleware and x/gateway adapters plus tests to expect the canonical `contract` error codes.
- Validation: `go test -timeout 20s ./middleware/... ./contract`, `go test -race -timeout 60s ./middleware`, `go vet ./middleware ./contract`.
