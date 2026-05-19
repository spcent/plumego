// Package log provides Plumego's stable logging interfaces and base logger
// implementations.
//
// The package owns the minimal StructuredLogger contract, field helpers, and
// the canonical NewLogger constructor. Exporter setup, aggregation pipelines,
// request policy, and feature-specific logging semantics belong in application
// code or observability extensions.
package log
