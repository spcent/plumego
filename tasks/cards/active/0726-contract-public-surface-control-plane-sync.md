# Card 0726

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- none

Goal:
Make the `contract` public surface and module documentation match the actual exported API before further stable hardening.

Scope:
- Compare the actual `go doc ./contract` surface with `contract/module.yaml` and the contract module README.
- Add missing public entrypoints that are intentionally stable.
- Clarify success/error envelope behavior that is already frozen.
- Fix stale or misleading wording, including trace wording on `WriteError`.

Non-goals:
- Do not remove or rename exported symbols in this card.
- Do not change runtime behavior.
- Do not migrate extension-owned error codes here.

Files:
- contract/module.yaml
- docs/modules/contract/README.md
- contract/errors.go

Tests:
- go test -timeout 20s ./contract/...
- go run ./internal/checks/module-manifests
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/contract/README.md` for implemented behavior only.
- Update `contract/module.yaml` public entrypoints to match intentional stable exports.

Done Definition:
- Public entrypoint documentation no longer omits intentional stable `contract` symbols.
- The README documents the empty success envelope, binding/cache semantics, and trace carrier boundary.
- Targeted contract tests and manifest/boundary checks pass.

Outcome:
