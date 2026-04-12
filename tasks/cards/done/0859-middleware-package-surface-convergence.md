# Card 0859

Priority: P1
State: done
Primary Module: middleware
Owned Files:
- `middleware/accesslog`
- `middleware/limits`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `reference/standard-service`
- `cmd/plumego/internal/scaffold`

Goal:
- Converge stable middleware packages on one clear responsibility and one canonical constructor path per package.
- Remove leftover package shapes that bundle unrelated transport behaviors or use non-canonical constructor naming.

Problem:
- `middleware/accesslog` still exposes `Logging(...)` instead of the repo-wide `Middleware(...)` constructor convention used by the rest of the stable middleware surface.
- `middleware/limits` currently owns two unrelated behaviors in one package: request body size enforcement and concurrency throttling.
- This leaves stable middleware with package-level responsibility drift and inconsistent constructor naming even after earlier surface cleanup.

Scope:
- Rename the canonical `accesslog` constructor to a single middleware-style entrypoint and remove the old name in the same change.
- Split `middleware/limits` into one package per middleware responsibility, or otherwise reduce it to one canonical middleware per package without leaving wrappers behind.
- Update reference apps, scaffolding, docs, and tests in the same change.

Non-goals:
- Do not redesign the access log field set or error schema.
- Do not change the actual body-limit or concurrency-limit semantics unless required by the package split.
- Do not move transport throttling into `security` or another non-middleware stable root.

Files:
- `middleware/accesslog`
- `middleware/limits`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `reference/standard-service`
- `cmd/plumego/internal/scaffold`

Tests:
- `go test -timeout 20s ./middleware/... ./reference/... ./cmd/plumego/internal/scaffold/...`
- `go test -race -timeout 60s ./middleware/...`
- `go vet ./middleware/... ./reference/... ./cmd/plumego/internal/scaffold/...`

Docs Sync:
- Keep middleware docs aligned on the rule that each stable middleware package has one transport responsibility and one canonical constructor path.

Done Definition:
- `middleware/accesslog` uses the canonical constructor naming used across stable middleware.
- The old `limits` umbrella shape is gone; each remaining stable middleware package owns one clear transport behavior.
- Reference apps and scaffold output only use the new canonical middleware package surfaces.
