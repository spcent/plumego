# Testing `core.App`

> **Package**: `github.com/spcent/plumego/core`

Canonical testing patterns for v1.

---

## Principles

- Prefer `httptest` and `app.ServeHTTP` for route and middleware tests.
- Keep tests deterministic and local (no real network unless required).
- Register middleware/routes before serving requests.
- Use `go test -race` for concurrency-sensitive behavior.

---

## Basic Route Test

```go
func TestHealthRoute(t *testing.T) {
    app := core.New()
    app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
    if strings.TrimSpace(rec.Body.String()) != "ok" {
        t.Fatalf("unexpected body: %q", rec.Body.String())
    }
}
```

---

## Path Parameter Test

```go
func TestPathParam(t *testing.T) {
    app := core.New()
    app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        id := plumego.Param(r, "id")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(id))
    })

    req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
    if strings.TrimSpace(rec.Body.String()) != "42" {
        t.Fatalf("unexpected id: %q", rec.Body.String())
    }
}
```

---

## Middleware Order Test

```go
func TestMiddlewareOrder(t *testing.T) {
    app := core.New()

    var calls []string
    mu := sync.Mutex{}

    mw := func(name string) middleware.Middleware {
        return func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                mu.Lock()
                calls = append(calls, name+":before")
                mu.Unlock()

                next.ServeHTTP(w, r)

                mu.Lock()
                calls = append(calls, name+":after")
                mu.Unlock()
            })
        }
    }

    if err := app.Use(mw("m1"), mw("m2")); err != nil {
        t.Fatalf("register middleware: %v", err)
    }

    app.Get("/x", func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        calls = append(calls, "handler")
        mu.Unlock()
        w.WriteHeader(http.StatusOK)
    })

    req := httptest.NewRequest(http.MethodGet, "/x", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    got := strings.Join(calls, ",")
    want := "m1:before,m2:before,handler,m2:after,m1:after"
    if got != want {
        t.Fatalf("order mismatch\nwant: %s\n got: %s", want, got)
    }
}
```

---

## Group Middleware Test (Router-Level)

```go
func TestGroupMiddleware(t *testing.T) {
    app := core.New()

    api := app.Router().Group("/api")
    api.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-Group", "api")
            next.ServeHTTP(w, r)
        })
    })
    api.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
    if rec.Header().Get("X-Group") != "api" {
        t.Fatalf("missing group middleware header")
    }
}
```

---

## Notes on `Boot()` in Tests

`Boot()` starts a real server and blocks until shutdown. Most HTTP behavior should be tested via `ServeHTTP` instead.

Use `Boot()` only when validating lifecycle wiring (components/runners/signals) and keep those tests isolated.

---

## Quality Gates

Run these before merging core changes:

```bash
go test -timeout 20s ./...
go test -race -timeout 60s ./...
go vet ./...
```
