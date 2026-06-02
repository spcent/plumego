# AGENTS.md - use-cases/dbadmin

Local operating guide for agents working on the `dbadmin` use-case app.

## Scope

`dbadmin` is an independent Plumego use-case application with its own Go module,
frontend package, Docker development stack, and release artifacts. It is a
production-scale internal-tool example, not a canonical template.

## Boundaries

- Keep application code under `use-cases/dbadmin`.
- Do not move shared framework behavior into Plumego stable roots unless the
  change is explicitly requested and independently justified.
- Preserve the app-local `go.mod` with `replace github.com/spcent/plumego => ../..`.
- Do not add new dependencies without approval.
- Treat database credentials, connection URIs, API keys, session tokens, and SQL
  text from user input as sensitive. Never log secrets.
- Keep dangerous operations fail-closed: readonly checks and explicit
  confirmation must be enforced by backend handlers, not only the UI.

## Default Read Path

1. `README.md`
2. `docs/regression-matrix.md`
3. Target handler/domain files under `internal/`
4. Frontend API and page files under `web/src/`
5. Docker/demo docs only when changing onboarding or local verification

## Validation

For backend changes:

```bash
cd use-cases/dbadmin
gofmt -w .
go test ./...
go vet ./...
```

For frontend changes:

```bash
cd use-cases/dbadmin/web
npm test
npm run build
```

For docs-only changes, skip Go and frontend gates unless examples, config, or
generated artifacts changed.
