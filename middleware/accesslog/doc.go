// Package accesslog provides HTTP access logging middleware for Plumego
// services.
//
// The middleware records transport-level request and response fields through an
// injected logger. It does not own business audit policy, persistence, or
// exporter setup.
package accesslog
