# Semantic Cache

> **Package**: `github.com/spcent/plumego/ai/semanticcache` | **Feature**: Embedding-based caching

## Overview

The `semanticcache` package provides embedding-based semantic caching for LLM responses. Unlike exact-match caching, semantic caching finds similar (not just identical) queries and returns cached responses — dramatically reducing API costs and latency.

**How it works**:
1. Query is embedded into a vector representation
2. Vector is compared against cached embeddings using cosine similarity
3. If similarity exceeds threshold, cached response is returned
4. Otherwise, LLM is called and response is cached with its embedding

## Quick Start

```go
import "github.com/spcent/plumego/ai/semanticcache"

// Create semantic cache
cache := semanticcache.New(
    semanticcache.WithEmbeddingProvider(embeddingProvider),
    semanticcache.WithVectorStore(vectorStore),
    semanticcache.WithSimilarityThreshold(0.92), // 92% similarity required
    semanticcache.WithTTL(24 * time.Hour),
)

// Wrap provider with caching
cachedClaude := semanticcache.Wrap(claude, cache)

// Use exactly like regular provider - caching is transparent
resp, err := cachedClaude.Complete(ctx, req)
```

## Embedding Providers

### OpenAI Embeddings

```go
import "github.com/spcent/plumego/ai/semanticcache/embedding"

embedder := embedding.NewOpenAI(
    embedding.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    embedding.WithModel("text-embedding-3-small"), // 1536 dimensions, low cost
)
```

### Cohere Embeddings

```go
embedder := embedding.NewCohere(
    embedding.WithAPIKey(os.Getenv("COHERE_API_KEY")),
    embedding.WithModel("embed-english-v3.0"),
)
```

### Local Embeddings

```go
// Use local model for privacy-sensitive data
embedder := embedding.NewLocal(
    embedding.WithModelPath("./models/sentence-transformers"),
)
```

## Vector Stores

### In-Memory (Development)

```go
vectorStore := semanticcache.NewInMemoryVectorStore()
```

### Redis with Vector Search

```go
vectorStore := semanticcache.NewRedisVectorStore(
    os.Getenv("REDIS_URL"),
    semanticcache.WithIndexName("ai-cache"),
    semanticcache.WithDimensions(1536), // Match embedding dimensions
)
```

### PostgreSQL with pgvector

```go
vectorStore := semanticcache.NewPGVectorStore(
    db,
    semanticcache.WithTableName("ai_cache_vectors"),
    semanticcache.WithDimensions(1536),
)
```

## Configuration

```go
cache := semanticcache.New(
    // Embedding provider (required)
    semanticcache.WithEmbeddingProvider(embedder),

    // Vector store (required)
    semanticcache.WithVectorStore(vectorStore),

    // Similarity threshold (0.0 - 1.0)
    // Higher = more precise matching, fewer cache hits
    // Lower  = more aggressive caching, potential inaccuracies
    semanticcache.WithSimilarityThreshold(0.92),

    // Cache TTL
    semanticcache.WithTTL(24 * time.Hour),

    // Maximum cached entries
    semanticcache.WithMaxEntries(10000),

    // Minimum query length to cache (skip very short queries)
    semanticcache.WithMinQueryLength(10),

    // Exclude certain queries from caching
    semanticcache.WithExcludePatterns([]string{
        "random", "generate a unique", "today's date",
    }),
)
```

## Cache Key Strategy

Semantic caches must account for context beyond the query text:

```go
cache := semanticcache.New(
    // Include system prompt in cache key
    semanticcache.WithSystemPromptHashing(true),

    // Include model in cache key (different models, different responses)
    semanticcache.WithModelAwareness(true),

    // Custom key function
    semanticcache.WithKeyFunc(func(req *provider.CompletionRequest) string {
        // Include temperature in key (deterministic vs creative requests)
        if req.Temperature > 0.5 {
            return "" // Don't cache non-deterministic requests
        }
        return fmt.Sprintf("%s:%s", req.Model, lastMessage(req.Messages))
    }),
)
```

## Exact-Match Cache (LLM Cache)

For pure exact-match caching without embeddings:

```go
import "github.com/spcent/plumego/ai/llmcache"

// Exact match only — faster but fewer cache hits
exactCache := llmcache.New(
    llmcache.WithStorage(redisStore),
    llmcache.WithTTL(1 * time.Hour),
    llmcache.WithKeyFunc(func(req *provider.CompletionRequest) string {
        h := sha256.New()
        json.NewEncoder(h).Encode(req)
        return hex.EncodeToString(h.Sum(nil))
    }),
)

cachedClaude := llmcache.Wrap(claude, exactCache)
```

## Cache Metrics

```go
stats := cache.Stats()

fmt.Printf("Hit rate: %.1f%%\n", stats.HitRate*100)
fmt.Printf("Total hits: %d\n", stats.Hits)
fmt.Printf("Total misses: %d\n", stats.Misses)
fmt.Printf("Avg similarity: %.3f\n", stats.AvgSimilarity)
fmt.Printf("Tokens saved: %d\n", stats.TokensSaved)
fmt.Printf("Cost saved: $%.4f\n", stats.CostSaved)

// Prometheus metrics
cache_hits_total{cache="semantic"}
cache_misses_total{cache="semantic"}
cache_similarity_score_histogram
cache_tokens_saved_total
```

## When to Use Semantic Caching

**Good use cases:**
- Customer support bots (similar questions → same answers)
- Documentation Q&A (stable knowledge base)
- Code explanation (same patterns → same explanations)
- FAQ systems

**Poor use cases:**
- Real-time data queries ("current stock price")
- Personalized responses (user-specific context)
- Creative generation (stories, poems — variety expected)
- Sensitive/private queries

## Cache Invalidation

```go
// Invalidate specific entry
cache.Delete(ctx, queryEmbedding)

// Clear all entries for a topic
cache.DeleteByTag(ctx, "product-docs")

// Full cache clear (use sparingly)
cache.Clear(ctx)
```

## Related Documentation

- [Provider Interface](provider.md) — Wrapped provider pattern
- [Tokenizer](tokenizer.md) — Track tokens saved by cache
- [Rate Limiting](rate-limit.md) — Combined with caching for cost control
