# Testing Plumego Applications

> **Package**: `github.com/spcent/plumego/core`

This document covers testing patterns, utilities, and best practices for Plumego applications.

---

## Table of Contents

- [Overview](#overview)
- [Testing Applications](#testing-applications)
- [Testing Components](#testing-components)
- [Testing Runners](#testing-runners)
- [Testing Handlers](#testing-handlers)
- [Testing Middleware](#testing-middleware)
- [Integration Testing](#integration-testing)
- [Best Practices](#best-practices)

---

## Overview

### Testing Philosophy

Plumego is designed for testability:
- **No global state**: All dependencies are explicit
- **Dependency injection**: Easy to mock dependencies
- **Minimal setup**: Quick test environment creation
- **Standard library**: Use standard `testing` package

### Test Types

1. **Unit Tests**: Test individual functions and methods
2. **Component Tests**: Test components in isolation
3. **Integration Tests**: Test complete application
4. **Handler Tests**: Test HTTP handlers
5. **Middleware Tests**: Test middleware behavior

---

## Testing Applications

### Basic Application Test

```go
package main

import (
    "context"
    "net/http"
    "testing"
    "time"

    "github.com/spcent/plumego/core"
)

func TestApplication(t *testing.T) {
    // Create test app
    app := core.New(
        core.WithAddr(":0"), // Random port
    )

    app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    // Start in background
    go func() {
        if err := app.Boot(); err != nil {
            t.Errorf("Boot failed: %v", err)
        }
    }()

    // Wait for server to start
    time.Sleep(100 * time.Millisecond)

    // Make request
    resp, err := http.Get("http://" + app.Addr() + "/ping")
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    // Verify response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }

    // Shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := app.Shutdown(ctx); err != nil {
        t.Errorf("Shutdown failed: %v", err)
    }
}
```

### Test Helper

```go
package testutil

import (
    "context"
    "net/http"
    "testing"
    "time"

    "github.com/spcent/plumego/core"
)

type TestApp struct {
    *core.App
    t *testing.T
}

func NewTestApp(t *testing.T, opts ...core.Option) *TestApp {
    // Add test defaults
    testOpts := []core.Option{
        core.WithAddr(":0"), // Random port
        core.WithServerTimeouts(
            5*time.Second,  // read
            5*time.Second,  // write
            5*time.Second,  // idle
            1*time.Second,  // shutdown
        ),
    }

    app := core.New(append(testOpts, opts...)...)

    return &TestApp{
        App: app,
        t:   t,
    }
}

func (ta *TestApp) Start() {
    go func() {
        if err := ta.Boot(); err != nil {
            ta.t.Errorf("Boot failed: %v", err)
        }
    }()

    // Wait for server
    time.Sleep(100 * time.Millisecond)
}

func (ta *TestApp) Stop() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := ta.Shutdown(ctx); err != nil {
        ta.t.Errorf("Shutdown failed: %v", err)
    }
}

func (ta *TestApp) Get(path string) (*http.Response, error) {
    return http.Get("http://" + ta.Addr() + path)
}

func (ta *TestApp) Post(path, contentType string, body io.Reader) (*http.Response, error) {
    return http.Post("http://"+ta.Addr()+path, contentType, body)
}

// Usage
func TestWithHelper(t *testing.T) {
    app := NewTestApp(t)
    app.App.Get("/test", handler)

    app.Start()
    defer app.Stop()

    resp, err := app.Get("/test")
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    // Assertions...
}
```

---

## Testing Components

### Basic Component Test

```go
package mycomponent

import (
    "context"
    "testing"
)

func TestComponent(t *testing.T) {
    comp := &Component{
        Config: "test-config",
    }

    // Test Start
    ctx := context.Background()
    if err := comp.Start(ctx); err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // Test functionality
    result := comp.DoSomething()
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }

    // Test Stop
    if err := comp.Stop(ctx); err != nil {
        t.Fatalf("Stop failed: %v", err)
    }
}
```

### Component with Mock Dependencies

```go
package mycomponent

import (
    "context"
    "testing"

    "github.com/spcent/plumego/core/di"
)

// Mock dependency
type MockDatabase struct {
    queryResult string
    queryError  error
}

func (m *MockDatabase) Query(sql string) (string, error) {
    return m.queryResult, m.queryError
}

func TestComponentWithMock(t *testing.T) {
    // Create DI container
    container := di.New()

    // Register mock
    mockDB := &MockDatabase{
        queryResult: "test-result",
    }
    container.Register(mockDB)

    // Create component
    comp := &Component{}

    // Resolve mock dependency
    var db *MockDatabase
    if err := container.Resolve(&db); err != nil {
        t.Fatalf("Failed to resolve: %v", err)
    }
    comp.db = db

    // Test component with mock
    result := comp.FetchData()
    if result != "test-result" {
        t.Errorf("Expected 'test-result', got '%s'", result)
    }
}
```

### Component Lifecycle Test

```go
func TestComponentLifecycle(t *testing.T) {
    comp := &Component{}

    // Test initial state
    if comp.started {
        t.Error("Component should not be started initially")
    }

    // Test Start
    ctx := context.Background()
    if err := comp.Start(ctx); err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    if !comp.started {
        t.Error("Component should be started after Start()")
    }

    // Test Health
    name, status := comp.Health()
    if name != "mycomponent" {
        t.Errorf("Expected name 'mycomponent', got '%s'", name)
    }
    if status != health.StatusUp {
        t.Errorf("Expected StatusUp, got %v", status)
    }

    // Test Stop
    if err := comp.Stop(ctx); err != nil {
        t.Fatalf("Stop failed: %v", err)
    }

    if comp.started {
        t.Error("Component should not be started after Stop()")
    }
}
```

---

## Testing Runners

### Basic Runner Test

```go
package myrunner

import (
    "context"
    "testing"
    "time"
)

func TestRunner(t *testing.T) {
    executed := false

    runner := &Worker{
        interval: 100 * time.Millisecond,
        task: func() {
            executed = true
        },
    }

    // Start runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // Wait for task execution
    time.Sleep(200 * time.Millisecond)

    // Verify task executed
    if !executed {
        t.Error("Task should have been executed")
    }

    // Stop runner
    stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    if err := runner.Stop(stopCtx); err != nil {
        t.Fatalf("Stop failed: %v", err)
    }
}
```

### Runner with Channels

```go
func TestRunnerWithChannels(t *testing.T) {
    events := make(chan string, 10)

    runner := &EventWorker{
        interval: 50 * time.Millisecond,
        events:   events,
    }

    ctx := context.Background()
    runner.Start(ctx)

    // Wait for events
    select {
    case event := <-events:
        if event != "expected-event" {
            t.Errorf("Expected 'expected-event', got '%s'", event)
        }
    case <-time.After(200 * time.Millisecond):
        t.Error("Timeout waiting for event")
    }

    runner.Stop(ctx)
}
```

---

## Testing Handlers

### HTTP Handler Test

```go
package handlers

import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestPingHandler(t *testing.T) {
    // Create request
    req := httptest.NewRequest(http.MethodGet, "/ping", nil)

    // Create response recorder
    w := httptest.NewRecorder()

    // Call handler
    PingHandler(w, req)

    // Get response
    resp := w.Result()
    defer resp.Body.Close()

    // Verify status
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }

    // Verify body
    body, _ := io.ReadAll(resp.Body)
    if string(body) != "pong" {
        t.Errorf("Expected 'pong', got '%s'", string(body))
    }
}
```

### Context Handler Test

```go
import (
    "github.com/spcent/plumego"
)

func TestContextHandler(t *testing.T) {
    // Create request
    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

    // Add path parameters
    req = plumego.WithParam(req, "id", "123")

    // Create response recorder
    w := httptest.NewRecorder()

    // Create context
    ctx := plumego.NewContext(w, req)

    // Call handler
    GetUserHandler(ctx)

    // Verify response
    if ctx.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", ctx.StatusCode)
    }
}
```

### Handler with Dependencies

```go
type UserHandler struct {
    db UserRepository
}

func TestUserHandler(t *testing.T) {
    // Mock repository
    mockDB := &MockUserRepository{
        users: map[string]User{
            "123": {ID: "123", Name: "Alice"},
        },
    }

    // Create handler with mock
    handler := &UserHandler{db: mockDB}

    // Test
    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
    req = plumego.WithParam(req, "id", "123")
    w := httptest.NewRecorder()

    handler.GetUser(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

---

## Testing Middleware

### Basic Middleware Test

```go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestLoggingMiddleware(t *testing.T) {
    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Wrap with middleware
    middleware := LoggingMiddleware(handler)

    // Create request
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()

    // Call middleware
    middleware.ServeHTTP(w, req)

    // Verify response
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }

    // Verify logging occurred (check logs or metrics)
}
```

### Middleware with Context

```go
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name       string
        token      string
        wantStatus int
    }{
        {
            name:       "valid token",
            token:      "valid-token",
            wantStatus: http.StatusOK,
        },
        {
            name:       "invalid token",
            token:      "invalid-token",
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "missing token",
            token:      "",
            wantStatus: http.StatusUnauthorized,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            })

            middleware := AuthMiddleware(handler)

            req := httptest.NewRequest(http.MethodGet, "/", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", "Bearer "+tt.token)
            }

            w := httptest.NewRecorder()
            middleware.ServeHTTP(w, req)

            if w.Code != tt.wantStatus {
                t.Errorf("Expected %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}
```

---

## Integration Testing

### Full Application Test

```go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"

    "github.com/spcent/plumego/core"
)

func TestUserAPI(t *testing.T) {
    // Setup
    app := core.New(
        core.WithAddr(":0"),
        core.WithRecommendedMiddleware(),
    )

    setupRoutes(app)

    go app.Boot()
    defer app.Shutdown(context.Background())

    time.Sleep(100 * time.Millisecond)

    baseURL := "http://" + app.Addr()

    // Test: Create user
    t.Run("CreateUser", func(t *testing.T) {
        user := map[string]string{"name": "Alice", "email": "alice@example.com"}
        body, _ := json.Marshal(user)

        resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(body))
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusCreated {
            t.Errorf("Expected 201, got %d", resp.StatusCode)
        }
    })

    // Test: List users
    t.Run("ListUsers", func(t *testing.T) {
        resp, err := http.Get(baseURL + "/users")
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Expected 200, got %d", resp.StatusCode)
        }

        var users []map[string]string
        if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
            t.Fatalf("Failed to decode response: %v", err)
        }

        if len(users) == 0 {
            t.Error("Expected at least one user")
        }
    })
}
```

### Database Integration Test

```go
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create app with real database
    app := core.New(
        core.WithAddr(":0"),
        core.WithComponent(&DatabaseComponent{DB: db}),
    )

    go app.Boot()
    defer app.Shutdown(context.Background())

    time.Sleep(100 * time.Millisecond)

    // Run tests against real database
    t.Run("CRUD Operations", func(t *testing.T) {
        // Create
        // Read
        // Update
        // Delete
    })
}

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://localhost/test?sslmode=disable")
    if err != nil {
        t.Fatalf("Failed to open database: %v", err)
    }

    // Run migrations
    // Seed test data

    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    // Drop test data
    db.Close()
}
```

---

## Best Practices

### ✅ Do

1. **Use Table-Driven Tests**
   ```go
   func TestHandler(t *testing.T) {
       tests := []struct {
           name       string
           input      string
           want       string
           wantStatus int
       }{
           {"valid", "input1", "output1", 200},
           {"invalid", "input2", "", 400},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test logic
           })
       }
   }
   ```

2. **Use Test Helpers**
   ```go
   func newTestApp(t *testing.T) *core.App {
       return core.New(core.WithAddr(":0"))
   }
   ```

3. **Clean Up Resources**
   ```go
   func TestSomething(t *testing.T) {
       app := newTestApp(t)
       defer app.Shutdown(context.Background())
       // Test logic
   }
   ```

4. **Use Subtests**
   ```go
   func TestAPI(t *testing.T) {
       t.Run("Create", func(t *testing.T) { ... })
       t.Run("Read", func(t *testing.T) { ... })
       t.Run("Update", func(t *testing.T) { ... })
       t.Run("Delete", func(t *testing.T) { ... })
   }
   ```

5. **Mock External Dependencies**
   ```go
   type MockService struct {
       response string
       err      error
   }

   func (m *MockService) Call() (string, error) {
       return m.response, m.err
   }
   ```

### ❌ Don't

1. **Don't Use Fixed Ports**
   ```go
   // ❌ Port might be in use
   app := core.New(core.WithAddr(":8080"))

   // ✅ Use random port
   app := core.New(core.WithAddr(":0"))
   ```

2. **Don't Leak Resources**
   ```go
   // ❌ No cleanup
   func TestApp(t *testing.T) {
       app := core.New()
       app.Boot()
   }

   // ✅ With cleanup
   func TestApp(t *testing.T) {
       app := core.New()
       go app.Boot()
       defer app.Shutdown(context.Background())
   }
   ```

3. **Don't Skip Error Checks**
   ```go
   // ❌ Ignoring errors
   app.Boot()

   // ✅ Check errors
   if err := app.Boot(); err != nil {
       t.Fatalf("Boot failed: %v", err)
   }
   ```

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestApplication ./core

# Run integration tests
go test -tags=integration ./...

# Skip integration tests
go test -short ./...

# Verbose output
go test -v ./...

# With timeout
go test -timeout 30s ./...
```

---

## Next Steps

- **[Application](application.md)** - Creating testable applications
- **[Components](components.md)** - Testing components
- **[Runners](runners.md)** - Testing background services

---

**Related**:
- [Application Configuration](application.md)
- [Component System](components.md)
- [Dependency Injection](dependency-injection.md)
