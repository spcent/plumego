# API Gateway - Phase 3: Advanced Features

Phase 3 introduces enterprise-grade advanced features including distributed tracing, response caching, canary deployments, and advanced routing capabilities.

## Overview

Phase 3 features:
- **Distributed Tracing** - OpenTelemetry-compatible request tracing
- **Response Caching** - Intelligent response caching with LRU eviction
- **Canary Deployments** - Traffic splitting for gradual rollouts
- **Advanced Routing** - Rule-based routing with conditions

## Features

### 1. Distributed Tracing

OpenTelemetry-compatible distributed tracing for request correlation across services.

#### Configuration

```bash
# Enable distributed tracing
TRACING_ENABLED=true

# Tracing backend endpoint
TRACING_ENDPOINT=http://localhost:14268/api/traces

# Service name for traces
TRACING_SERVICE_NAME=api-gateway

# Sampling rate (0.0 to 1.0)
TRACING_SAMPLE_RATE=1.0

# Export format (jaeger, zipkin, otlp)
TRACING_EXPORT_FORMAT=jaeger
```

#### How It Works

1. **Trace ID Propagation**
   - Extracts `X-Trace-ID` from incoming requests
   - Generates new trace ID if not present
   - Propagates trace ID to downstream services

2. **Span Creation**
   - Creates a span for each request
   - Records span duration and status
   - Generates unique span IDs

3. **Sampling**
   - Configurable sampling rate (e.g., 0.1 = 10% of requests)
   - Reduces overhead for high-traffic systems
   - Always samples when trace ID is provided

#### Headers Added

| Header | Direction | Description |
|--------|-----------|-------------|
| `X-Trace-ID` | Request & Response | Unique trace identifier |
| `X-Span-ID` | Request | Current span identifier |
| `X-Parent-Span-ID` | Request | Parent span identifier |

#### Example with Jaeger

**1. Start Jaeger**

```bash
docker run -d --name jaeger \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  jaegertracing/all-in-one:latest
```

**2. Configure Gateway**

```bash
TRACING_ENABLED=true
TRACING_ENDPOINT=http://localhost:14268/api/traces
TRACING_SERVICE_NAME=api-gateway
TRACING_SAMPLE_RATE=1.0
TRACING_EXPORT_FORMAT=jaeger
```

**3. View Traces**

Open Jaeger UI at http://localhost:16686 and search for traces by:
- Service: `api-gateway`
- Operation: HTTP method + path
- Trace ID: From `X-Trace-ID` header

#### Example with Zipkin

**1. Start Zipkin**

```bash
docker run -d --name zipkin \
  -p 9411:9411 \
  openzipkin/zipkin:latest
```

**2. Configure Gateway**

```bash
TRACING_ENABLED=true
TRACING_ENDPOINT=http://localhost:9411/api/v2/spans
TRACING_SERVICE_NAME=api-gateway
TRACING_SAMPLE_RATE=0.1  # Sample 10% of requests
TRACING_EXPORT_FORMAT=zipkin
```

**3. View Traces**

Open Zipkin UI at http://localhost:9411

#### Sampling Strategies

```bash
# Development: Trace everything
TRACING_SAMPLE_RATE=1.0

# Production: Sample 10% of requests
TRACING_SAMPLE_RATE=0.1

# High-traffic: Sample 1% of requests
TRACING_SAMPLE_RATE=0.01

# Disable: No sampling
TRACING_ENABLED=false
```

#### Propagating Traces to Backend Services

Backend services should read and propagate trace headers:

```go
// In backend service
traceID := r.Header.Get("X-Trace-ID")
spanID := r.Header.Get("X-Span-ID")

// Add to outgoing requests
req.Header.Set("X-Trace-ID", traceID)
req.Header.Set("X-Parent-Span-ID", spanID)
```

### 2. Response Caching

Intelligent HTTP response caching with LRU eviction and TTL expiration.

#### Configuration

```bash
# Enable response caching
CACHE_ENABLED=true

# Cache TTL (time to live)
CACHE_TTL=5m

# Maximum cache size in MB
CACHE_MAX_SIZE=100

# Paths to exclude from caching (wildcards supported)
CACHE_EXCLUDE_PATHS=/admin/*,/metrics,/health

# Only cache these HTTP methods
CACHE_ONLY_METHODS=GET,HEAD

# Headers that affect cache key
CACHE_VARY_HEADERS=Accept,Accept-Encoding,Accept-Language
```

#### How It Works

1. **Cache Key Generation**
   - Combines: Method + Path + Query + Vary Headers
   - Hashed using SHA-256 for fixed-length keys
   - Example: `GET:/api/v1/users?page=1:Accept=application/json`

2. **Cache Lookup**
   - Checks if valid cached response exists
   - Returns cached response with `X-Cache: HIT` header
   - Falls through to backend with `X-Cache: MISS` header

3. **Cache Storage**
   - Only caches successful responses (2xx status codes)
   - Respects configured TTL
   - Stores: Status code, body, and Content-Type header

4. **Cache Eviction**
   - LRU (Least Recently Used) when size limit exceeded
   - Automatic cleanup of expired entries every minute
   - Per-entry size tracking

#### Cache Headers

| Header | Value | Description |
|--------|-------|-------------|
| `X-Cache` | `HIT` or `MISS` | Indicates cache hit or miss |

#### Example Usage

**Basic Caching**

```bash
# Enable caching for GET requests
CACHE_ENABLED=true
CACHE_TTL=5m
CACHE_MAX_SIZE=100
```

```bash
# First request - cache miss
curl -i http://localhost:8080/api/v1/products
# X-Cache: MISS

# Second request within 5 minutes - cache hit
curl -i http://localhost:8080/api/v1/products
# X-Cache: HIT
```

**Content Negotiation**

With vary headers, different content types are cached separately:

```bash
CACHE_VARY_HEADERS=Accept,Accept-Language

# JSON response cached separately
curl -H "Accept: application/json" http://localhost:8080/api/v1/users

# XML response cached separately
curl -H "Accept: application/xml" http://localhost:8080/api/v1/users
```

**Excluding Paths**

```bash
# Don't cache admin or real-time endpoints
CACHE_EXCLUDE_PATHS=/admin/*,/metrics,/health,/api/v1/realtime/*
```

#### Cache Invalidation

Cache entries are automatically invalidated:
- After TTL expires
- When LRU eviction occurs
- On gateway restart

For manual invalidation:
```bash
# Restart gateway to clear all cache
# Or implement a cache flush endpoint in Phase 4
```

#### Performance Considerations

**Memory Usage**

```bash
# For 100 MB cache:
# - ~1000 cached responses at 100 KB each
# - ~10,000 cached responses at 10 KB each
CACHE_MAX_SIZE=100

# For high-traffic sites:
CACHE_MAX_SIZE=500  # 500 MB
```

**TTL Selection**

```bash
# Static content (product catalog)
CACHE_TTL=15m

# Semi-static content (user profiles)
CACHE_TTL=5m

# Dynamic content (search results)
CACHE_TTL=1m
```

#### Cache Statistics

Monitor cache effectiveness:
- Cache hit rate: Hits / (Hits + Misses)
- Target: 70-90% hit rate for cacheable content

### 3. Canary Deployments / Traffic Splitting

Gradually roll out new versions by splitting traffic between primary and canary targets.

#### Configuration

```bash
# Enable canary deployments
CANARY_ENABLED=true

# Default weight for primary targets
CANARY_DEFAULT_WEIGHT=100
```

#### Canary Rules

Canary rules are configured per service in code or JSON config:

```go
cfg.Canary.Rules = []CanaryRule{
    {
        ServiceName:   "user",
        CanaryTargets: []string{"http://user-service-canary:8081"},
        CanaryWeight:  10,  // 10% of traffic
    },
    {
        ServiceName:   "order",
        CanaryTargets: []string{"http://order-service-v2:9001"},
        CanaryWeight:  20,  // 20% of traffic
        HeaderMatch: map[string]string{
            "X-Beta-User": "true",  // Beta users only
        },
    },
}
```

#### Routing Methods

**1. Weight-Based Routing**

Routes percentage of traffic to canary:

```go
CanaryWeight: 10  // 10% canary, 90% primary
```

Traffic split is random but statistically consistent.

**2. Header-Based Routing**

Routes requests with specific headers to canary:

```go
HeaderMatch: map[string]string{
    "X-Beta-User": "true",
    "X-Version": "v2",
}
```

All conditions must match to route to canary.

**3. Cookie-Based Routing**

Routes requests with specific cookie to canary:

```go
CookieMatch: "canary_group"  // Cookie name
```

Cookie value must be "canary" to route to canary targets.

#### Example: Gradual Rollout

**Phase 1: Internal Testing (1%)**

```go
CanaryRule{
    ServiceName:   "user",
    CanaryTargets: []string{"http://user-service-v2:8081"},
    CanaryWeight:  1,
    HeaderMatch: map[string]string{
        "X-Internal": "true",
    },
}
```

**Phase 2: Beta Users (10%)**

```go
CanaryRule{
    ServiceName:   "user",
    CanaryTargets: []string{"http://user-service-v2:8081"},
    CanaryWeight:  10,
    HeaderMatch: map[string]string{
        "X-Beta-User": "true",
    },
}
```

**Phase 3: Gradual Rollout (50%)**

```go
CanaryRule{
    ServiceName:   "user",
    CanaryTargets: []string{"http://user-service-v2:8081"},
    CanaryWeight:  50,
}
```

**Phase 4: Full Rollout (100%)**

```go
// Update primary targets to v2
// Remove canary rule
```

#### Monitoring Canary Deployments

**Request Headers**

The gateway adds `X-Canary-Target` header:
- `X-Canary-Target: primary` - Routed to primary
- `X-Canary-Target: canary` - Routed to canary

**Metrics**

Monitor canary vs primary:
```promql
# Request rate by target
rate(http_requests_total{canary_target="canary"}[5m])
rate(http_requests_total{canary_target="primary"}[5m])

# Error rate by target
rate(http_requests_total{canary_target="canary", status=~"5.."}[5m])
rate(http_requests_total{canary_target="primary", status=~"5.."}[5m])

# Latency by target
histogram_quantile(0.95, http_request_duration_seconds{canary_target="canary"})
histogram_quantile(0.95, http_request_duration_seconds{canary_target="primary"})
```

#### Rollback Strategy

If canary shows issues:

1. **Immediate**: Set `CanaryWeight: 0`
2. **Graceful**: Gradually decrease weight
3. **Emergency**: Disable canary: `CANARY_ENABLED=false`

### 4. Advanced Routing Rules

Rule-based routing with header, query parameter, and path matching.

#### Configuration

```bash
# Enable advanced routing
ADVANCED_ROUTING_ENABLED=true
```

#### Routing Rules

Define routing rules in code or JSON config:

```go
cfg.Advanced.RoutingRules = []RoutingRule{
    {
        Path: "/api/v1/users/*",
        HeaderMatch: map[string]string{
            "X-API-Version": "v2",
        },
        TargetService: "user-v2",
        Priority: 10,
    },
    {
        Path: "/api/v1/orders/*",
        QueryMatch: map[string]string{
            "beta": "true",
        },
        TargetURL: "http://order-service-beta:9001",
        Priority: 5,
    },
}
```

#### Rule Matching

Rules are evaluated in priority order (highest first):

1. **Path Matching**
   - Exact: `/api/v1/users/123`
   - Wildcard: `/api/v1/users/*`

2. **Header Matching**
   - All conditions must match
   - Case-sensitive values

3. **Query Parameter Matching**
   - All conditions must match
   - Case-sensitive values

#### Example: API Versioning

Route by API version header:

```go
RoutingRule{
    Path: "/api/users/*",
    HeaderMatch: map[string]string{
        "X-API-Version": "v2",
    },
    TargetService: "user-v2",
    Priority: 10,
},
RoutingRule{
    Path: "/api/users/*",
    TargetService: "user-v1",  // Default
    Priority: 1,
}
```

Usage:
```bash
# Route to v2
curl -H "X-API-Version: v2" http://localhost:8080/api/users/123

# Route to v1 (default)
curl http://localhost:8080/api/users/123
```

#### Example: Feature Flags

Route based on feature flags:

```go
RoutingRule{
    Path: "/api/v1/search/*",
    QueryMatch: map[string]string{
        "use_new_search": "true",
    },
    TargetURL: "http://search-service-v2:7001",
    Priority: 10,
}
```

Usage:
```bash
# New search engine
curl "http://localhost:8080/api/v1/search?q=laptop&use_new_search=true"

# Old search engine
curl "http://localhost:8080/api/v1/search?q=laptop"
```

#### Example: Mobile vs Web

Route based on User-Agent:

```go
RoutingRule{
    Path: "/api/v1/*",
    HeaderMatch: map[string]string{
        "User-Agent": "MobileApp/1.0",
    },
    TargetURL: "http://mobile-api:8080",
    Priority: 10,
}
```

#### Request Headers Added

| Header | Description |
|--------|-------------|
| `X-Target-Service` | Which service to route to |
| `X-Target-URL` | Specific URL override |

## Architecture

### Middleware Stack

Phase 3 middleware positioning:

```
1. Tracing          ← First (capture full request)
2. Security Headers ← Security policy
3. Logging          ← Log all requests
4. Cache            ← Return cached response if available
5. JWT Auth         ← Verify authentication
6. Rate Limiting    ← Enforce limits
7. Timeout          ← Apply timeouts
8. CORS             ← Handle CORS
9. Advanced Routing ← Apply routing rules
10. Canary          ← Traffic splitting
11. Proxy           ← Forward to backend
```

### Request Flow with Phase 3

```
┌─────────────┐
│   Request   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│    Tracing      │ ← Generate/propagate trace ID
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Cache Check   │
└────────┬────────┘
         │
    Hit  │  Miss
    ┌────┴────┐
    │         ▼
    │   ┌──────────────┐
    │   │   Auth,      │
    │   │ Rate Limit,  │
    │   │   etc.       │
    │   └──────┬───────┘
    │          │
    │          ▼
    │   ┌──────────────┐
    │   │  Advanced    │
    │   │  Routing     │
    │   └──────┬───────┘
    │          │
    │          ▼
    │   ┌──────────────┐
    │   │   Canary     │
    │   │  Selection   │
    │   └──────┬───────┘
    │          │
    │          ▼
    │   ┌──────────────┐
    │   │ Proxy Request│
    │   └──────┬───────┘
    │          │
    ▼          ▼
┌─────────────────┐
│     Response    │
└─────────────────┘
```

## Deployment

### Development Configuration

```bash
# Tracing: Full sampling for development
TRACING_ENABLED=true
TRACING_SAMPLE_RATE=1.0
TRACING_ENDPOINT=http://localhost:14268/api/traces

# Caching: Short TTL for development
CACHE_ENABLED=true
CACHE_TTL=1m
CACHE_MAX_SIZE=50

# Canary: Disabled in development
CANARY_ENABLED=false

# Advanced routing: Disabled in development
ADVANCED_ROUTING_ENABLED=false
```

### Production Configuration

```bash
# Tracing: Sample 10% of requests
TRACING_ENABLED=true
TRACING_SAMPLE_RATE=0.1
TRACING_ENDPOINT=http://jaeger-collector:14268/api/traces
TRACING_SERVICE_NAME=api-gateway-prod

# Caching: Longer TTL, larger cache
CACHE_ENABLED=true
CACHE_TTL=10m
CACHE_MAX_SIZE=500
CACHE_EXCLUDE_PATHS=/admin/*,/metrics,/health

# Canary: Enabled with rules
CANARY_ENABLED=true
CANARY_DEFAULT_WEIGHT=100

# Advanced routing: Enabled with rules
ADVANCED_ROUTING_ENABLED=true
```

### Docker Compose with Tracing

```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "5775:5775/udp"
      - "6831:6831/udp"
      - "6832:6832/udp"
      - "5778:5778"
      - "16686:16686"  # UI
      - "14268:14268"  # Collector
      - "14250:14250"
      - "9411:9411"    # Zipkin compatible
    environment:
      COLLECTOR_ZIPKIN_HOST_PORT: ":9411"

  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      TRACING_ENABLED: "true"
      TRACING_ENDPOINT: "http://jaeger:14268/api/traces"
      TRACING_SERVICE_NAME: "api-gateway"
      TRACING_SAMPLE_RATE: "1.0"

      CACHE_ENABLED: "true"
      CACHE_TTL: "5m"
      CACHE_MAX_SIZE: "100"

      CANARY_ENABLED: "true"

      # Other configurations...
    depends_on:
      - jaeger
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-phase3-config
data:
  TRACING_ENABLED: "true"
  TRACING_ENDPOINT: "http://jaeger-collector.tracing:14268/api/traces"
  TRACING_SERVICE_NAME: "api-gateway"
  TRACING_SAMPLE_RATE: "0.1"
  TRACING_EXPORT_FORMAT: "jaeger"

  CACHE_ENABLED: "true"
  CACHE_TTL: "10m"
  CACHE_MAX_SIZE: "500"
  CACHE_EXCLUDE_PATHS: "/admin/*,/metrics,/health"
  CACHE_ONLY_METHODS: "GET,HEAD"
  CACHE_VARY_HEADERS: "Accept,Accept-Encoding"

  CANARY_ENABLED: "true"
  CANARY_DEFAULT_WEIGHT: "100"

  ADVANCED_ROUTING_ENABLED: "true"

---
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
        image: plumego-gateway:phase3
        ports:
        - containerPort: 8080
          name: http
        envFrom:
        - configMapRef:
            name: gateway-phase3-config
        - configMapRef:
            name: gateway-config  # Phase 1 & 2 configs
        resources:
          requests:
            memory: "256Mi"  # Increased for cache
            cpu: "100m"
          limits:
            memory: "1Gi"    # Account for cache size
            cpu: "1000m"
```

## Best Practices

### Distributed Tracing

1. **Sampling Strategy**
   - Development: 100% sampling (`TRACING_SAMPLE_RATE=1.0`)
   - Production: 10-20% sampling (`TRACING_SAMPLE_RATE=0.1`)
   - High-traffic: 1-5% sampling (`TRACING_SAMPLE_RATE=0.01`)

2. **Trace Propagation**
   - Always propagate trace IDs to downstream services
   - Include trace ID in application logs
   - Use consistent header names

3. **Performance**
   - Sampling reduces overhead
   - Async export to avoid blocking requests
   - Monitor tracing backend capacity

### Response Caching

1. **Cache Only Appropriate Content**
   - GET/HEAD requests only
   - Successful responses (2xx status codes)
   - Exclude user-specific or sensitive data

2. **TTL Selection**
   - Balance freshness vs hit rate
   - Static content: 15-30 minutes
   - Semi-dynamic: 5-10 minutes
   - Dynamic: 1-2 minutes

3. **Memory Management**
   - Monitor cache size and hit rate
   - Adjust `CACHE_MAX_SIZE` based on available memory
   - Typical: 10-20% of available memory

4. **Vary Headers**
   - Include headers that affect response content
   - Common: Accept, Accept-Encoding, Accept-Language
   - Avoid user-specific headers (Authorization, etc.)

### Canary Deployments

1. **Gradual Rollout**
   - Start with 1-5% traffic
   - Monitor metrics carefully
   - Increase by 10-20% increments
   - Full rollout at 100%

2. **Monitoring**
   - Compare error rates: canary vs primary
   - Compare latencies: p50, p95, p99
   - Monitor business metrics
   - Set up alerts for anomalies

3. **Rollback Plan**
   - Automated rollback on error rate increase
   - Manual rollback capability
   - Document rollback procedures

4. **Duration**
   - Each phase: 15-30 minutes minimum
   - Full rollout: 2-4 hours typical
   - Critical services: Extend duration

### Advanced Routing

1. **Rule Priority**
   - Higher priority for specific rules
   - Lower priority for catch-all rules
   - Test rule precedence thoroughly

2. **Performance**
   - Minimize number of rules
   - Cache rule evaluation results if possible
   - Profile rule matching performance

3. **Testing**
   - Test all routing paths
   - Verify header/query matching
   - Test fallback behavior

## Performance Considerations

### Memory Usage

| Feature | Memory Impact | Recommendation |
|---------|---------------|----------------|
| Tracing | Low (~1-5 MB) | Enable for all environments |
| Caching | High (up to configured limit) | Monitor carefully |
| Canary | Low (~1 MB) | Negligible |
| Advanced Routing | Low (~1-5 MB) | Negligible |

### CPU Usage

| Feature | CPU Impact | Optimization |
|---------|------------|--------------|
| Tracing | Low (1-2%) | Reduce sampling rate |
| Caching | Medium (5-10%) | Worth it for cache hits |
| Canary | Low (1-2%) | Optimize rule matching |
| Advanced Routing | Low (1-2%) | Minimize rule count |

### Latency Impact

| Feature | Latency Add | Notes |
|---------|-------------|-------|
| Tracing | <1ms | Negligible |
| Caching (hit) | <1ms | Actually reduces latency |
| Caching (miss) | 1-5ms | Worth it for hit rate |
| Canary | <1ms | Random selection is fast |
| Advanced Routing | 1-2ms | Depends on rule complexity |

## Monitoring

### Metrics to Track

**Tracing**
```promql
# Sampling rate
rate(traces_sampled_total[5m])

# Trace ID generation failures
rate(trace_id_generation_errors_total[5m])
```

**Caching**
```promql
# Cache hit rate
sum(rate(cache_hits_total[5m])) /
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))

# Cache size
cache_size_bytes

# Cache evictions
rate(cache_evictions_total[5m])
```

**Canary**
```promql
# Traffic distribution
rate(http_requests_total{canary_target="canary"}[5m]) /
rate(http_requests_total[5m])

# Error rate comparison
rate(http_requests_total{canary_target="canary",status=~"5.."}[5m])
rate(http_requests_total{canary_target="primary",status=~"5.."}[5m])
```

**Advanced Routing**
```promql
# Rule evaluations
rate(routing_rule_evaluations_total[5m])

# Rule matches
rate(routing_rule_matches_total[5m])
```

## Troubleshooting

### Tracing Issues

**Problem**: Traces not appearing in Jaeger

**Solutions**:
1. Verify `TRACING_ENDPOINT` is correct
2. Check sampling rate: `TRACING_SAMPLE_RATE`
3. Verify Jaeger collector is running
4. Check for network connectivity issues
5. Verify trace ID format is correct

```bash
# Test tracing endpoint
curl http://localhost:14268/api/traces

# Check trace headers
curl -v http://localhost:8080/api/v1/users
# Look for X-Trace-ID in response
```

### Cache Issues

**Problem**: Low cache hit rate

**Solutions**:
1. Check `CACHE_ONLY_METHODS` includes request methods
2. Verify paths aren't in `CACHE_EXCLUDE_PATHS`
3. Check TTL isn't too short
4. Verify vary headers are appropriate
5. Monitor cache evictions

```bash
# Check cache status
curl -I http://localhost:8080/api/v1/products
# Look for X-Cache: HIT or MISS
```

**Problem**: High memory usage

**Solutions**:
1. Reduce `CACHE_MAX_SIZE`
2. Decrease `CACHE_TTL`
3. Add more paths to `CACHE_EXCLUDE_PATHS`
4. Increase cache eviction frequency

### Canary Issues

**Problem**: Canary not receiving traffic

**Solutions**:
1. Verify `CANARY_ENABLED=true`
2. Check canary rule configuration
3. Verify `CanaryWeight` is > 0
4. Check header/cookie matching conditions
5. Verify canary targets are reachable

```bash
# Test canary routing
curl -H "X-Beta-User: true" http://localhost:8080/api/v1/users
# Check X-Canary-Target header in response
```

**Problem**: Uneven traffic distribution

**Solutions**:
1. Verify weight configuration
2. Check for sticky sessions interfering
3. Monitor over longer time period
4. Verify load balancer configuration

### Advanced Routing Issues

**Problem**: Rules not matching

**Solutions**:
1. Verify `ADVANCED_ROUTING_ENABLED=true`
2. Check rule priority order
3. Verify header/query values match exactly (case-sensitive)
4. Check path pattern matching
5. Test rules in isolation

```bash
# Test routing rule
curl -H "X-API-Version: v2" http://localhost:8080/api/users/123
# Check X-Target-Service header
```

## Integration Examples

### Example 1: Complete Observability Stack

```yaml
version: '3.8'

services:
  # Tracing
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # Collector

  # Metrics
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  # Gateway
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      # Phase 1: Metrics & Logging
      METRICS_ENABLED: "true"

      # Phase 2: Security
      AUTH_ENABLED: "true"
      SECURITY_HEADERS_ENABLED: "true"

      # Phase 3: Advanced
      TRACING_ENABLED: "true"
      TRACING_ENDPOINT: "http://jaeger:14268/api/traces"
      CACHE_ENABLED: "true"
      CANARY_ENABLED: "true"
```

### Example 2: Multi-Region Deployment

```bash
# Region 1: Primary
CANARY_ENABLED=false
CACHE_ENABLED=true
CACHE_MAX_SIZE=500

# Region 2: Canary
CANARY_ENABLED=true
CANARY_DEFAULT_WEIGHT=100
# Route 10% of Region 1 traffic here

# Both regions trace to central Jaeger
TRACING_ENDPOINT=http://central-jaeger:14268/api/traces
```

## Conclusion

Phase 3 transforms the API Gateway into a feature-complete enterprise solution:

- ✅ **Distributed Tracing** - Full request correlation
- ✅ **Response Caching** - Intelligent caching with LRU
- ✅ **Canary Deployments** - Safe gradual rollouts
- ✅ **Advanced Routing** - Flexible rule-based routing

Combined with Phase 1 (Observability) and Phase 2 (Security), the gateway now provides:
- Complete observability stack
- Enterprise security features
- Advanced traffic management
- Production-ready scalability

The API Gateway is now ready for high-scale production deployments with advanced operational capabilities.

## Next Steps

Future enhancements (Phase 4+):
- Circuit breakers and bulkheads
- GraphQL gateway support
- gRPC proxying
- WebSocket routing
- API composition/aggregation
- Request/response transformations
- Plugin system for custom logic
