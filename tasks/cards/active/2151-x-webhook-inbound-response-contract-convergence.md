# Card 2151: x/webhook Inbound Response Contract Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/webhook
Owned Files:
- `x/webhook/in.go`
- `x/webhook/webhook_component_test.go`
- `docs/modules/x-webhook/README.md`
Depends On: none

Goal:
Converge inbound webhook HTTP success responses around local DTO structs and
focused response-shape tests.

Problem:
Outbound webhook handlers now use typed local response DTOs, but inbound GitHub
and Stripe handlers still build success payloads with repeated
`map[string]any` literals for deduped and published events. The fields are
stable endpoint contracts (`ok`, `provider`, `topic`, `event_type`, delivery or
event IDs, `deduped`, and `body_bytes`), so the contract is clearer and easier
to test as typed DTOs.

Scope:
- Replace inbound GitHub and Stripe success maps with local DTO structs.
- Keep existing verification, dedupe, topic selection, and publish semantics.
- Add focused tests for successful GitHub and Stripe response shapes, including
  dedupe behavior.
- Keep existing canonical error helpers and error codes unchanged.
- Update docs only to note that inbound success bodies are typed local DTOs.

Non-goals:
- Do not change HMAC/GitHub/Stripe verification behavior.
- Do not change topic naming or pubsub message metadata.
- Do not add dependencies.
- Do not change outbound webhook handlers.

Files:
- `x/webhook/in.go`
- `x/webhook/webhook_component_test.go`
- `docs/modules/x-webhook/README.md`

Tests:
- `go test -race -timeout 60s ./x/webhook/...`
- `go test -timeout 20s ./x/webhook/...`
- `go vet ./x/webhook/...`

Docs Sync:
Update `docs/modules/x-webhook/README.md` with the inbound success response
shape policy if code changes are made.

Done Definition:
- Inbound success responses no longer use one-off maps.
- GitHub and Stripe published and deduped response shapes are covered by tests.
- Existing verification and error semantics remain unchanged.
- The listed validation commands pass.

Outcome:
