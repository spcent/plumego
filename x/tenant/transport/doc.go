// Package transport contains tenant-specific transport mapping helpers.
//
// These helpers keep tenant-facing HTTP details explicit: default header names,
// Retry-After propagation, and remaining-budget response headers for quota and
// rate-limit middleware.
package transport
