// Package contract provides Plumego's canonical HTTP response, error, and
// request metadata contracts.
//
// The package owns transport primitives only: structured API errors, success
// response envelopes, request ID context accessors, route metadata carriers,
// and lightweight trace metadata. It deliberately avoids route matching,
// application bootstrap, authentication policy, persistence, and business
// response shapes.
//
// Handlers should use WriteResponse for successful JSON responses and
// WriteError with values from NewErrorBuilder for failures. This keeps one
// success path and one error-construction path across stable Plumego packages.
package contract
