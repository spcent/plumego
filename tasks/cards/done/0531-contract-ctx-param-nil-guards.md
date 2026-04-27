# Card 0531

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/context_core.go
- contract/context_extended_test.go
- contract/context_mustparam_test.go
Depends On: 2240

Goal:
Make `Ctx.Param` and `Ctx.MustParam` nil-receiver safe.

Scope:
- Return an absent param from `Param` when called on a nil `*Ctx`.
- Return `ErrMissingParam` from `MustParam` when called on a nil `*Ctx`.
- Add focused regression coverage.

Non-goals:
- Do not change whitespace semantics for existing params.
- Do not make `Ctx.Params` private.
- Do not add request bag behavior.

Files:
- `contract/context_core.go`
- `contract/context_extended_test.go`
- `contract/context_mustparam_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens legacy compatibility helpers.

Done Definition:
- Nil `*Ctx` param access never panics.
- Existing param tests continue to pass.

Outcome:
- `Ctx.Param` now treats a nil receiver as an absent parameter map.
- `Ctx.MustParam` keeps returning an error that wraps `ErrMissingParam` for nil receivers.
- Added focused regression coverage for both helper paths.
