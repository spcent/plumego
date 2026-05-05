# Card 0764: x/frontend Variant Downgrade Observability

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
Depends On: 0763

Goal:
Make the decision around precompressed variant downgrade observability explicit.

Scope:
- Record that per-request compressed variant open/stat failures currently emit
  no log or metric.
- Record that adding observability would require a deliberate public or
  module-local design decision.
- Keep current behavior as best-effort downgrade unless owner decides otherwise.

Non-goals:
- Do not add logging or metrics in this card.
- Do not add dependencies.
- Do not change response behavior.

Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
This is a governance and behavior contract documentation card.

Done Definition:
- Observability gap is explicit and owner-reviewable.
- Current best-effort downgrade behavior remains documented.
- The listed validation commands pass.
