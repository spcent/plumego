# Card 5403: Health JSON Contract Coverage

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: health
Owned Files:
- `health/json_contract_test.go`
Depends On: 5402

Goal:
Add focused JSON contract coverage for the public health DTOs so stable payload
shape changes are intentional.

Scope:
- Cover `HealthStatus` JSON fields, `omitempty` behavior, duration encoding,
  and dependency output.
- Cover `ComponentHealth` embedded field flattening.
- Cover `ReadinessStatus` component readiness output and omitted empty fields.

Non-goals:
- Do not change the public DTO fields or tags.
- Do not add custom JSON marshalers.
- Do not change HTTP handlers or extension behavior.

Files:
- `health/json_contract_test.go`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; this card records existing behavior in tests.

Done Definition:
- Public JSON shape for health DTOs is covered by focused tests.
- The listed validation commands pass.

Outcome:
- Added JSON contract coverage for `HealthStatus`, `ComponentHealth`, and
  `ReadinessStatus`.
- Locked existing `omitempty`, embedded field flattening, `time.Time`, and
  `time.Duration` JSON behavior without changing public DTOs.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
