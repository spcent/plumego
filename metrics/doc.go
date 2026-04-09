// Package metrics provides the stable metrics contracts and small in-memory base collectors.
//
// The stable surface is intentionally narrow:
//   - metric record and label contracts
//   - the shared HTTP observer contract
//   - BaseMetricsCollector, NoopCollector, and MultiCollector
//   - generic record/stat/reset collector composition
//
// Prometheus exposition, tracing implementations, and dev-dashboard collectors
// live in their owning extension packages:
//   - x/observability for generic Prometheus, tracing adapters, buffered record inspection, and non-HTTP feature metric helpers
//   - x/devtools for dev-only collectors
//   - x/messaging for messaging-specific metrics reporters and exporters
package metrics
