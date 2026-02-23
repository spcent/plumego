# Dependency Injection

> **Package**: `github.com/spcent/plumego/core/di`

Plumego includes a built-in dependency injection (DI) container that manages component dependencies, resolves initialization order, and prevents circular dependencies.

---

## Table of Contents

- [Overview](#overview)
- [Container API](#container-api)
- [Registration](#registration)
- [Resolution](#resolution)
- [Topological Sorting](#topological-sorting)
- [Lifecycle Integration](#lifecycle-integration)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## Overview

### What is Dependency Injection?

Dependency Injection is a design pattern where objects receive their dependencies from external sources rather than creating them internally.

### Why DI in Plumego?

The DI container provides:
- **Automatic dependency resolution**: Components declare what they need
- **Lifecycle management**: Dependencies start before dependents
- **Type safety**: Compile-time type checking
- **Testability**: Easy to mock dependencies
- **No circular dependencies**: Detected at boot time

---

## Container API

### Core Types

```go
package di

// Container manages dependencies
type Container struct { ... }

// New creates a new container
func New() *Container

// Register adds a dependency to the container
func (c *Container) Register(value interface{}) error

// Resolve retrieves a dependency from the container
func (c *Container) Resolve(target interface{}) error

// Has checks if a type is registered
func (c *Container) Has(typ reflect.Type) bool

// MustResolve resolves or panics
func (c *Container) MustResolve(target interface{})
```

---

## Registration

### Registering Components

Components are automatically registered when added to the application:

```go
app := core.New(
    core.WithComponent(&DatabaseComponent{}),
    core.WithComponent(&CacheComponent{}),
)
// Both components are registered in the DI container
```

### Manual Registration

You can also register dependencies manually:

```go
container := app.DI()

// Register by value
db := &DatabaseConnection{}
container.Register(db)

// Register interface
var cache Cache = &RedisCache{}
container.Register(cache)
```

### Registration Rules

1. **One instance per type**: Only one instance of each type can be registered
2. **Type-based**: Registration is by concrete type, not interface
3. **Pointer types**: Must register pointer types for components
4. **Thread-safe**: Registration is safe for concurrent access

---

## Resolution

### Resolving Dependencies

```go
// In component Start() method
func (c *MyComponent) Start(ctx context.Context) error {
    var db *DatabaseComponent
    if err := app.DI().Resolve(&db); err != nil {
        return fmt.Errorf("failed to resolve database: %w", err)
    }

    c.db = db
    return nil
}
```

### MustResolve

For dependencies that must exist:

```go
func (c *MyComponent) Start(ctx context.Context) error {
    var db *DatabaseComponent
    app.DI().MustResolve(&db) // Panics if not found

    c.db = db
    return nil
}
```

### Checking Existence

```go
if app.DI().Has(reflect.TypeOf(&OptionalComponent{})) {
    var comp *OptionalComponent
    app.DI().Resolve(&comp)
    // Use comp
}
```

---

## Topological Sorting

### How It Works

The DI container uses topological sorting to determine component start order based on dependencies:

```go
// Component A has no dependencies
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type {
    return nil
}

// Component B depends on A
type ComponentB struct {}
func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentA{})}
}

// Component C depends on B
type ComponentC struct {}
func (c *ComponentC) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentB{})}
}

app := core.New(
    core.WithComponent(&ComponentC{}),
    core.WithComponent(&ComponentA{}),
    core.WithComponent(&ComponentB{}),
)

// Start order: A → B → C (regardless of registration order)
// Stop order:  C → B → A (reverse)
```

### Dependency Graph

```
Registration Order:  C, A, B
Dependency Graph:    A ← B ← C
Start Order:         A → B → C
Stop Order:          C → B → A
```

### Circular Dependency Detection

Circular dependencies are detected at boot time:

```go
// Component A depends on B
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentB{})}
}

// Component B depends on A (circular!)
type ComponentB struct {}
func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentA{})}
}

app := core.New(
    core.WithComponent(&ComponentA{}),
    core.WithComponent(&ComponentB{}),
)

app.Boot() // panic: circular dependency detected: A → B → A
```

---

## Lifecycle Integration

### Boot Sequence with DI

```
1. Component Registration
   ├── App.WithComponent(&ComponentA{})
   ├── App.WithComponent(&ComponentB{})
   └── App.WithComponent(&ComponentC{})

2. Dependency Resolution
   ├── Build dependency graph
   ├── Detect circular dependencies
   └── Topological sort

3. Component Start (in dependency order)
   ├── ComponentA.Start()
   ├── ComponentB.Start()
   └── ComponentC.Start()
```

### Accessing Dependencies in Start()

```go
type CacheComponent struct {
    db *DatabaseComponent
}

func (c *CacheComponent) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&DatabaseComponent{})}
}

func (c *CacheComponent) Start(ctx context.Context) error {
    // DatabaseComponent is guaranteed to be started
    var db *DatabaseComponent
    if err := app.DI().Resolve(&db); err != nil {
        return err
    }

    c.db = db
    return c.initialize()
}
```

---

## Best Practices

### ✅ Do

1. **Declare Dependencies Explicitly**
   ```go
   func (c *MyComponent) Dependencies() []reflect.Type {
       return []reflect.Type{
           reflect.TypeOf(&RequiredComponent{}),
       }
   }
   ```

2. **Resolve in Start() Method**
   ```go
   func (c *MyComponent) Start(ctx context.Context) error {
       var dep *RequiredComponent
       if err := app.DI().Resolve(&dep); err != nil {
           return err
       }
       c.dep = dep
       return nil
   }
   ```

3. **Use Pointer Types**
   ```go
   // ✅ Correct
   reflect.TypeOf(&ComponentA{})

   // ❌ Wrong
   reflect.TypeOf(ComponentA{})
   ```

4. **Handle Resolution Errors**
   ```go
   var dep *OptionalComponent
   if err := app.DI().Resolve(&dep); err != nil {
       log.Printf("Optional component not available: %v", err)
       // Continue without it
   }
   ```

### ❌ Don't

1. **Don't Create Circular Dependencies**
   ```go
   // ❌ Circular dependency
   type A struct {}
   func (a *A) Dependencies() []reflect.Type {
       return []reflect.Type{reflect.TypeOf(&B{})}
   }

   type B struct {}
   func (b *B) Dependencies() []reflect.Type {
       return []reflect.Type{reflect.TypeOf(&A{})}
   }
   ```

2. **Don't Resolve Before Start()**
   ```go
   // ❌ Don't resolve in constructor
   func NewComponent(app *core.App) *Component {
       var dep *Dependency
       app.DI().Resolve(&dep) // Too early!
       return &Component{dep: dep}
   }

   // ✅ Resolve in Start()
   func (c *Component) Start(ctx context.Context) error {
       var dep *Dependency
       app.DI().Resolve(&dep)
       c.dep = dep
       return nil
   }
   ```

3. **Don't Use Global Variables**
   ```go
   // ❌ Global dependency
   var globalDB *DatabaseComponent

   // ✅ Component field
   type Component struct {
       db *DatabaseComponent
   }
   ```

4. **Don't Register Non-Pointer Types**
   ```go
   // ❌ Value type
   container.Register(ComponentA{})

   // ✅ Pointer type
   container.Register(&ComponentA{})
   ```

---

## Examples

### Basic Dependency

```go
package main

import (
    "context"
    "github.com/spcent/plumego/core"
)

// Database component (no dependencies)
type DatabaseComponent struct {
    connected bool
}

func (c *DatabaseComponent) Start(ctx context.Context) error {
    c.connected = true
    return nil
}

func (c *DatabaseComponent) Dependencies() []reflect.Type {
    return nil
}

// ... other methods

// Cache component (depends on Database)
type CacheComponent struct {
    db *DatabaseComponent
}

func (c *CacheComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&DatabaseComponent{}),
    }
}

func (c *CacheComponent) Start(ctx context.Context) error {
    var db *DatabaseComponent
    if err := app.DI().Resolve(&db); err != nil {
        return err
    }

    c.db = db
    return nil
}

// ... other methods

func main() {
    app := core.New(
        core.WithComponent(&CacheComponent{}),
        core.WithComponent(&DatabaseComponent{}),
    )

    // Start order: Database → Cache
    app.Boot()
}
```

### Multiple Dependencies

```go
type AnalyticsComponent struct {
    db    *DatabaseComponent
    cache *CacheComponent
    queue *QueueComponent
}

func (c *AnalyticsComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&DatabaseComponent{}),
        reflect.TypeOf(&CacheComponent{}),
        reflect.TypeOf(&QueueComponent{}),
    }
}

func (c *AnalyticsComponent) Start(ctx context.Context) error {
    // Resolve all dependencies
    if err := app.DI().Resolve(&c.db); err != nil {
        return err
    }
    if err := app.DI().Resolve(&c.cache); err != nil {
        return err
    }
    if err := app.DI().Resolve(&c.queue); err != nil {
        return err
    }

    return nil
}
```

### Optional Dependencies

```go
type NotificationComponent struct {
    email *EmailComponent // Optional
    sms   *SMSComponent   // Optional
}

func (c *NotificationComponent) Dependencies() []reflect.Type {
    // No hard dependencies
    return nil
}

func (c *NotificationComponent) Start(ctx context.Context) error {
    // Try to resolve optional dependencies
    if app.DI().Has(reflect.TypeOf(&EmailComponent{})) {
        app.DI().Resolve(&c.email)
    }

    if app.DI().Has(reflect.TypeOf(&SMSComponent{})) {
        app.DI().Resolve(&c.sms)
    }

    return nil
}

func (c *NotificationComponent) Send(message string) {
    if c.email != nil {
        c.email.Send(message)
    }

    if c.sms != nil {
        c.sms.Send(message)
    }
}
```

### Dependency Chain

```go
// Level 1: No dependencies
type ConfigComponent struct {}
func (c *ConfigComponent) Dependencies() []reflect.Type { return nil }

// Level 2: Depends on Config
type DatabaseComponent struct {
    config *ConfigComponent
}
func (c *DatabaseComponent) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ConfigComponent{})}
}

// Level 3: Depends on Database
type CacheComponent struct {
    db *DatabaseComponent
}
func (c *CacheComponent) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&DatabaseComponent{})}
}

// Level 4: Depends on Cache
type APIComponent struct {
    cache *CacheComponent
}
func (c *APIComponent) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&CacheComponent{})}
}

// Start order: Config → Database → Cache → API
```

### Testing with DI

```go
package mycomponent

import (
    "context"
    "testing"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/core/di"
)

// Mock dependency
type MockDatabase struct {
    data map[string]string
}

func (m *MockDatabase) Get(key string) string {
    return m.data[key]
}

func TestComponent(t *testing.T) {
    // Create DI container
    container := di.New()

    // Register mock dependency
    mockDB := &MockDatabase{
        data: map[string]string{"key": "value"},
    }
    container.Register(mockDB)

    // Create component
    comp := &MyComponent{}

    // Resolve mock dependency
    var db *MockDatabase
    if err := container.Resolve(&db); err != nil {
        t.Fatalf("Failed to resolve: %v", err)
    }

    comp.db = db

    // Test component behavior
    result := comp.DoSomething()
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}
```

---

## Advanced Usage

### Custom Container

```go
// Create custom container
container := di.New()

// Register services
container.Register(&DatabaseService{})
container.Register(&CacheService{})

// Use in application
app := core.New(
    core.WithDIContainer(container),
)
```

### Lazy Resolution

```go
type MyComponent struct {
    getDB func() *DatabaseComponent
}

func (c *MyComponent) Start(ctx context.Context) error {
    // Store resolver function instead of resolving immediately
    c.getDB = func() *DatabaseComponent {
        var db *DatabaseComponent
        app.DI().MustResolve(&db)
        return db
    }
    return nil
}

func (c *MyComponent) Query() {
    // Resolve lazily
    db := c.getDB()
    db.Query("SELECT ...")
}
```

### Factory Pattern

```go
type DatabaseFactory struct {}

func (f *DatabaseFactory) Create(config Config) *DatabaseComponent {
    return &DatabaseComponent{
        DSN: config.DSN,
    }
}

// Register factory
app.DI().Register(&DatabaseFactory{})

// Resolve factory
type MyComponent struct {
    factory *DatabaseFactory
}

func (c *MyComponent) Start(ctx context.Context) error {
    app.DI().Resolve(&c.factory)
    return nil
}

func (c *MyComponent) NewConnection() *DatabaseComponent {
    return c.factory.Create(c.config)
}
```

---

## Troubleshooting

### Problem: Circular Dependency

**Error**: `panic: circular dependency detected: A → B → A`

**Solution**: Refactor to remove circular dependency
```go
// ❌ Circular
A depends on B
B depends on A

// ✅ Fixed option 1: Introduce interface
A depends on Interface
B implements Interface

// ✅ Fixed option 2: Extract common dependency
A depends on C
B depends on C
```

### Problem: Dependency Not Found

**Error**: `failed to resolve dependency: type not registered`

**Solution**: Ensure component is registered
```go
app := core.New(
    core.WithComponent(&RequiredComponent{}), // Must register first
    core.WithComponent(&DependentComponent{}),
)
```

### Problem: Wrong Resolution Order

**Error**: Component starts before its dependency

**Solution**: Declare dependency explicitly
```go
func (c *MyComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&RequiredComponent{}), // Must declare
    }
}
```

---

## Next Steps

- **[Components](components.md)** - Building components with dependencies
- **[Lifecycle](lifecycle.md)** - Understanding component start order
- **[Testing](testing.md)** - Testing with mock dependencies

---

**Related**:
- [Component System](components.md)
- [Lifecycle Management](lifecycle.md)
- [Application Configuration](application.md)
