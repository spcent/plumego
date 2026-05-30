# AGENTS.md - reference/with-ai

Operational guide for agents working in `reference/with-ai`.

This file is intentionally short. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read `reference/standard-service` first — this service has the same core shape
and extends it with AI capabilities.

## 2. Purpose And Boundaries

`reference/with-ai` demonstrates the minimal wiring to add LLM capabilities to a
plumego service: an offline mock provider, in-memory session management, a tool
registry, and streaming. The mock provider makes the demo runnable without API
keys; swap it with a real provider from `x/ai/provider` for production use.

Hard rules:

- `x/ai/*` packages are allowed. No other `x/*` imports.
- No new third-party dependencies.
- No hidden globals, `init()` registration, reflection routing, or controller scanning.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all routes explicit in `internal/app/routes.go`.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- The AI handler must not embed provider credentials. Credentials belong in config
  and must not be logged.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, and environment.
- `internal/app/app.go`: middleware chain (requestid → recovery), graceful shutdown.
- `internal/app/routes.go`: route table — POST /api/chat and GET /api/ai/status.
- `internal/handler`: AIHandler with Chat() and Status() methods.

Dependency direction:

```text
main.go → config → app → handler
handler → contract, router, x/ai/*
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Swap the mock provider for a real one:

1. Construct the real provider in `internal/app/routes.go` (e.g. `provider.NewOpenAI(...)`).
2. Pass credentials from config, not hardcoded.
3. The `AIHandler.Provider` field is the interface — no other change needed.

Add a tool:

1. Implement `x/ai/tool.Tool` in `internal/handler` or a new `internal/tools` package.
2. Register it with the `ToolRegistry` in `routes.go`.
3. Add a focused test for the tool's `Execute` method.

Add an endpoint:

1. Add a method to `AIHandler` (or a new handler struct) in `internal/handler`.
2. Register the route in `routes.go`.
3. Add a focused test in `internal/app/app_test.go` or the handler test file.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-ai && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- No provider credentials are hardcoded or logged.
- `x/ai/*` imports only — no other `x/*` packages.
- Tool registry uses an allowlist policy; deny by default.
- Response envelope uses `contract.WriteResponse` / `contract.WriteError`.
- Session manager cleanup is triggered on graceful shutdown.
