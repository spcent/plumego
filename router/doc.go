// Package router provides Plumego's HTTP route matching, path parameter,
// grouping, and reverse URL primitives.
//
// Router implements http.Handler and keeps route structure separate from
// response writing, authentication policy, service construction, and business
// validation. Routes are registered explicitly with a method, path, handler,
// and optional metadata.
//
// Use NewRouter for standalone routing, Group for shared prefixes, Param for
// matched path parameters, and WithRouteName with URL for reverse URL
// generation.
package router
