package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
)

// Group creates a new router group with the given prefix.
// Groups allow you to share a common path prefix and middleware across multiple routes.
// Child groups inherit the parent's prefix and can add their own middleware.
func (r *Router) Group(prefix string) *Router {
	fullPrefix := normalizeGroupPrefix(r.prefix, prefix)

	return &Router{
		prefix:            fullPrefix,
		parent:            r,
		middlewareManager: NewMiddlewareManager(),
		state:             r.state,
	}
}

// GroupFunc creates a new route group with the given prefix and invokes
// the callback function with the group router.
func (r *Router) GroupFunc(prefix string, fn func(*Router)) *Router {
	g := r.Group(prefix)
	fn(g)
	return g
}

func normalizeGroupPrefix(parent, child string) string {
	if child != "" && child[0] != '/' {
		child = "/" + child
	}
	child = strings.TrimRight(child, "/")
	parent = strings.TrimRight(parent, "/")

	combined := parent + child
	if combined == "" {
		return ""
	}
	return combined
}

// AddRoute adds a route to the router with the given method, path and handler.
func (r *Router) AddRoute(method, path string, handler Handler) error {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	if r.state.frozen {
		return contract.WrapError(
			fmt.Errorf("router is frozen, cannot add route after freeze"),
			"add_route",
			"router",
			map[string]any{
				"method": method,
				"path":   path,
			},
		)
	}

	fullPath := r.fullPath(path)
	if fullPath == "" {
		fullPath = "/"
	}

	if r.state.trees[method] == nil {
		r.state.trees[method] = &node{}
	}

	current := r.state.trees[method]
	if fullPath == "/" {
		if current.handler != nil {
			return contract.WrapError(
				fmt.Errorf("duplicate route registration: %s /", method),
				"add_route",
				"router",
				map[string]any{
					"method": method,
					"path":   fullPath,
				},
			)
		}
		current.handler = handler
		current.fullPath = fullPath
		current.validation = r.validationFor(method, fullPath)
		r.state.routes[method] = append(r.state.routes[method], route{Method: method, Path: fullPath})
		return nil
	}

	segments := compilePathSegments(fullPath)
	paramKeys := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg.isParam {
			paramKeys = append(paramKeys, seg.paramName)
			child := r.findParamChild(current)
			if child != nil {
				if child.paramName == "" {
					child.paramName = seg.paramName
				} else if child.paramName != seg.paramName {
					return contract.WrapError(
						fmt.Errorf("route conflict: parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName),
						"add_route",
						"router",
						map[string]any{
							"method": method,
							"path":   fullPath,
						},
					)
				}
				current = child
				continue
			}

			child = &node{path: ":", paramName: seg.paramName}
			r.insertChild(current, child)
			current = child
			continue
		}

		if seg.isWild {
			paramKeys = append(paramKeys, seg.paramName)
			child := r.findWildChild(current)
			if child != nil {
				if child.paramName == "" {
					child.paramName = seg.paramName
				} else if child.paramName != seg.paramName {
					return contract.WrapError(
						fmt.Errorf("route conflict: wildcard parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName),
						"add_route",
						"router",
						map[string]any{
							"method": method,
							"path":   fullPath,
						},
					)
				}
				current = child
				continue
			}

			child = &node{path: "*", paramName: seg.paramName}
			r.insertChild(current, child)
			current = child
			continue
		}

		child := r.findChild(current, seg.raw)
		if child == nil {
			child = &node{path: seg.raw}
			r.insertChild(current, child)
		}

		current = child
	}

	if current.handler != nil {
		return contract.WrapError(
			fmt.Errorf("duplicate route registration: %s %s", method, fullPath),
			"add_route",
			"router",
			map[string]any{
				"method": method,
				"path":   fullPath,
			},
		)
	}
	current.handler = handler
	current.paramKeys = paramKeys
	current.fullPath = fullPath
	current.middlewares = r.routeMiddlewares()
	current.validation = r.validationFor(method, fullPath)

	r.state.routes[method] = append(r.state.routes[method], route{Method: method, Path: fullPath})
	return nil
}

// AddRouteWithOptions adds a route and attaches metadata options.
func (r *Router) AddRouteWithOptions(method, path string, handler Handler, opts ...RouteOption) error {
	if err := r.AddRoute(method, path, handler); err != nil {
		return err
	}
	if len(opts) == 0 {
		return nil
	}
	meta := RouteMeta{}
	for _, opt := range opts {
		if opt != nil {
			opt(&meta)
		}
	}
	r.SetRouteMeta(method, path, meta)
	return nil
}

// compilePathSegments parses a URL path into route segments.
func compilePathSegments(path string) []segment {
	if path == "/" {
		return nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]segment, 0, len(parts))
	for _, segmentPart := range parts {
		if strings.HasPrefix(segmentPart, ":") {
			segments = append(segments, segment{
				raw:       segmentPart,
				isParam:   true,
				paramName: segmentPart[1:],
			})
		} else if strings.HasPrefix(segmentPart, "*") {
			segments = append(segments, segment{
				raw:       segmentPart,
				isWild:    true,
				paramName: segmentPart[1:],
			})
		} else {
			segments = append(segments, segment{raw: segmentPart})
		}
	}
	return segments
}

// HandleFunc registers a standard http.HandlerFunc for the given path and method.
func (r *Router) HandleFunc(method, path string, h http.HandlerFunc) error {
	return r.AddRoute(method, path, h)
}

// Handle registers a standard http.Handler for the given path and method.
func (r *Router) Handle(method, path string, h http.Handler) error {
	return r.AddRoute(method, path, h)
}

// HandleWithOptions registers a standard http.Handler for the given path and method with metadata.
func (r *Router) HandleWithOptions(method, path string, h http.Handler, opts ...RouteOption) error {
	return r.AddRouteWithOptions(method, path, h, opts...)
}
