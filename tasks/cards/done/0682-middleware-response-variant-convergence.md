# Card 0682

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/transport_test.go
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
Depends On: 0681

Goal:
Converge response variant handling so middleware does not duplicate `Vary` tokens and request coalescing does not merge common representation or credential variants.

Scope:
- Add an internal transport helper that appends `Vary` values only when they are not already present, including comma-joined existing values.
- Route CORS and gzip `Vary` writes through that helper.
- Extend the default coalescing key with common representation and credential headers.
- Canonicalize configured header names in `HeaderAwareKeyFunc`.

Non-goals:
- Do not change CORS authorization policy.
- Do not change gzip compression eligibility.
- Do not add business policy, tenant resolution, or response-cache semantics to coalescing.

Files:
- `middleware/internal/transport/http.go`
- `middleware/internal/transport/transport_test.go`
- `middleware/cors/cors.go`
- `middleware/cors/cors_test.go`
- `middleware/compression/gzip.go`
- `middleware/compression/gzip_test.go`
- `middleware/coalesce/coalesce.go`
- `middleware/coalesce/coalesce_test.go`

Tests:
- `go test -timeout 20s ./middleware/internal/transport ./middleware/cors ./middleware/compression ./middleware/coalesce`
- `go test -race -timeout 60s ./middleware/coalesce`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this tightens internal transport behavior and conservative defaults.

Done Definition:
- Existing `Vary` tokens are preserved without duplicate additions.
- Default coalescing key partitions credentialed and negotiated request variants.
- Targeted tests, coalesce race test, and vet pass.

Outcome:
- Added `internal/transport.AddVary` with case-insensitive token de-duplication and comma-value awareness.
- Routed CORS and gzip `Vary` writes through the shared helper.
- Extended the default coalesce key with common representation and credential headers: `Accept`, `Accept-Encoding`, `Accept-Language`, `Authorization`, `Cookie`, and `Range`.
- Reused canonicalized header hashing for `HeaderAwareKeyFunc`.
- Validation run: `go test -timeout 20s ./middleware/internal/transport ./middleware/cors ./middleware/compression ./middleware/coalesce`; `go test -race -timeout 60s ./middleware/coalesce`; `go vet ./middleware/...`.
