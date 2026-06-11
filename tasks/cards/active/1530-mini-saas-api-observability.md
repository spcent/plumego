# Card 1530

Milestone: M-025
Recipe: (none — observability capability)
Context Package: use-cases/mini-saas-api
Priority: P1
State: active
Primary Module: use-cases/mini-saas-api
Owned Files: internal/handler/{audit,metricsguard}.go, app/routes observability wiring
Depends On: 1527

## Goal

Request metrics flow into a Prometheus collector exposed on GET /metrics
(optionally token-guarded with constant-time comparison), and admins can read
the tenant audit trail newest-first with a limit parameter.

## Scope

x/observability PrometheusCollector into httpmetrics; PrometheusExporter on
/metrics; RequireMetricsToken guard; AuditHandler on GET /api/v1/tenant/audit.

## Non-goals

No tracing/OTel, no x/observability/ops or devtools surfaces, no dashboards.

## Files

internal/handler/audit.go
internal/handler/metricsguard.go
internal/app/app.go
internal/app/routes.go

## Acceptance Tests

internal/app/observability_test.go: TestAcceptanceMetricsEndpoint
internal/app/observability_test.go: TestAcceptanceAuditTrail

## Tests

Metrics token guard (none/wrong/right); audit member 403; audit limit +
invalid limit; audit content free of passwords and refresh tokens.

## Docs Sync

(env.example already documents APP_METRICS_TOKEN)

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 180s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (n/a).

## Outcome

httpmetrics now feeds observability.NewPrometheusCollector("mini_saas_api");
exporter mounted at GET /metrics behind RequireMetricsToken (subtle constant-
time compare, no-op when APP_METRICS_TOKEN empty — startup warning logged).
GET /api/v1/tenant/audit (admin) returns newest-first entries with ?limit=N.
Audit content verified to exclude passwords and refresh tokens.
