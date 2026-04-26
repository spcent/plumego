# Card 3105: Log Documentation and Boundary Sync

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: log
Owned Files:
- `docs/modules/log/README.md`
- `log/module.yaml` only if checks or stable guidance need tightening
- `tasks/cards/active/README.md`
Depends On:
- `tasks/cards/done/3101-log-callsite-depth.md`
- `tasks/cards/done/3102-log-field-semantics.md`
- `tasks/cards/done/3103-log-json-resilience.md`
- `tasks/cards/done/3104-log-file-backend-lifecycle.md`

Goal:
Align the log module documentation and active queue with the implemented
contracts from the cleanup pass.

Problem:
The code already contains several important log contract choices, but the module
docs only state them at a high level:
- context-aware methods preserve call shape and must not infer transport
  metadata;
- `LoggerConfig.Format` is the one stable construction path;
- field precedence and text escaping behavior need to be explicit after cleanup;
- file/flag/default backend behavior remains internal and must not be treated
  as stable application bootstrap.

Scope:
- Update log module docs with implemented behavior and non-goals.
- Tighten module manifest test commands only if the cleanup changed required
  checks.
- Update the active-card queue after the cleanup cards are completed.

Non-goals:
- Do not document unimplemented behavior.
- Do not promote internal glog helpers or file backend details into public API.
- Do not change code unless a doc inconsistency exposes a small missed test.

Tests:
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`

Done Definition:
- Docs describe only the stable `log` behavior that exists after the cleanup.
- Active queue no longer lists completed log cards.
- Boundary and manifest checks pass.
