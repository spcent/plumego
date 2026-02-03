# API Gateway Example

This example demonstrates how to use plumego as an API gateway with reverse proxy, load balancing, and service discovery.

## Features

- ✅ **Reverse Proxy**: Forward requests to backend services
- ✅ **Load Balancing**: Round-robin, weighted, least-connections, IP-hash
- ✅ **Service Discovery**: Static, Consul (extensible)
- ✅ **Health Checking**: Active health probes with auto-recovery
- ✅ **Path Rewriting**: Strip/add/replace URL prefixes
- ✅ **Request/Response Modification**: Add headers, transform data
- ✅ **Custom Error Handling**: Graceful degradation
- ✅ **CORS Support**: Cross-origin resource sharing

## Quick Start

### 1. Start Mock Backend Services

In separate terminals, start 3 backend services:

```bash
# Terminal 1: User service backend 1
go run mock-service.go 8081 user-service-1

# Terminal 2: User service backend 2
go run mock-service.go 8082 user-service-2

# Terminal 3: Product service
go run mock-service.go 7001 product-service
```

### 2. Start the API Gateway

```bash
# Terminal 4: API Gateway
go run main.go
```

The gateway will start on http://localhost:8080

### 3. Test the Gateway

```bash
# Gateway status
curl http://localhost:8080/status

# Gateway health
curl http://localhost:8080/health

# Get users (proxied to user service with round-robin)
curl http://localhost:8080/api/v1/users

# Refresh multiple times to see load balancing:
for i in {1..5}; do
  curl http://localhost:8080/api/v1/users
  sleep 1
done

# Get products (proxied to product service)
curl http://localhost:8080/api/v1/products
```

## Architecture

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     ▼
┌─────────────────────────────────┐
│   API Gateway (:8080)           │
│                                 │
│  - Reverse Proxy                │
│  - Load Balancing               │
│  - Health Checking              │
│  - Path Rewriting               │
│  - Service Discovery            │
└─────────────────────────────────┘
     │
     ├──────────┬──────────────┬─────────────┐
     │          │              │             │
     ▼          ▼              ▼             ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐
│ User    │ │ User    │ │ Order   │ │ Product     │
│ Service │ │ Service │ │ Service │ │ Service     │
│ :8081   │ │ :8082   │ │ :9001   │ │ :7001       │
└─────────┘ └─────────┘ └─────────┘ └─────────────┘
```

## Configuration

### Static Backends

```go
app.Use("/api/v1/users/*", proxy.New(proxy.Config{
    Targets: []string{
        "http://localhost:8081",
        "http://localhost:8082",
    },
    LoadBalancer: proxy.NewRoundRobinBalancer(),
    PathRewrite:  proxy.StripPrefix("/api/v1"),
}))
```

### Service Discovery

```go
sd := discovery.NewStatic(map[string][]string{
    "user-service": {
        "http://localhost:8081",
        "http://localhost:8082",
    },
})

app.Use("/api/v1/orders/*", proxy.New(proxy.Config{
    ServiceName:  "order-service",
    Discovery:    sd,
    LoadBalancer: proxy.NewWeightedRoundRobinBalancer(),
}))
```

### Health Checking

```go
app.Use("/api/*", proxy.New(proxy.Config{
    Targets: []string{"http://localhost:8081"},
    HealthCheck: &proxy.HealthCheckConfig{
        Interval:       10 * time.Second,
        Timeout:        5 * time.Second,
        Path:           "/health",
        ExpectedStatus: http.StatusOK,
        OnHealthChange: func(backend *proxy.Backend, healthy bool) {
            log.Printf("Backend %s is now %s", 
                backend.URL, 
                map[bool]string{true: "healthy", false: "unhealthy"}[healthy])
        },
    },
}))
```

### Path Rewriting

```go
// Strip prefix: /api/v1/users -> /users
PathRewrite: proxy.StripPrefix("/api/v1")

// Add prefix: /users -> /internal/users
PathRewrite: proxy.AddPrefix("/internal")

// Replace prefix: /api/v1/users -> /api/v2/users
PathRewrite: proxy.ReplacePrefix("/api/v1", "/api/v2")

// Chain rewriters
PathRewrite: proxy.Chain(
    proxy.StripPrefix("/api"),
    proxy.AddPrefix("/internal"),
)
```

### Request/Response Modification

```go
proxy.Config{
    // Modify outgoing request
    ModifyRequest: proxy.ChainRequestModifiers(
        proxy.AddHeader("X-Gateway", "plumego"),
        proxy.AddHeader("X-Request-ID", uuid.New().String()),
    ),

    // Modify incoming response
    ModifyResponse: proxy.ChainResponseModifiers(
        proxy.AddResponseHeader("X-Served-By", "Gateway"),
        proxy.AddResponseHeader("X-Cache-Status", "MISS"),
    ),
}
```

### Custom Error Handling

```go
proxy.Config{
    ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
        log.Printf("Proxy error: %v", err)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Service unavailable",
            "message": "Please try again later",
        })
    },
}
```

## Load Balancing Strategies

### Round Robin

Distributes requests evenly across backends:

```go
LoadBalancer: proxy.NewRoundRobinBalancer()
```

### Weighted Round Robin

Distributes requests based on backend weights:

```go
LoadBalancer: proxy.NewWeightedRoundRobinBalancer()

// Set weights on backends (requires manual backend configuration)
```

### Random

Randomly selects a backend:

```go
LoadBalancer: proxy.NewRandomBalancer()
```

### Least Connections

Routes to the backend with fewest active connections:

```go
LoadBalancer: proxy.NewLeastConnectionsBalancer()
```

### IP Hash

Routes requests from the same client IP to the same backend:

```go
LoadBalancer: proxy.NewIPHashBalancer()
```

## Service Discovery Providers

### Static (In-Memory)

Best for: Development, testing, simple deployments

```go
sd := discovery.NewStatic(map[string][]string{
    "service-name": {"http://host1:8080", "http://host2:8080"},
})
```

### Consul

Best for: Production, microservices, dynamic environments

```go
sd, err := discovery.NewConsul("localhost:8500", discovery.ConsulConfig{
    Datacenter:  "dc1",
    Token:       "secret-token",
    OnlyHealthy: true,
})
```

### Kubernetes (Future)

Coming in v1.1:

```go
sd, err := discovery.NewKubernetes(discovery.K8sConfig{
    Namespace: "default",
    LabelSelector: "app=myapp",
})
```

## Advanced Features

### WebSocket Proxying

```go
app.Use("/ws/*", proxy.New(proxy.Config{
    Targets: []string{"ws://localhost:8080"},
    WebSocketEnabled: true,
}))
```

### Connection Pooling

```go
proxy.Config{
    Transport: &proxy.TransportConfig{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DialTimeout:         30 * time.Second,
    },
}
```

### Retry Logic

```go
proxy.Config{
    RetryCount:   3,
    RetryBackoff: 100 * time.Millisecond,
}
```

### Timeout Configuration

```go
proxy.Config{
    Timeout: 30 * time.Second,
}
```

## Production Deployment

### With Docker Compose

```yaml
version: '3.8'
services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      - consul

  consul:
    image: consul:latest
    ports:
      - "8500:8500"

  user-service:
    build: ./services/user
    environment:
      - CONSUL_ADDRESS=consul:8500
```

### With Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: gateway
        image: plumego-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DISCOVERY_TYPE
          value: "kubernetes"
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: api-gateway
```

## Monitoring

The gateway automatically tracks:

- Request count per backend
- Success/failure rates
- Backend health status
- Connection pool statistics

Access metrics at `/health` endpoint or integrate with Prometheus.

## Troubleshooting

### All backends unhealthy

1. Check backend services are running
2. Check health check configuration
3. Verify network connectivity
4. Check logs for error messages

### High latency

1. Check backend response times
2. Adjust timeout settings
3. Review connection pool configuration
4. Consider adding more backends

### Load balancing not working

1. Verify multiple backends are configured
2. Check backends are healthy
3. Review load balancer strategy
4. Check logs for selection errors

## Next Steps

- Integrate with Consul for dynamic service discovery
- Add circuit breaker for fault tolerance
- Implement response caching
- Add rate limiting per service
- Set up centralized logging
- Configure distributed tracing

## License

MIT
