# Card 0544

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_test.go
Depends On: 2253

Goal:
Reject malformed `query` tags with an empty parameter name instead of binding from an empty query key.

Scope:
- Trim the query tag name before binding.
- Treat `query:",omitempty"` and whitespace-only names as invalid bind destinations.
- Keep `query:"-"` and absent tags unchanged.
- Add focused regression coverage.

Non-goals:
- Do not add default field-name inference.
- Do not change existing scalar or slice query binding behavior.
- Do not change error envelope mapping.

Files:
- `contract/context_bind.go`
- `contract/context_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing tag parsing.

Done Definition:
- Empty query tag names return errors wrapping `ErrInvalidBindDst`.
- Existing query binding tests continue to pass.

Outcome:
- Query tag names are now trimmed before use.
- Empty query tag names now fail explicitly with `ErrInvalidBindDst` instead of reading from the empty query key.
