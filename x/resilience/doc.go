// Package resilience provides reusable circuit breaker and rate limiter
// primitives for the extension layer.
//
// Use circuitbreaker for wrapping external calls that may fail and need
// automatic recovery. Use ratelimit for token-bucket and sliding-window
// rate limiting that can be composed into middleware or service calls.
//
// Both sub-packages expose explicit, configurable primitives with no hidden
// globals. Feature-specific resilience orchestration (e.g. AI provider
// fallback) belongs in the owning extension, not here.
//
// Sub-packages:
//
//   - circuitbreaker: Three-state circuit breaker (closed/open/half-open)
//   - ratelimit: Token-bucket and sliding-window rate limiter primitives
package resilience
