# Caching Layers

Plumego has five distinct caching layers. Each serves a specific purpose and is
intentionally kept separate. Picking the wrong layer leads to redundant
implementations or missed capabilities — this doc shows which layer to reach
for and why.

## Decision table

| Layer | Package | Purpose | When to use |
|---|---|---|---|
| **In-process primitive** | `store/cache` | TTL-bounded, memory-bounded in-process cache | Caching anything inside a single process: computed results, config values, parsed tokens |
| **Distributed / Redis** | `x/data/cache` | Consistent-hashing distributed cache, Redis adapter, leaderboard | Multi-instance deployments that need a shared cache tier |
| **HTTP response** | `x/gateway/cache` | Keyed HTTP response cache for the reverse proxy | Gateway-level response caching; vary by method, URL, or header |
| **LLM exact-key** | `x/ai/llmcache` | Exact-key AI response cache | Deduplicating identical LLM requests (same prompt, same params) |
| **LLM semantic** | `x/ai/semanticcache` | Embedding-based similarity cache built on llmcache | Deduplicating near-duplicate prompts across vector distance |

## Layer details

### `store/cache` — in-process primitive

```
store/cache.MemoryCache
```

- stdlib-only, no external dependencies
- TTL expiration, memory-bounded eviction
- Thread-safe
- **Do not** use for shared state across processes or instances

### `x/data/cache` — distributed and Redis

```
x/data/cache/distributed   consistent-hashing cluster
x/data/cache/redis         store/cache.Cache adapter for Redis
x/data/cache/leaderboard   ranked-data cache on top of store/cache
```

- Builds on the `store/cache.Cache` interface
- Brings external dependencies (Redis client, consistent-hash library)
- Start with `store/cache` first; switch to this layer only when multi-instance
  sharing is required

### `x/gateway/cache` — HTTP response cache

```
x/gateway/cache.Cache
x/gateway/cache.Store
```

- HTTP-level caching, lives inside the gateway reverse proxy
- Keyed by request method + URL + configurable response headers
- **Not** a general-purpose data cache; scoped to gateway middleware only

### `x/ai/llmcache` — LLM exact-key cache

```
x/ai/llmcache.Cache
```

- Stores AI provider responses keyed by exact prompt hash
- Used internally by `x/ai/semanticcache` for fallback lookups
- Forbidden from importing `x/data/**` to keep the AI family isolated

### `x/ai/semanticcache` — LLM semantic cache

```
x/ai/semanticcache.Cache
x/ai/semanticcache/embedding    embedding provider interface
x/ai/semanticcache/vectorstore  vector similarity backend interface
x/ai/semanticcache/cachemanager lifecycle management
```

- Calls an embedding model to convert prompts to vectors, then searches for
  similar cached responses above a configurable similarity threshold
- Falls back to `llmcache` for exact-key matches
- Requires an external embedding provider and vector store

## What does NOT belong here

- **Tenant-aware caching**: lives in `x/tenant/store/cache`
- **Feature-level counters or rollout rates**: those are metrics, not cache;
  see `x/observability/featuremetrics`
- **Session storage**: use `store/kv` or `x/data/kvengine`

## Import rules summary

```
store/cache       ← no external deps; usable anywhere
x/data/cache      ← imports store/cache; never import from x/ai or x/tenant
x/gateway/cache   ← imports x/gateway only; never from x/data or x/ai
x/ai/llmcache     ← imports x/ai only; forbidden from x/data/**
x/ai/semanticcache← imports x/ai/llmcache; uses embedding and vector backends
```

See `specs/dependency-rules.yaml` for the machine-checked version of these
constraints.
