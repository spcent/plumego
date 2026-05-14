# Card 1396

Milestone: M-004
Recipe: specs/change-recipes/add-middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/response_buffer.go
- middleware/debug/debug_errors.go
- middleware/coalesce/coalesce.go
- middleware/conformance/response_writer_contract_test.go
Depends On:
- 1394

Goal:
Converge stable middleware response helper usage without changing middleware behavior.

Scope:
- Audit `SafeWrite`, `EnsureNoSniff`, header copy/replace, flush forwarding, hijack forwarding, and buffered-response usage.
- Replace local duplicate helper logic with `middleware/internal/transport` helpers where behavior is identical.
- Preserve specialized wrappers for gzip, timeout, coalesce, recovery, body limit, and debug error capture when they carry distinct state or replay semantics.
- Strengthen conformance coverage for `Unwrap`, `Flush`, `Hijack`, panic finalization, and stale-header replacement.

Non-goals:
- Do not create a new public middleware helper package.
- Do not change middleware constructor names.
- Do not change streaming, timeout, gzip, or coalescing semantics.

Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/response_buffer.go
- middleware/debug/debug_errors.go
- middleware/coalesce/coalesce.go
- middleware/conformance/response_writer_contract_test.go

Tests:
- go test -race -timeout 60s ./middleware/...
- go test -timeout 20s ./middleware/...
- go vet ./middleware/...

Docs Sync:
- Update `docs/modules/middleware/README.md` only if helper ownership guidance changes.

Done Definition:
- Duplicate transport helper logic is reduced where behavior matches exactly.
- Middleware conformance tests cover the affected response-writer contracts.
- Target checks pass.

Outcome:
