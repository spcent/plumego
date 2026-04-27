# Card 5404: Health Public Examples

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: health
Owned Files:
- `health/example_test.go`
Depends On: 5403

Goal:
Add executable examples for the stable `health` public surface so expected usage
is easier to discover and harder to drift.

Scope:
- Add examples for `HealthState.IsReady`.
- Add examples for `HealthStatus`, `ComponentHealth`, and `ReadinessStatus`.

Non-goals:
- Do not add HTTP or manager examples to `health`.
- Do not change public APIs or runtime behavior.
- Do not introduce dependencies.

Files:
- `health/example_test.go`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; examples are tested package documentation.

Done Definition:
- The public health DTOs and readiness predicate have executable examples.
- The listed validation commands pass.

Outcome:
- Added executable examples for readiness state checks and each public health
  DTO.
- Kept examples transport-free so HTTP ownership stays outside stable `health`.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
