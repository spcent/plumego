# P2 Implementation Plan - API Gateway Advanced Features

> **Version**: v1.0.0-rc.1
> **Status**: Planning Phase
> **Dependencies**: Zero (Go stdlib only)
> **Approach**: Contract-based extensibility

---

## Overview

This document outlines the P2 (Priority 2) feature implementation for plumego's API gateway capabilities. All features follow the **zero external dependencies** principle and use **contract-based architecture** for protocol extensions.

### Design Principles

1. **No Third-Party Dependencies**: All core implementations use only Go standard library
2. **Contract-Based Extensibility**: Protocol support (gRPC, GraphQL) via interface contracts
3. **User-Provided Adapters**: Users implement adapters using their chosen libraries
4. **Backward Compatibility**: All additions are non-breaking
5. **Performance First**: Minimal overhead, efficient resource usage

---

## P2 Features Summary

| Feature | Priority | Complexity | Dependencies | Status |
|---------|----------|------------|--------------|--------|
| Response Caching | P2-A | Medium | stdlib only | Planned |
| API Version Negotiation | P2-B | Low | stdlib only | Planned |
| Request Coalescing | P2-C | High | stdlib only | Planned |
| Protocol Adapter Contracts | P2-D | Medium | contracts only | Planned |
| Request/Response Transformation | P2-E | Medium | stdlib only | Planned |
| Rate Limiting (Advanced) | P2-F | Medium | stdlib only | Planned |

---

## P2-A: Response Caching Middleware

### Objective
Implement HTTP response caching to reduce backend load and improve latency.

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ GET /api/users/123
       ▼
┌─────────────────────────────┐
│  Cache Middleware           │
│  1. Generate cache key      │
│  2. Check cache (hit/miss)  │
│  3. Return or forward       │
└──────┬──────────────────────┘
       │ (cache miss)
       ▼
┌─────────────────────────────┐
│  Proxy Middleware           │
│  Forward to backend         │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────┐
│   Backend   │
└─────────────┘
```

### Implementation Plan

**File**: `middleware/cache/http_cache.go`

```go
package cache

import (
    "bytes"
    "hash/fnv"
    "io"
    "net/http"
    "sync"
    "time"
)

// Cache key strategy
type KeyStrategy interface {
    Generate(r *http.Request) string
}

// Cache store interface (allows LRU, Redis adapter, etc.)
type Store interface {
    Get(key string) (*CachedResponse, bool)
    Set(key string, resp *CachedResponse, ttl time.Duration)
    Delete(key string)
    Clear()
}

// Cached response
type CachedResponse struct {
    StatusCode int
    Header     http.Header
    Body       []byte
    CachedAt   time.Time
    ExpiresAt  time.Time
}

// Cache configuration
type Config struct {
    // Key strategy (default: URL + Method + Accept header)
    KeyStrategy KeyStrategy

    // Cache store (default: in-memory LRU)
    Store Store

    // Default TTL
    DefaultTTL time.Duration

    // Max cacheable size (default: 1MB)
    MaxSize int64

    // Cacheable methods (default: GET, HEAD)
    Methods []string

    // Cacheable status codes (default: 200, 301, 404)
    StatusCodes []int

    // Cache control parser
    RespectCacheControl bool

    // Conditional requests (ETag, Last-Modified)
    EnableConditionalRequests bool
}

// Middleware function
func Middleware(config Config) func(http.Handler) http.Handler {
    // Implementation
}
```

**File**: `middleware/cache/store_memory.go`

```go
package cache

import (
    "container/list"
    "sync"
    "time"
)

// In-memory LRU cache using stdlib only
type MemoryStore struct {
    capacity int
    items    map[string]*list.Element
    evictList *list.List
    mu       sync.RWMutex
}

type entry struct {
    key   string
    value *CachedResponse
}

func NewMemoryStore(capacity int) *MemoryStore {
    return &MemoryStore{
        capacity:  capacity,
        items:     make(map[string]*list.Element),
        evictList: list.New(),
    }
}

func (s *MemoryStore) Get(key string) (*CachedResponse, bool) {
    // LRU get with expiration check
}

func (s *MemoryStore) Set(key string, resp *CachedResponse, ttl time.Duration) {
    // LRU set with eviction
}
```

**File**: `middleware/cache/key_strategy.go`

```go
package cache

// Default key: Method + URL + Accept + Accept-Encoding
type DefaultKeyStrategy struct{}

func (s *DefaultKeyStrategy) Generate(r *http.Request) string {
    h := fnv.New64a()
    h.Write([]byte(r.Method))
    h.Write([]byte(r.URL.String()))
    h.Write([]byte(r.Header.Get("Accept")))
    h.Write([]byte(r.Header.Get("Accept-Encoding")))
    return fmt.Sprintf("%x", h.Sum64())
}

// Custom key strategy example
type CustomKeyStrategy struct {
    IncludeHeaders []string
}
```

**Integration Example**:

```go
app := core.New(
    core.WithAddr(":8080"),
)

// Add caching before proxy
app.Use(cache.Middleware(cache.Config{
    DefaultTTL:  5 * time.Minute,
    MaxSize:     1 << 20, // 1MB
    Methods:     []string{"GET", "HEAD"},
    StatusCodes: []int{200, 301, 404},
    RespectCacheControl: true,
}))

// Then proxy
app.Use("/api/*", proxy.New(proxy.Config{
    Targets: []string{"http://backend:8080"},
}))
```

### Testing Plan

- [ ] Cache hit/miss scenarios
- [ ] TTL expiration
- [ ] LRU eviction
- [ ] Cache-Control header parsing
- [ ] Conditional requests (ETag, If-None-Match)
- [ ] Large response handling (> MaxSize)
- [ ] Concurrent access (race detection)

---

## P2-B: API Version Negotiation

### Objective
Support multiple API versions through content negotiation and routing.

### Architecture

```
Client Request:
  GET /users/123
  Accept: application/vnd.myapi.v2+json

Version Middleware:
  1. Parse Accept header
  2. Extract version (v2)
  3. Add to context
  4. Route to versioned handler
```

### Implementation Plan

**File**: `middleware/versioning/version.go`

```go
package versioning

import (
    "context"
    "net/http"
    "regexp"
    "strconv"
    "strings"
)

type Strategy int

const (
    // Accept header: application/vnd.myapi.v2+json
    StrategyAcceptHeader Strategy = iota

    // URL path: /v2/users/123
    StrategyURLPath

    // Query param: /users/123?version=2
    StrategyQueryParam

    // Custom header: X-API-Version: 2
    StrategyCustomHeader
)

type Config struct {
    Strategy       Strategy
    DefaultVersion int

    // For Accept header strategy
    VendorPrefix string // e.g., "application/vnd.myapi"

    // For URL path strategy
    PathPrefix string // e.g., "/v"

    // For query param strategy
    ParamName string // e.g., "version"

    // For custom header strategy
    HeaderName string // e.g., "X-API-Version"

    // Version validator
    SupportedVersions []int
}

type contextKey string

const versionContextKey contextKey = "api_version"

func Middleware(config Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            version := extractVersion(r, config)

            // Validate version
            if !isSupportedVersion(version, config.SupportedVersions) {
                http.Error(w, "Unsupported API version", http.StatusNotAcceptable)
                return
            }

            // Add to context
            ctx := context.WithValue(r.Context(), versionContextKey, version)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetVersion(ctx context.Context) int {
    if v, ok := ctx.Value(versionContextKey).(int); ok {
        return v
    }
    return 0
}
```

**File**: `middleware/versioning/extractor.go`

```go
package versioning

func extractVersion(r *http.Request, config Config) int {
    switch config.Strategy {
    case StrategyAcceptHeader:
        return extractFromAccept(r, config.VendorPrefix)
    case StrategyURLPath:
        return extractFromPath(r, config.PathPrefix)
    case StrategyQueryParam:
        return extractFromQuery(r, config.ParamName)
    case StrategyCustomHeader:
        return extractFromHeader(r, config.HeaderName)
    }
    return config.DefaultVersion
}

func extractFromAccept(r *http.Request, vendorPrefix string) int {
    // Parse: application/vnd.myapi.v2+json -> 2
    accept := r.Header.Get("Accept")
    re := regexp.MustCompile(vendorPrefix + `\.v(\d+)`)
    matches := re.FindStringSubmatch(accept)
    if len(matches) >= 2 {
        if v, err := strconv.Atoi(matches[1]); err == nil {
            return v
        }
    }
    return 0
}
```

**Integration with Proxy**:

```go
// Route different versions to different backends
app.Use(versioning.Middleware(versioning.Config{
    Strategy:          versioning.StrategyAcceptHeader,
    VendorPrefix:      "application/vnd.myapi",
    DefaultVersion:    1,
    SupportedVersions: []int{1, 2, 3},
}))

// Version-aware proxy
app.Use("/api/*", func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        version := versioning.GetVersion(r.Context())

        var targets []string
        switch version {
        case 1:
            targets = []string{"http://api-v1:8080"}
        case 2:
            targets = []string{"http://api-v2:8080"}
        case 3:
            targets = []string{"http://api-v3:8080"}
        }

        proxy.New(proxy.Config{Targets: targets}).ServeHTTP(w, r)
    })
})
```

---

## P2-C: Request Coalescing

### Objective
Deduplicate identical in-flight requests to reduce backend load.

### Architecture

```
Concurrent Requests:
  Client A: GET /api/users/123
  Client B: GET /api/users/123
  Client C: GET /api/users/123

Coalescing Middleware:
  1. Generate request key
  2. Check if in-flight
  3. If yes: wait for result
  4. If no: execute and broadcast

Backend:
  Only receives 1 request

Response:
  All 3 clients receive same response
```

### Implementation Plan

**File**: `middleware/coalesce/coalesce.go`

```go
package coalesce

import (
    "bytes"
    "context"
    "hash/fnv"
    "net/http"
    "sync"
)

// In-flight request tracker
type inFlightRequest struct {
    key      string
    response *capturedResponse
    done     chan struct{}
    err      error
}

type capturedResponse struct {
    StatusCode int
    Header     http.Header
    Body       []byte
}

type Coalescer struct {
    mu        sync.RWMutex
    inFlight  map[string]*inFlightRequest
    keyFunc   KeyFunc
}

type KeyFunc func(r *http.Request) string

func New(keyFunc KeyFunc) *Coalescer {
    if keyFunc == nil {
        keyFunc = DefaultKeyFunc
    }
    return &Coalescer{
        inFlight: make(map[string]*inFlightRequest),
        keyFunc:  keyFunc,
    }
}

func (c *Coalescer) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only coalesce safe methods
        if r.Method != "GET" && r.Method != "HEAD" {
            next.ServeHTTP(w, r)
            return
        }

        key := c.keyFunc(r)

        // Check if request is in-flight
        c.mu.RLock()
        inflight, exists := c.inFlight[key]
        c.mu.RUnlock()

        if exists {
            // Wait for in-flight request to complete
            <-inflight.done

            // Write cached response
            if inflight.err == nil {
                writeResponse(w, inflight.response)
            } else {
                http.Error(w, "Request failed", http.StatusBadGateway)
            }
            return
        }

        // Create new in-flight request
        inflight = &inFlightRequest{
            key:  key,
            done: make(chan struct{}),
        }

        c.mu.Lock()
        c.inFlight[key] = inflight
        c.mu.Unlock()

        // Capture response
        recorder := &responseRecorder{
            ResponseWriter: w,
            statusCode:     http.StatusOK,
            body:           &bytes.Buffer{},
        }

        // Execute request
        next.ServeHTTP(recorder, r)

        // Store response
        inflight.response = &capturedResponse{
            StatusCode: recorder.statusCode,
            Header:     recorder.Header().Clone(),
            Body:       recorder.body.Bytes(),
        }

        // Cleanup and broadcast
        c.mu.Lock()
        delete(c.inFlight, key)
        c.mu.Unlock()
        close(inflight.done)
    })
}

func DefaultKeyFunc(r *http.Request) string {
    h := fnv.New64a()
    h.Write([]byte(r.Method))
    h.Write([]byte(r.URL.String()))
    return fmt.Sprintf("%x", h.Sum64())
}
```

**File**: `middleware/coalesce/recorder.go`

```go
package coalesce

type responseRecorder struct {
    http.ResponseWriter
    statusCode int
    body       *bytes.Buffer
    written    bool
}

func (r *responseRecorder) WriteHeader(code int) {
    if !r.written {
        r.statusCode = code
        r.written = true
        r.ResponseWriter.WriteHeader(code)
    }
}

func (r *responseRecorder) Write(b []byte) (int, error) {
    if !r.written {
        r.WriteHeader(http.StatusOK)
    }
    r.body.Write(b)
    return r.ResponseWriter.Write(b)
}
```

### Testing Plan

- [ ] Concurrent identical requests
- [ ] Different requests (no coalescing)
- [ ] POST/PUT requests (should not coalesce)
- [ ] Error propagation
- [ ] Timeout handling
- [ ] Race condition testing

---

## P2-D: Protocol Adapter Contracts (gRPC, GraphQL)

### Objective
Define **interface contracts** for protocol adapters without importing protocol libraries into plumego core.

### Design Philosophy

**Problem**: We want to support gRPC and GraphQL gateways, but:
- Cannot add protobuf/gRPC deps to plumego core
- Cannot add GraphQL libs to plumego core
- Must remain zero-dependency

**Solution**: Contract-based architecture
- Plumego defines **interfaces** for protocol adapters
- Users implement adapters using their chosen libraries
- Plumego provides examples but not implementations

### Architecture

```
┌─────────────────────────────────────────┐
│  plumego/contract/protocol              │
│  - Defines adapter interfaces           │
│  - No protocol-specific code            │
│  - Pure Go stdlib                       │
└─────────────────────────────────────────┘
                  ▲
                  │ implements
                  │
┌─────────────────────────────────────────┐
│  User Application                       │
│  - Import grpc.io/grpc                  │
│  - Implement GRPCAdapter interface      │
│  - Register with plumego                │
└─────────────────────────────────────────┘
```

### Implementation Plan

**File**: `contract/protocol/adapter.go`

```go
package protocol

import (
    "context"
    "io"
    "net/http"
)

// ProtocolAdapter converts HTTP requests to/from other protocols
type ProtocolAdapter interface {
    // Name returns the protocol name (e.g., "grpc", "graphql")
    Name() string

    // Handles checks if this adapter can handle the request
    Handles(r *http.Request) bool

    // Transform converts HTTP request to protocol-specific request
    Transform(ctx context.Context, r *http.Request) (Request, error)

    // Execute sends the protocol request to backend
    Execute(ctx context.Context, req Request) (Response, error)

    // Encode writes the protocol response as HTTP response
    Encode(ctx context.Context, w http.ResponseWriter, resp Response) error
}

// Generic request/response interfaces
type Request interface {
    Method() string
    Headers() map[string][]string
    Body() io.Reader
    Metadata() map[string]interface{}
}

type Response interface {
    StatusCode() int
    Headers() map[string][]string
    Body() io.Reader
    Metadata() map[string]interface{}
}
```

**File**: `contract/protocol/grpc.go`

```go
package protocol

// GRPCAdapter contract for gRPC gateway adapters
// Users implement this using google.golang.org/grpc
type GRPCAdapter interface {
    ProtocolAdapter

    // Service returns the gRPC service name
    Service() string

    // Methods returns supported RPC methods
    Methods() []string
}

// Example metadata for gRPC
type GRPCMetadata struct {
    Service    string
    Method     string
    Authority  string
    Timeout    string
    ContentType string
}
```

**File**: `contract/protocol/graphql.go`

```go
package protocol

// GraphQLAdapter contract for GraphQL gateway adapters
// Users implement this using github.com/graphql-go/graphql
type GraphQLAdapter interface {
    ProtocolAdapter

    // Schema returns the GraphQL schema definition
    Schema() string

    // Validate validates a GraphQL query
    Validate(query string) error
}

// Example GraphQL request structure
type GraphQLRequest struct {
    Query         string                 `json:"query"`
    OperationName string                 `json:"operationName,omitempty"`
    Variables     map[string]interface{} `json:"variables,omitempty"`
}
```

**File**: `contract/protocol/registry.go`

```go
package protocol

import (
    "fmt"
    "net/http"
    "sync"
)

// Registry holds protocol adapters
type Registry struct {
    mu       sync.RWMutex
    adapters []ProtocolAdapter
}

func NewRegistry() *Registry {
    return &Registry{
        adapters: make([]ProtocolAdapter, 0),
    }
}

func (r *Registry) Register(adapter ProtocolAdapter) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.adapters = append(r.adapters, adapter)
}

func (r *Registry) Find(req *http.Request) (ProtocolAdapter, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, adapter := range r.adapters {
        if adapter.Handles(req) {
            return adapter, nil
        }
    }

    return nil, fmt.Errorf("no adapter found for request")
}
```

**File**: `middleware/protocol/middleware.go`

```go
package protocol

import (
    "net/http"
    "github.com/spcent/plumego/contract/protocol"
)

// Middleware creates a protocol gateway middleware
func Middleware(registry *protocol.Registry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Try to find adapter
            adapter, err := registry.Find(r)
            if err != nil {
                // No adapter, pass through to next handler
                next.ServeHTTP(w, r)
                return
            }

            // Transform request
            req, err := adapter.Transform(r.Context(), r)
            if err != nil {
                http.Error(w, "Transform failed", http.StatusBadRequest)
                return
            }

            // Execute protocol request
            resp, err := adapter.Execute(r.Context(), req)
            if err != nil {
                http.Error(w, "Execution failed", http.StatusBadGateway)
                return
            }

            // Encode response
            if err := adapter.Encode(r.Context(), w, resp); err != nil {
                http.Error(w, "Encoding failed", http.StatusInternalServerError)
                return
            }
        })
    }
}
```

### User Implementation Example (NOT in plumego)

**User's code** (in their application):

```go
// user-app/adapters/grpc_adapter.go
package adapters

import (
    "context"
    "io"
    "net/http"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "github.com/spcent/plumego/contract/protocol"
)

type MyGRPCAdapter struct {
    conn    *grpc.ClientConn
    service string
}

func NewGRPCAdapter(target string, service string) (*MyGRPCAdapter, error) {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }

    return &MyGRPCAdapter{
        conn:    conn,
        service: service,
    }, nil
}

func (a *MyGRPCAdapter) Name() string {
    return "grpc"
}

func (a *MyGRPCAdapter) Handles(r *http.Request) bool {
    // Check if request is for gRPC endpoint
    return r.Header.Get("Content-Type") == "application/grpc" ||
           r.URL.Path == "/grpc/"+a.service
}

func (a *MyGRPCAdapter) Transform(ctx context.Context, r *http.Request) (protocol.Request, error) {
    // Parse HTTP request into gRPC call
    // User has full access to grpc library here
}

func (a *MyGRPCAdapter) Execute(ctx context.Context, req protocol.Request) (protocol.Response, error) {
    // Execute gRPC call using conn
}

func (a *MyGRPCAdapter) Encode(ctx context.Context, w http.ResponseWriter, resp protocol.Response) error {
    // Encode gRPC response as HTTP
}
```

**User's main.go**:

```go
package main

import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/contract/protocol"
    protomw "github.com/spcent/plumego/middleware/protocol"
    "myapp/adapters"
)

func main() {
    // Create protocol registry
    registry := protocol.NewRegistry()

    // Register user's gRPC adapter (uses grpc library)
    grpcAdapter, _ := adapters.NewGRPCAdapter("backend:50051", "UserService")
    registry.Register(grpcAdapter)

    // Register user's GraphQL adapter (uses graphql library)
    graphqlAdapter, _ := adapters.NewGraphQLAdapter("http://backend:4000")
    registry.Register(graphqlAdapter)

    // Create app with protocol middleware
    app := core.New(
        core.WithAddr(":8080"),
    )

    // Add protocol gateway middleware
    app.Use(protomw.Middleware(registry))

    app.Boot()
}
```

### Benefits

1. **Zero Dependencies**: Plumego core has zero deps
2. **User Choice**: Users pick their preferred gRPC/GraphQL libraries
3. **Type Safety**: Contract interfaces enforce structure
4. **Testability**: Users can mock adapters
5. **Extensibility**: Can add WebSocket, Thrift, SOAP adapters

### Documentation Plan

Create comprehensive examples in `examples/`:

- `examples/grpc-gateway/` - Full gRPC gateway with grpc-go
- `examples/graphql-gateway/` - Full GraphQL gateway with graphql-go
- `examples/multi-protocol-gateway/` - Combined HTTP/gRPC/GraphQL

Each example includes:
- Adapter implementation
- Backend service
- Client test code
- README with explanation

---

## P2-E: Request/Response Transformation

### Objective
Transform request/response payloads without protocol-specific knowledge.

### Implementation Plan

**File**: `middleware/transform/transform.go`

```go
package transform

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
)

// RequestTransformer modifies request before proxying
type RequestTransformer func(*http.Request) error

// ResponseTransformer modifies response before returning
type ResponseTransformer func(*http.Response) error

type Config struct {
    RequestTransformers  []RequestTransformer
    ResponseTransformers []ResponseTransformer
}

func Middleware(config Config) func(http.Handler) http.Handler {
    // Apply transformations
}

// Built-in transformers

// AddHeader adds a header to request
func AddHeader(key, value string) RequestTransformer {
    return func(r *http.Request) error {
        r.Header.Set(key, value)
        return nil
    }
}

// RenameJSONField renames a JSON field in request body
func RenameJSONField(from, to string) RequestTransformer {
    return func(r *http.Request) error {
        // Read body
        body, _ := io.ReadAll(r.Body)
        r.Body.Close()

        // Parse JSON
        var data map[string]interface{}
        json.Unmarshal(body, &data)

        // Rename field
        if val, exists := data[from]; exists {
            data[to] = val
            delete(data, from)
        }

        // Write back
        newBody, _ := json.Marshal(data)
        r.Body = io.NopCloser(bytes.NewReader(newBody))
        r.ContentLength = int64(len(newBody))

        return nil
    }
}
```

---

## P2-F: Advanced Rate Limiting

### Objective
Enhance existing rate limiting with:
- Token bucket algorithm
- Distributed rate limiting (contract-based)
- Per-endpoint quotas
- Quota sharing across cluster

### Implementation Plan

**File**: `middleware/ratelimit/token_bucket.go`

```go
package ratelimit

import (
    "sync"
    "time"
)

// TokenBucket implements token bucket algorithm
type TokenBucket struct {
    capacity  int64
    tokens    int64
    refillRate int64 // tokens per second
    lastRefill time.Time
    mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
    return &TokenBucket{
        capacity:   capacity,
        tokens:     capacity,
        refillRate: refillRate,
        lastRefill: time.Now(),
    }
}

func (tb *TokenBucket) Allow(tokens int64) bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    // Refill tokens based on elapsed time
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill).Seconds()
    newTokens := int64(elapsed * float64(tb.refillRate))

    if newTokens > 0 {
        tb.tokens = min(tb.capacity, tb.tokens+newTokens)
        tb.lastRefill = now
    }

    // Check if enough tokens
    if tb.tokens >= tokens {
        tb.tokens -= tokens
        return true
    }

    return false
}
```

**File**: `contract/ratelimit/distributed.go`

```go
package ratelimit

// DistributedStore contract for distributed rate limiting
// Users implement using Redis, etcd, etc.
type DistributedStore interface {
    // Increment increments the counter and returns new value
    Increment(key string, window time.Duration) (int64, error)

    // Get returns current counter value
    Get(key string) (int64, error)

    // Reset resets the counter
    Reset(key string) error
}
```

---

## Implementation Order

### Phase 1: Foundation (Week 1)
1. Response Caching (P2-A)
2. API Version Negotiation (P2-B)

### Phase 2: Advanced (Week 2)
3. Request Coalescing (P2-C)
4. Request/Response Transformation (P2-E)

### Phase 3: Extensibility (Week 3)
5. Protocol Adapter Contracts (P2-D)
6. Advanced Rate Limiting (P2-F)

### Phase 4: Documentation & Examples (Week 4)
7. Create comprehensive examples
8. Write integration tests
9. Update documentation

---

## Testing Strategy

### Unit Tests
- All features have standalone unit tests
- Table-driven test patterns
- Edge case coverage

### Integration Tests
- End-to-end gateway tests
- Multi-feature composition tests
- Performance benchmarks

### Example Applications
- Each P2 feature has working example
- Examples demonstrate best practices
- Examples include performance metrics

---

## Success Criteria

- [ ] All features use **zero external dependencies**
- [ ] All features have **>80% test coverage**
- [ ] All features pass `go test -race`
- [ ] Protocol adapters work with user-provided implementations
- [ ] Documentation is comprehensive
- [ ] Examples are production-ready
- [ ] Performance benchmarks show <5% overhead

---

## Next Steps

1. **Review this plan** - Confirm approach aligns with vision
2. **Prioritize features** - Decide implementation order
3. **Begin implementation** - Start with P2-A (caching)
4. **Iterative review** - Review each feature before next

---

## Questions for Consideration

1. **Caching**: Should we support distributed cache stores via contract?
2. **Versioning**: Should we support sunset headers for deprecated versions?
3. **Coalescing**: Should we add TTL for coalesced responses?
4. **Protocols**: Should we provide reference gRPC adapter implementation?
5. **Transformation**: Should we support Lua/JavaScript for custom transforms?

---

**Document Status**: Ready for Review
**Estimated LOC**: ~2,500 lines total
**Estimated Dependencies**: 0 (zero)
**Backward Compatibility**: 100% (all additions)
