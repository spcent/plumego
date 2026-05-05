# Card 0743: x/frontend Option Contract Clarity

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
Depends On: 0742

Goal:
Make the intentionally sealed `Option` API contract explicit before future
stable evaluation.

Scope:
- Document that external callers should use the exported `With*` options rather
  than constructing custom options against the unexported config type.
- Record why the public surface remains small and sealed.
- Update evidence docs so API snapshot interpretation matches this contract.

Non-goals:
- Do not export `config`.
- Do not introduce a new builder API.
- Do not change runtime behavior.

Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
This is a docs sync card.

Done Definition:
- The sealed option contract is explicit in package docs and module primer.
- Evidence docs describe the API snapshot as a sealed option surface.
- The listed validation commands pass.

Outcome:
- Documented `Option` as an intentionally sealed constructor input.
- Clarified that applications should compose exported `With*` helpers and that
  future stable knobs should appear as explicit helpers.
- Updated extension evidence to interpret the API snapshot accordingly.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
  - `go run ./internal/checks/extension-maturity`
