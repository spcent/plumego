# CLAUDE.md — plumego

> Go 1.24+ | Toolchain: go1.24.4 | Main module: standard-library only

This file is a practical AI-assistant playbook for working in `spcent/plumego`.
Use it together with `AGENTS.md`.

## 1) What Plumego Is

Plumego is a standard-library-first HTTP toolkit.

Primary goals:

- Keep HTTP flow explicit (`http.Handler`, `http.Request`, `http.ResponseWriter`)
- Keep app lifecycle explicit (`core.New(...)` + startup/shutdown)
- Keep middleware composable and transport-focused
- Keep architecture easy to read and refactor safely

Non-goals:

- Heavy framework abstractions
- Hidden magic for routing/DI/binding
- Dependency-heavy core module

---

## 2) Canonical Style Source (Mandatory)

`docs/CANONICAL_STYLE_GUIDE.md` is the canonical style authority for examples,
documentation snippets, templates, and AI-generated code.

Prefer these defaults:

- Handler: `func(w http.ResponseWriter, r *http.Request)`
- Route registration: explicit method + path + handler
- Request decoding: explicit JSON decoding in handler
- Error path: consistent structured errors via `contract.WriteError`
- Dependency flow: constructor injection, not request-context service lookup
- Middleware: `func(http.Handler) http.Handler`, no business logic

If multiple APIs exist, choose canonical style unless the task explicitly requires
compatibility behavior.

---

## 3) High-Value Boundaries

- `core/`: lifecycle/configuration/startup
- `router/`: routing behavior only
- `middleware/`: transport concerns only
- `contract/`: context + structured errors + response helpers
- `security/`: verification and security-critical logic
- `store/`: persistence abstractions
- `tenant/`: multi-tenant config/policy/quota primitives

Do not blur these boundaries.

---

## 4) Required Checks Before Final Output

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Recommended when relevant:

```bash
go test -race ./...
```

---

## 5) Change Rules

Do:

- Make focused, reversible changes
- Preserve existing public behavior unless asked to change it
- Add/update tests near the changed behavior
- Update docs when changing API/config/security/default behavior

Do not:

- Add non-stdlib dependencies to the main module without strong reason
- Introduce new style variants when a canonical one exists
- Hide dependencies in request context lookup
- Put business logic into middleware
- Log secrets

---

## 6) Minimal Canonical Example

```go
app := core.New()

app.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
})
```

Use one handler family consistently in one tutorial/example flow.

---

## 7) Documentation Sync Targets

When behavior/config/API changes, sync:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `env.example`

---

Default strategy: prefer explicitness over convenience and consistency over novelty.
