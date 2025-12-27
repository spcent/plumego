# AGENTS.md — plumego (Best Practice)

This document defines **strict operational guidance** for automated coding agents
(Codex, Copilot, ChatGPT, etc.) working in this repository.

The goal is to preserve this project as a **small, explicit, production-grade Go server**
built on the **Go standard library only**, with predictable behavior and long-term maintainability.

---

## 1. Project overview

* Repository: `spcent/plumego`
* Language: Go (module-based; see `go.mod`)
* Entry point: `main.go`

Expected structure (verify locally):

* `handlers/` — HTTP handlers (request/response boundary only)
* `pkg/` — reusable internal packages (router, middleware, auth, storage, etc.)
* `docs/` — design notes, API contracts, operational docs
* `scripts/` — dev / build / release helpers
* `.github/workflows/` — CI workflows
* `env.example` — environment variable template

### Server purpose

This server provides a **minimal but production-ready web runtime** built exclusively on the Go standard library.

It is designed to support:

* REST-style APIs
* webhook receivers (verification + deduplication)
* lightweight WebSocket-based real-time services
* JWT-based authentication
* in-process pub-sub
* small local persistence for tokens, deduplication, and config

**Non-goals**:

* No full-stack framework
* No ORM
* No automatic schema generation
* No reflection-heavy magic
* No hidden background goroutines
* No implicit global state

---

## 2. Golden rules (non-negotiable)

1. **Minimal change radius**

   * Fix the smallest possible surface.
   * Do not refactor unrelated code to “clean things up”.

2. **Explicit behavior**

   * Control flow must be readable without tracing macros or codegen.
   * Avoid clever abstractions.

3. **No external dependencies**

   * Do not add third-party packages unless explicitly instructed.
   * `golang.org/x/*` counts as external unless approved.

4. **Behavior changes require tests + docs**

   * Any change affecting:

     * HTTP status codes
     * response payloads
     * auth requirements
     * routing
       must update tests and documentation.

5. **Security-first defaults**

   * Never log secrets, tokens, Authorization headers, cookies, or user identifiers.
   * Redact or hash where visibility is required.

6. **Idiomatic Go**

   * Error-first returns
   * Context-aware APIs
   * Small, cohesive packages
   * No panic for expected runtime conditions

---

## 3. Local setup & command discovery

Agents **must not assume** commands.

Required discovery process:

1. Inspect `Makefile`:

   ```sh
   cat Makefile
   ```
2. Inspect CI workflows:

   ```sh
   ls .github/workflows
   ```

Only then may agents run:

```sh
go test ./...
go test -race ./...    # if CI budget allows
go vet ./...
gofmt -w .
```

If canonical Make targets exist, prefer them:

* `make test`
* `make lint`
* `make fmt`
* `make run`

---

## 4. Code organization rules

### Handlers (`handlers/`)

Handlers must:

* Parse and validate input
* Call domain/service logic
* Map errors to HTTP responses

Handlers must NOT:

* Contain business logic
* Perform persistence directly
* Spawn goroutines without explicit lifecycle control

### Internal packages (`pkg/`)

* Packages must be cohesive and small
* Avoid circular imports
* Prefer concrete types internally
* Use interfaces only at boundaries

### Context usage

* All request-scoped operations must accept `context.Context`
* Respect cancellation and deadlines
* `context.Value` is allowed **only** for request-scoped metadata
  (e.g., request ID, auth claims)

---

## 5. Error handling & HTTP responses

* Do not panic for expected errors
* Wrap errors explicitly:

  ```go
  return fmt.Errorf("parse token: %w", err)
  ```
* Centralize HTTP error → status mapping where possible
* Ensure JSON error responses are consistent across handlers

---

## 6. Testing guidelines

* Prefer table-driven tests
* For HTTP:

  * Use `net/http/httptest`
  * Do not bind real ports in unit tests
* Integration tests (if any):

  * Must be clearly separated (e.g. build tags)

### When behavior changes

You must update:

* Tests (new or modified)
* Documentation:

  * new routes
  * env vars
  * auth requirements
  * scripts or Make targets

---

## 7. Security & safety checklist (mandatory before completion)

Before submitting changes, agents must verify:

* [ ] No secrets or PII logged
* [ ] Auth middleware applied where required
* [ ] JSON / query / path input validated
* [ ] Request size limits considered
* [ ] Timeouts set on servers and outbound calls
* [ ] No unsafe deserialization
* [ ] No shell execution without justification

---

## 8. Required change summary (agent output format)

Every change submission **must include**:

* **Summary**: What changed
* **Reason**: Why it was needed
* **How to test**: Exact commands
* **Risk**: What could break / compatibility notes

If routing or middleware changed, explicitly list:

* affected endpoints
* auth changes
* backward compatibility impact

---

## 9. Directory-specific rules

Subdirectories may define stricter rules via `AGENTS.override.md`.

Root rules always apply.
