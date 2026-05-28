// Package health defines transport-agnostic health state, readiness status, and
// component health models.
//
// The package owns stable status values and lightweight DTOs only. It does not
// own HTTP handlers, endpoint registration, component orchestration, retry
// policy, health history, or ops reporting.
//
// HTTP health endpoints (/healthz, /readyz) and orchestration logic live in
// x/observability/ops. Reference apps show canonical wiring examples.
package health
