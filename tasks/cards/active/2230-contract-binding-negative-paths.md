# Card 2230

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_test.go
- contract/write_bind_error_test.go
Depends On: 2229

Goal:
Make request binding fail explicitly for nil request state and unsupported tagged query fields instead of panicking or silently dropping data.

Scope:
- Add nil guards for `Ctx.BindJSON`, `Ctx.BindQuery`, and body reads.
- Return structured bind errors for tagged query fields with unsupported destination kinds.
- Preserve existing supported scalar, pointer-to-scalar, and primitive slice behavior.
- Add focused negative-path tests.

Non-goals:
- Do not add mixed-source binding.
- Do not add new validation tags.
- Do not change the `Ctx` public shape.

Files:
- `contract/context_bind.go`
- `contract/context_test.go`
- `contract/write_bind_error_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this makes documented binding limits enforceable.

Done Definition:
- Nil request/context binding returns errors instead of panicking.
- Unsupported tagged query fields return a bind error satisfying `errors.Is(err, ErrInvalidBindDst)`.
- Existing bind and validation tests continue to pass.

Outcome:
