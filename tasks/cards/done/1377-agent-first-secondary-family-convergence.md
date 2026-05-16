# Card 1377

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: specs
Owned Files:
- `x/gateway/**`
- `x/observability/**`
- `x/data/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
Depends On: Cards 1374, 1375, and 1376

Goal:
- Apply the same canonical-root convergence pattern to gateway, observability, and data subordinate roots.

Scope:
- Split into implementation subcards before moving code.
- Converge `x/discovery` and `x/ipc` under `x/gateway`.
- Converge `x/ops` and `x/devtools` under `x/observability`.
- Converge `x/cache` under `x/data`.

Non-goals:
- Do not move messaging roots from this card.
- Do not change runtime behavior.
- Do not widen stable root imports.

Files:
- `x/gateway/**`
- `x/observability/**`
- `x/data/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Required for affected module primers after each subcard.

Done Definition:
- Gateway, observability, and data convergence is represented by small executable cards or completed moves.
- Taxonomy and hotspot checks continue to pass.
- No subordinate root competes with its canonical family as a default agent entrypoint.

Outcome:
- Split secondary-family convergence into executable cards:
  - Card 1378: `x/discovery` and `x/ipc` into `x/gateway`
  - Card 1379: `x/ops` and `x/devtools` into `x/observability`
  - Card 1380: `x/cache` into `x/data`
- Validation passed:
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/reference-layout`
  - `go run ./internal/checks/dependency-rules`
