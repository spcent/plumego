# Card 0145

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`
Depends On:
- `0144-core-handler-preparation-state-convergence.md`

Goal:
- Remove the leftover private server-start path so `core` has one serving model
  after the earlier `Boot`/`Run` removal.

Problem:
- The public first-party lifecycle is `Prepare` + `Start` + `Server`, with
  callers invoking `srv.ListenAndServe()` themselves.
- `core/lifecycle.go` still keeps a second private serve path in `startServer`
  plus `serverStartConfig`, startup logging branches, TLS validation, and dead
  runtime placeholders such as `stopRuntime` and `signalHandling`.
- That private path is now exercised only by tests, so `core` still maintains a
  second internal serving model that production callers do not use.

Scope:
- Remove `startServer` and any dead supporting state/config that exists only
  for that path.
- Move any still-needed validation or logging to the canonical explicit
  lifecycle path instead of keeping a hidden alternative.
- Update lifecycle tests to serve through the same public path used by
  first-party applications.

Non-goals:
- Do not reintroduce `Run`/`Boot`-style convenience APIs.
- Do not change HTTP shutdown semantics or drain behaviour in this card.
- Do not redesign TLS configuration itself.

Files:
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`

Tests:
- Remove direct test coverage of `startServer` and assert serving behaviour via
  `Prepare` + `Start` + `Server`.
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Keep the core primer aligned on the single explicit serve path.

Done Definition:
- `core` has no private alternate serve path competing with `Server()`.
- Dead lifecycle/runtime placeholders tied to that path are removed.
- Lifecycle tests and docs describe one serving model only.

Outcome:
- Removed `startServer`, `serverStartConfig`, `stopRuntime`, and the unused
  `signalHandling` runtime shadow state so `Server()` is the only serve path.
- Moved TLS certificate loading into the canonical preparation path, so callers
  keep using `Prepare` + `Start` + `Server` and choose
  `ListenAndServe()`/`ListenAndServeTLS("", "")` on the returned server.
- Updated lifecycle tests, reference apps, scaffold output, and core/root docs
  to describe and exercise the same explicit serving flow.
