# Card 0102

Priority: P1

Goal:
- Tighten the boundary between `x/gateway` and `x/rest` so edge transport and reusable resource-interface work never compete for the same discovery role.

Scope:
- gateway and rest metadata
- architecture wording for edge transport versus resource APIs

Non-goals:
- Do not change runtime gateway or rest implementations in this card.
- Do not add new transport adapters.

Files:
- `specs/extension-entrypoints.yaml`
- `x/gateway/module.yaml`
- `x/rest/module.yaml`
- `docs/modules/x-gateway/README.md`
- `x/rest/README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep the gateway and rest primers aligned with the capability split.

Done Definition:
- `x/gateway` is clearly the edge transport surface.
- `x/rest` is clearly the reusable resource-interface surface.
- No top-level docs imply they are interchangeable entrypoints.
