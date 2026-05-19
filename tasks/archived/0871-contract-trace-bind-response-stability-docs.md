# Card 0871

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
- contract/module.yaml
- contract/trace.go
- contract/response.go
- contract/context_bind.go
Depends On:
- 0729

Goal:
Make trace carrier, binding cache, and response no-body semantics explicit enough for stable use.

Scope:
- Tighten comments for trace carrier behavior without adding tracing infrastructure.
- Clarify that `BindJSON` is a cache-oriented helper and `BindOptions.MaxBodySize` is post-read.
- Clarify no-body success status behavior and empty success envelope behavior in the module README.
- Add focused tests only if an existing documented behavior lacks coverage.

Non-goals:
- Do not add new trace propagation infrastructure.
- Do not change `Ctx` from legacy compatibility carrier into a framework request bag.
- Do not introduce a new decode helper or response envelope.

Files:
- docs/modules/contract/README.md
- contract/module.yaml
- contract/trace.go
- contract/response.go
- contract/context_bind.go
- contract/*_test.go

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- All changes in this card are documentation or comment sync for implemented behavior.

Done Definition:
- Stable behavior docs describe trace, bind, and response semantics without implying unsupported infrastructure.
- Any missing regression coverage discovered during implementation is added.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Tightened comments on `WriteJSON`, `WriteResponse`, `BindJSON`, `WithTraceContext`, and `WithSpanIDString`.
- Added module hints keeping trace propagation in `x/observability` and treating `Ctx.BindJSON` as legacy compatibility.
- Added stable semantics notes for trace carrier, binding cache, and no-body/empty success responses.
- Confirmed existing regression coverage already covers the documented response and trace behaviors.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests
