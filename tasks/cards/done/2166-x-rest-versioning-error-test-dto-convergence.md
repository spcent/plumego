# Card 2166: x/rest/versioning Error Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: done
Primary Module: x/rest/versioning
Owned Files:
- `x/rest/versioning/version_test.go`
Depends On: none

Goal:
Make the unsupported-version error test decode a typed response shape.

Problem:
`TestUnsupportedVersionUsesVersioningCode` decodes the fixed structured error
response through `map[string]any`. The test exists to lock a stable versioning
error code, so the response shape should be explicit.

Scope:
- Replace the generic map decode with a local typed error response struct.
- Keep existing status and code assertions.

Non-goals:
- Do not change versioning middleware behavior or public APIs.
- Do not add dependencies.

Files:
- `x/rest/versioning/version_test.go`

Tests:
- `go test -race -timeout 60s ./x/rest/versioning/...`
- `go test -timeout 20s ./x/rest/versioning/...`
- `go vet ./x/rest/versioning/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- The unsupported-version error test no longer decodes through `map[string]any`.
- The listed validation commands pass.

Outcome:
- Added a local typed error response fixture for the unsupported-version test.
- Replaced the generic map decode while preserving the status and code checks.

Validation:
- `go test -race -timeout 60s ./x/rest/versioning/...`
- `go test -timeout 20s ./x/rest/versioning/...`
- `go vet ./x/rest/versioning/...`
