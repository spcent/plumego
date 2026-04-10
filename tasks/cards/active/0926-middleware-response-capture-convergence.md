# Card 0926: Middleware Response Capture Convergence

Priority: P1
State: active
Primary Module: middleware

## Goal

Converge stable middleware response capture/buffering behavior so timeout, compression, and coalescing do not maintain subtly different response-writer semantics.

## Problem

Several stable middleware packages capture or buffer downstream responses independently:

- `middleware/timeout` has `timeoutResponseWriter` with custom header/body/status handling and bypass behavior.
- `middleware/compression` has `gzipResponseWriter` with separate header/write/buffer logic.
- `middleware/coalesce` uses `middleware/internal/transport.ResponseRecorder` for captured responses.
- `middleware/internal/transport` already owns shared response-recorder and safe-write helpers.

This creates repeated behavior around:

- header cloning/copying
- status defaulting
- buffering limits
- write-after-timeout or write-after-overflow behavior
- `nosniff` header enforcement
- large/streaming response fallback semantics

The stable middleware layer should have one small internal capture primitive and each middleware should adapt policy around it, not re-implement writer semantics.

## Scope

- Audit `timeout`, `compression`, and `coalesce` response capture behavior.
- Keep shared mechanical response capture in `middleware/internal/transport`.
- Remove duplicated header/status/body capture logic where possible.
- Preserve each middleware's distinct policy: timeout deadline, gzip negotiation, request coalescing.
- Add regression tests for headers, status codes, body writes, flush behavior, overflow/bypass, and canonical error shapes.
- Keep public middleware constructors unchanged unless a duplicate constructor is proven unnecessary.

## Non-Goals

- Do not change middleware execution order semantics.
- Do not add business response helpers to middleware.
- Do not move observability exporter behavior into stable middleware.
- Do not add external dependencies.

## Expected Files

- `middleware/internal/transport/http.go`
- `middleware/timeout/timeout.go`
- `middleware/compression/gzip.go`
- `middleware/coalesce/coalesce.go`
- `middleware/*/*_test.go`
- `middleware/conformance/*`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./middleware/...
go test -race -timeout 60s ./middleware/...
go vet ./middleware/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- Shared response capture mechanics live in one internal middleware primitive.
- Timeout, compression, and coalescing tests prove preserved behavior.
- No middleware emits non-canonical JSON error shapes.
- Middleware conformance tests pass.
- Focused gates and repo-wide gates pass.
