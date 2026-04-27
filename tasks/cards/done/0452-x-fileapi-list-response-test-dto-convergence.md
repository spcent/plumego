# Card 0452: x/fileapi List Response Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: done
Primary Module: x/fileapi
Owned Files:
- `x/fileapi/handler_test.go`
Depends On: none

Goal:
Make the file list handler test decode the existing typed list response instead
of a generic map.

Problem:
`TestHandler_List` still decodes the contract envelope data into
`map[string]any` even though `handler.go` already defines `listResponse`. This
keeps the test less precise than the handler contract.

Scope:
- Replace the generic map decode with `listResponse`.
- Add assertions for item count and pagination fields already returned by the
  handler.

Non-goals:
- Do not change file API behavior, response fields, storage behavior, or route
  wiring.
- Do not add dependencies.

Files:
- `x/fileapi/handler_test.go`

Tests:
- `go test -race -timeout 60s ./x/fileapi/...`
- `go test -timeout 20s ./x/fileapi/...`
- `go vet ./x/fileapi/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- `TestHandler_List` no longer decodes through `map[string]any`.
- The listed validation commands pass.

Outcome:
- Updated `TestHandler_List` to decode `listResponse` from the success
  envelope.
- Added assertions for items and pagination fields.

Validation:
- `go test -race -timeout 60s ./x/fileapi/...`
- `go test -timeout 20s ./x/fileapi/...`
- `go vet ./x/fileapi/...`
