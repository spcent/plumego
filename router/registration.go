package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Group creates a new router group with the given prefix.
// Groups allow sharing a common path prefix across multiple routes.
func (r *Router) Group(prefix string) *Router {
	fullPrefix := normalizeGroupPrefix(r.prefix, prefix)

	return &Router{
		prefix: fullPrefix,
		parent: r,
		state:  r.state,
	}
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

// AddRoute adds a route to the router with the given method, path, handler, and
// optional route metadata.
func (r *Router) AddRoute(method, path string, handler http.Handler, opts ...RouteOption) error {
	if handler == nil {
		return fmt.Errorf("router add_route %s %s: %w", method, path, errors.New("nil handler"))
	}

	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	if r.state.frozen {
		return fmt.Errorf("router add_route %s %s: %w", method, path, fmt.Errorf("router is frozen, cannot add route after freeze"))
	}
	meta := routeMetaFromOptions(opts...)

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
			return fmt.Errorf("router add_route %s %s: %w", method, fullPath, fmt.Errorf("duplicate route registration: %s /", method))
		}
		current.handler = handler
		current.fullPath = fullPath
		r.storeRouteMetaLocked(method, fullPath, meta)
		r.state.routes[method] = append(r.state.routes[method], route{Method: method, Path: fullPath})
		r.clearMatchCacheLocked()
		return nil
	}

	segments := compilePathSegments(fullPath)
	if err := validateRouteSegments(fullPath, segments); err != nil {
		return fmt.Errorf("router add_route %s %s: %w", method, fullPath, err)
	}
	paramKeys := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg.isParam {
			paramKeys = append(paramKeys, seg.paramName)
			child := r.findParamChild(current)
			if child != nil {
				if child.paramName == "" {
					child.paramName = seg.paramName
				} else if child.paramName != seg.paramName {
					return fmt.Errorf("router add_route %s %s: %w", method, fullPath, fmt.Errorf("route conflict: parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName))
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
					return fmt.Errorf("router add_route %s %s: %w", method, fullPath, fmt.Errorf("route conflict: wildcard parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName))
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
		return fmt.Errorf("router add_route %s %s: %w", method, fullPath, fmt.Errorf("duplicate route registration: %s %s", method, fullPath))
	}
	current.handler = handler
	current.paramKeys = paramKeys
	current.fullPath = fullPath
	r.storeRouteMetaLocked(method, fullPath, meta)

	r.state.routes[method] = append(r.state.routes[method], route{Method: method, Path: fullPath})
	r.clearMatchCacheLocked()
	return nil
}

func (r *Router) clearMatchCacheLocked() {
	if r.state.matchCache != nil {
		r.state.matchCache.Clear()
	}
}

func routeMetaFromOptions(opts ...RouteOption) RouteMeta {
	meta := RouteMeta{}
	if len(opts) == 0 {
		return meta
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&meta)
		}
	}
	return meta
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

func validateRouteSegments(path string, segments []segment) error {
	for i, seg := range segments {
		switch {
		case seg.isParam && seg.paramName == "":
			return fmt.Errorf("route pattern %s has empty parameter name", path)
		case seg.isWild && seg.paramName == "":
			return fmt.Errorf("route pattern %s has empty wildcard name", path)
		case seg.isWild && i != len(segments)-1:
			return fmt.Errorf("route pattern %s has non-terminal wildcard segment", path)
		}
	}
	return nil
}
