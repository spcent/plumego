# Card 0437: x/webhook Outbound Handler Transport Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/webhook
Owned Files:
- `x/webhook/out.go`
- `x/webhook/outbound_test.go`
- `docs/modules/x-webhook/README.md`
Depends On: none

Goal:
Converge outbound webhook HTTP handlers around one route-param path, typed
response DTOs, and structured query validation.

Problem:
`x/webhook/out.go` has already centralized many error helpers, but outbound
handlers still read route params directly from `contract.RequestContext...` at
each call site and build several success payloads with ad hoc maps. The
`enabled` list filter silently treats any non-`1`/`true` value as false, which
means malformed query input changes filtering behavior instead of returning a
structured validation error. Trigger, list, delivery detail, and replay
responses also have endpoint contracts that are clearer as local DTO structs.

Scope:
- Add private helpers for required route params and optional boolean query
  parsing.
- Make malformed `enabled` query values return a structured validation error
  rather than silently filtering disabled targets.
- Replace outbound success maps with local DTO structs for target lists,
  trigger responses, delivery lists/details, and replay responses.
- Add focused tests for invalid `enabled`, missing route params, and response
  shapes.

Non-goals:
- Do not change inbound webhook verification.
- Do not change delivery retry or signing behavior.
- Do not change route paths.
- Do not add a new response envelope or per-feature error family.

Files:
- `x/webhook/out.go`
- `x/webhook/outbound_test.go`
- `docs/modules/x-webhook/README.md`

Tests:
- `go test -race -timeout 60s ./x/webhook/...`
- `go test -timeout 20s ./x/webhook/...`
- `go vet ./x/webhook/...`

Docs Sync:
Update `docs/modules/x-webhook/README.md` only if the outbound query validation
contract is documented.

Done Definition:
- Outbound route-param extraction has one private implementation path.
- Malformed boolean query values fail through `contract.WriteError`.
- Success responses no longer use one-off maps where local DTOs are clearer.
- The listed validation commands pass.

Outcome:
- Added private outbound helpers for required route params and optional boolean
  query parsing.
- Made malformed `enabled` query values return structured validation errors
  instead of silently filtering as disabled.
- Replaced outbound target list, trigger, delivery list/detail, and replay
  success maps with local DTO response structs.
- Added tests for invalid `enabled`, missing route params, and typed outbound
  response shapes.
- Documented the outbound `enabled` query validation contract.

Validation:
- `go test -race -timeout 60s ./x/webhook/...`
- `go test -timeout 20s ./x/webhook/...`
- `go vet ./x/webhook/...`
