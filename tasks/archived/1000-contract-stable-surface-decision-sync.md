# Card 1000

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- 0740

Goal:
Record final stable decisions for the remaining broad contract surfaces: validation, response envelope, request context, and trace baggage.

Scope:
- Document the support boundary for `ValidateStruct` rules as compatibility behavior.
- Record the empty success envelope decision for body-eligible nil responses.
- Clarify `RequestContext` must remain route metadata only.
- Clarify `TraceContext.Baggage` is copied carrier data only; extraction, injection, limits, and policy live in observability owners.
- Update module manifest guardrails where useful without expanding public API.

Non-goals:
- Do not remove public symbols.
- Do not change runtime behavior.
- Do not add tracing or validation infrastructure.

Files:
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow

Docs Sync:
- This card is documentation/control-plane sync only.

Done Definition:
- Stable support boundaries are explicit for validation, response envelope, request context, and trace baggage.
- Future expansion paths are described as separate symbol-change or owner-module work.
- Manifest and workflow checks pass.

Outcome:
- Documented the retained `ValidateStruct` compatibility rules and placed new validation expansion outside `contract`.
- Recorded the empty success-envelope `{}` behavior as the stable body-eligible nil response representation.
- Clarified that `RequestContext` remains router metadata only.
- Clarified that `TraceContext.Baggage` is copied carrier data only and that propagation policy lives with observability owners.
- Updated module manifest guardrails for these broad compatibility surfaces.

Validation:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
