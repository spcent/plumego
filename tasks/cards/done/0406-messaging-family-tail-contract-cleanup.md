# Card 0406: Messaging Family Tail Contract Cleanup

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/messaging
Owned Files:
- x/messaging/sms_metrics.go
- x/messaging/sms_metrics_test.go
- x/pubsub/distributed.go
- x/pubsub/distributed_test.go
- docs/modules/x-messaging/README.md
Depends On: none

Goal:
- Clean up remaining messaging-family contract inconsistencies after the broader admin HTTP convergence work.
- Make SMS metrics exporter construction explicit for nil dependencies and finish stable error-code coverage in subordinate pubsub cluster HTTP helpers.

Scope:
- Audit `NewSMSPrometheusExporter` for panic-only nil metric source behavior.
- Audit `x/pubsub/distributed.go` cluster/admin helper errors that still rely on type-only errors without explicit codes.
- Add focused tests for nil SMS metric source and pubsub method/auth error responses.
- Keep `x/messaging` as the family entrypoint and `x/pubsub` as a subordinate package.

Non-goals:
- Do not change message delivery, broker, scheduler, webhook, or queue semantics.
- Do not add hidden provider globals or tenant bootstrap ownership.
- Do not make `x/pubsub` compete with `x/messaging` as the app-facing discovery root.

Files:
- `x/messaging/sms_metrics.go`: add explicit non-panic construction behavior for nil metric source.
- `x/messaging/sms_metrics_test.go`: cover nil and valid SMS metrics exporter construction.
- `x/pubsub/distributed.go`: add stable codes to remaining cluster/admin helper errors.
- `x/pubsub/distributed_test.go`: cover method/auth error responses.
- `docs/modules/x-messaging/README.md`: document family constructor/error expectations if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/messaging/... ./x/pubsub/...`
- `go test -timeout 20s ./x/messaging/... ./x/pubsub/...`
- `go vet ./x/messaging/... ./x/pubsub/...`

Docs Sync:
- Required if public constructor behavior or public error codes/messages change.

Done Definition:
- SMS metrics exporter has a clear non-panic construction path for nil dependencies.
- Remaining pubsub cluster/admin HTTP errors in scope use explicit stable codes.
- Focused tests cover both cleanup areas.
- The three listed validation commands pass.

Outcome:
- Added `NewSMSPrometheusExporterE` and `ErrNilMetricRecordSource` for explicit non-panic SMS metrics exporter construction.
- Preserved `NewSMSPrometheusExporter` compatibility by delegating to the error-returning constructor and panicking on invalid input.
- Made pubsub cluster method/auth helper error codes explicit while preserving existing response status and messages.
- Added focused SMS exporter tests for nil source, panic compatibility, valid construction, and default namespace behavior.
- Documented SMS exporter constructor guidance in `docs/modules/x-messaging/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./x/messaging/... ./x/pubsub/...`
  - `go test -timeout 20s ./x/messaging/... ./x/pubsub/...`
  - `go vet ./x/messaging/... ./x/pubsub/...`
