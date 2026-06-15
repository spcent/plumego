# Which Extension Should I Use?

Plumego has 15 extension families covering different needs. This guide helps you choose the right one(s) for your service.

## Decision Tree

Start at the top and follow the path that matches your use case.

```
Do you need non-HTTP functionality?
├─ YES → Real-time bidirectional communication?
│  ├─ YES → x/websocket (beta)
│  └─ NO → Other protocol?
│     ├─ gRPC → x/rpc (experimental)
│     └─ Other → Check docs/modules/
│
├─ HTTP API extensions?
│  ├─ REST with CRUD conventions → x/rest (beta)
│  ├─ Multi-tenant isolation → x/tenant (beta)
│  ├─ API gateway or reverse proxy → x/gateway (beta)
│  └─ Static/embedded asset serving → x/frontend (beta)
│
├─ Message/event processing?
│  └─ x/messaging (beta) for queues, pub/sub, scheduling, webhooks
│
├─ Observability?
│  └─ x/observability (beta) for Prometheus, OpenTelemetry, tracing
│
└─ Advanced/specialized features?
   ├─ AI/LLM providers → x/ai (experimental, stable subpackages exist)
   ├─ Database topology/caching → x/data (experimental, stable subpackages exist)
   ├─ File upload/download → x/fileapi (experimental)
   ├─ Circuit breaker/rate limiting → x/resilience (experimental)
   ├─ Request validation → x/validate (experimental)
   └─ OpenAPI spec generation → x/openapi (experimental)
```

## Beta Extensions (Production-Ready)

These are safe for production adoption. APIs are stable across release references with deprecation notices for any breaking changes.

### `x/rest` — REST CRUD Conventions

**What it does:** Provides resource controllers and standard REST route patterns.

**When to use:** Building JSON APIs with standard GET, POST, PUT, DELETE patterns.

**Example:** 
```go
users := app.Group("/users")
users.Get("", listUsers)
users.Post("", createUser)
users.Get("/{id}", getUser)
users.Put("/{id}", updateUser)
users.Delete("/{id}", deleteUser)
```

**Learn more:** `reference/with-rest/` + `docs/modules/x/rest/README.md`

### `x/websocket` — Real-Time Communication

**What it does:** WebSocket hub with explicit route registration.

**When to use:** Real-time features like chat, notifications, live updates.

**Example:**
```go
hub := websocket.NewHub()
app.Get("/ws", hub.Upgrade(handleMessage))
```

**Learn more:** `reference/with-websocket/` + `docs/modules/x/websocket/README.md`

### `x/gateway` — API Gateway & Reverse Proxy

**What it does:** Proxy, rewrite, balancing, and edge transport.

**When to use:** API gateway, reverse proxy, load balancing, request rewriting.

**Example:**
```go
gw := gateway.New(
    gateway.Route("/api/*", "http://backend1:8080"),
    gateway.Route("/admin/*", "http://backend2:8080"),
)
app.Get("/*", gw.Handler())
```

**Learn more:** `reference/with-gateway/` + `docs/modules/x/gateway/README.md`

### `x/observability` — Metrics & Tracing

**What it does:** Prometheus exporters, OpenTelemetry tracers, collectors, and adapters.

**When to use:** Exporting metrics to Prometheus or Datadog, distributed tracing.

**Example:**
```go
exporter := observability.NewPrometheusExporter()
defer exporter.Export()
app.Use(exporter.Middleware())
```

**Learn more:** `reference/with-observability/` + `docs/modules/x/observability/README.md`

### `x/tenant` — Multi-Tenancy

**What it does:** Tenant resolution, policy, quota, rate limiting, session, isolation.

**When to use:** Multi-tenant SaaS APIs with per-tenant isolation.

**Example:**
```go
app.Use(tenant.Resolver(tenantFromHeader))
app.Use(tenant.IsolationMiddleware())

// In handlers:
tenantID := tenant.IDFromContext(r)
```

**Learn more:** `reference/with-tenant/` + `docs/modules/x/tenant/README.md`

### `x/frontend` — Static Assets

**What it does:** Serve static assets, embedded with the binary, optional compression.

**When to use:** Serving frontend code (JavaScript, CSS, HTML) from the same binary.

**Example:**
```go
app.Get("/*", frontend.Handler(fs))
```

**Learn more:** `reference/with-frontend/` + `docs/modules/x/frontend/README.md`

### `x/messaging` — Queues & Pub/Sub

**What it does:** App-facing messaging service for queues, pub/sub, scheduling, webhooks.

**When to use:** Asynchronous task processing, event publishing, webhook delivery.

**Example:**
```go
messaging.Publish("user.created", userData)
messaging.Subscribe("user.created", handleUserCreated)
```

**Learn more:** `reference/with-messaging/` + `docs/modules/x/messaging/README.md`

## Experimental Extensions (APIs May Change)

These are suitable for exploration and learning, but may change in minor versions. Use in production only with explicit acceptance of the risk.

### `x/ai` — LLM Providers & Agents

**What it does:** LLM provider integration, sessions, streaming, tools.

**When to use:** AI-powered features (chatbots, code generation, etc.).

**Note:** Root family is experimental, but subpackages `provider`, `session`, `streaming`, and `tool` have beta evidence.

**Learn more:** `docs/modules/x/ai/README.md` (start here for the family overview)

### `x/data` — Storage & Topology

**What it does:** Storage topologies, caching layers, migrations, sharding.

**When to use:** Complex database patterns, multi-tier caching.

**Note:** Root family is experimental, but subpackages `file` and `idempotency` have beta evidence.

**Learn more:** `docs/modules/x/data/README.md`

### `x/fileapi` — HTTP File Transport

**What it does:** HTTP endpoints for file upload, download, and temporary URLs.

**When to use:** File operations over HTTP.

**Learn more:** `docs/modules/x/fileapi/README.md`

### `x/resilience` — Fault Tolerance Primitives

**What it does:** Circuit breaker, rate limiter, retry logic.

**When to use:** Protecting against cascading failures, rate limiting external APIs.

**Learn more:** `docs/modules/x/resilience/README.md`

### `x/rpc` — gRPC Integration

**What it does:** gRPC server, client, HTTP-gRPC gateway.

**When to use:** gRPC services, mixed gRPC/HTTP deployments.

**Learn more:** `docs/modules/x/rpc/README.md`

### `x/validate` — Request Validation

**What it does:** Validation helpers and binding utilities.

**When to use:** Complex request validation beyond basic type checking.

**Learn more:** `docs/modules/x/validate/README.md`

### `x/openapi` — API Documentation

**What it does:** OpenAPI 3.1 spec generation from route metadata.

**When to use:** Auto-generating API docs, OpenAPI compliance.

**Learn more:** `docs/modules/x/openapi/README.md`

## Combining Extensions

You can combine extensions in the same service. Here are common patterns:

### REST API with Observability

```
✓ Use: x/rest + x/observability
├─ x/rest for CRUD routes
├─ x/observability for Prometheus metrics
```

**Example:** `reference/with-rest/` can be extended with `reference/with-observability/`

### Multi-Tenant REST API

```
✓ Use: x/rest + x/tenant + x/observability
├─ x/rest for CRUD routes
├─ x/tenant for isolation
├─ x/observability for tracking per-tenant metrics
```

**Example:** Combine patterns from `reference/with-rest/`, `reference/with-tenant/`, `reference/with-observability/`

### Real-Time SaaS Platform

```
✓ Use: x/rest + x/websocket + x/tenant + x/observability
├─ x/rest for API
├─ x/websocket for real-time notifications
├─ x/tenant for multi-tenancy
├─ x/observability for monitoring
```

### API Gateway with Multiple Backends

```
✓ Use: x/gateway + x/observability + x/resilience
├─ x/gateway for proxying
├─ x/resilience for circuit breaker (optional)
├─ x/observability for tracing requests across services
```

### AI-Powered Service

```
✓ Use: x/ai/provider + x/ai/session + x/observability
├─ x/ai/provider for LLM calls (beta surface)
├─ x/ai/session for conversation state (beta surface)
├─ x/observability for token usage metrics
```

## Avoiding Wrong Choices

### Don't use...

- **`x/*` for core service needs** — The 9 stable roots handle config, logging, routing, errors, health, etc. Extensions are for specialized capabilities, not core wiring.

- **Multiple similar extensions** — Don't use both `x/gateway` and a third-party proxy library. Pick one.

- **Experimental extensions in production without acceptance** — If you need `x/ai` in production, document the risk and have a migration plan.

- **Extensions to replace stable root functionality** — For example, don't add an external validation library instead of using manual validation in handlers. Keep it explicit.

## Still Unsure?

1. **Read the quick start**: `docs/start/POSITIONING.md` explains why Plumego exists
2. **Check the adoption path**: `docs/start/adoption-path.md` has a 30-minute step to add one extension
3. **Look at reference apps**: Each `reference/with-*` shows how to use an extension
4. **Check maturity**: `docs/concepts/extension-maturity.md` shows what's stable vs experimental
5. **Ask in the community**: Open an issue or discussion if you're stuck

## Next Steps

1. Choose one extension family that matches your immediate need
2. Copy the corresponding `reference/with-*` directory
3. Read the extension's README in `docs/modules/x/*/`
4. Add one route or handler using the extension
5. Test it locally
6. Once it works, add another extension if needed

---

For extension maturity levels, see `docs/concepts/extension-maturity.md`.  
For the stable learning path, see `docs/start/adoption-path.md`.  
For all modules, see `docs/modules/INDEX.md`.
