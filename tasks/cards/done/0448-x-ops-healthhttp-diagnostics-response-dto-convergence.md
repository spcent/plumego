# Card 0448: x/ops/healthhttp Diagnostics Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/ops/healthhttp
Owned Files:
- `x/ops/healthhttp/runtime.go`
- `x/ops/healthhttp/handlers_test.go`
- `x/ops/healthhttp/history_test.go`
- `docs/modules/x-ops/README.md`
Depends On: none

Goal:
Converge healthhttp diagnostics responses and tests on typed DTOs instead of
ad hoc maps.

Problem:
`DiagnosticsHandler` still builds its JSON success payload with
`map[string]any`, adding optional manager fields by mutation. Nearby
healthhttp responses already use typed payloads, and history stats already has
a `HistoryStats` struct, but the stats test still decodes through a generic
map. This makes the ops diagnostics surface less explicit than the rest of the
package.

Scope:
- Replace the diagnostics response map with a local typed DTO struct.
- Preserve existing JSON field names and omission behavior for nil managers.
- Decode diagnostics and history stats tests into typed payloads.
- Update the x/ops primer with the typed healthhttp diagnostics response
  policy.

Non-goals:
- Do not change liveness text responses or CSV history export.
- Do not change health manager execution, readiness, config, or component
  behavior.
- Do not add dependencies.

Files:
- `x/ops/healthhttp/runtime.go`
- `x/ops/healthhttp/handlers_test.go`
- `x/ops/healthhttp/history_test.go`
- `docs/modules/x-ops/README.md`

Tests:
- `go test -race -timeout 60s ./x/ops/healthhttp/...`
- `go test -timeout 20s ./x/ops/healthhttp/...`
- `go vet ./x/ops/healthhttp/...`

Docs Sync:
Update `docs/modules/x-ops/README.md` for typed healthhttp diagnostics JSON
success payloads.

Done Definition:
- `DiagnosticsHandler` no longer assembles success data through a one-off map.
- Tests decode diagnostics and history stats through typed DTOs.
- The listed validation commands pass.

Outcome:
- Added a local typed diagnostics response DTO with optional manager-owned
  fields.
- Replaced diagnostics response map mutation with direct typed response
  construction.
- Updated diagnostics and history stats tests to decode typed payloads.
- Documented the x/ops healthhttp diagnostics DTO policy.

Validation:
- `go test -race -timeout 60s ./x/ops/healthhttp/...`
- `go test -timeout 20s ./x/ops/healthhttp/...`
- `go vet ./x/ops/healthhttp/...`
