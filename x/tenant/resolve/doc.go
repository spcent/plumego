// Package resolve contains tenant resolution helpers and middleware.
//
// The canonical resolution order in Middleware is:
//   - authenticated principal tenant ID first
//   - custom extractor when configured
//   - configured tenant header fallback
//
// Use this package for transport-level tenant identification only. Keep tenant
// policy, quota, and persistence decisions in their owning x/tenant packages.
package resolve
