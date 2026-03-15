# Card 0402

Priority: P1

Goal:
- Expand negative-path and integration test coverage for critical stable roots
  before the v1 freeze, targeting error-path code that was exercised only by
  happy-path tests or not at all.

Scope:
- `contract`: add `WriteBindError` negative-path tests covering each sentinel
  error (`ErrRequestBodyTooLarge`, `ErrEmptyRequestBody`, `ErrInvalidJSON`,
  `ErrUnexpectedExtraData`) and validation field errors; verify HTTP status codes
  and JSON error codes are correct.
- `router`: add negative tests for frozen-router registration (must panic),
  duplicate route registration (must panic), route-param validation failure
  (must return 400), unknown path (must return 404), and double-slash path
  (must return 404 without matching the parameterised route).

Non-goals:
- Do not change the behaviour of any existing code.
- Do not add tests for `x/*` packages in this card.
- Do not add benchmarks.

Files:
- `contract/write_bind_error_test.go` (new)
- `router/negative_test.go` (new)

Tests:
- `go test -race -timeout 60s ./contract/... ./router/...`

Docs Sync:
- No prose changes required.

Done Definition:
- Both new test files compile and pass with race detection enabled.
- No existing tests are broken.
