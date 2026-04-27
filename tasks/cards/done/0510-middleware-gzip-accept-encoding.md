# Card 0510

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/compression/gzip.go; middleware/compression/gzip_test.go; tasks/cards/done/0510-middleware-gzip-accept-encoding.md
Depends On: 2219

Goal:
Make gzip compression honor HTTP `Accept-Encoding` tokens instead of substring matching.

Scope:
- Replace `strings.Contains(..., "gzip")` request negotiation with token parsing.
- Do not compress when `gzip;q=0` or only unrelated encodings are present.
- Keep wildcard encoding support when not explicitly disabled by `gzip;q=0`.
- Add tests for `gzip;q=0`, false-positive tokens such as `xgzip`, and wildcard encoding.

Non-goals:
- Do not redesign response buffering or compression thresholds.
- Do not add third-party content negotiation dependencies.
- Do not change public constructor names.

Files:
- `middleware/compression/gzip.go`
- `middleware/compression/gzip_test.go`

Tests:
- `go test -timeout 20s ./middleware/compression`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this is protocol-correct behavior under existing gzip support.

Done Definition:
- Gzip is only selected when explicitly acceptable by token or wildcard.
- `gzip;q=0` prevents compression.
- Targeted middleware tests and vet pass.

Outcome:
- Replaced substring matching with `Accept-Encoding` token parsing and q-value handling.
- Added coverage for `gzip;q=0`, false-positive `xgzip`, wildcard, and explicit gzip-over-wildcard cases.
- Validation run: `go test -timeout 20s ./middleware/compression`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
