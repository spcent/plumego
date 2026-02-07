# Plumego Documentation Plan

## Existing Documentation Status

| File / Directory | Status | Quality | Action |
|---|---|---|---|
| README.md (437 lines) | exists | good | minor update — add doc site links once docs land |
| CLAUDE.md / AGENTS.md | exists | good | keep as-is (AI-assistant facing, not user docs) |
| docs/*.md (32 files) | exists | internal-only | keep — design notes, not user-facing; no action |
| examples/docs/en/ (~25 files, ~6.9K lines) | exists | moderate | restructure into unified docs/ tree |
| examples/docs/zh/ (~35 files) | exists | moderate | keep in sync after en/ restructure |
| examples/ (19 apps) | exists | no READMEs | add per-example README.md stubs |

## Proposed Documentation Structure

```
docs/
├── getting-started.md          # P0 — 5-min tutorial
├── core-concepts.md            # P0 — lifecycle, components, options
├── api-reference/
│   ├── app.md                  # P0 — core.New, Boot, options
│   ├── router.md               # P0 — routes, groups, params
│   ├── context.md              # P1 — request/response helpers
│   ├── middleware.md            # P1 — chain, registry, built-ins
│   ├── errors.md               # P1 — contract errors, categories
│   └── response.md             # P2 — response helpers
├── guides/
│   ├── routing.md              # P0 — routing patterns, groups, static
│   ├── middleware.md            # P1 — writing custom middleware
│   ├── error-handling.md       # P1 — structured errors, WriteError
│   ├── configuration.md        # P1 — env vars, .env, config manager
│   ├── security.md             # P1 — JWT, passwords, headers, abuse
│   ├── multi-tenancy.md        # P1 — tenant setup, quota, policy, DB
│   └── testing.md              # P2 — test patterns, mocking
└── advanced/
    ├── ai-gateway.md           # P2 — LLM providers, sessions, SSE
    ├── websocket.md            # P2 — hub, JWT auth, broadcast
    ├── scheduler.md            # P2 — cron, delayed jobs, retries
    ├── pubsub-webhooks.md      # P2 — pub/sub, inbound/outbound hooks
    ├── observability.md        # P2 — metrics, tracing, health
    └── store.md                # P2 — KV, cache, DB wrapper, file
```

## Documentation Tasks (Prioritized)

### P0 — Must Have (Core Understanding)

| Document | Purpose | Key Sections | Lines | Tokens | Deps |
|---|---|---|---|---|---|
| getting-started.md | 5-min tutorial | Install, Hello World, Add Route, Run | 100-120 | 8-10K | None |
| core-concepts.md | Architecture overview | Lifecycle, Components, Options, Handlers | 120-150 | 10-12K | None |
| api-reference/app.md | App creation & boot | New(), WithX options, Boot(), Handler | 100-130 | 8-10K | core-concepts |
| api-reference/router.md | Route registration | Methods, params, groups, reverse routing | 100-120 | 8-10K | app.md |
| guides/routing.md | Routing patterns | Static, param, group, middleware per-group | 80-100 | 6-8K | router.md |

### P1 — Should Have (Full API Coverage)

| Document | Purpose | Key Sections | Lines | Tokens | Deps |
|---|---|---|---|---|---|
| api-reference/context.md | Request/response API | Bind, JSON, Query, Param, Header, Stream | 100-120 | 8-10K | router.md |
| api-reference/middleware.md | Middleware system | Chain, Registry, built-in list (19 pkgs) | 120-140 | 10-12K | context.md |
| api-reference/errors.md | Error types | Categories, NewError, WriteError, options | 80-100 | 6-8K | context.md |
| guides/middleware.md | Writing middleware | Signature, chain order, testing middleware | 80-100 | 6-8K | middleware.md |
| guides/error-handling.md | Error patterns | Validation, auth, rate limit, custom errors | 80-100 | 6-8K | errors.md |
| guides/configuration.md | Config system | Env vars, .env, LoadEnv, ConfigManager | 80-100 | 6-8K | core-concepts |
| guides/security.md | Security features | JWT, password, headers, abuse guard, input | 100-120 | 8-10K | middleware.md |
| guides/multi-tenancy.md | Tenant isolation | Config, quota, policy, TenantDB, hooks | 120-140 | 10-12K | security.md |

### P2 — Nice to Have (Advanced & Polish)

| Document | Purpose | Key Sections | Lines | Tokens | Deps |
|---|---|---|---|---|---|
| api-reference/response.md | Response helpers | WriteResponse, WriteError, JSON, Stream | 60-80 | 5-6K | context.md |
| guides/testing.md | Test patterns | Table-driven, ResetForTesting, mock clock | 60-80 | 5-6K | P1 done |
| advanced/ai-gateway.md | AI integration | Providers, sessions, SSE, tools, cache | 120-150 | 10-12K | P1 done |
| advanced/websocket.md | WebSocket hub | Hub setup, JWT, rooms, broadcast | 80-100 | 6-8K | security.md |
| advanced/scheduler.md | Task scheduling | Cron, delay, retry, persistence, admin | 80-100 | 6-8K | core-concepts |
| advanced/pubsub-webhooks.md | Events & webhooks | PubSub, GitHub/Stripe inbound, outbound | 100-120 | 8-10K | core-concepts |
| advanced/observability.md | Monitoring | Prometheus, OpenTelemetry, health probes | 80-100 | 6-8K | P1 done |
| advanced/store.md | Data persistence | KV with WAL, cache, TenantDB, file store | 100-120 | 8-10K | core-concepts |

## Code Examples Required

| Example | Purpose | Used In | Complexity | Lines |
|---|---|---|---|---|
| Hello World | Simplest app | README, getting-started | Simple | 15-20 |
| Route with params | Path parameters | router.md, routing guide | Simple | 20-25 |
| Route groups | Grouped routes with middleware | routing guide | Simple | 25-30 |
| JSON API endpoint | REST handler with Context | context.md, getting-started | Medium | 30-40 |
| Custom middleware | Writing a middleware func | middleware guide | Medium | 25-35 |
| Middleware chain | Auth + logging + recovery | middleware.md | Medium | 35-45 |
| Structured errors | Error categories and responses | error-handling guide | Medium | 30-40 |
| Tenant setup | Multi-tenant configuration | multi-tenancy guide | Complex | 50-70 |
| Scheduler jobs | Cron + delayed tasks | scheduler.md | Medium | 30-40 |
| AI provider | LLM streaming endpoint | ai-gateway.md | Complex | 60-80 |

## Module Documentation Order

1. **App** (core/) — Foundation; create app, configure, boot. Depends on: nothing
2. **Router** (router/) — Define routes, groups, params. Depends on: App
3. **Context** (contract/) — Read requests, write responses. Depends on: Router
4. **Middleware** (middleware/) — Extend pipeline. Depends on: Router, Context
5. **Errors** (contract/) — Structured error handling. Depends on: Context
6. **Config** (config/) — Environment, .env loading. Depends on: App
7. **Security** (security/) — JWT, passwords, headers. Depends on: Middleware
8. **Tenant** (tenant/) — Multi-tenancy. Depends on: Security, Config
9. **Scheduler** (scheduler/) — Background tasks. Depends on: App
10. **AI Gateway** (ai/) — LLM integration. Depends on: Middleware, Security

## Effort Estimation

| Priority | Files | Lines | Tokens | Scope |
|---|---|---|---|---|
| P0 | 5 | 500-620 | 40-50K | Get users started |
| P1 | 8 | 760-920 | 60-76K | Full API coverage |
| P2 | 8 | 680-850 | 55-68K | Advanced topics |
| **Total** | **21** | **1,940-2,390** | **155-194K** | |

Estimated generation cost: ~$0.97 (at $5/M input tokens, assuming 194K tokens)

## Key Decisions

1. **Style**: Technical, concise, code-heavy — match Go ecosystem conventions
2. **Audience**: Intermediate Go developers familiar with net/http
3. **Examples**: Realistic (compilable snippets, not pseudocode)
4. **Location**: Unified under docs/ (restructure away from examples/docs/)
5. **Existing en/ docs**: Migrate usable content; rewrite where outdated
6. **Chinese docs**: Update zh/ in a follow-up pass after en/ stabilizes
