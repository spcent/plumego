# x/ai

## Purpose

`x/ai` is the experimental capability family for AI-related adapters and workflows.
Import the owning subpackage directly; the root `x/ai` package is only a
module marker and not a canonical bootstrap surface.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is clearly AI capability work
- provider, session, streaming, tool, or orchestration work is involved

## Do not use this module for

- stable root abstractions
- generic HTTP bootstrap

## Stability tiers

The family is still experimental overall, but `x/ai/module.yaml` already
declares subpackage-level tiers:

- stable-tier subpackages: `provider`, `session`, `streaming`, `tool`
- experimental subpackages: `orchestration`, `semanticcache`, `marketplace`, `distributed`, `resilience`

Treat additional supporting subpackages under `x/ai/*` as experimental unless
the manifest explicitly promotes them.

Treat the manifest as the canonical source when these tiers change.

## Subpackage beta evidence

The root `x/ai` family remains experimental. Beta evaluation is tracked only at
the stable-tier subpackage level:

- `provider`: `docs/extension-evidence/x-ai-provider.md`
- `session`: `docs/extension-evidence/x-ai-session.md`
- `streaming`: `docs/extension-evidence/x-ai-streaming.md`
- `tool`: `docs/extension-evidence/x-ai-tool.md`

These records are evidence for compatibility-oriented adoption of the named
subpackage only. They do not promote `x/ai` as a family and do not cover
`orchestration`, `semanticcache`, `marketplace`, `distributed`, `resilience`, or
other experimental AI packages.

The subpackage evidence records are tracked in
`specs/extension-beta-evidence.yaml` under `subpackage_candidates`. The
promotion check validates those entries against the `stability_tiers.stable`
section in `x/ai/module.yaml`; root `x/ai` remains `experimental`.
Release comparison commands in the evidence records apply only to the named
subpackage and must not be used as root-family promotion evidence.

## Recommended stable-tier adoption path

For new AI service work that needs a compatibility-oriented path, start with
only the stable-tier subpackages:

1. Use `x/ai/provider` for provider-neutral request, response, streaming, and
   adapter selection contracts.
2. Use `x/ai/session` for explicit session state, message lifecycle, and
   storage ownership.
3. Use `x/ai/tool` for local tool registration, allow-list policy, and
   invocation boundaries.
4. Add `x/ai/streaming` only when the service needs streaming coordination or
   SSE-backed progress updates.

Wire these packages in the owning application, handler, or extension
constructor. Do not hide provider selection, session storage, stream managers,
or tool registries behind package globals or implicit registration.

Keep `x/ai/orchestration`, `x/ai/semanticcache`, `x/ai/marketplace`,
`x/ai/distributed`, and `x/ai/resilience` on the experimental path until their
contracts have separate promotion evidence. If a service needs those packages,
document the experimental dependency in the consuming module instead of treating
the whole `x/ai` family as beta-ready.

## Common entrypoints

- `x/ai/provider` — provider abstraction and provider adapters
- `x/ai/session` — AI session lifecycle and state handling
- `x/ai/streaming` — streaming primitives and stream coordination
- `x/ai/tool` — tool registration and execution policy
- `x/ai/orchestration` — multi-step agent workflow composition
- `x/ai/semanticcache` — embedding/vector-backed semantic cache flows

## Resilience primitive relationship

`x/resilience` owns reusable circuit breaker and rate-limit primitives for
cross-extension use. The older `x/ai/circuitbreaker` and `x/ai/ratelimit`
packages remain AI-provider-specific compatibility primitives used by
`x/ai/resilience`.

New cross-family resilience work should start in `x/resilience`. New AI provider
wrapping may use `x/ai/resilience`, but dynamic composition should prefer
`NewResilientProviderE` so invalid provider wiring returns an error instead of
panicking. `x/ai/ratelimit.TokenBucketLimiter` owns a cleanup goroutine only when
constructed with a cleanup interval, and callers should call `Close` when that
background cleanup is enabled.

## Streaming error contract

- `x/ai/sse` and `x/ai/streaming` HTTP handlers use structured `contract.WriteError` responses for setup and request failures.
- Default SSE stream setup, invalid JSON, workflow callback, and handler error-event paths use safe public messages and do not expose provider, callback, decoder, or stream creation internals.
- `x/ai/streaming` terminal SSE payloads use local typed DTO structs instead of one-off maps.
- `provider.Manager.RegisterE` and `streaming.StreamManager.RegisterE` are the explicit validation paths for dynamic registration input; the legacy `Register` methods remain compatibility wrappers for known-good inputs.

The stable-tier entrypoints are `provider`, `session`, `streaming`, and `tool`.
Use other `x/ai/*` packages with experimental-module expectations.

## Runnable offline examples

- `x/ai/provider/example_test.go` — mock-backed provider usage without network calls
- `x/ai/session/example_test.go` — in-memory session lifecycle and active-context retrieval
- `x/ai/streaming/example_test.go` — SSE-backed progress updates with an in-memory recorder
- `x/ai/tool/example_test.go` — offline tool registration, policy filtering, and execution

## First files to read

- `x/ai/module.yaml`
- the owning subpackage under `x/ai/*`
- `specs/repo.yaml`

## Boundary rules

- keep AI wiring explicit in handlers or owning extensions
- do not add hidden provider globals or registration side effects
- use `provider.AutoToolChoice`, `provider.NoneToolChoice`, and `provider.AnyToolChoice` for fresh tool-choice values; the older exported `ToolChoiceAuto`, `ToolChoiceNone`, and `ToolChoiceAny` variables remain only for stable API compatibility
- keep transport-only concerns out of `x/ai`
- do not require stable roots to know AI internals

## Validation focus

- `go test -race -timeout 60s ./x/ai/...`
- `go test -timeout 20s ./x/ai/...`
- `go vet ./x/ai/...`
- when stable-tier APIs change, add coverage at the provider, session, streaming, or tool boundary instead of documenting speculative guarantees

## Current test coverage

Stable-tier packages have unit and contract tests alongside their example files:

- `x/ai/provider` — Manager routing delegation, Complete/CompleteStream dispatch, custom Router injection via `WithRouter`; `ClaudeProvider` and `OpenAIProvider` adapter tests cover constructor defaults, option overrides (BaseURL, HTTPClient), successful completion (offline via `httptest.NewServer`), and API error paths (401/429 responses)
- `x/ai/session` — session CRUD, TTL expiry, message append with token counting, auto-trim, Update, pagination, context key/value store
- `x/ai/tool` — Registry register/get/list/execute, policy filtering via `AllowListPolicy`, builtin tools (echo, calculator, timestamp, read-file, write-file, bash), error result metrics

Experimental-tier packages have basic contract tests:

- `x/ai/semanticcache/cachemanager` — cache stats, cleanup all, warming from queries, and unsupported maintenance operations (`Compact`, selective cleanup policies, `WarmFromFile`) return `ErrUnsupportedMaintenance`
- `x/ai/marketplace` — Manager contract tests covering PublishAgent, GetAgent, SearchAgents, ListAgentVersions, RateAgent, IsAgentInstalled, ListInstalledAgents, InstallAgent error path, install/uninstall round-trip; two pre-existing bugs fixed (mutex deadlock in `UpdateDownloadCount`, `InstallationRecord.Metadata` type assertion after JSON round-trip)
