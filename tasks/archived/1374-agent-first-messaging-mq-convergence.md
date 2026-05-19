# Card 1374

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/**`
- `x/mq/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
Depends On: Card 1373 not required; beta evidence closure is separate.

Goal:
- Collapse the durable queue primitive into the messaging family so queue work starts from `x/messaging`.

Scope:
- Move `x/mq` under the `x/messaging` family path.
- Update repository imports and tests for the new queue primitive path.
- Update taxonomy, dependency rules, package hotspots, and module manifests.

Non-goals:
- Do not change queue behavior.
- Do not promote `x/messaging` maturity.
- Do not move `x/messaging/pubsub`, `x/messaging/scheduler`, or `x/messaging/webhook` in this card.

Files:
- `x/messaging/**`
- `x/mq/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`

Tests:
- `go test -timeout 20s ./x/messaging/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Updated current module docs and control-plane references for the new `x/messaging/mq` path.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/mq` path.
- Queue tests pass under the messaging family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/mq` to `x/messaging/mq`.
- Updated Go imports, taxonomy, dependency rules, hotspot routing, maturity metadata, module docs, and current evidence docs.
- Validation passed:
  - `go test -timeout 20s ./x/messaging/... ./x/ai/distributed/...`
  - `go vet ./x/messaging/... ./x/ai/distributed/...`
  - `go test -timeout 20s ./internal/checks/...`
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/dependency-rules`
