# Card 2250

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/request_id.go
- contract/errors_test.go
- contract/request_id_test.go
Depends On: 2249

Goal:
Apply the canonical request-id safety rules to explicit API error request ids.

Scope:
- Reuse one request-id normalization path for context ids and API error ids.
- Trim explicit API error request ids.
- Drop explicit API error request ids containing control characters.
- Add focused coverage for builder and `WriteError` behavior.

Non-goals:
- Do not add request-id generation policy.
- Do not add length limits.
- Do not change the error envelope shape.

Files:
- `contract/errors.go`
- `contract/request_id.go`
- `contract/errors_test.go`
- `contract/request_id_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this reuses the existing request-id safety contract.

Done Definition:
- Explicit unsafe error request ids are not emitted.
- Valid explicit error request ids remain preserved.

Outcome:
- Context request ids and API error request ids now use the same normalization helper.
- `ErrorBuilder.RequestID` and `WriteError` trim valid ids and drop unsafe ids containing control characters.
