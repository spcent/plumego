package router

import (
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

// SetRouteMeta sets metadata for a route.
func (r *Router) SetRouteMeta(method, path string, meta RouteMeta) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	pattern := r.fullPath(path)
	r.setMeta(method, pattern, meta)
	if meta.Name != "" {
		r.registerNamedRoute(meta.Name, method, pattern)
	}
}

func (r *Router) registerNamedRoute(name, method, pattern string) {
	if r.state.namedRoutes == nil {
		r.state.namedRoutes = make(map[string]*NamedRoute)
	}

	paramPos := make(map[string]int)
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	pos := 0
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramPos[part[1:]] = pos
			pos++
		} else if strings.HasPrefix(part, "*") {
			paramPos[part[1:]] = pos
			pos++
		}
	}

	r.state.namedRoutes[name] = &NamedRoute{
		Method:   method,
		Pattern:  pattern,
		ParamPos: paramPos,
	}
}

// URL generates a URL for a named route with the given parameters.
func (r *Router) URL(name string, params ...string) string {
	r.state.mu.RLock()
	namedRoute, exists := r.state.namedRoutes[name]
	r.state.mu.RUnlock()
	if !exists {
		return ""
	}

	paramMap := make(map[string]string)
	for i := 0; i < len(params)-1; i += 2 {
		paramMap[params[i]] = params[i+1]
	}

	result := namedRoute.Pattern
	parts := strings.Split(strings.Trim(result, "/"), "/")
	resultParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			if val, ok := paramMap[paramName]; ok {
				resultParts = append(resultParts, url.PathEscape(val))
			} else {
				resultParts = append(resultParts, part)
			}
		} else if strings.HasPrefix(part, "*") {
			paramName := part[1:]
			if val, ok := paramMap[paramName]; ok {
				segments := strings.Split(val, "/")
				for i, seg := range segments {
					segments[i] = url.PathEscape(seg)
				}
				resultParts = append(resultParts, strings.Join(segments, "/"))
			} else {
				resultParts = append(resultParts, "")
			}
		} else {
			resultParts = append(resultParts, part)
		}
	}

	return "/" + strings.Join(resultParts, "/")
}

// URLMust generates a URL for a named route and panics if the route doesn't exist.
func (r *Router) URLMust(name string, params ...string) string {
	url := r.URL(name, params...)
	if url == "" {
		panic(fmt.Sprintf("named route %q not found", name))
	}
	return url
}

// HasRoute checks if a named route exists.
func (r *Router) HasRoute(name string) bool {
	r.state.mu.RLock()
	_, exists := r.state.namedRoutes[name]
	r.state.mu.RUnlock()
	return exists
}

// NamedRoutes returns a copy of all registered named routes.
func (r *Router) NamedRoutes() map[string]*NamedRoute {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	result := make(map[string]*NamedRoute, len(r.state.namedRoutes))
	for k, v := range r.state.namedRoutes {
		paramPos := make(map[string]int, len(v.ParamPos))
		for pk, pv := range v.ParamPos {
			paramPos[pk] = pv
		}
		result[k] = &NamedRoute{
			Method:   v.Method,
			Pattern:  v.Pattern,
			ParamPos: paramPos,
		}
	}
	return result
}

// Routes returns a snapshot of all registered routes with metadata.
func (r *Router) Routes() []RouteInfo {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	infos := make([]RouteInfo, 0)
	for _, routes := range r.state.routes {
		for _, entry := range routes {
			meta := r.metaFor(entry.Method, entry.Path)
			infos = append(infos, RouteInfo{
				Method: entry.Method,
				Path:   entry.Path,
				Meta:   meta,
			})
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Method == infos[j].Method {
			return infos[i].Path < infos[j].Path
		}
		return infos[i].Method < infos[j].Method
	})

	return infos
}

// Print prints all registered routes grouped by method.
func (r *Router) Print(w io.Writer) {
	fmt.Fprintln(w, "Registered Routes:")

	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	methods := make([]string, 0, len(r.state.routes))
	for method := range r.state.routes {
		methods = append(methods, method)
	}
	sort.Strings(methods)

	for _, method := range methods {
		routes := append([]route(nil), r.state.routes[method]...)
		sort.Slice(routes, func(i, j int) bool { return routes[i].Path < routes[j].Path })

		for _, route := range routes {
			label := route.Path
			if strings.Contains(route.Path, "/*") {
				label += "   [wildcard]"
			}
			fmt.Fprintf(w, "%-6s %s\n", route.Method, label)
		}
	}
}
