# Abuse Guard

> **Package**: `github.com/spcent/plumego/security/abuse` | **Stability**: High

## Overview

The `abuse/` package provides anti-abuse protection through rate limiting, IP blocking, and anomaly detection. It protects your application from:

- **Brute-force attacks**: Login/registration abuse
- **API abuse**: Excessive requests from single clients
- **DoS attacks**: Resource exhaustion
- **Scraping**: Automated data harvesting
- **Spam**: Bulk form submissions

The abuse guard uses **token bucket algorithm** for rate limiting with per-client tracking.

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/security/abuse"

// Create abuse guard
guard := abuse.NewGuard(abuse.Config{
    RequestsPerSecond: 10,                 // 10 requests per second
    BurstSize:         20,                 // Allow burst of 20
    CleanupInterval:   time.Minute,        // Clean up old clients every minute
})

// Check if client is allowed
clientIP := "192.168.1.1"
if !guard.Allow(clientIP) {
    http.Error(w, "Too many requests", http.StatusTooManyRequests)
    return
}

// Process request...
```

### With Middleware

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/ratelimit"
)

app := core.New(
    // Enable abuse guard globally
    core.WithAbuseGuardEnabled(true),
    core.WithAbuseGuardConfig(abuse.Config{
        RequestsPerSecond: 100,    // 100 req/s per IP
        BurstSize:         200,    // Allow burst of 200
    }),
)
```

## Core Types

### Config

```go
type Config struct {
    RequestsPerSecond float64       // Requests per second per client
    BurstSize         int            // Maximum burst size
    CleanupInterval   time.Duration  // Interval to clean up old clients
}
```

### Guard

```go
type Guard struct {
    // ... internal fields
}

func NewGuard(config Config) *Guard
func (g *Guard) Allow(clientID string) bool
func (g *Guard) Block(clientID string)
func (g *Guard) Unblock(clientID string)
func (g *Guard) IsBlocked(clientID string) bool
func (g *Guard) Reset(clientID string)
```

## Token Bucket Algorithm

### How It Works

Each client has a "bucket" that:
1. Fills with tokens at a constant rate (RequestsPerSecond)
2. Has maximum capacity (BurstSize)
3. Each request consumes 1 token
4. Request denied if bucket is empty

**Example**:
```
Config: 10 req/s, burst=20

Time 0s:  Bucket has 20 tokens (full)
Time 0s:  Client makes 15 requests → 5 tokens left
Time 0s:  Client makes 10 requests → DENIED (only 5 tokens)
Time 1s:  Bucket refills by 10 → 15 tokens
Time 1s:  Client makes 15 requests → 0 tokens left
```

### Configuration Strategies

| Use Case | Requests/Second | Burst | Rationale |
|----------|----------------|-------|-----------|
| **API Gateway** | 100 | 200 | High throughput, allow bursts |
| **Login Endpoint** | 1 | 5 | Prevent brute-force |
| **Registration** | 0.1 | 3 | Prevent spam signups |
| **Password Reset** | 0.2 | 2 | Prevent abuse |
| **Public API** | 10 | 20 | Moderate usage |

## Rate Limiting Patterns

### Per-Endpoint Rate Limits

```go
var (
    // Strict for authentication
    loginGuard = abuse.NewGuard(abuse.Config{
        RequestsPerSecond: 1,
        BurstSize:         5,
        CleanupInterval:   time.Minute,
    })

    // Moderate for API
    apiGuard = abuse.NewGuard(abuse.Config{
        RequestsPerSecond: 10,
        BurstSize:         20,
        CleanupInterval:   time.Minute,
    })

    // Lenient for static assets
    staticGuard = abuse.NewGuard(abuse.Config{
        RequestsPerSecond: 100,
        BurstSize:         200,
        CleanupInterval:   time.Minute,
    })
)

func rateLimit(guard *abuse.Guard) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)

            if !guard.Allow(clientIP) {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Apply different limits
app.Post("/login", handleLogin, rateLimit(loginGuard))
app.Get("/api/users", handleUsers, rateLimit(apiGuard))
app.Get("/static/*", handleStatic, rateLimit(staticGuard))
```

### Per-User Rate Limits

```go
type UserGuard struct {
    guards map[string]*abuse.Guard
    mu     sync.RWMutex
    config abuse.Config
}

func NewUserGuard(config abuse.Config) *UserGuard {
    return &UserGuard{
        guards: make(map[string]*abuse.Guard),
        config: config,
    }
}

func (ug *UserGuard) Allow(userID string) bool {
    ug.mu.Lock()
    guard, ok := ug.guards[userID]
    if !ok {
        guard = abuse.NewGuard(ug.config)
        ug.guards[userID] = guard
    }
    ug.mu.Unlock()

    return guard.Allow(userID)
}

// Usage
var userGuard = NewUserGuard(abuse.Config{
    RequestsPerSecond: 100,
    BurstSize:         200,
})

func apiMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserIDFromToken(r)
        if userID == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if !userGuard.Allow(userID) {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## IP Blocking

### Manual IP Blocking

```go
// Block abusive IP
guard.Block("192.168.1.100")

// Check if IP is blocked
if guard.IsBlocked("192.168.1.100") {
    http.Error(w, "Blocked", http.StatusForbidden)
    return
}

// Unblock IP
guard.Unblock("192.168.1.100")
```

### Automatic Blocking

```go
type SmartGuard struct {
    guard          *abuse.Guard
    violations     map[string]int
    mu             sync.RWMutex
    blockThreshold int
}

func NewSmartGuard(config abuse.Config, blockThreshold int) *SmartGuard {
    return &SmartGuard{
        guard:          abuse.NewGuard(config),
        violations:     make(map[string]int),
        blockThreshold: blockThreshold,
    }
}

func (sg *SmartGuard) Allow(clientIP string) bool {
    // Check if already blocked
    if sg.guard.IsBlocked(clientIP) {
        return false
    }

    // Check rate limit
    if !sg.guard.Allow(clientIP) {
        // Increment violation count
        sg.mu.Lock()
        sg.violations[clientIP]++
        violations := sg.violations[clientIP]
        sg.mu.Unlock()

        // Auto-block after threshold violations
        if violations >= sg.blockThreshold {
            sg.guard.Block(clientIP)
            log.Printf("Auto-blocked IP: %s (violations: %d)", clientIP, violations)
        }

        return false
    }

    return true
}

// Usage
smartGuard := NewSmartGuard(abuse.Config{
    RequestsPerSecond: 10,
    BurstSize:         20,
}, 5) // Block after 5 violations
```

## Response Headers

### Standard Rate Limit Headers

```go
func rateLimitMiddleware(guard *abuse.Guard) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)

            // Add rate limit headers
            w.Header().Set("X-RateLimit-Limit", "100")
            w.Header().Set("X-RateLimit-Remaining", getRemainingTokens(guard, clientIP))
            w.Header().Set("X-RateLimit-Reset", getResetTime(guard, clientIP))

            if !guard.Allow(clientIP) {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Client Identification

### Extract Client IP

```go
func getClientIP(r *http.Request) string {
    // 1. Check X-Forwarded-For (behind proxy)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        // Take first IP
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }

    // 2. Check X-Real-IP (nginx)
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }

    // 3. Use RemoteAddr (direct connection)
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return ip
}
```

### Composite Keys

```go
// Rate limit by IP + User-Agent (detect bots)
func getClientKey(r *http.Request) string {
    ip := getClientIP(r)
    ua := r.Header.Get("User-Agent")
    return fmt.Sprintf("%s:%s", ip, ua)
}

// Rate limit by IP + Endpoint (different limits per endpoint)
func getEndpointKey(r *http.Request) string {
    ip := getClientIP(r)
    path := r.URL.Path
    return fmt.Sprintf("%s:%s", ip, path)
}
```

## Monitoring and Logging

### Log Rate Limit Events

```go
func rateLimitWithLogging(guard *abuse.Guard) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)

            if !guard.Allow(clientIP) {
                // Log rate limit violation
                log.Printf("Rate limit exceeded: ip=%s path=%s method=%s ua=%s",
                    clientIP,
                    r.URL.Path,
                    r.Method,
                    r.Header.Get("User-Agent"),
                )

                // Send to metrics (Prometheus, etc.)
                metrics.RateLimitExceeded.WithLabelValues(clientIP, r.URL.Path).Inc()

                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Metrics Integration

```go
import "github.com/spcent/plumego/metrics"

var (
    rateLimitExceeded = metrics.NewCounter("rate_limit_exceeded", "Total rate limit violations")
    blockedRequests   = metrics.NewCounter("blocked_requests", "Total blocked requests")
)

func (sg *SmartGuard) Allow(clientIP string) bool {
    if sg.guard.IsBlocked(clientIP) {
        blockedRequests.Inc()
        return false
    }

    if !sg.guard.Allow(clientIP) {
        rateLimitExceeded.Inc()
        return false
    }

    return true
}
```

## Advanced Patterns

### Sliding Window Rate Limit

```go
type SlidingWindow struct {
    requests map[string][]time.Time
    mu       sync.RWMutex
    limit    int
    window   time.Duration
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
    return &SlidingWindow{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (sw *SlidingWindow) Allow(clientID string) bool {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-sw.window)

    // Remove old requests
    times := sw.requests[clientID]
    var valid []time.Time
    for _, t := range times {
        if t.After(cutoff) {
            valid = append(valid, t)
        }
    }

    // Check limit
    if len(valid) >= sw.limit {
        sw.requests[clientID] = valid
        return false
    }

    // Add new request
    valid = append(valid, now)
    sw.requests[clientID] = valid
    return true
}
```

### Geographic Rate Limiting

```go
import "github.com/oschwald/geoip2-golang"

type GeoGuard struct {
    db     *geoip2.Reader
    guards map[string]*abuse.Guard // country code -> guard
}

func (gg *GeoGuard) Allow(clientIP string) bool {
    // Lookup country
    ip := net.ParseIP(clientIP)
    record, err := gg.db.Country(ip)
    if err != nil {
        // Unknown country, use default strict limit
        return gg.guards["default"].Allow(clientIP)
    }

    // Use country-specific guard
    country := record.Country.IsoCode
    guard, ok := gg.guards[country]
    if !ok {
        guard = gg.guards["default"]
    }

    return guard.Allow(clientIP)
}
```

## Testing

### Unit Tests

```go
func TestAbuseGuard(t *testing.T) {
    guard := abuse.NewGuard(abuse.Config{
        RequestsPerSecond: 10,
        BurstSize:         20,
        CleanupInterval:   time.Minute,
    })

    clientIP := "192.168.1.1"

    // Should allow first 20 requests (burst)
    for i := 0; i < 20; i++ {
        if !guard.Allow(clientIP) {
            t.Errorf("Request %d denied, expected allowed", i+1)
        }
    }

    // 21st request should be denied
    if guard.Allow(clientIP) {
        t.Error("Request 21 allowed, expected denied")
    }

    // After 1 second, should refill by 10 tokens
    time.Sleep(time.Second)

    // Next 10 requests should be allowed
    for i := 0; i < 10; i++ {
        if !guard.Allow(clientIP) {
            t.Errorf("Request denied after refill, expected allowed")
        }
    }
}

func TestIPBlocking(t *testing.T) {
    guard := abuse.NewGuard(abuse.Config{
        RequestsPerSecond: 10,
        BurstSize:         20,
    })

    clientIP := "192.168.1.100"

    // Block IP
    guard.Block(clientIP)

    // Should deny all requests
    if guard.Allow(clientIP) {
        t.Error("Blocked IP allowed request")
    }

    // Unblock IP
    guard.Unblock(clientIP)

    // Should allow requests again
    if !guard.Allow(clientIP) {
        t.Error("Unblocked IP denied request")
    }
}
```

## Performance Considerations

### Memory Usage

Each tracked client consumes memory. Clean up inactive clients:

```go
config := abuse.Config{
    CleanupInterval: 5 * time.Minute, // Clean up every 5 minutes
}
```

### Concurrent Access

The guard uses `sync.RWMutex` internally for thread safety. No additional locking needed.

### Distributed Systems

For multi-server deployments, use centralized rate limiting (Redis):

```go
import "github.com/go-redis/redis/v8"

type RedisGuard struct {
    client *redis.Client
    limit  int
    window time.Duration
}

func (rg *RedisGuard) Allow(clientIP string) bool {
    ctx := context.Background()
    key := fmt.Sprintf("ratelimit:%s", clientIP)

    // Increment counter
    count, err := rg.client.Incr(ctx, key).Result()
    if err != nil {
        return false
    }

    // Set expiry on first request
    if count == 1 {
        rg.client.Expire(ctx, key, rg.window)
    }

    return count <= int64(rg.limit)
}
```

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [Middleware: Rate Limiting](../middleware/ratelimit.md) — Rate limiting middleware
- [Best Practices](best-practices.md) — Security best practices

## Reference Implementation

See examples:
- `examples/reference/` — Abuse guard with login rate limiting
- `examples/api-gateway/` — API gateway with per-client limits
