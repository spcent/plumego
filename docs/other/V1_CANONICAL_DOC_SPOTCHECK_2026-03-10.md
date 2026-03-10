# v1 Canonical Docs Spot-Check (2026-03-10)

Status: Active  
Scope: v1 canonical docs only (not historical planning docs)

## Checked Targets

- `README.md`
- `README_CN.md`
- `docs/README.md`
- `docs/getting-started.md`
- `docs/modules/core/*.md`
- `docs/modules/router/*.md`
- `docs/modules/middleware/*.md`
- `docs/modules/ai/*.md`
- `docs/modules/log/*.md`
- `docs/modules/store/*.md`

## Gate Rule

Disallow removed/pre-freeze APIs in canonical docs:

- `WithRecommendedMiddleware`
- `WithTenantMiddleware`
- `WithTenantConfigManager`
- `core.WithRecovery(...)`
- `core.WithLogging(...)`
- `app.Group(...)`
- `core.WithServer(...)`

## Result (2026-03-10)

- Script: `scripts/check-doc-api-drift.sh`
- Outcome: PASS (`doc-api-drift: ok`)

## Operational Notes

- Historical docs may still include old APIs for migration context. They must be clearly labeled as historical/non-canonical.
- CI gate is enforced by `.github/workflows/quality-gates.yml` via `bash scripts/check-doc-api-drift.sh`.
- Full pre-release aggregate check is available at `scripts/v1-release-readiness-check.sh`.
