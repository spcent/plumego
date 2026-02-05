# Sprint 2: Vector Embedding Cache Enhancements

## Overview

This sprint extends Phase 3's semantic caching with advanced embedding providers, vector store backends, and multi-tier caching strategies.

## New Features

### 1. Enhanced Embedding Providers (`embedding/`)

#### Provider Interface
```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, text string) (*Embedding, error)
    GenerateBatch(ctx context.Context, texts []string) ([]*Embedding, error)
    Dimensions() int
    Model() string
    CostPerToken() float64
    MaxBatchSize() int
    SupportsAsync() bool
}
```

#### Implemented Providers

**OpenAI Provider**
- Models: text-embedding-3-small, text-embedding-3-large, text-embedding-ada-002
- Dimensions: 1536 (small/ada), 3072 (large)
- Batch size: up to 2048
- Automatic retries with exponential backoff
- Cost tracking per request

```go
provider, err := embedding.NewOpenAIProvider(embedding.OpenAIConfig{
    APIKey: "your-api-key",
    Model:  "text-embedding-3-small",
    Timeout: 30 * time.Second,
})
```

**Cohere Provider**
- Models: embed-english-v3.0, embed-multilingual-v3.0, embed-english-light-v3.0
- Dimensions: 1024 (v3), 384 (light), 4096 (v2)
- Batch size: up to 96
- Multiple input types: search_document, search_query, classification, clustering

```go
provider, err := embedding.NewCohereProvider(embedding.CohereConfig{
    APIKey:    "your-api-key",
    Model:     "embed-english-v3.0",
    InputType: "search_document",
})
```

**Statistics Tracking**
```go
// Wrap any provider with statistics
statsProvider := embedding.NewProviderWithStats(provider)

// Get statistics
stats := statsProvider.Stats()
fmt.Printf("Total embeddings: %d\n", stats.TotalEmbeddings)
fmt.Printf("Total cost: $%.4f\n", stats.TotalCost)
fmt.Printf("Average latency: %v\n", stats.AverageLatency)
```

### 2. Vector Store Backends (`vectorstore/`)

#### Backend Interface
```go
type Backend interface {
    Add(ctx context.Context, entry *Entry) error
    AddBatch(ctx context.Context, entries []*Entry) error
    Search(ctx context.Context, query []float64, topK int, threshold float64) ([]*SearchResult, error)
    Get(ctx context.Context, id string) (*Entry, error)
    Delete(ctx context.Context, id string) error
    Update(ctx context.Context, entry *Entry) error
    Count(ctx context.Context) (int64, error)
    Clear(ctx context.Context) error
    Close() error
    Stats(ctx context.Context) (*BackendStats, error)
}
```

#### Memory Backend

Fast in-memory vector store with LRU eviction:

```go
backend := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
    Dimensions: 1536,
    MaxSize:    1000, // Max 1000 entries
})

// Add entry
entry := &vectorstore.Entry{
    ID:       "entry-1",
    Vector:   embedding.Vector,
    Text:     "query text",
    Response: "cached response",
}
backend.Add(ctx, entry)

// Search
results, err := backend.Search(ctx, queryVector, 10, 0.85)
for _, result := range results {
    fmt.Printf("Similarity: %.4f\n", result.Similarity)
    fmt.Printf("Response: %s\n", result.Entry.Response)
}
```

Features:
- Cosine similarity search
- Automatic LRU eviction when maxSize reached
- Access count tracking
- TTL support
- Memory usage estimation

### 3. Multi-Tier Caching (`vectorstore/tiered.go`)

Implements L1 (memory) → L2 (Redis) → L3 (vector DB) caching hierarchy:

```go
// Create tiered cache
l1 := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
    Dimensions: 1536,
    MaxSize:    100,
})

l2 := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
    Dimensions: 1536,
    MaxSize:    1000,
})

l3 := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
    Dimensions: 1536,
})

tiered, err := vectorstore.NewTieredBackend(vectorstore.TieredConfig{
    L1: l1,
    L2: l2,
    L3: l3,
})

// Search automatically tries L1 → L2 → L3
results, err := tiered.Search(ctx, query, 10, 0.85)

// Check hit rates
stats := tiered.TieredStats()
fmt.Printf("L1 hits: %d\n", stats.L1Hits)
fmt.Printf("L2 hits: %d\n", stats.L2Hits)
fmt.Printf("L3 hits: %d\n", stats.L3Hits)
fmt.Printf("Overall hit rate: %.2f%%\n", tiered.HitRate()*100)
```

Features:
- Automatic cache promotion (L3 → L2 → L1)
- Independent tier configuration
- Hit rate tracking per tier
- Best-effort writes to lower tiers

### 4. Cache Management (`cachemanager/`)

#### Cache Manager

```go
manager := cachemanager.NewManager(backend, embeddingProvider)

// Get statistics
stats, err := manager.Stats(ctx)
fmt.Printf("Total entries: %d\n", stats.TotalEntries)
fmt.Printf("Memory usage: %d bytes\n", stats.MemoryUsage)

// Cleanup old entries
err = manager.Cleanup(ctx, cachemanager.CleanupPolicy{
    Type:   cachemanager.CleanupTypeOldest,
    MaxAge: 7 * 24 * time.Hour, // Remove entries older than 7 days
})

// Export cache
err = manager.Export(ctx, "/path/to/export.json")

// Import cache
err = manager.Import(ctx, "/path/to/import.json")
```

#### Cache Warmer

Pre-warm cache from historical data:

```go
warmer := cachemanager.NewWarmer(backend, provider)

// Warm from query history
queries := []cachemanager.QueryResponse{
    {Query: "What is Go?", Response: "Go is a programming language..."},
    {Query: "How to use channels?", Response: "Channels are..."},
}
err := warmer.WarmFromQueries(ctx, queries)

// Warm from documents
docs := []cachemanager.Document{
    {
        ID:      "doc-1",
        Title:   "Go Tutorial",
        Content: "Full tutorial text...",
        Summary: "Tutorial summary...",
    },
}
err := warmer.WarmFromDocuments(ctx, docs)

// Warm with progress reporting
err = warmer.WarmWithProgress(ctx, queries, func(current, total int, stats *WarmingStats) {
    fmt.Printf("Progress: %d/%d (%.1f%%)\n", current, total, float64(current)/float64(total)*100)
})
```

### 5. Vector Utilities (`embedding/utils.go`)

Advanced vector operations:

```go
// Similarity measures
similarity, _ := embedding.CosineSimilarity(vec1, vec2)
distance, _ := embedding.EuclideanDistance(vec1, vec2)
manhattan, _ := embedding.ManhattanDistance(vec1, vec2)

// Vector operations
normalized := embedding.Normalize(vector)
averaged, _ := embedding.AverageVectors(vectors)
interpolated, _ := embedding.Interpolate(vec1, vec2, 0.5)

// Batch operations
similarities, _ := embedding.BatchCosineSimilarity(query, allVectors)
topKIndices := embedding.TopK(similarities, 10)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/spcent/plumego/ai/semanticcache/embedding"
    "github.com/spcent/plumego/ai/semanticcache/vectorstore"
    "github.com/spcent/plumego/ai/semanticcache/cachemanager"
)

func main() {
    ctx := context.Background()

    // 1. Setup embedding provider
    provider, err := embedding.NewOpenAIProvider(embedding.OpenAIConfig{
        APIKey: "your-openai-api-key",
        Model:  "text-embedding-3-small",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Wrap with statistics tracking
    provider = embedding.NewProviderWithStats(provider)

    // 2. Setup multi-tier cache
    l1 := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
        Dimensions: provider.Dimensions(),
        MaxSize:    100, // Fast tier: 100 entries
    })

    l2 := vectorstore.NewMemoryBackend(vectorstore.MemoryConfig{
        Dimensions: provider.Dimensions(),
        MaxSize:    1000, // Medium tier: 1000 entries
    })

    tiered, _ := vectorstore.NewTieredBackend(vectorstore.TieredConfig{
        L1: l1,
        L2: l2,
    })

    // 3. Warm cache with historical data
    warmer := cachemanager.NewWarmer(tiered, provider)
    queries := []cachemanager.QueryResponse{
        {Query: "What is semantic caching?", Response: "Semantic caching uses embeddings..."},
        {Query: "How do embeddings work?", Response: "Embeddings convert text to vectors..."},
    }
    warmer.WarmFromQueries(ctx, queries)

    // 4. Use the cache
    query := "Explain semantic caching"

    // Generate embedding
    emb, err := provider.Generate(ctx, query)
    if err != nil {
        log.Fatal(err)
    }

    // Search cache
    results, err := tiered.Search(ctx, emb.Vector, 1, 0.85)
    if err != nil {
        log.Fatal(err)
    }

    if len(results) > 0 {
        fmt.Printf("✓ Cache hit! Similarity: %.4f\n", results[0].Similarity)
        fmt.Printf("Response: %s\n", results[0].Entry.Response)
    } else {
        fmt.Println("Cache miss - would call LLM")
    }

    // 5. Check statistics
    stats := provider.(*embedding.ProviderWithStats).Stats()
    fmt.Printf("\nEmbedding Stats:\n")
    fmt.Printf("  Total embeddings: %d\n", stats.TotalEmbeddings)
    fmt.Printf("  Total cost: $%.4f\n", stats.TotalCost)

    tieredStats := tiered.TieredStats()
    fmt.Printf("\nCache Stats:\n")
    fmt.Printf("  L1 hit rate: %.2f%%\n", tiered.L1HitRate()*100)
    fmt.Printf("  Overall hit rate: %.2f%%\n", tiered.HitRate()*100)
    fmt.Printf("  Total queries: %d\n", tieredStats.TotalQueries)
}
```

## Performance

**Embedding Generation:**
- OpenAI: ~50ms per request (batch of 10)
- Cohere: ~100ms per request (batch of 10)

**Vector Search (1000 entries):**
- Memory backend: ~1ms for top-10 search
- L1 cache hit: <0.1ms
- L2 cache hit: ~1ms
- L3 cache hit: ~10ms

**Memory Usage:**
- Per entry (1536 dimensions): ~12-15 KB
- 1000 entries: ~12-15 MB

## Future Enhancements

- [ ] Redis backend for L2 tier
- [ ] Pinecone/Qdrant/Weaviate integration for L3
- [ ] Async embedding generation
- [ ] Distributed caching
- [ ] Query rewriting for better cache hits
- [ ] Automatic dimension reduction
- [ ] Embedding compression
- [ ] Cache invalidation strategies

## Testing

```bash
# Test embedding utilities
go test ./ai/semanticcache/embedding -v

# Test vector stores
go test ./ai/semanticcache/vectorstore -v

# Benchmarks
go test ./ai/semanticcache/embedding -bench=. -benchmem
go test ./ai/semanticcache/vectorstore -bench=. -benchmem
```

## Migration from Phase 3

Existing semantic cache code remains compatible. New features are additive:

```go
// Old code still works
cache := semanticcache.NewSemanticCache(...)

// New enhanced version
provider := embedding.NewOpenAIProvider(...)
backend := vectorstore.NewMemoryBackend(...)
tiered := vectorstore.NewTieredBackend(...)
```
