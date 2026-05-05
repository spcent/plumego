# Card 0757

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- 0756

Goal:
Make the `contract` public surface inventory match the actual exported API before stable.

Scope:
- Add exported `RequestContext` to the module manifest and README public entrypoint lists.
- Record the retention decision for legacy exported sentinels and codes that are not strongly used in production.
- Keep runtime behavior unchanged.

Non-goals:
- Do not remove exported symbols in this card.
- Do not add new public API.
- Do not change request context semantics.

Files:
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go run ./internal/checks/module-manifests
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- This card is docs and manifest sync.

Done Definition:
- Public inventory explicitly includes `RequestContext`.
- Retained legacy exported symbols have a documented stable decision.
- Target checks pass.
