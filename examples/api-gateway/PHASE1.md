# Phase 1: Production-Ready Features

This document describes the Phase 1 production features implemented for the API Gateway.

## ✅ Implemented Features

### 1. Prometheus Metrics (Observability)

**What it does:**
- Exposes metrics endpoint compatible with Prometheus
- Tracks HTTP requests, latency, and status codes
- Per-route metrics tracking
- Built-in Prometheus exposition format

**Configuration:**
```bash
METRICS_ENABLED=true
METRICS_PATH=/metrics
METRICS_NAMESPACE=api_gateway
```

**Usage:**
```bash
# Access metrics
curl http://localhost:8080/metrics

# Prometheus scrape config
scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

**Metrics Exposed:**
- `api_gateway_http_requests_total{method,path,status}` - Total HTTP requests
- `api_gateway_http_request_duration_seconds{method,path}` - Request duration histogram
- `api_gateway_http_request_duration_seconds_sum` - Sum of all request durations
- `api_gateway_http_request_duration_seconds_count` - Count of requests

### 2. Structured Access Logging

**What it does:**
- Logs every request with detailed information
- Structured JSON format
- Integration with Prometheus metrics

**Logged Fields:**
- `method` - HTTP method
- `path` - Request path
- `query` - Query string
- `status` - HTTP status code
- `duration_ms` - Request duration in milliseconds
- `remote_addr` - Client IP address
- `user_agent` - Client user agent
- `bytes_written` - Response size in bytes

**Example Log Output:**
```json
{
  "level": "info",
  "msg": "access",
  "method": "GET",
  "path": "/api/v1/users/123",
  "query": "",
  "status": 200,
  "duration_ms": 45,
  "remote_addr": "127.0.0.1:54321",
  "user_agent": "curl/7.79.1",
  "bytes_written": 256
}
```

### 3. Gateway-Level Rate Limiting

**What it does:**
- Protects backend services from traffic spikes
- Token bucket algorithm implementation
- Configurable rate and burst size
- Returns 429 Too Many Requests when exceeded

**Configuration:**
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=1000        # Requests per second
RATE_LIMIT_BURST=2000      # Burst size
```

**Behavior:**
- Allows sustained rate of `RATE_LIMIT_RPS` requests/second
- Can handle bursts up to `RATE_LIMIT_BURST` requests
- When limit exceeded:
  - Returns HTTP 429
  - Adds `Retry-After: 1` header
  - Adds `X-RateLimit-Limit` header
  - Adds `X-RateLimit-Remaining: 0` header

**Example Response:**
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
Retry-After: 1

Rate limit exceeded
```

### 4. Enhanced Timeout Configuration

**What it does:**
- Gateway-level timeout for all requests
- Per-service timeout configuration
- Prevents hung requests
- Configurable via environment variables

**Configuration:**
```bash
# Gateway-level timeout (applied to all requests)
TIMEOUT_GATEWAY=30s

# Default service timeout
TIMEOUT_SERVICE=10s

# Per-service timeouts (override default)
USER_SERVICE_TIMEOUT=10s
ORDER_SERVICE_TIMEOUT=15s
PRODUCT_SERVICE_TIMEOUT=10s
```

**Duration Format:**
Supports both duration strings and seconds:
- `30s` - 30 seconds
- `1m` - 1 minute
- `30` - 30 seconds (integer)

**Behavior:**
- Gateway timeout is applied first (outermost layer)
- Service timeout is applied to backend calls
- If gateway timeout expires: returns 504 Gateway Timeout
- Service timeout must be ≤ gateway timeout (validated)

### 5. Retry Configuration

**What it does:**
- Configurable retry count per service
- Automatic retry on transient failures
- Exponential backoff between retries

**Configuration:**
```bash
USER_SERVICE_RETRY=3      # Retry up to 3 times
ORDER_SERVICE_RETRY=2     # Retry up to 2 times
PRODUCT_SERVICE_RETRY=3   # Retry up to 3 times
```

**Behavior:**
- Retries only on network errors or 5xx responses
- Uses exponential backoff: 0ms, 100ms, 200ms, 400ms...
- Respects circuit breaker if enabled
- Total attempts = 1 + retry_count

### 6. Enhanced Configuration Validation

**What it validates:**

**Server Configuration:**
- Server address is not empty

**Metrics Configuration:**
- If metrics enabled, path and namespace must be set

**Rate Limit Configuration:**
- Requests per second must be positive
- Burst size must be positive
- Burst size should be ≥ requests per second

**Timeout Configuration:**
- Gateway timeout must be positive
- Service timeout must be positive
- Service timeout must not exceed gateway timeout

**Service Configuration:**
- At least one service must be enabled
- Enabled services must have at least one target
- Target URLs must start with http:// or https://
- Service timeouts must be positive
- Service timeouts must not exceed gateway timeout
- Retry counts must be non-negative

**Example Validation Error:**
```
Invalid configuration: order service timeout exceeds gateway timeout
```

## Configuration Summary

### New Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| **Metrics** | | |
| `METRICS_ENABLED` | `true` | Enable Prometheus metrics |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `METRICS_NAMESPACE` | `api_gateway` | Prometheus namespace |
| **Rate Limiting** | | |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_RPS` | `1000` | Requests per second |
| `RATE_LIMIT_BURST` | `2000` | Burst size |
| **Timeouts** | | |
| `TIMEOUT_GATEWAY` | `30s` | Gateway timeout |
| `TIMEOUT_SERVICE` | `10s` | Default service timeout |
| **Per-Service** | | |
| `{SERVICE}_TIMEOUT` | `10s` | Service-specific timeout |
| `{SERVICE}_RETRY` | `3` | Service-specific retry count |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Client Request                     │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Access Logging Middleware                   │
│  (logs request start, records metrics)               │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Rate Limiting Middleware                    │
│  (checks token bucket, returns 429 if exceeded)      │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Gateway Timeout Middleware                  │
│  (enforces gateway-level timeout)                    │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          CORS Middleware                             │
│  (adds CORS headers if enabled)                      │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Router (match route)                        │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Proxy Handler                               │
│  - Load balancing                                    │
│  - Service timeout                                   │
│  - Retries with backoff                              │
│  - Health checking                                   │
│  - Service discovery                                 │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│          Backend Service                             │
└─────────────────────────────────────────────────────┘
```

## Monitoring & Observability

### Prometheus Metrics

**Scrape the metrics endpoint:**
```bash
curl http://localhost:8080/metrics
```

**Sample Metrics Output:**
```
# HELP api_gateway_http_requests_total Total HTTP requests
# TYPE api_gateway_http_requests_total counter
api_gateway_http_requests_total{method="GET",path="/api/v1/users",status="200"} 1523

# HELP api_gateway_http_request_duration_seconds HTTP request duration
# TYPE api_gateway_http_request_duration_seconds histogram
api_gateway_http_request_duration_seconds_bucket{method="GET",path="/api/v1/users",le="0.01"} 1200
api_gateway_http_request_duration_seconds_bucket{method="GET",path="/api/v1/users",le="0.05"} 1480
api_gateway_http_request_duration_seconds_bucket{method="GET",path="/api/v1/users",le="0.1"} 1523
api_gateway_http_request_duration_seconds_sum{method="GET",path="/api/v1/users"} 45.67
api_gateway_http_request_duration_seconds_count{method="GET",path="/api/v1/users"} 1523
```

### Grafana Dashboard

**Key Metrics to Monitor:**
1. **Request Rate** - Requests per second
2. **Error Rate** - 5xx responses
3. **Latency** - P50, P95, P99 response times
4. **Rate Limit Hits** - 429 responses
5. **Timeout Errors** - 504 responses

**Sample PromQL Queries:**
```promql
# Request rate
rate(api_gateway_http_requests_total[5m])

# Error rate
rate(api_gateway_http_requests_total{status=~"5.."}[5m])

# P95 latency
histogram_quantile(0.95, rate(api_gateway_http_request_duration_seconds_bucket[5m]))

# Rate limit hits
rate(api_gateway_http_requests_total{status="429"}[5m])
```

## Testing

### Load Testing

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Test rate limiting
hey -n 10000 -c 100 http://localhost:8080/api/v1/users

# Expected: Some requests will get 429 responses
```

### Metrics Verification

```bash
# Check metrics endpoint
curl http://localhost:8080/metrics | grep api_gateway

# Check health
curl http://localhost:8080/health
```

### Timeout Testing

```bash
# Test gateway timeout (should timeout after 30s)
curl -i http://localhost:8080/api/v1/slow-endpoint

# Expected: 504 Gateway Timeout after configured timeout
```

## Production Deployment

### Docker Compose Example

```yaml
version: '3.8'

services:
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      # Server
      GATEWAY_ADDR: ":8080"
      GATEWAY_DEBUG: "false"

      # Metrics
      METRICS_ENABLED: "true"
      METRICS_PATH: "/metrics"
      METRICS_NAMESPACE: "api_gateway"

      # Rate Limiting
      RATE_LIMIT_ENABLED: "true"
      RATE_LIMIT_RPS: "5000"
      RATE_LIMIT_BURST: "10000"

      # Timeouts
      TIMEOUT_GATEWAY: "30s"
      TIMEOUT_SERVICE: "10s"

      # Services
      USER_SERVICE_TARGETS: "http://user-api:8081"
      ORDER_SERVICE_TARGETS: "http://order-api:9001"
      PRODUCT_SERVICE_TARGETS: "http://product-api:7001"

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['api-gateway:8080']
    metrics_path: '/metrics'
```

## Next Steps: Phase 2

Phase 2 will add:
1. JWT Authentication
2. Distributed Tracing (OpenTelemetry)
3. TLS/mTLS Support
4. Hot Configuration Reload
5. Admin API for runtime management

See the main implementation plan for details.
