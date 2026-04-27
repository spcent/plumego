// Package metrics provides the stable metrics contracts and small in-memory base collectors.
//
// The stable surface is intentionally narrow:
//   - metric record and label contracts
//   - NewHTTPRecord for the shared HTTP request record shape
//   - the shared HTTP observer contract
//   - BaseMetricsCollector, NoopCollector, and MultiCollector
//   - generic record/stat/reset collector composition
//
// BaseMetricsCollector records aggregate counts only. It classifies
// MetricRecord.Error values as errors for every metric name and classifies HTTP
// records with status codes >= 400 as HTTP errors. Response byte aggregation,
// raw record retention, exporters, and feature-specific metrics belong in
// owning extension packages.
//
// Prometheus exposition, tracing implementations, and dev-dashboard collectors
// live in their owning extension packages:
//   - x/observability for generic Prometheus, tracing adapters, buffered record inspection, and non-HTTP feature metric helpers
//   - x/devtools for dev-only collectors
//   - x/messaging for messaging-specific metrics reporters and exporters
package metrics
