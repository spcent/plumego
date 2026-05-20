# Card 2031

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/app
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/config_test.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/alerts.md`
Depends On:
- `tasks/cards/active/2030-workerfleet-domain-policy-surface.md`

## Goal

Expose worker status, step stuck, and alert threshold policies through validated config objects so workerfleet can run with explicit dev and production-ready settings.

## Scope

- Add app config fields and env parsing for status and alert policy inputs.
- Validate unsafe or contradictory values at startup.
- Introduce explicit profile-oriented defaults without changing the module boundary.

## Non-goals

- Do not implement multi-environment config file loading.
- Do not change notifier transport semantics.
- Do not add UI or admin APIs for runtime policy editing.

## Files

- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/config_test.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/alerts.md`

## Acceptance Tests

- `reference/workerfleet/internal/app/config_test.go: TestLoadConfigParsesStatusAndAlertPolicies`
- `reference/workerfleet/internal/app/config_test.go: TestValidateConfigRejectsUnsafeThresholds`
- `reference/workerfleet/internal/app/bootstrap_test.go: TestBootstrapUsesConfiguredPolicies`

## Tests

- Add boundary tests for minimum, maximum, and zero-value threshold handling.
- Keep existing env parsing tests green.

## Docs Sync

- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/alerts.md`
- `reference/workerfleet/docs/design/technical-design.md`

## Validation

- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./...`
- `gofmt -l reference/workerfleet/internal/app`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added explicit app-level profile and policy configuration for worker status and alert thresholds. `WORKERFLEET_PROFILE` now selects dev or prod defaults, status policy inputs are parsed and validated from env, alert thresholds derive from the effective status policy unless explicitly overridden, and bootstrap now injects configured status and alert policies into ingest, loop, and alert runners. README, alerts, and technical design docs were updated to document the profile and fail-closed startup validation behavior.
