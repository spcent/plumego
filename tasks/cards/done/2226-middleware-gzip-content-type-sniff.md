# Card 2226

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files: middleware/compression/gzip.go; middleware/compression/gzip_test.go; tasks/cards/done/2226-middleware-gzip-content-type-sniff.md
Depends On: 2225

Goal:
Let gzip middleware make the same content-type decision net/http would make when handlers write a body without setting `Content-Type`.

Scope:
- Detect content type from the first body bytes before deciding whether to compress when no header is set.
- Preserve explicit content type behavior and skip binary/streaming/error responses.
- Add tests for text responses without explicit content type and binary responses without explicit content type.

Non-goals:
- Do not redesign gzip buffering or threshold behavior.
- Do not compress streaming responses.
- Do not add non-stdlib dependencies.

Files:
- `middleware/compression/gzip.go`
- `middleware/compression/gzip_test.go`

Tests:
- `go test -timeout 20s ./middleware/compression`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this aligns middleware behavior with net/http defaults.

Done Definition:
- Text responses without explicit content type can be compressed.
- Binary responses without explicit content type are still skipped.
- Targeted middleware tests and vet pass.

Outcome:
- Added first-write content type detection using `http.DetectContentType` before gzip compression decision.
- Added tests for implicit text compression and implicit binary passthrough.
- Validation run: `go test -timeout 20s ./middleware/compression`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
