# API Gateway Configuration

The API Gateway supports configuration via environment variables, making it easy to deploy in different environments.

## Configuration Files

- `.env.example` - Example configuration file with all available options
- `.env` - Your local configuration (copy from `.env.example`)

## Quick Start

1. Copy the example configuration:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` to match your environment:
   ```bash
   # Server configuration
   GATEWAY_ADDR=:8080
   GATEWAY_DEBUG=false

   # Enable/disable services
   USER_SERVICE_ENABLED=true
   USER_SERVICE_TARGETS=http://localhost:8081,http://localhost:8082
   ```

3. Run the gateway:
   ```bash
   cd server
   go run .
   ```

## Configuration Options

### Server Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_ADDR` | `:8080` | Server listen address |
| `GATEWAY_DEBUG` | `false` | Enable debug mode |

### CORS Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ENABLED` | `true` | Enable/disable CORS |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated list of allowed origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | `Content-Type,Authorization` | Allowed headers |

### User Service

| Variable | Default | Description |
|----------|---------|-------------|
| `USER_SERVICE_ENABLED` | `true` | Enable user service proxy |
| `USER_SERVICE_TARGETS` | `http://localhost:8081,http://localhost:8082` | Backend targets (comma-separated) |
| `USER_SERVICE_PATH_PREFIX` | `/api/v1/users` | Path prefix for routing |
| `USER_SERVICE_LB` | `round-robin` | Load balancer algorithm |
| `USER_SERVICE_HEALTH_CHECK` | `false` | Enable health checking |

### Order Service

| Variable | Default | Description |
|----------|---------|-------------|
| `ORDER_SERVICE_ENABLED` | `true` | Enable order service proxy |
| `ORDER_SERVICE_TARGETS` | `http://localhost:9001,http://localhost:9002` | Backend targets |
| `ORDER_SERVICE_PATH_PREFIX` | `/api/v1/orders` | Path prefix |
| `ORDER_SERVICE_LB` | `weighted-round-robin` | Load balancer algorithm |
| `ORDER_SERVICE_HEALTH_CHECK` | `true` | Enable health checking |

### Product Service

| Variable | Default | Description |
|----------|---------|-------------|
| `PRODUCT_SERVICE_ENABLED` | `true` | Enable product service proxy |
| `PRODUCT_SERVICE_TARGETS` | `http://localhost:7001` | Backend targets |
| `PRODUCT_SERVICE_PATH_PREFIX` | `/api/v1/products` | Path prefix |
| `PRODUCT_SERVICE_LB` | `round-robin` | Load balancer algorithm |
| `PRODUCT_SERVICE_HEALTH_CHECK` | `false` | Enable health checking |

## Load Balancer Algorithms

Available load balancer algorithms:

- `round-robin` - Distribute requests evenly across all backends
- `weighted-round-robin` - Distribute based on weights (WIP)
- `least-connections` - Send to backend with fewest active connections
- `ip-hash` - Consistent hashing based on client IP

## Examples

### Disable a Service

```bash
# In .env
PRODUCT_SERVICE_ENABLED=false
```

### Add More Backend Targets

```bash
# In .env
USER_SERVICE_TARGETS=http://backend-1:8081,http://backend-2:8082,http://backend-3:8083
```

### Change Load Balancer

```bash
# In .env
USER_SERVICE_LB=least-connections
```

### Enable Health Checks

```bash
# In .env
USER_SERVICE_HEALTH_CHECK=true
```

### Production Configuration

```bash
# .env.production
GATEWAY_ADDR=:80
GATEWAY_DEBUG=false

CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

USER_SERVICE_ENABLED=true
USER_SERVICE_TARGETS=http://user-service-1:8081,http://user-service-2:8081
USER_SERVICE_LB=least-connections
USER_SERVICE_HEALTH_CHECK=true

ORDER_SERVICE_ENABLED=true
ORDER_SERVICE_TARGETS=http://order-service-1:9001,http://order-service-2:9001
ORDER_SERVICE_LB=weighted-round-robin
ORDER_SERVICE_HEALTH_CHECK=true

PRODUCT_SERVICE_ENABLED=true
PRODUCT_SERVICE_TARGETS=http://product-service:7001
PRODUCT_SERVICE_LB=round-robin
PRODUCT_SERVICE_HEALTH_CHECK=true
```

## Docker Deployment

You can also use environment variables directly with Docker:

```bash
docker run -p 8080:8080 \
  -e GATEWAY_ADDR=:8080 \
  -e USER_SERVICE_TARGETS=http://user-api:8081,http://user-api:8082 \
  -e ORDER_SERVICE_TARGETS=http://order-api:9001 \
  plumego-gateway
```

Or with docker-compose:

```yaml
services:
  gateway:
    image: plumego-gateway
    ports:
      - "8080:8080"
    environment:
      GATEWAY_ADDR: ":8080"
      GATEWAY_DEBUG: "false"
      USER_SERVICE_TARGETS: "http://user-service:8081"
      ORDER_SERVICE_TARGETS: "http://order-service:9001"
      PRODUCT_SERVICE_TARGETS: "http://product-service:7001"
```

## Validation

The gateway will validate the configuration on startup and fail fast if:

- Server address is empty
- No services are enabled
- Enabled services have no targets configured

Check the startup logs for configuration details:

```
Starting API Gateway with configuration:
  Server: :8080 (Debug: false)
  CORS: true
  User Service: true (2 targets)
  Order Service: true (2 targets)
  Product Service: true (1 targets)
Registering User Service: /api/v1/users -> [http://localhost:8081 http://localhost:8082]
Registering Order Service: /api/v1/orders -> [http://localhost:9001 http://localhost:9002] (with service discovery)
Registering Product Service: /api/v1/products -> [http://localhost:7001]
```
