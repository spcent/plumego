# Card 1376

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/**`
- `x/scheduler/**`
- `x/webhook/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
Depends On: Card 1375

Goal:
- Collapse scheduling and webhook delivery primitives into the messaging family.

Scope:
- Move `x/scheduler` and `x/webhook` under the `x/messaging` family path.
- Update imports, tests, taxonomy, package hotspots, and dependency rules.
- Preserve explicit app wiring through `x/messaging` entrypoints.

Non-goals:
- Do not change scheduling, retry, verification, or delivery behavior.
- Do not change stable root packages.
- Do not promote maturity status.

Files:
- `x/messaging/**`
- `x/scheduler/**`
- `x/webhook/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`

Tests:
- `go test -timeout 20s ./x/messaging/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Updated current module docs and control-plane references for the new `x/messaging/scheduler` and `x/messaging/webhook` paths.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/scheduler` or `github.com/spcent/plumego/x/webhook` paths.
- Scheduler and webhook tests pass under the messaging family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/scheduler` to `x/messaging/scheduler`.
- Moved `x/webhook` to `x/messaging/webhook`.
- Updated Go imports, taxonomy, dependency rules, hotspot routing, maturity metadata, module docs, and current evidence docs.
- Validation passed:
  - `go test -timeout 20s ./x/messaging/... ./x/ai/distributed/... ./reference/with-webhook/...`
  - `go vet ./x/messaging/... ./x/ai/distributed/... ./reference/with-webhook/...`
  - `go test -timeout 20s ./...` from `cmd/plumego`
  - `go vet ./...` from `cmd/plumego`
  - `go test -timeout 20s ./internal/checks/...`
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/module-manifests`
