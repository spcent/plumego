# Card 0740

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_test.go
- contract/context_extended_test.go
- docs/modules/contract/README.md
Depends On:
- 0739

Goal:
Make query binding more stable for custom scalar types by supporting `encoding.TextUnmarshaler`.

Scope:
- Support scalar destination fields implementing `encoding.TextUnmarshaler`.
- Support pointer-to-TextUnmarshaler fields without introducing nested pointer support.
- Preserve existing primitive and primitive-slice behavior.
- Add focused positive and negative tests.
- Document the query binding support matrix.

Non-goals:
- Do not add a general query binding framework.
- Do not support maps, nested pointers, or arbitrary struct parsing.
- Do not change missing query value semantics.

Files:
- contract/context_bind.go
- contract/context_test.go
- contract/context_extended_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Update contract docs with the query binding type support matrix.

Done Definition:
- Custom scalar types can parse themselves through `encoding.TextUnmarshaler`.
- Unsupported types still fail with `ErrInvalidBindDst`.
- Targeted tests, vet, and manifest checks pass.

Outcome:
