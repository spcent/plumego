# Card 0756

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- docs/modules/contract/README.md
Depends On:
- 0755

Goal:
Record the final stable decision for the intentionally broad `contract` public surface.

Scope:
- Document that the current public surface remains supported for stable compatibility.
- Record the accepted compatibility cost for `Ctx`, binding helpers, validation helpers, trace carrier data, response helpers, and error helpers.
- Add final rules for future public API additions, breaking changes, and deprecated-symbol removal.
- Reference the active conformance checks that guard the most risky misuse patterns.

Non-goals:
- Do not remove public symbols in this card.
- Do not change runtime behavior.
- Do not widen `contract` ownership beyond transport primitives.

Files:
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- This card is the docs sync.

Done Definition:
- Stable public surface decision is explicit and actionable.
- Future compatibility rules are documented next to the module guidance.
- Targeted checks pass.
