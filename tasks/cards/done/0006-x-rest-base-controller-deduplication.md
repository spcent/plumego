# Card 0006

Priority: P1
State: done
Primary Module: x/rest
Owned Files:
  - x/rest/resource.go

Depends On: —

Goal:
`BaseResourceController` and `BaseContextResourceController` share 10 identical handler methods
(Index, Show, Create, Update, Delete, Patch, Options, Head, BatchCreate, BatchDelete)
that are pure copy-paste duplicates. The duplication should be eliminated via embedding.

Scope:
- Have `BaseContextResourceController` embed `BaseResourceController` and remove the 10 duplicate methods
- Adjust `BaseContextResourceController.resourceName()` calls to read from the embedded field `ResourceName` (already public in base)
- Keep the public API of both types unchanged, with no impact on any existing callers

Non-goals:
- Do not merge the two structs into one
- Do not change the ResourceController interface definition
- Do not modify QueryParams / QueryBuilder / ResourceOptions

Files:
  - x/rest/resource.go (lines 58–95 BaseResourceController methods, lines 487–524 BaseContextResourceController duplicate methods)

Tests:
  - go test ./x/rest/...
  - Confirm both controller types can be asserted against the ResourceController interface

Docs Sync: —

Done Definition:
- `BaseContextResourceController` no longer declares any methods that duplicate `BaseResourceController`
- All 10 methods satisfy the ResourceController interface via embedding automatically
- `go build ./x/rest/...` and `go test ./x/rest/...` both pass
- `grep -n "func (c \*BaseContextResourceController)" x/rest/resource.go` shows only methods that differ from base

Outcome:
The code already had the embedding in place. `BaseContextResourceController` only declares `ApplySpec` and `ParseQueryParams` (unique to it); the 10 handler methods (Index, Show, Create, Update, Delete, Patch, Options, Head, BatchCreate, BatchDelete) are inherited via embedding of `BaseResourceController`. All tests pass.
