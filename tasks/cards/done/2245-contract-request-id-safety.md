# Card 2245

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/request_id.go
- contract/request_id_test.go
Depends On: 2244

Goal:
Reject unsafe request id values that contain control characters.

Scope:
- Keep trimming surrounding whitespace.
- Ignore request ids containing ASCII control characters or DEL.
- Add regression coverage for newline, tab, NUL, and valid punctuation.

Non-goals:
- Do not add request-id generation policy to `contract`.
- Do not impose a length or charset policy beyond transport safety.
- Do not change `RequestIDHeader`.

Files:
- `contract/request_id.go`
- `contract/request_id_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; request-id generation policy remains middleware-owned.

Done Definition:
- Unsafe ids are not stored in context.
- Existing request-id round trips continue to pass.

Outcome:
- `WithRequestID` now ignores request ids containing ASCII control characters or DEL after trimming.
- Added regression coverage for whitespace trimming, unsafe ids, and visible punctuation.
