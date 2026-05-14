# Card 1392

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/README.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/docs/metrics.md
- reference/workerfleet/docs/alerts.md
Depends On:
- 1387
- 1391

Goal:
- Synchronize workerfleet docs with implemented runtime loops, alert loop, metrics, and shutdown behavior.

Scope:
- Remove stale notes that say the HTTP entrypoint does not start Kubernetes sync or alert evaluation loops after implementation proves otherwise.
- Document implemented error visibility from Card 1387.
- Keep English and Chinese technical design docs consistent.
- Confirm README runtime configuration matches actual env variables and defaults.
- Document implemented behavior only.

Non-goals:
- Do not document planned auth behavior before Card 1393 implements it.
- Do not change API examples unless implementation payloads changed.
- Do not add roadmap-scale planning prose to docs.

Files:
- reference/workerfleet/README.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/docs/metrics.md
- reference/workerfleet/docs/alerts.md

Tests:
- rg -n "does not start|should be added|Current implementation note" reference/workerfleet/docs
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- This card is docs sync.

Done Definition:
- Runtime-loop and alert-loop docs match code.
- English and Chinese technical design docs describe the same implemented behavior.
- No stale implementation notes remain.
- Lightweight checks pass.

Outcome:
