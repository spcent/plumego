# Card 0752: x/frontend Header Option Validation

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/config.go`
- `x/frontend/response_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0751

Goal:
Apply consistent safety validation to all options that write response headers.

Scope:
- Validate `WithCacheControl` and `WithIndexCacheControl` values for unsafe
  control characters.
- Validate `WithMIMETypes` values before they can be written as
  `Content-Type`.
- Keep normalization behavior for empty values and extension keys.
- Add focused regression tests.

Non-goals:
- Do not parse Cache-Control directives semantically.
- Do not reject valid MIME parameters such as charsets.
- Do not change `WithHeaders` policy beyond reuse.

Files:
- `x/frontend/config.go`
- `x/frontend/response_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -race -timeout 60s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document that cache and MIME options reject unsafe header values.

Done Definition:
- Cache and MIME option values cannot inject CR/LF/control characters.
- Existing header behavior remains passing.
- The listed validation commands pass.
