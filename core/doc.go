// Package core provides Plumego's HTTP application kernel.
//
// The package owns application construction, explicit route and middleware
// attachment, handler preparation, and the prepared http.Server lifecycle. It
// preserves net/http compatibility: App implements http.Handler through
// ServeHTTP, and Prepare exposes a caller-owned *http.Server through Server.
//
// Core is deliberately narrow. It does not own route matching internals,
// readiness signaling, environment loading, feature catalogs, persistence,
// debug routes, logger lifecycle, or advanced TLS policy. Applications should
// compose those concerns explicitly in their main package, reference app, or
// owning extension package.
//
// The supported construction path is:
//
//	cfg := core.DefaultConfig()
//	app := core.New(cfg, core.AppDependencies{Logger: logger})
//	app.Get("/healthz", http.HandlerFunc(handler))
//	if err := app.Prepare(); err != nil {
//		return err
//	}
//	server, err := app.Server()
//	if err != nil {
//		return err
//	}
//
// ServeHTTP prepares only the handler path and intentionally skips server-only
// validation such as TLS certificate loading. Use Prepare before Server when
// the application needs the full server lifecycle.
//
// Core errors are ordinary wrapped Go errors. They include stable diagnostic
// operation context such as "core prepare_server" or "core add_route", and may
// include operation parameters before the final cause. The package intentionally
// does not export lifecycle sentinel errors. Callers should use errors.Is for
// wrapped causes exported by the package that produced the cause, and avoid
// branching on complete error strings.
package core
