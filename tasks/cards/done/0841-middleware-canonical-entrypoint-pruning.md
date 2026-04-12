# Card 0841

Priority: P1
State: done
Primary Module: middleware
Owned Files:
- `middleware/accesslog/accesslog.go`
- `middleware/cors/cors.go`
- `middleware/auth/auth.go`
- `middleware/ratelimit/limiter.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

Goal:
- Converge stable middleware packages on one canonical constructor shape per package: exported functions should return `middleware.Middleware` directly without parallel convenience wrappers.
- Remove legacy or duplicate entrypoints that preserve multiple ways to do the same transport behavior.

Problem:
- `middleware/accesslog` exports both `Middleware(...)` and `Logging(...)`, with `Logging(...)` explicitly described as the “old combined convenience shape”.
- `middleware/cors` exports `CORS(...)`, `CORSWithOptions(...)`, and `CORSWithOptionsFunc(...)`, which spreads one CORS behavior across three parallel entrypoints.
- `middleware/auth` still exports `SimpleAuth(...)` alongside the contract-driven `Authenticate(...)` / `Authorize(...)` path, which keeps a one-off auth style in the stable middleware root.
- `middleware/ratelimit` keeps both `(*RateLimiter).Middleware()` and `RateLimitMiddleware(...)`, which duplicates constructor ownership for the same behavior.
- The module README says middleware should favor narrow, explicit transport behavior, but these duplicate entrypoints preserve legacy convenience instead of one canonical path.

Scope:
- Decide the canonical constructor shape for each affected middleware package and delete parallel convenience wrappers.
- Update call sites and tests to use the canonical entrypoint only.
- Keep behavior and ordering semantics unchanged while collapsing duplicate APIs.
- Sync middleware docs and manifest examples to the reduced constructor surface.

Non-goals:
- Do not redesign middleware internals, error codes, or ordering semantics.
- Do not add compatibility aliases or deprecated wrappers.
- Do not move auth verification logic into middleware; keep it injected from security-owned implementations.
- Do not introduce a new umbrella middleware factory package.

Files:
- `middleware/accesslog/accesslog.go`
- `middleware/cors/cors.go`
- `middleware/auth/auth.go`
- `middleware/ratelimit/limiter.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go test -race -timeout 60s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Keep the middleware manifest and primer aligned on the rule that each middleware package exposes one canonical transport constructor, not multiple wrapper families.

Done Definition:
- Each affected middleware package has one clear exported constructor path for its behavior.
- Legacy convenience wrappers and duplicate exported entrypoints are removed with no residual call sites.
- Middleware tests and docs use the converged constructor names only.
- Middleware docs and manifest describe the same canonical constructor surface that the code implements.

Outcome:
- Completed. `accesslog` now converges on `Logging(...)`, `cors` now converges on `Middleware(CORSOptions)`, simple fixed-token auth moved behind `security/authn.StaticToken(...)` + `middleware/auth.Authenticate(...)`, and duplicate `RateLimitMiddleware(...)` has been removed.
