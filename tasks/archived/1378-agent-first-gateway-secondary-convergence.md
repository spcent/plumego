# Card 1378

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/**`
- `x/discovery/**`
- `x/ipc/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
Depends On: Card 1377

Goal:
- Collapse gateway subordinate primitives into the gateway family path.

Scope:
- Move `x/discovery` under `x/gateway`.
- Move `x/ipc` under `x/gateway`.
- Update imports, tests, taxonomy, package hotspots, dependency rules, and module manifests.

Non-goals:
- Do not change discovery or IPC behavior.
- Do not change `x/rest` or `reference/standard-service`.
- Do not widen stable root imports.

Files:
- `x/gateway/**`
- `x/discovery/**`
- `x/ipc/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Update `docs/modules/x-gateway/README.md`, `docs/modules/x-discovery/README.md`, and `docs/modules/x-ipc/README.md` if discovery or import paths change.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/discovery` or `github.com/spcent/plumego/x/ipc` paths.
- Gateway, discovery, and IPC tests pass under the gateway family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/discovery` to `x/gateway/discovery` and `x/ipc` to `x/gateway/ipc`.
- Updated import comments, module manifests, taxonomy, hotspots, maturity signals, routing, docs, and active evidence references to the gateway family paths.
- Verified `go test -timeout 20s ./x/gateway/...`, `go vet ./x/gateway/...`, `go run ./internal/checks/agent-workflow`, `go run ./internal/checks/dependency-rules`, and `go run ./internal/checks/module-manifests`.
