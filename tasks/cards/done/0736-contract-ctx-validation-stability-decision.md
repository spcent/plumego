# Card 0736

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0735

Goal:
Record the stable decision for the broad `Ctx` and reflection validation helpers so their support burden is explicit.

Scope:
- Document that `Ctx` and `ValidateStruct` are retained as legacy compatibility helpers, not the preferred new endpoint style.
- Clarify that new stable transport code should prefer standard `http.Request` handling and explicit decode/validation where practical.
- Record remaining release caveats around future breaking changes without changing public symbols.

Non-goals:
- Do not remove `Ctx`, `BindJSON`, or `ValidateStruct`.
- Do not change runtime behavior.
- Do not add a new validation framework.

Files:
- docs/modules/contract/README.md
- contract/module.yaml

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow

Docs Sync:
- This card is documentation/control-plane sync only.

Done Definition:
- The contract module docs distinguish canonical transport primitives from retained legacy compatibility helpers.
- Future breaking work is described as a planned symbol-change task rather than hidden in runtime behavior.
- Manifest and workflow checks pass.

Outcome:
- Documented `Ctx`, `BindJSON`, `BindQuery`, and `ValidateStruct` as retained compatibility helpers rather than preferred expansion points.
- Clarified that complex validation policy belongs in owning modules.
- Added module manifest guardrails against expanding `Ctx` or `ValidateStruct` into framework surfaces.
- Kept runtime behavior unchanged.

Validation:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
