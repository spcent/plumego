# Troubleshooting

Common problems encountered when using Plumego, with their causes and fixes.

---

## Route Registration

### "route registration failed after Prepare"

**Symptom:** `app.Get(...)` returns a non-nil error with a message like
"router is frozen".

**Cause:** `app.Prepare()` freezes the router. Any `app.Get`, `app.Post`,
`app.Use`, or `app.Group` call after `Prepare` is rejected.

**Fix:** Register all routes and middleware before calling `app.Prepare()`.

```go
// correct order
app.Use(requestid.Middleware())
app.Get("/ping", pingHandler)
app.Prepare()           // freeze happens here
srv, _ := app.Server()
srv.ListenAndServe()
```

---

### Route not matching — 404 for a path I registered

**Causes (check in order):**

1. **Method mismatch.** `app.Get("/users", h)` will 404 on `POST /users`.
   Use `app.Post` or `app.Any` if multiple methods share a handler.

2. **Trailing slash.** `/users` and `/users/` are distinct routes. Register
   both if you need both, or redirect in a middleware.

3. **Param syntax.** Path parameters use `{param}` notation in some contexts
   and `:param` in others — confirm which syntax your router version accepts
   by checking `docs/modules/router/README.md`.

4. **Group prefix collision.** A group registered with `/api/v1` and a route
   registered with `/api/v1/users` both produce `/api/v1/api/v1/users` if
   the group prefix is repeated in the route path.

---

### "route already registered" panic or error

**Cause:** The same `method + path` combination was registered twice, usually
when an `init()`-style helper silently registers routes.

**Fix:** Search for all registration sites with `rg -n 'app\.(Get|Post|Put|Delete|Patch|Any|AddRoute)'`
and remove duplicates. Plumego has no hidden auto-registration; every route
must appear in your explicit wiring.

---

## Middleware

### Middleware executes in unexpected order

Middleware runs in registration order. The first `app.Use(A)` call wraps
outermost; the last wraps innermost.

```
app.Use(A)  // outermost — runs first on request, last on response
app.Use(B)
app.Use(C)  // innermost — runs last on request, first on response
```

Recovery middleware should be registered before (outer to) logging middleware
so that panics are caught before the access log entry is written.

---

### Recovery middleware does not catch panics from a goroutine

Recovery only catches panics on the goroutine that calls `ServeHTTP`. A panic
in a goroutine spawned inside a handler is not caught. Handle those with
`defer/recover` inside the goroutine or use a supervised worker pattern.

---

### CORS preflight returns 404

`cors.Middleware` must be registered before route matching. If the CORS
middleware is only applied to a route group, OPTIONS requests that do not
match any group prefix will fall through to the default 404 handler.

**Fix:** Register CORS middleware at the top-level `app.Use` call, not inside
a group.

---

### `contract.WriteError` response body is empty

**Cause:** `w.WriteHeader(statusCode)` was already called before
`contract.WriteError`. Once the header is written, the body written by
`contract.WriteError` is discarded by `http.ResponseWriter`.

**Fix:** Do not call `w.WriteHeader` directly before delegating to
`contract.WriteError`. Let `contract.WriteError` set the status code.

---

## Response Helpers

### JSON response is sent but status code is 200 instead of expected code

`contract.WriteResponse(w, r, http.StatusCreated, body, nil)` sets the
status code only if it has not already been written. If any other code path
called `w.WriteHeader` or `w.Write` before this, the status code is already
locked at 200.

---

### `contract.WriteResponse` and a direct `json.NewEncoder` both run

Only call one response writer per handler execution path. Two writes produce
corrupted JSON. Use early returns after each write:

```go
if err := doSomething(); err != nil {
    _ = contract.WriteError(w, r, appErr)
    return   // <-- essential
}
_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
```

---

## Path Parameters

### `router.Param(r, "id")` returns empty string

**Causes:**

1. The route was not matched through Plumego's router (e.g. the handler was
   registered directly on `http.ServeMux`). Parameters are only set by
   Plumego's router.

2. The parameter name in the route pattern and the name passed to `Param`
   differ. `/users/:userId` must be read as `router.Param(r, "userId")`, not
   `"id"`.

3. The request reached the handler via a test that creates the request
   directly with `httptest.NewRequest` without going through the router. In
   tests, use `httptest.NewServer` with the full `app` handler to ensure
   routing runs.

---

## Security

### JWT verification always returns an error

**Causes (check in order):**

1. **Algorithm mismatch.** The signing algorithm used to create the token
   must match the algorithm passed to the verifier. HS256 tokens cannot be
   verified with RS256 config.

2. **Key type mismatch.** HMAC algorithms (`HS*`) require a `[]byte` key.
   RSA algorithms (`RS*`, `PS*`) require `*rsa.PrivateKey` / `*rsa.PublicKey`.
   EC algorithms (`ES*`) require `*ecdsa.PrivateKey` / `*ecdsa.PublicKey`.

3. **Clock skew.** Tokens with a strict `exp` claim will fail verification
   if the server clock is ahead of the issuing clock. Allow a small leeway or
   synchronise clocks.

4. **Token prefix.** The `auth` middleware extracts the token from the
   `Authorization: Bearer <token>` header. Pass only the token value to
   `security.VerifyToken`, not the `Bearer ` prefix.

---

### bcrypt comparison is slow in tests

`security.HashPassword` uses a configurable cost. Default cost is suitable
for production but can slow test suites significantly. Use
`security.HashPasswordWithCost(password, bcrypt.MinCost)` in test helpers.

---

## Health and Readiness

### Health check always returns degraded

The `health` package owns models and the `ComponentChecker` interface, but
does not own the HTTP handler. Your handler must call each registered
`ComponentChecker.Check(ctx)` and aggregate results. If a checker always
returns `Degraded`, investigate the checker implementation, not the
`health` package.

---

## Lifecycle

### Graceful shutdown does not drain in-flight requests

`app.Shutdown(ctx)` calls `(*http.Server).Shutdown(ctx)`. The timeout is
controlled by the context you pass. Use a context with a deadline:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := app.Shutdown(shutdownCtx); err != nil {
    log.Printf("shutdown: %v", err)
}
```

If requests are still active after the deadline, they are forcibly closed.

---

### `app.Server()` returns an error after `Prepare()` succeeds

`Server()` constructs the `*http.Server` from the config supplied to
`core.New`. Common causes: an invalid `Addr` string (e.g. missing colon),
or a TLS config that references a certificate file that does not exist.

---

## Extensions (x/*)

### An `x/*` package cannot be imported from a stable root

This is a hard boundary enforced by `internal/checks/dependency-rules`.
Stable roots (`core`, `router`, `contract`, `middleware`, `security`,
`store`, `health`, `log`, `metrics`) must not import any `x/*` package.

Move the dependency-bearing code to the calling application or into the
appropriate `x/*` family. See `docs/concepts/core-boundary.md`.

---

### `x/tenant` middleware is not isolating requests

Tenant isolation requires the tenant resolver to run before any
tenant-scoped handler. Ensure `x/tenant` resolution middleware is registered
before (outer to) the handler that reads tenant context. Verify with a test
that routes two requests with different tenant identifiers and confirms
different tenant context values.

---

## Testing

### `httptest.NewRecorder` body is empty after handler call

`contract.WriteResponse` writes JSON to `w`. If the recorder body is empty,
the handler returned before writing (usually due to an early-return error
path). Add `t.Log(w.Body.String())` to see what was actually written, or
check that the handler's error paths all write a response before returning.

---

### Tests pass locally but fail in CI with race detector

The race detector is enabled in CI (`go test -race`). Common causes:

- A package-level variable written in one test and read in another.
- A goroutine started in a handler that outlives the request context.
- An `httptest.Server` that is not closed after each test.

Use `t.Cleanup(func() { server.Close() })` to close test servers and avoid
resource leaks between tests.

---

## Getting More Help

- Package-specific behaviour: `docs/modules/<package>/README.md`
- Canonical app layout: `reference/standard-service`
- Style and wiring conventions: `docs/reference/canonical-style-guide.md`
- When Plumego is not the right fit: `docs/start/when-not-to-use-plumego.md`
