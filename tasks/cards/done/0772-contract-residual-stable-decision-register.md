# Card 0772

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0771

Goal:
Record the remaining v1 stable tradeoffs as explicit residual decisions instead of re-discovering them in repeated audits.

Scope:
- Add a concise residual decision register for the intentionally broad public API, exported `APIError`, permissive `WriteResponse` status behavior, `BindJSON`, and `ValidateStruct`.
- Tie each accepted tradeoff to its guardrail or future breaking-change path.

Non-goals:
- Do not remove or narrow stable public APIs.
- Do not hide `APIError` fields in v1.
- Do not change response, bind, validation, or error behavior.

Files:
- docs/modules/contract/README.md
- contract/module.yaml

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract stable decision docs.

Done Definition:
- Residual stable tradeoffs are explicitly listed and tied to guardrails.
- Target checks pass.

Outcome:
- Added a residual stable decision register to contract docs covering the broad API surface, exported `APIError`, permissive `WriteResponse`, `BindJSON`, and `ValidateStruct`.
- Mirrored the accepted residual v1 tradeoffs in `contract/module.yaml`.
- Tied each accepted tradeoff to an existing guardrail or future breaking-change path.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
