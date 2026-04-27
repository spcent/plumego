# Card 0445: x/messaging API Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/api.go`
- `x/messaging/api_test.go`
- `docs/modules/x-messaging/README.md`
Depends On: none

Goal:
Converge x/messaging HTTP success payloads on local typed DTO structs while
keeping route paths, status codes, and response field names unchanged.

Problem:
`x/messaging/api.go` still uses one-off map literals for several app-facing HTTP
success responses: send accepted, receipt list, and channel health. Neighboring
response contracts already use named structs such as `BatchResult`,
`ServiceStats`, and `Receipt`, so these maps make the API surface less explicit
and harder to test consistently.

Scope:
- Replace the remaining messaging HTTP success maps with local DTO structs.
- Keep existing JSON field names: `id`, `status`, `receipts`, `count`, and
  `channels`.
- Add focused endpoint tests that decode the contract envelope into typed DTOs.
- Update the x/messaging primer with the typed response DTO policy.

Non-goals:
- Do not change route registration.
- Do not change send, quota, receipt, provider, scheduler, or monitor behavior.
- Do not expose subordinate mq, pubsub, scheduler, or webhook internals.
- Do not add dependencies.

Files:
- `x/messaging/api.go`
- `x/messaging/api_test.go`
- `docs/modules/x-messaging/README.md`

Tests:
- `go test -race -timeout 60s ./x/messaging/...`
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`

Docs Sync:
Update `docs/modules/x-messaging/README.md` to state that app-facing HTTP
success responses use typed local DTOs instead of one-off maps.

Done Definition:
- `x/messaging/api.go` no longer writes app-facing success payloads with
  one-off map literals.
- Focused tests cover send accepted, receipt list, and channel health response
  shapes through the contract envelope.
- The listed validation commands pass.

Outcome:
- Added local typed DTOs for send accepted, receipt list, and channel health
  success responses.
- Replaced the remaining one-off app-facing success maps in `x/messaging/api.go`.
- Added focused handler tests that decode the contract envelope into typed
  messaging API payloads.
- Documented the x/messaging local DTO policy for app-facing HTTP success
  payloads.

Validation:
- `go test -race -timeout 60s ./x/messaging/...`
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`
