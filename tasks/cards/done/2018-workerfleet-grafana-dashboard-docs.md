# Card 2018

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/docs
Owned Files:
- `reference/workerfleet/docs/grafana.md`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/README.md`
Depends On:
- `tasks/cards/done/2017-workerfleet-metrics-instrumentation.md`
Blocked By: —

Goal:
- Document the Grafana dashboard plan for workerfleet Prometheus metrics, including panel groups, PromQL examples, and alerting recommendations.
- Give operators a direct path from `/metrics` scrape data to useful pod/node/case visualizations.

Scope:
- Add dashboard documentation for fleet overview, node heatmap, case phase latency, throughput, and alert panels.
- Provide PromQL examples for active cases, worker status, phase duration p95/p99, case finish rate, and non-online worker distribution.
- Document scrape configuration assumptions and cardinality cautions.

Non-goals:
- Do not generate a Grafana JSON dashboard unless separately requested.
- Do not add Prometheus deployment manifests.
- Do not change metrics implementation.

Files:
- `reference/workerfleet/docs/grafana.md`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/README.md`

Tests:
- `cd reference/workerfleet && go test ./...`
- Documentation examples should be syntactically coherent and reference implemented metric names.

Docs Sync:
- Link `grafana.md` from `reference/workerfleet/README.md`.

Done Definition:
- Operators have documented PromQL and dashboard panel guidance for workerfleet.
- The docs reference only implemented metric names and approved labels.
- Cardinality and retention assumptions are explicit.

Outcome:
- Added Grafana dashboard planning documentation with scrape assumptions, panel groups, PromQL examples, alerting recommendations, cardinality constraints, and README/metrics links.
