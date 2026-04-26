# Card 2222

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files: middleware/requestid/helpers.go; middleware/requestid/request_id.go; middleware/requestid/request_id_test.go; tasks/cards/done/2222-middleware-requestid-empty-generator.md
Depends On: 2221

Goal:
Prevent empty or whitespace-only custom request ID generator output from becoming the canonical request ID.

Scope:
- Trim generated request IDs before attaching them.
- Fall back to the package generator when a custom generator returns an empty value.
- Avoid writing blank request ID headers.
- Add tests for empty and whitespace generator output.

Non-goals:
- Do not change the request ID format or decode algorithm.
- Do not add validation policy beyond non-empty trimming.
- Do not remove custom generator support.

Files:
- `middleware/requestid/helpers.go`
- `middleware/requestid/request_id.go`
- `middleware/requestid/request_id_test.go`

Tests:
- `go test -timeout 20s ./middleware/requestid`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; custom generators remain supported with defensive fallback.

Done Definition:
- Empty custom generator output no longer creates empty context/header IDs.
- Whitespace request IDs are normalized.
- Targeted middleware tests and vet pass.

Outcome:
- Routed middleware generation through `EnsureRequestID`.
- Trimmed generated IDs and fell back to package generation when custom output is empty.
- Made `AttachRequestID` skip blank IDs instead of writing empty context/header values.
- Added tests for empty generator fallback, trim behavior, and blank attach skip.
- Validation run: `go test -timeout 20s ./middleware/requestid`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
