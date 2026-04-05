# Card 0803

Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- `x/ai/provider/provider_test.go`
- `x/ai/session/session_test.go`
- `x/ai/streaming/streaming_test.go`
- `x/ai/tool/tool_test.go`
Depends On:
- `0802-x-ai-offline-stable-tier-examples.md`

Goal:
- Deepen contract coverage around the stable-tier `x/ai` packages before any broader stability claims.

Scope:
- Add focused negative-path and boundary-contract tests for provider, session, streaming, and tool packages.
- Cover explicit invariants such as error propagation, storage edge cases, stream lifecycle behavior, and tool policy enforcement where current tests are thin.
- Keep the assertions package-local and fast enough for routine iteration.

Non-goals:
- Do not broaden coverage across experimental `x/ai` packages in this card.
- Do not redesign stable-tier APIs to fit tests.
- Do not add external-service integration requirements.

Files:
- `x/ai/provider/provider_test.go`
- `x/ai/session/session_test.go`
- `x/ai/streaming/streaming_test.go`
- `x/ai/tool/tool_test.go`

Tests:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go test -race -timeout 60s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go vet ./x/ai/...`

Docs Sync:
- Touch `docs/modules/x-ai/README.md` only if the strengthened tests expose a mismatch in documented stable-tier guarantees.

Done Definition:
- Stable-tier package tests cover the key success and failure contracts the primer relies on.
- New assertions stay local to provider, session, streaming, and tool instead of leaking orchestration concerns into the card.
- Test coverage strengthens confidence without introducing hidden globals or network-dependent paths.

Outcome:
