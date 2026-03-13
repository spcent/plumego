# Service Discovery

> **Package**: `github.com/spcent/plumego/x/discovery` | **Backends**: Static, Consul

## Overview

The `discovery` package provides service discovery for microservice architectures. It resolves service names to network addresses, enabling dynamic routing without hardcoded endpoints.

## Discovery Interface

```go
type Discovery interface {
    Register(service Service) error
    Deregister(serviceID string) error
    Resolve(name string) (string, error)
    ResolveAll(name string) ([]string, error)
    Watch(name string, onChange func([]Service)) error
}
```

## Static Discovery

Best for simple setups with known, fixed endpoints:

```go
import "github.com/spcent/plumego/x/discovery"

registry := discovery.NewStatic(
    discovery.Service{
        ID:   "payment-1",
        Name: "payment-service",
        Addr: "payment-1:8080",
        Tags: []string{"v2"},
    },
    discovery.Service{
        ID:   "payment-2",
        Name: "payment-service",
        Addr: "payment-2:8080",
        Tags: []string{"v2"},
    },
    discovery.Service{
        ID:   "email-1",
        Name: "email-service",
        Addr: "email-service:8080",
    },
)

// Resolve single instance (round-robin)
addr, err := registry.Resolve("payment-service")
fmt.Println(addr) // "payment-1:8080" or "payment-2:8080"

// Resolve all instances
addrs, err := registry.ResolveAll("payment-service")
// ["payment-1:8080", "payment-2:8080"]
```

### Environment-Based Static Config

```go
// Load from environment variables
registry := discovery.NewStaticFromEnv(
    "PAYMENT_SERVICE_ADDR",
    "EMAIL_SERVICE_ADDR",
    "NOTIFICATION_SERVICE_ADDR",
)
```

## Consul Discovery

For dynamic environments where services register/deregister automatically:

```go
registry, err := discovery.NewConsul(
    os.Getenv("CONSUL_ADDR"), // "consul:8500"
    discovery.ConsulConfig{
        Datacenter: "dc1",
        Token:      os.Getenv("CONSUL_TOKEN"),
        Namespace:  "default", // Consul Enterprise namespace (optional)

        // IncludeUnhealthy: false (default) — only passing/healthy instances returned
        // Set to true to include unhealthy instances
        IncludeUnhealthy: false,

        WaitTime: 30 * time.Second, // polling interval for Watch
        Timeout:  60 * time.Second, // HTTP request timeout
    },
)
```

### Register Current Service

```go
err := registry.Register(discovery.Service{
    ID:      "my-service-" + hostname,
    Name:    "my-service",
    Addr:    hostname + ":8080",
    Tags:    []string{"production", "v2.1"},
    Meta:    map[string]string{"version": "2.1.0"},
    Check: &discovery.HealthCheck{
        HTTP:     "http://localhost:8080/health",
        Interval: "10s",
        Timeout:  "3s",
    },
})
```

### Deregister on Shutdown

```go
// Register as Plumego runner for automatic lifecycle management
app := core.New(
    core.WithRunner(discovery.NewConsulRunner(registry, service)),
)
// Automatically registers on Start, deregisters on Stop
```

### Watch for Changes

```go
registry.Watch("payment-service", func(services []discovery.Service) {
    log.Infof("Payment service instances changed: %d", len(services))
    loadBalancer.Update(services)
})
```

## Load Balancing

### Round-Robin

```go
lb := discovery.NewRoundRobin(registry)

// Get next instance (cycles through all healthy instances)
addr, err := lb.Next("payment-service")
```

### Random

```go
lb := discovery.NewRandom(registry)
addr, err := lb.Next("payment-service")
```

### Least Connections

```go
lb := discovery.NewLeastConnections(registry)
addr, err := lb.Next("payment-service")
```

## HTTP Client Integration

Use service discovery to resolve an address, then build the outbound request with the standard library:

```go
addr, err := registry.Resolve("payment-service")
if err != nil {
    return err
}

client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get("http://" + addr + "/api/payments")
// Resolves payment-service -> payment-1:8080 or payment-2:8080
```

## Service Health

```go
// Check service health directly
healthy, err := registry.Health("payment-service")
fmt.Printf("Healthy instances: %d\n", len(healthy))

// Filter by tag
instances, err := registry.ResolveByTag("payment-service", "v2")
```

## Docker Compose Configuration

```yaml
services:
  app:
    environment:
      CONSUL_ADDR: consul:8500
    depends_on:
      - consul

  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    command: agent -server -bootstrap-expect=1 -ui -client=0.0.0.0
```

## Kubernetes Integration

In Kubernetes, use the built-in DNS for service discovery:

```go
// Kubernetes DNS automatically resolves service names
registry := discovery.NewStatic(
    discovery.Service{Name: "payment-service", Addr: "payment-service.default.svc.cluster.local:8080"},
    discovery.Service{Name: "email-service", Addr: "email-service.default.svc.cluster.local:8080"},
)
```

## Related Documentation

- [Net Overview](../README.md) — Network module overview
- [Middleware: Proxy](../../middleware/proxy.md) — Proxy with service discovery
