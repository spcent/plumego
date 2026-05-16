# Card 1375

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/**`
- `x/pubsub/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
Depends On: Card 1374

Goal:
- Collapse the pub/sub broker primitive into the messaging family so broker work is subordinate to `x/messaging`.

Scope:
- Move `x/pubsub` under the `x/messaging` family path.
- Update imports, tests, taxonomy, package hotspots, and dependency rules.
- Keep app-facing messaging entrypoints in `x/messaging`.

Non-goals:
- Do not change broker behavior.
- Do not move `x/messaging/scheduler` or `x/messaging/webhook`.
- Do not change stable middleware or core.

Files:
- `x/messaging/**`
- `x/pubsub/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`

Tests:
- `go test -timeout 20s ./x/messaging/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Updated current module docs and control-plane references for the new `x/messaging/pubsub` path.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/pubsub` path.
- Pub/sub tests pass under the messaging family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/pubsub` to `x/messaging/pubsub`.
- Updated Go imports, taxonomy, dependency rules, hotspot routing, maturity metadata, module docs, and current evidence docs.
- Validation passed:
  - `go test -timeout 20s ./x/messaging/... ./x/messaging/webhook/... ./x/devtools/pubsubdebug ./x/ai/distributed/... ./reference/with-webhook/...`
  - `go vet ./x/messaging/... ./x/messaging/webhook/... ./x/devtools/pubsubdebug ./x/ai/distributed/... ./reference/with-webhook/...`
  - `go test -timeout 20s ./...` from `cmd/plumego`
  - `go vet ./...` from `cmd/plumego`
  - `go test -timeout 20s ./internal/checks/...`
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/module-manifests`
