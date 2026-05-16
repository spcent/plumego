# Card 1449: Post-v1 Convergence Plan Closure

Milestone: post-v1-convergence
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: tasks
Owned Files:
- `tasks/cards/active/README.md`
- `tasks/cards/blocked/README.md`
- `tasks/cards/done/1449-post-v1-convergence-plan-closure.md`
Depends On: Cards 1374 through 1380 and card 71907ae7 queue lifecycle commit

Goal:
- Record the post-v1 Codex iteration plan in `tasks/` after extension family
  path convergence and queue lifecycle cleanup.

Scope:
- Treat `v1.0.0` as already published and preserve stable-root freeze.
- Keep completed convergence implementation in cards 1374 through 1380.
- Keep incomplete extension promotion evidence in `tasks/cards/blocked/`.
- Keep the active queue empty until the next executable post-v1 card exists.

Non-goals:
- Do not change runtime behavior.
- Do not promote any extension maturity status.
- Do not add stable-root public API.
- Do not switch branches or rewrite release tags.

Files:
- `tasks/cards/done/1449-post-v1-convergence-plan-closure.md`
- `tasks/cards/active/README.md`
- `tasks/cards/blocked/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./internal/checks/...`

Docs Sync:
- Not required beyond this task record; this card records already-implemented
  task queue state.

Done Definition:
- Completed convergence cards are discoverable in `tasks/cards/done/`.
- Blocked evidence cards are discoverable in `tasks/cards/blocked/`.
- The active queue does not advertise blocked work as executable.
- Evidence and workflow checks pass.

Outcome:
- Completed implementation cards:
  - 1374: moved `x/mq` to `x/messaging/mq`
  - 1375: moved `x/pubsub` to `x/messaging/pubsub`
  - 1376: moved `x/scheduler` and `x/webhook` to `x/messaging`
  - 1377: split remaining secondary-family convergence into executable cards
  - 1378: moved `x/discovery` and `x/ipc` to `x/gateway`
  - 1379: moved `x/ops` and `x/devtools` to `x/observability`
  - 1380: moved `x/cache` to `x/data/cache`
- Blocked evidence cards:
  - 1367: `x/tenant` beta evidence closure
  - 1370: `x/ai` stable-tier beta evidence closure
  - 1371: selected `x/data` beta evidence closure
  - 1372: `x/gateway/discovery` surface beta evidence closure
  - 1373: `x/messaging` app-facing service beta evidence closure
- Verification recorded for this closure:
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/reference-layout`
  - `go test -timeout 20s ./internal/checks/...`
