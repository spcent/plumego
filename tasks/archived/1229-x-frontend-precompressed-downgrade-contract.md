# Card 1229: x/frontend Precompressed Downgrade Contract

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
Depends On: 0759

Goal:
Make precompressed variant downgrade behavior explicit before stable adoption.

Scope:
- Document that missing or unreadable compressed variants downgrade to the
  original asset when identity is acceptable.
- Document that directory scan errors still fail mount construction.
- Record that compressed variant downgrade observability is a stable decision
  point if future users need logging or metrics.

Non-goals:
- Do not add logging or metrics.
- Do not change response behavior.
- Do not promote module status.

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
This is a behavior contract documentation card.

Done Definition:
- Best-effort compressed variant downgrade is explicit.
- Scan fail-fast behavior remains distinct from per-request downgrade behavior.
- The listed validation commands pass.

Outcome:
- Documented request-time missing or unreadable compressed variants as
  best-effort misses that downgrade to the original asset when `identity` is
  acceptable.
- Kept directory precompressed scan errors documented as mount construction
  failures.
- Recorded the lack of compressed-variant downgrade observability as a module
  risk and stable decision point.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
