# Card 1395

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: router
Owned Files:
- router/dispatch.go
- router/router_contract_test.go
- docs/modules/router/README.md
Depends On:
- 1394

Goal:
Decide and freeze router response-shape behavior for nil router, malformed request, not found, and method-not-allowed paths.

Scope:
- Add or tighten tests for `ServeHTTP` nil/zero-value router behavior, nil request, nil URL, stdlib 404, and structured 405.
- Decide whether router initialization and malformed request errors stay as direct stdlib responses or move to `contract.WriteError`.
- If response shape changes, update docs and keep the change isolated to router dispatch.

Non-goals:
- Do not change route matching semantics.
- Do not change static file serving behavior.
- Do not add auth, validation, or business policy to router.

Files:
- router/dispatch.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -race -timeout 60s ./router
- go test -timeout 20s ./router
- go vet ./router

Docs Sync:
- Required only if visible response shape changes.

Done Definition:
- Router miss/error response behavior is explicitly tested.
- Router docs state any intentional stdlib response exceptions.
- Target checks pass.

Outcome:
