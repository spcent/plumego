// Package frontend provides explicit static and embedded asset serving for
// single-page applications and other frontend distributions.
//
// The package exposes a single handler constructor that serves files from an
// http.FileSystem with content negotiation, pre-compressed asset support
// (Brotli and gzip), configurable cache headers, and SPA index fallback.
//
// Typical usage:
//
//	handler := frontend.New(frontend.Config{
//	    Prefix:    "/app",
//	    IndexFile: "index.html",
//	    Fallback:  true,
//	})
//	router.Mount("/app", handler)
//
// The package does not own route registration or middleware wiring; callers
// mount the returned handler at an explicit path.
package frontend
