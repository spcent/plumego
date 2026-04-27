# Card 0175

Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/provider/example_test.go`
- `x/ai/session/example_test.go`
- `x/ai/streaming/example_test.go`
- `x/ai/tool/example_test.go`
- `docs/modules/x-ai/README.md`
Depends On:
- `0174-x-ai-stability-tier-doc-sync.md`

Goal:
- Add runnable offline examples for the stable-tier `x/ai` packages without requiring live network calls.

Scope:
- Add example tests that exercise provider, session, streaming, and tool composition using existing local mocks and in-memory helpers.
- Keep examples explicit and handler-level so they demonstrate composition without hidden registration or bootstrap magic.
- Update the module primer only enough to point readers to the new runnable examples.

Non-goals:
- Do not add real provider credentials, network calls, or integration-test dependencies.
- Do not introduce a new top-level examples app for `x/ai`.
- Do not cover experimental `x/ai` subpackages in this card.

Files:
- `x/ai/provider/example_test.go`
- `x/ai/session/example_test.go`
- `x/ai/streaming/example_test.go`
- `x/ai/tool/example_test.go`
- `docs/modules/x-ai/README.md`

Tests:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go test -race -timeout 60s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`

Docs Sync:
- Keep `docs/modules/x-ai/README.md` aligned with the actual example entrypoints and the manifest-declared stable tier.

Done Definition:
- Each stable-tier package has at least one runnable offline example.
- Examples compose existing package APIs explicitly and compile under `go test`.
- The primer points readers at the example-backed stable-tier entrypoints instead of abstract promises.

Outcome:
- Added runnable offline examples for `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, and `x/ai/tool` using stable-tier public APIs only.
- The examples rely on in-memory helpers and mock-backed providers, so they execute under `go test` without network calls or bootstrap wiring.
- `docs/modules/x-ai/README.md` now points readers at the new example entrypoints instead of only describing them abstractly.

Validation Run:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go test -race -timeout 60s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
