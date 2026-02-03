# Multi-Tenant SaaS Example

This example demonstrates a complete multi-tenant SaaS application built with Plumego's tenant features.

## Features

- **Tenant Management**: Admin API for creating, updating, and deleting tenants
- **Quota Enforcement**: Per-tenant request and token rate limiting
- **Policy Controls**: Per-tenant allowed models and tools
- **Database Isolation**: Automatic tenant filtering for all database queries
- **Analytics**: Per-tenant request analytics and monitoring
- **Caching**: LRU cache for tenant configurations with TTL

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Requests                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
         ┌─────────────────────────────┐
         │  Tenant Middleware Chain    │
         │  1. Resolve tenant ID       │
         │  2. Check quota limits      │
         │  3. Validate policies       │
         └──────────┬──────────────────┘
                    │
         ┌──────────▼──────────────────┐
         │   Business Logic Handlers   │
         │   - User Management         │
         │   - Analytics               │
         │   - ... other endpoints     │
         └──────────┬──────────────────┘
                    │
         ┌──────────▼──────────────────┐
         │   TenantDB (auto-filtering) │
         │   All queries scoped to     │
         │   current tenant            │
         └──────────┬──────────────────┘
                    │
         ┌──────────▼──────────────────┐
         │      SQLite Database        │
         │   - tenants (config)        │
         │   - users (tenant-scoped)   │
         │   - api_requests (logs)     │
         └─────────────────────────────┘
```

## Quick Start

### 1. Run the application

```bash
cd examples/multi-tenant-saas
go run .
```

The server will start on `http://localhost:8080`.

### 2. Create a tenant

```bash
curl -X POST http://localhost:8080/admin/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "quota_requests_per_minute": 100,
    "quota_tokens_per_minute": 10000,
    "allowed_models": ["gpt-4", "gpt-3.5-turbo"],
    "allowed_tools": ["search", "calculator"],
    "metadata": {
      "company": "Acme Corporation",
      "plan": "enterprise"
    }
  }'
```

### 3. Create a user (tenant-scoped)

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: acme-corp" \
  -d '{
    "email": "john@acme.com",
    "name": "John Doe"
  }'
```

### 4. List users (tenant-scoped)

```bash
curl http://localhost:8080/api/users \
  -H "X-Tenant-ID: acme-corp"
```

### 5. Get analytics (tenant-scoped)

```bash
curl http://localhost:8080/api/analytics/requests \
  -H "X-Tenant-ID: acme-corp"
```

## API Endpoints

### Admin Endpoints (No tenant required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | `/admin/tenants` | Create a new tenant |
| GET    | `/admin/tenants/:id` | Get tenant configuration |
| PUT    | `/admin/tenants/:id` | Update tenant configuration |
| DELETE | `/admin/tenants/:id` | Delete a tenant |
| GET    | `/admin/tenants` | List all tenants |

### Business API Endpoints (Requires `X-Tenant-ID` header)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/api/users` | List users for current tenant |
| POST   | `/api/users` | Create a user for current tenant |
| GET    | `/api/users/:id` | Get user by ID (tenant-scoped) |
| DELETE | `/api/users/:id` | Delete user (tenant-scoped) |
| GET    | `/api/analytics/requests` | Get request analytics |

### Health Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/health` | Health check |
| GET    | `/` | Service information |

## Quota Enforcement

Tenants are rate-limited based on their quota configuration:

```json
{
  "quota_requests_per_minute": 100,
  "quota_tokens_per_minute": 10000
}
```

When a tenant exceeds their quota:
- Returns `429 Too Many Requests`
- Includes `Retry-After` header with seconds to wait
- Resets at the start of each minute

## Policy Controls

Tenants can be restricted to specific models and tools:

```json
{
  "allowed_models": ["gpt-4", "gpt-3.5-turbo"],
  "allowed_tools": ["search", "calculator"]
}
```

Requests using disallowed models/tools:
- Returns `403 Forbidden`
- Includes reason in response body

## Database Isolation

The `TenantDB` wrapper automatically filters all queries by tenant ID:

```go
// Your query
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT * FROM users WHERE active = ?", true)

// Automatically becomes
"SELECT * FROM users WHERE tenant_id = 'acme-corp' AND active = true"
```

This prevents:
- Cross-tenant data leaks
- Accidental data exposure
- SQL injection attacks via tenant ID

## Configuration

Environment variables:

```bash
APP_ADDR=:8080              # Server address
DB_PATH=./tenants.db        # SQLite database path
```

## Testing Multi-Tenancy

### Create two tenants

```bash
# Tenant 1: High quota
curl -X POST http://localhost:8080/admin/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "enterprise",
    "quota_requests_per_minute": 1000,
    "quota_tokens_per_minute": 100000
  }'

# Tenant 2: Low quota
curl -X POST http://localhost:8080/admin/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "starter",
    "quota_requests_per_minute": 10,
    "quota_tokens_per_minute": 1000
  }'
```

### Create users for each tenant

```bash
# Enterprise tenant user
curl -X POST http://localhost:8080/api/users \
  -H "X-Tenant-ID: enterprise" \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@enterprise.com", "name": "Alice"}'

# Starter tenant user
curl -X POST http://localhost:8080/api/users \
  -H "X-Tenant-ID: starter" \
  -H "Content-Type: application/json" \
  -d '{"email": "bob@starter.com", "name": "Bob"}'
```

### Verify isolation

```bash
# Enterprise tenant sees only their user
curl http://localhost:8080/api/users -H "X-Tenant-ID: enterprise"
# Returns: [{"email": "alice@enterprise.com", ...}]

# Starter tenant sees only their user
curl http://localhost:8080/api/users -H "X-Tenant-ID: starter"
# Returns: [{"email": "bob@starter.com", ...}]
```

### Test quota enforcement

```bash
# Exhaust starter tenant's quota (10 req/min)
for i in {1..15}; do
  curl http://localhost:8080/api/users -H "X-Tenant-ID: starter"
done

# Last 5 requests will return 429 Too Many Requests
```

## Production Considerations

### Security

1. **Admin Authentication**: Add authentication to admin endpoints
2. **HTTPS**: Use TLS in production
3. **API Keys**: Replace header-based tenant ID with signed tokens
4. **Input Validation**: Validate all user inputs
5. **Rate Limiting**: Add per-IP rate limiting

### Performance

1. **Database**: Use PostgreSQL or MySQL for production
2. **Caching**: Tune cache size and TTL based on tenant count
3. **Connection Pooling**: Configure database connection pool
4. **Horizontal Scaling**: Run multiple instances behind load balancer

### Monitoring

1. **Metrics**: Export tenant-specific metrics to Prometheus
2. **Logging**: Centralize logs with tenant ID tagging
3. **Alerting**: Set up alerts for quota exceeded events
4. **Tracing**: Add distributed tracing for request flows

## Next Steps

- Add JWT authentication for API endpoints
- Implement tenant-specific feature flags
- Add billing integration based on usage metrics
- Implement tenant data export/import
- Add WebSocket support with tenant isolation
- Implement sliding window quotas for smoother rate limiting

## License

This example is part of the Plumego project and uses the same license.
