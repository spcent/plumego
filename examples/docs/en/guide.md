# Plumego Module Playbook (English)

> This guide follows a modern technical-writing layout: every module comes with concepts, configuration checklists, operational tips, troubleshooting tables, and runnable snippets. Examples reuse the demo service in `examples/reference`, so you can run `go run .` and validate the behavior immediately.

## Core: lifecycle and application wiring

The Core package orchestrates the entire HTTP server lifecycle. It wires routers, middleware chains, Pub/Sub, metrics collectors, tracers, and webhook services into a predictable startup flow.

### Concepts and responsibilities

- **Explicit construction**: `core.New` receives the address, debug switch, Pub/Sub bus, Prometheus collector, and OpenTelemetry tracer up front, avoiding hidden globals.
- **Composable entry point**: Core does not contain business logic; it simply binds Router, Middleware, Metrics, and Webhook modules, keeping boundaries clear.
- **Validation before serving**: `Boot()` freezes the router and builds middleware stacks before listening, returning errors instead of silently ignoring misconfigurations.

### Configuration checklist

1. **Debug mode**: `core.WithDebug()` relaxes some guards and emits verbose logs; keep it disabled in production.
2. **Address and graceful shutdown**: override the default `:8080` with `core.WithAddr` or env vars. Core uses `http.Server` shutdown to drain in-flight requests on SIGTERM.
3. **Inbound webhooks**: configure GitHub/Stripe secrets, tolerated clock skew, and `MaxBodyBytes` through `core.WithWebhookIn` to protect memory and security budgets.
4. **Outbound webhooks**: `core.WithWebhookOut` enables the delivery service, including trigger token, base path, stats toggle, and pagination limits.
5. **Observability hooks**: plug `metrics.NewPrometheusCollector` and `metrics.NewOpenTelemetryTracer` into Core to enrich logging middleware automatically.

### Startup order and safety

- Core copies critical pointers at construction to prevent external mutation at runtime.
- Routes and middleware must be registered before `Boot()`; the router becomes immutable afterward and will panic on late registrations.
- Long-running goroutines (e.g., outbound webhook scheduler) should start before `Boot()` and be stopped with `defer` to align with HTTP lifecycle events.

### Troubleshooting

- **Port conflicts**: verify `WithAddr` values and use `netstat` to inspect listeners inside containers.
- **Missing secrets**: webhook signature mismatches often come from absent env vars; search for `signature mismatch` in logs.
- **Noisy output**: remove `WithDebug()` in production and drive verbosity from your logging backend.

### Example

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
    core.WithPubSub(pubsub.New()),
    core.WithMetricsCollector(metrics.NewPrometheusCollector("plumego_example")),
)
app.EnableRecovery()
app.EnableLogging()
if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

### Advanced practices

- **Multi-instance consistency**: keep initialization code identical across pods and push `WithAddr`, timeout values, and middleware toggles via configuration services to prevent drift.
- **Blue/green rollouts**: prebuild the full route tree before traffic cutover; the freeze-on-`Boot()` behavior prevents partially initialized handlers from receiving traffic.
- **Graceful shutdown choreography**: stop outbound webhook retries first, then close HTTP listeners, and finally drain Pub/Sub to avoid dropping internal events.
- **Configuration rollback**: log the resolved values of helpers like `envOr` so operators can confirm which defaults are active during rollback windows.

## Router: declarative routing and parameter handling

The Router package implements a prefix-tree matcher with support for static paths, named parameters, and wildcards. It keeps APIs predictable across code, documentation, and monitoring dashboards.

### Routing strategy

- **Prefer RESTful shapes**: paths like `GET /resources/:id` or `POST /resources` are cache-friendly and easy to group in metrics.
- **Named parameters**: `:id` values are injected into the request context; `*filepath` matches multi-level resources such as documentation trees.
- **Reusable groups**: `Group` composes prefixes and middleware, letting child routes focus solely on business logic.

### Debugging and freezing

- **Route visibility**: the router keeps a method-to-path map; export it in debug routes to audit what is actually registered.
- **Freeze guard**: once `Boot()` runs the router is frozen, and additional registrations panic—this catches initialization-order mistakes early.
- **Stackable middleware**: group-level middleware is inherited by child routes; handlers can append more layers to build deterministic execution order.

### Performance and maintainability

- The prefix tree avoids linear scans; keep dynamic segments shallow for faster lookups.
- For gray releases or throttling, attach middleware at the group level instead of branching inside handlers.
- Use descriptive verbs or actions in paths (`/disable`, `/resume`) so tracing and logs remain human-readable.

### Example

```go
api := app.Router().Group("/api")
api.Get("/users/:id", userHandler)
api.Post("/users", createUserHandler)
assets := app.Router().Group("/docs")
assets.Get("/*filepath", docsHandler)
```

### Advanced practices

- **Observable naming**: standardized path patterns (for example `/v1/users/:id/disable`) let dashboards aggregate by action or resource.
- **Boundary-driven grouping**: isolate internal admin routes from public APIs with separate Groups and attach IP or auth middleware at the boundary.
- **Debug export**: add a `/debug/routes` endpoint in dev to dump the `routes` registry, helping newcomers map the surface area quickly.
- **Profiling hotspots**: keep the first path segment static on high-traffic routes; split heavy handlers into preprocessing and postprocessing routes when necessary.

## Middleware: cross-cutting orchestration

Middleware powers recovery, structured logging, CORS, timeouts, compression, and custom guards. This section emphasizes order of execution, composition rules, safety notes, and customization patterns.

### Execution order

- Middleware is bound during route registration and executes in the sequence: **global → group → route-specific**.
- Recovery should sit near the outer layer to capture panics from downstream handlers.
- Logging and metrics middleware should wrap the stack to record complete request lifecycles.

### Customization pattern

1. **Respect the signature**: implement `middleware.Middleware` by wrapping and returning an `http.Handler`.
2. **Avoid global state**: capture configuration in closures or dedicated structs to prevent data races.
3. **Consistent errors**: return uniform JSON or text envelopes so monitoring tools can categorize failures reliably.

### Safety notes

- Lock down CORS origins instead of defaulting to `*`.
- Apply body-size or timeout middleware to upload endpoints to prevent memory exhaustion.
- Mask sensitive headers (Authorization, cookies) before writing logs.

### Example

```go
app.EnableRecovery()
app.EnableLogging()
app.EnableCORS()
app.Use(middleware.NewConcurrencyLimiter(100))
```

### Advanced practices

- **Composed stacks**: combine rate limits, auth, and audit middleware at the group level; isolate WebSocket or webhook paths into separate groups with tailored policies.
- **Error semantics**: return structured errors containing trace IDs and machine-readable codes so dashboards can group failures.
- **Defensive timeouts**: wrap downstream calls with `context.WithTimeout` and add headers indicating timeout decisions to guide client retries.
- **Security auditing**: log role, resource, and action from access-control middleware and forward to Pub/Sub for centralized compliance reporting.

## Pub/Sub and WebSocket: event-driven workflows

Plumego ships with an in-process Pub/Sub bus and a WebSocket hub, letting teams build lightweight real-time features without extra brokers.

### Design highlights

- **Topic prefixes**: adopt consistent prefixes such as `in.github.*`, `in.stripe.*`, or `ws.broadcast` to simplify subscription filters.
- **Concurrency safety**: the bus manages locks internally, but long blocking subscribers should still be avoided to keep latency predictable.
- **Heartbeat and cleanup**: WebSocket defaults include heartbeat intervals; tune `PingInterval` and `WriteWait` to match client expectations.

### Working with webhooks

Inbound webhooks can publish directly onto the bus, while WebSocket subscribers broadcast the same payload to browsers or dashboards. Outbound webhooks can consume internal events as well, enabling multi-target fan-out.

### Example

```go
bus := pubsub.New()
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(envOr("WS_SECRET", "dev-secret"))
_, _ = app.ConfigureWebSocketWithOptions(wsCfg)
bus.Subscribe("in.github.*", func(ctx context.Context, evt pubsub.Event) error {
    return bus.Publish(ctx, "ws.broadcast", evt.Payload)
})
```

### Advanced practices

- **Replay readiness**: add idempotency checks inside subscribers (for example, dedup by event ID) and persist critical events for later replay when consumers recover.
- **Broadcast vs. unicast**: include channel or user identifiers in payloads so the WebSocket hub can push selectively instead of broadcasting everything.
- **Timeout discipline**: wrap subscriber work with `context.WithTimeout`; surface errors so upstream monitors catch slow external dependencies.
- **Load testing**: simulate high-frequency webhook input to observe memory and goroutine counts, ensuring subscriber logic does not bottleneck the bus.

## Metrics and Health: observability by default

Observability is built in. Plumego integrates Prometheus metrics, OpenTelemetry tracing, and standard health probes with minimal setup.

### Metrics collection

- Create a collector with `metrics.NewPrometheusCollector(namespace)` and expose `/metrics` for scraping.
- Logging middleware automatically records latency, status codes, and panic counts—no extra boilerplate required.
- Distinguish instances with namespaces or custom labels in multi-replica deployments.

### Tracing

- Use `metrics.NewOpenTelemetryTracer(serviceName)` and let logging middleware inject trace IDs into logs.
- Start new spans at ingress points (e.g., webhook handlers) and propagate `context.Context` through business logic to preserve the trace.

### Health probes

- `health.ReadinessHandler()` should reflect dependencies such as databases or caches; wrap checks with timeouts to avoid probe stalls.
- `health.BuildInfoHandler()` reports build metadata for deployment comparisons.
- Combine with Kubernetes readiness and liveness probes to restart unhealthy pods automatically.

### Advanced practices

- **Layered metrics**: design prefixes for ingress, business, and external dependency metrics; tag routes, callers, and regions for richer slicing.
- **Error budgets**: use status distribution and P99 latency to enforce SLOs; throttle new rollouts when budgets approach exhaustion.
- **Probe extensions**: add database/cache checks to `/health/ready` with context deadlines so probes never dominate request processing time.
- **Tracing conventions**: standardize header keys and log formats for trace propagation across polyglot services.

## Security and Webhook: boundaries and signature verification

Security centers on webhook ingress/egress and request signing. The module offers sensible defaults plus hooks for audit requirements.

### Inbound security

- Configure shared secrets and allowed clock skew to prevent replay attacks for GitHub/Stripe webhooks.
- Use `MaxBodyBytes` to cap payload size and avoid memory pressure.
- Enable logging with care—record event metadata and signature verdicts, not raw payloads.

### Outbound security

- Protect trigger endpoints with `TriggerToken` and consider rate limiting on the caller side.
- Tune retry counts and backoff strategies in `webhookout.Config` so downstream outages do not exhaust worker pools.
- Swap `NewMemStore` with a persistent implementation when reliability requirements demand it.

### Audit and compliance

- Include trace IDs and request IDs in audit trails for sensitive actions.
- Introduce event allow-lists on the Pub/Sub layer to reject unexpected webhook types.
- Rotate secrets regularly and allow dual-acceptance windows during rotation to avoid downtime.

### Advanced practices

- **Signature drills**: craft forged requests in dev to confirm signature failures are logged with the right status codes and messages; test timestamp replay rejection.
- **Fault injection**: simulate 5xx responses or timeouts for outbound targets to validate backoff policies and ensure trigger endpoints do not accumulate stalled jobs.
- **Secret hygiene**: store secrets in a managed vault and pass them via env vars or mounted files rather than baking them into images.
- **Access tiers**: separate public, partner, and internal webhook prefixes with different trigger tokens to reduce accidental cross-traffic.

## Frontend and static assets: embedded UI and documentation delivery

The Frontend module serves embedded static assets through a consistent prefix. This section explains how to host UI bundles and Markdown docs (like this guide) alongside APIs.

### Embedding and hosting assets

- Use `//go:embed` to package `ui/*` or documentation directories into the binary, eliminating extra deployment steps.
- Call `frontend.RegisterFS(router, http.FS(staticFS), frontend.WithPrefix("/"))` to expose the assets as a static site.
- Place APIs under `/api` and static assets under `/` or `/docs` to avoid path collisions.

### Documentation navigation

- The example now exposes `/docs` with automatic language and module navigation; Markdown is rendered into clean HTML with safe escaping.
- Wildcard routing lets users jump directly to `/docs/zh/core` or `/docs/en/router` while keeping a shared navigation panel.
- The renderer provides base styling and escapes user content to avoid script injection.

### Operational tips

- Add Markdown linting to CI to keep `/docs` synchronized with releases (headings, links, examples).
- Customize the navigation template or CSS in the docs directory to brand the embedded site without running a separate doc portal.
- Include WebSocket playgrounds in embedded UI assets so readers can test real-time flows described in the docs.
- Set sensible cache headers for assets and rotate filenames (hash or version prefix) to invalidate outdated caches after updates.
