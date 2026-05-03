# Card 0119

Milestone:
Recipe: specs/change-recipes/stable-api.yaml
Priority: P0
State: active
Primary Module: middleware
Owned Files:
  - internal/httputil/http_response.go
  - middleware/bodylimit/body_limit.go
  - middleware/recovery/recover.go
  - middleware/compression/gzip.go
  - middleware/internal/transport/transport_test.go
Depends On: 0118

Goal:
Preserve modern `http.ResponseWriter` optional-interface behavior through
middleware wrappers by exposing `Unwrap()` consistently where wrappers hold an
underlying writer.

Scope:
- Add `Unwrap() http.ResponseWriter` to response wrappers that embed or hold an underlying writer.
- Add focused tests proving `http.NewResponseController` can reach the underlying writer where appropriate.
- Keep existing `Flush` and `Hijack` behavior intact.

Non-goals:
- Do not add broad new wrapper abstractions.
- Do not change compression, recovery, or body-limit response semantics.
- Do not support optional interfaces that cannot be safely replayed after buffering.

Files:
- `internal/httputil/http_response.go`
- `middleware/bodylimit/body_limit.go`
- `middleware/recovery/recover.go`
- `middleware/compression/gzip.go`
- targeted tests

Tests:
- `go test -timeout 20s ./middleware/... ./internal/httputil/...`
- `go test -race -timeout 60s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Not required unless behavior changes beyond optional interface preservation.

Done Definition:
- Wrapped writers expose `Unwrap`.
- Existing middleware tests and race tests pass.

Outcome:

