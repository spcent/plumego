# Card 2035

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/deploy
Priority: P1
State: done
Primary Module: reference/workerfleet/deploy
Owned Files:
- `reference/workerfleet/deploy/README.md`
- `reference/workerfleet/deploy/deployment.yaml`
- `reference/workerfleet/deploy/rbac.yaml`
- `reference/workerfleet/deploy/servicemonitor.yaml`
- `reference/workerfleet/deploy/prometheusrule.yaml`
Depends On:
- `tasks/cards/done/2031-workerfleet-config-profile-and-policy-wiring.md`
- `tasks/cards/done/2032-workerfleet-metrics-contract-stabilization.md`

## Goal

Provide runnable Kubernetes deployment assets for workerfleet, including RBAC, scraping, and baseline alert rule examples.

## Scope

- Add reference manifests for deployment, service account and RBAC, ServiceMonitor, and PrometheusRule.
- Document required secrets and config env variables in a deploy README.
- Keep manifests aligned with the current HTTP, readiness, metrics, and runtime loop behavior.

## Non-goals

- Do not add Helm charts or Kustomize overlays in this card.
- Do not implement multi-cluster deployment shapes.
- Do not add production-specific secret values.

## Files

- `reference/workerfleet/deploy/README.md`
- `reference/workerfleet/deploy/deployment.yaml`
- `reference/workerfleet/deploy/rbac.yaml`
- `reference/workerfleet/deploy/servicemonitor.yaml`
- `reference/workerfleet/deploy/prometheusrule.yaml`

## Acceptance Tests

<!-- Manifest/doc card: no Go acceptance test function. -->

## Tests

- Validate YAML shape manually against the current workerfleet ports, probes, and env names.

## Docs Sync

- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/alerts.md`
- `reference/workerfleet/docs/metrics.md`

## Validation

- `git diff --check`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added a reference Kubernetes deployment bundle for workerfleet, including pod-read RBAC, a single-replica deployment, service exposure, Prometheus Operator scraping, and baseline Prometheus alert rules. The deploy README explains required secrets, production-oriented defaults, and the current single-replica assumption for runtime loops.
