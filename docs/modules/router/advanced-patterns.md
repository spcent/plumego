# Advanced Routing Patterns

> **Package**: `github.com/spcent/plumego/router`

This document covers advanced routing techniques, custom matchers, constraints, and sophisticated URL patterns.

---

## Table of Contents

- [Subdomain Routing](#subdomain-routing)
- [Method Override](#method-override)
- [Content Negotiation](#content-negotiation)
- [Conditional Routing](#conditional-routing)
- [Route Constraints](#route-constraints)
- [Catch-All Routes](#catch-all-routes)
- [Fallback Routes](#fallback-routes)

---

## Subdomain Routing

Route requests based on subdomain.

### Subdomain Middleware

```go
func subdomainRouter(routes map[string]http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        host := r.Host
        parts := strings.Split(host, ".")

        if len(parts) < 2 {
            http.Error(w, "Invalid host", http.StatusBadRequest)
            return
        }

        subdomain := parts[0]

        handler, ok := routes[subdomain]
        if !ok {
            http.Error(w, "Unknown subdomain", http.StatusNotFound)
            return
        }

        handler.ServeHTTP(w, r)
    })
}
```

### Usage

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
)

func main() {
    // Create apps for different subdomains
    api := core.New()
    api.Get("/users", apiListUsers)

    admin := core.New()
    admin.Get("/dashboard", adminDashboard)

    www := core.New()
    www.Get("/", homepage)

    // Main app with subdomain routing
    app := core.New()
    app.Use(subdomainRouter(map[string]http.Handler{
        "api":   api.Router(),
        "admin": admin.Router(),
        "www":   www.Router(),
    }))

    app.Boot()
}
```

### Dynamic Subdomains

```go
func tenantRouter(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        host := r.Host
        parts := strings.Split(host, ".")

        // Extract tenant from subdomain
        tenant := parts[0]

        // Add to context
        ctx := context.WithValue(r.Context(), "tenant", tenant)
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    })
}

// Usage
app := core.New()
app.Use(tenantRouter)

app.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
    tenant := r.Context().Value("tenant").(string)
    fmt.Fprintf(w, "Dashboard for tenant: %s", tenant)
})
```

---

## Method Override

Support method override for clients that can't send PUT/DELETE.

### Middleware

```go
func methodOverride(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            // Check _method parameter
            method := r.FormValue("_method")
            if method == "" {
                // Check X-HTTP-Method-Override header
                method = r.Header.Get("X-HTTP-Method-Override")
            }

            // Override method
            if method != "" {
                r.Method = strings.ToUpper(method)
            }
        }

        next.ServeHTTP(w, r)
    })
}
```

### Usage

```go
app := core.New()
app.Use(methodOverride)

// These can now be called via POST with _method parameter
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)

// HTML form:
// <form method="POST" action="/users/123?_method=PUT">
//   ...
// </form>

// Or with header:
// POST /users/123
// X-HTTP-Method-Override: PUT
```

---

## Content Negotiation

Route based on Accept header.

### Accept Header Routing

```go
func contentNegotiationRouter(handlers map[string]http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        accept := r.Header.Get("Accept")

        // Try exact match
        if handler, ok := handlers[accept]; ok {
            handler(w, r)
            return
        }

        // Try prefix match
        for contentType, handler := range handlers {
            if strings.Contains(accept, contentType) {
                handler(w, r)
                return
            }
        }

        // Default to first handler
        for _, handler := range handlers {
            handler(w, r)
            return
        }

        http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
    }
}
```

### Usage

```go
app := core.New()

app.Get("/users", contentNegotiationRouter(map[string]http.HandlerFunc{
    "application/json": func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(users)
    },
    "application/xml": func(w http.ResponseWriter, r *http.Request) {
        xml.NewEncoder(w).Encode(users)
    },
    "text/html": func(w http.ResponseWriter, r *http.Request) {
        tmpl.Execute(w, users)
    },
}))
```

---

## Conditional Routing

Route based on custom conditions.

### Feature Flag Routing

```go
func featureFlag(flagName string, enabled, disabled http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if isFeatureEnabled(flagName) {
            enabled.ServeHTTP(w, r)
        } else {
            disabled.ServeHTTP(w, r)
        }
    })
}

// Usage
app.Get("/new-dashboard",
    featureFlag("new-dashboard",
        http.HandlerFunc(newDashboard),
        http.HandlerFunc(oldDashboard),
    ),
)
```

### A/B Testing Router

```go
func abTest(testName string, variants map[string]http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get user's variant (from cookie, database, etc.)
        variant := getUserVariant(testName, r)

        handler, ok := variants[variant]
        if !ok {
            // Default to first variant
            for _, h := range variants {
                handler = h
                break
            }
        }

        handler.ServeHTTP(w, r)
    })
}

// Usage
app.Get("/landing",
    abTest("landing-page", map[string]http.Handler{
        "control": http.HandlerFunc(landingV1),
        "variant-a": http.HandlerFunc(landingV2),
        "variant-b": http.HandlerFunc(landingV3),
    }),
)
```

### Time-Based Routing

```go
func timeBasedRouter(schedule map[string]http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        now := time.Now()
        hour := now.Hour()

        var handler http.Handler

        // Match time range
        switch {
        case hour >= 0 && hour < 6:
            handler = schedule["night"]
        case hour >= 6 && hour < 12:
            handler = schedule["morning"]
        case hour >= 12 && hour < 18:
            handler = schedule["afternoon"]
        default:
            handler = schedule["evening"]
        }

        if handler == nil {
            handler = schedule["default"]
        }

        handler.ServeHTTP(w, r)
    })
}

// Usage
app.Get("/maintenance",
    timeBasedRouter(map[string]http.Handler{
        "night": http.HandlerFunc(maintenanceMode),
        "default": http.HandlerFunc(normalMode),
    }),
)
```

---

## Route Constraints

Add validation to route parameters.

### Integer Constraint

```go
func intConstraint(min, max int) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            idStr := plumego.Param(r, "id")

            id, err := strconv.Atoi(idStr)
            if err != nil || id < min || id > max {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
app.Get("/pages/:id",
    intConstraint(1, 1000)(http.HandlerFunc(getPage)),
)
```

### Regex Constraint

```go
func regexConstraint(param string, pattern *regexp.Regexp) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            value := plumego.Param(r, param)

            if !pattern.MatchString(value) {
                http.Error(w, "Invalid parameter format", http.StatusBadRequest)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

app.Get("/resources/:uuid",
    regexConstraint("uuid", uuidPattern)(http.HandlerFunc(getResource)),
)
```

### Enum Constraint

```go
func enumConstraint(param string, values []string) func(http.Handler) http.Handler {
    validValues := make(map[string]bool)
    for _, v := range values {
        validValues[v] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            value := plumego.Param(r, param)

            if !validValues[value] {
                http.Error(w, "Invalid value", http.StatusBadRequest)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
app.Get("/users/status/:status",
    enumConstraint("status", []string{"active", "inactive", "pending"})(
        http.HandlerFunc(getUsersByStatus),
    ),
)
```

---

## Catch-All Routes

Handle unmatched routes.

### Custom 404 Handler

```go
app := core.New(
    core.WithRouter(router.New(
        router.WithNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Route not found",
                "path":  r.URL.Path,
            })
        })),
    )),
)
```

### SPA Fallback

```go
app := core.New()

// API routes
api := app.Group("/api")
api.Get("/users", listUsers)
api.Get("/products", listProducts)

// Catch-all for SPA (serve index.html)
app.Get("/*path", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./public/index.html")
})
```

### Proxy Fallback

```go
func proxyFallback(target *url.URL) http.Handler {
    proxy := httputil.NewSingleHostReverseProxy(target)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        proxy.ServeHTTP(w, r)
    })
}

app := core.New()

// Local routes
app.Get("/health", healthCheck)

// Proxy everything else
target, _ := url.Parse("http://backend:8080")
app.Get("/*path", proxyFallback(target))
```

---

## Fallback Routes

Provide fallbacks for different scenarios.

### Language Fallback

```go
func languageFallback(handlers map[string]http.Handler, defaultLang string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get language from Accept-Language header
        lang := parseAcceptLanguage(r.Header.Get("Accept-Language"))

        handler, ok := handlers[lang]
        if !ok {
            handler = handlers[defaultLang]
        }

        handler.ServeHTTP(w, r)
    })
}

// Usage
app.Get("/",
    languageFallback(map[string]http.Handler{
        "en": http.HandlerFunc(homeEN),
        "zh": http.HandlerFunc(homeZH),
        "es": http.HandlerFunc(homeES),
    }, "en"),
)
```

### Version Fallback

```go
func versionFallback(versions map[string]http.Handler, latest string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get version from header or query
        version := r.Header.Get("API-Version")
        if version == "" {
            version = r.URL.Query().Get("version")
        }
        if version == "" {
            version = latest
        }

        handler, ok := versions[version]
        if !ok {
            http.Error(w, "Unsupported API version", http.StatusBadRequest)
            return
        }

        handler.ServeHTTP(w, r)
    })
}

// Usage
app.Get("/api/users",
    versionFallback(map[string]http.Handler{
        "v1": http.HandlerFunc(listUsersV1),
        "v2": http.HandlerFunc(listUsersV2),
        "v3": http.HandlerFunc(listUsersV3),
    }, "v3"),
)
```

---

## Complete Example

### Advanced API Gateway

```go
package main

import (
    "net/http"
    "regexp"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/router"
)

func main() {
    app := core.New(
        core.WithRouter(router.New(
            router.WithNotFoundHandler(customNotFound),
        )),
    )

    // Health checks (no constraints)
    app.Get("/health", healthCheck)

    // API v1
    v1 := app.Group("/api/v1")

    // Users (with integer ID constraint)
    v1.Get("/users/:id",
        intConstraint(1, 999999)(http.HandlerFunc(getUserV1)),
    )

    // Resources (with UUID constraint)
    uuidPattern := regexp.MustCompile(`^[0-9a-f-]{36}$`)
    v1.Get("/resources/:uuid",
        regexConstraint("uuid", uuidPattern)(http.HandlerFunc(getResource)),
    )

    // Status (with enum constraint)
    v1.Get("/status/:type",
        enumConstraint("type", []string{"health", "metrics", "debug"})(
            http.HandlerFunc(getStatus),
        ),
    )

    // API v2 (content negotiation)
    v2 := app.Group("/api/v2")
    v2.Get("/users", contentNegotiationRouter(map[string]http.HandlerFunc{
        "application/json": listUsersJSON,
        "application/xml":  listUsersXML,
    }))

    // SPA fallback
    app.Get("/*path", spaFallback)

    app.Boot()
}

func customNotFound(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Not found",
        "path":  r.URL.Path,
    })
}

func spaFallback(w http.ResponseWriter, r *http.Request) {
    // Only serve SPA for non-API routes
    if !strings.HasPrefix(r.URL.Path, "/api") {
        http.ServeFile(w, r, "./public/index.html")
        return
    }

    http.NotFound(w, r)
}
```

---

## Best Practices

### ✅ Do

1. **Use Constraints for Validation**
   ```go
   app.Get("/users/:id",
       intConstraint(1, 999999)(handler),
   )
   ```

2. **Provide Clear Error Messages**
   ```go
   if !valid {
       http.Error(w, "Invalid UUID format", http.StatusBadRequest)
       return
   }
   ```

3. **Document Advanced Patterns**
   ```go
   // Route supports method override via _method parameter
   app.Put("/users/:id", updateUser)
   ```

### ❌ Don't

1. **Don't Overcomplicate**
   ```go
   // ❌ Too complex
   app.Get("/route",
       middleware1(
           middleware2(
               middleware3(
                   middleware4(handler),
               ),
           ),
       ),
   )

   // ✅ Use groups
   group := app.Group("/api", middleware1, middleware2, middleware3, middleware4)
   group.Get("/route", handler)
   ```

2. **Don't Ignore Performance**
   ```go
   // ❌ Expensive regex on every request
   app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
       pattern := regexp.MustCompile("...") // Compiled every request!
   })

   // ✅ Compile once
   var userPattern = regexp.MustCompile("...")
   ```

---

## Next Steps

- **[Router Overview](README.md)** - Core routing concepts
- **[Middleware](../middleware/)** - Built-in middleware
- **[Performance Guide](../../guides/performance.md)** - Optimization

---

**Related**:
- [Router Overview](README.md)
- [Middleware Binding](middleware-binding.md)
- [Trie Router](trie-router.md)
