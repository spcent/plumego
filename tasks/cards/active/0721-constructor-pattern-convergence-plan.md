# Card 0721

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: specs
Owned Files:
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `specs/package-hotspots.yaml`
- `tasks/cards/active/README.md`
Depends On: —

Goal:
- Turn constructor-pattern drift into a concrete convergence plan before changing public APIs.

Problem:
The codebase currently uses several constructor/error styles: `New/NewE`, `NewGateway/NewGatewayE`, `Middleware/MiddlewareE`, and panic convenience wrappers. Some are intentional compatibility paths, but the rule is not explicit enough for new extension work.

Scope:
- Inventory public constructors and panic wrappers in stable roots and `x/*`.
- Classify each pattern as canonical, legacy-compatible, or candidate for later migration.
- Update the style/control-plane guidance with a concise rule for:
  - middleware constructors
  - app-facing extension constructors
  - panic convenience wrappers
- Add follow-up cards only for modules that need implementation changes.

Non-goals:
- Do not remove or rename exported symbols in this card.
- Do not change runtime behavior.
- Do not widen stable public APIs.

Files:
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `specs/package-hotspots.yaml`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Required. This is a control-plane clarification card.

Done Definition:
- Constructor-pattern guidance is explicit enough for future cards.
- Existing panic wrappers are classified rather than casually removed.
- Any required implementation migrations are split into separate module-owned cards.

Outcome:
-
