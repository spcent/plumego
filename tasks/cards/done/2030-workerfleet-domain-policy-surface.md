# Card 2030

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/domain
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- `reference/workerfleet/internal/domain/status.go`
- `reference/workerfleet/internal/domain/alert_rules.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/domain/alert_engine_test.go`
- `reference/workerfleet/internal/domain/README.md`
Depends On:
- `tasks/cards/done/2022-workerfleet-alert-notification-loop.md`

## Goal

Turn worker status and alert thresholds into explicit domain policy objects instead of scattered defaults so runtime config can later bind to a stable policy surface.

## Scope

- Make status and alert thresholds explicit in domain policy types.
- Remove threshold scattering across alert rules and status evaluation helpers.
- Keep existing default behavior unchanged unless tests intentionally lock a bug fix.

## Non-goals

- Do not wire environment variables in this card.
- Do not change HTTP APIs.
- Do not add tenant-specific or cluster-specific policy behavior.

## Files

- `reference/workerfleet/internal/domain/status.go`
- `reference/workerfleet/internal/domain/alert_rules.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/domain/alert_engine_test.go`
- `reference/workerfleet/internal/domain/README.md`

## Acceptance Tests

- `reference/workerfleet/internal/domain/status_test.go: TestEvaluateWorkerStatus`
- `reference/workerfleet/internal/domain/alert_engine_test.go: TestAlertEngineUsesConfiguredThresholds`
- `reference/workerfleet/internal/domain/alert_engine_test.go: TestAlertEngineDefaultPolicyMatchesCurrentBehavior`

## Tests

- Add negative tests for zero or invalid policy thresholds.
- Keep dedupe and alert resolution tests green.

## Docs Sync

- `reference/workerfleet/internal/domain/README.md`
- `reference/workerfleet/docs/design/technical-design.md`

## Validation

- `cd reference/workerfleet && go test ./internal/domain/...`
- `cd reference/workerfleet && go vet ./internal/domain/...`
- `gofmt -l reference/workerfleet/internal/domain`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

- Introduced explicit `AlertPolicy` alongside `StatusPolicy` and added validation helpers so domain thresholds now live behind stable policy objects instead of scattered rule defaults.
- Updated `AlertEngine` to derive its default alert thresholds from `StatusPolicy` while still allowing explicit alert-policy overrides through a domain option.
- Added domain tests for invalid thresholds, configured alert thresholds, and default-policy parity with prior behavior.

Validation Run:

- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go test ./internal/domain/...`
- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go vet ./internal/domain/...`
- `gofmt -l reference/workerfleet/internal/domain`
