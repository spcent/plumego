# Card 0769

Milestone: M-003
Recipe: specs/change-recipes/api-contract.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/stream.go`
- `x/websocket/conn.go`
- `x/websocket/protocol_compliance_test.go`
- `x/websocket/advanced_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0768

Goal:
- Make large-message and streaming semantics accurate and stable-worthy.

Problem:
`ReadMessageReader` is a bounded reader, but it still buffers frame payloads into a `bytes.Buffer`. The name and documentation can be mistaken for true zero-copy or unbounded streaming.

Scope:
- Decide whether to keep `ReadMessageReader` as a bounded buffered reader or replace/add a true streaming API.
- Ensure `ReadMessage` remains clearly documented as full in-memory read.
- Keep read limits enforced across fragmented messages.
- Add tests for fragmented large messages, limit boundaries, and early close behavior.

Non-goals:
- Do not remove read limits.
- Do not support unbounded messages by default.
- Do not add non-stdlib buffering dependencies.

Files:
- `x/websocket/stream.go`
- `x/websocket/conn.go`
- `x/websocket/protocol_compliance_test.go`
- `x/websocket/advanced_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for large-message, read-limit, and streaming API semantics.

Done Definition:
- API names and docs match actual memory behavior.
- Large fragmented messages remain bounded by configured limits.
- Users can tell which API buffers and which API streams.

Outcome:
- Kept `ReadMessageReader` as the stable-candidate bounded buffered reader.
- Clarified Go doc that `ReadMessageReader` is not zero-copy or unbounded
  streaming.
- Clarified `ReadMessage` as a full in-memory read returning an owned payload
  copy.
- Confirmed existing fragmented-message, read-limit, and early-close coverage
  remains the required behavior.
