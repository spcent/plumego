# Card 0712

Milestone: M-002
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: store
Owned Files: store/module.yaml, health/module.yaml, log/module.yaml, metrics/module.yaml, tasks/cards/active/0712-stable-support-roots-regression.md
Depends On: 0710-middleware-chain-error-regression, 0711-security-fail-closed-regression

Goal:
Confirm `store`, `health`, `log`, and `metrics` remain stable support roots with no feature-catalog or extension drift in M-002.

Scope:
Review module manifests and current regression coverage for `store`, `health`, `log`, and `metrics`; create follow-up implementation cards only if concrete gaps are found.

Non-goals:
Do not change runtime code in this card.
Do not move tenant-aware storage, topology-heavy data, or exporter wiring into stable roots.
Do not update `x/*`.

Files:
store/module.yaml
health/module.yaml
log/module.yaml
metrics/module.yaml
tasks/cards/active/0712-stable-support-roots-regression.md

Tests:
go test -timeout 20s ./store/... ./health/... ./log/... ./metrics/...
go run ./internal/checks/module-manifests
go run ./internal/checks/dependency-rules

Docs Sync:
Record review outcome in this card. User-facing docs only change if they currently overstate support-root behavior.

Done Definition:
Outcome states whether support roots need follow-up implementation cards.
Support-root tests and manifest/boundary checks pass.
No runtime behavior changes are made in this review card.

Outcome:
To be filled during execution.

