// Package recovery provides panic recovery middleware for HTTP request
// handlers.
//
// The middleware converts downstream panics into sanitized 500 responses and
// logs server-side panic metadata through an injected logger. It does not expose
// panic details to clients.
package recovery
