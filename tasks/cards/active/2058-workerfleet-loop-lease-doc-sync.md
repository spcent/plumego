# Card 2058

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/README.md
- reference/workerfleet/deploy/README.md
- reference/workerfleet/docs/storage.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
Depends On: 2057

Goal:
Synchronize workerfleet documentation with the implemented loop lease families, including `notification_deliver`.

Scope:
Update README, deployment notes, storage docs, and English/Chinese technical design docs so every place that lists Mongo-backed loop leases includes Kubernetes sync, status sweep, alert evaluation, and notification delivery. Keep the deployment guidance conservative: `replicas: 1` remains the default, and multi-replica operation still requires Mongo storage, unique owners, and reviewed notification expectations.

Non-goals:
- Do not change code or deployment replica counts.
- Do not add lease metrics.
- Do not claim default multi-replica production readiness.

Files:
- reference/workerfleet/README.md
- reference/workerfleet/deploy/README.md
- reference/workerfleet/docs/storage.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md

Acceptance Tests:
- Documentation mentions `notification_deliver` or notification delivery wherever loop lease families are enumerated.

Tests:
- Docs-only diff check.

Docs Sync:
- reference/workerfleet/README.md
- reference/workerfleet/deploy/README.md
- reference/workerfleet/docs/storage.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md

Validation:
- git diff --check
- rg -n "kube_sync|status_sweep|alert_evaluate|notification_deliver|notification delivery" reference/workerfleet/README.md reference/workerfleet/deploy/README.md reference/workerfleet/docs/storage.md reference/workerfleet/docs/design/technical-design.md reference/workerfleet/docs/design/technical-design.zh-CN.md

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
