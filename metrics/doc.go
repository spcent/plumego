// Package metrics provides the stable metrics contracts and small in-memory base collectors.
//
// The stable surface is intentionally narrow:
//   - metric record and label contracts
//   - the shared HTTP observer contract
//   - BaseMetricsCollector, NoopCollector, and MultiCollector
//   - helper functions for recording and measuring operations
//
// Prometheus exposition, tracing implementations, and dev-dashboard collectors
// live in their owning extension packages:
//   - x/observability for generic Prometheus and tracing adapters
//   - x/devtools for dev-only collectors
//   - x/messaging for messaging-specific metrics reporters and exporters
package metrics
