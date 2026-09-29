// Package timeout provides request deadline enforcement middleware.
//
// [Middleware] enforces a maximum duration for request processing. When the
// downstream handler misses the deadline, the request context is canceled and
// a 504 Gateway Timeout response is returned. Responses are buffered up to a
// configured limit so a timed-out request cannot emit a partial response body;
// handlers must observe the request context to stop work promptly.
package timeout
