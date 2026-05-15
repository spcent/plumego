package router

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/spcent/plumego/contract"
)

// noBodyWriter wraps http.ResponseWriter and discards all body writes.
// It is used for matched HEAD requests, including explicit HEAD routes, MethodAny
// fallback, and GET fallback (RFC 7231 §4.3.2).
type noBodyWriter struct {
	http.ResponseWriter
}

func (noBodyWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w noBodyWriter) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(io.Discard, r)
}
func (w noBodyWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// ServeHTTP implements http.Handler and handles incoming HTTP requests.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !r.ready() || req == nil {
		http.Error(w, "router not initialized", http.StatusServiceUnavailable)
		return
	}
	if req.URL == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	path := fastNormalizePath(req.URL.Path)
	cachePath := path
	if path != "/" {
		cachePath = "/" + path
	}

	cacheKey := fastBuildCacheKey(req.Method, cachePath)
	if cachedResult, exists := r.state.matchCache.Get(cacheKey); exists {
		effectiveW := responseWriterForMatch(w, req, cachedResult)
		params := buildParamMap(cachedResult.ParamValues, cachedResult.ParamKeys)
		r.attachRouteContextAndServe(effectiveW, req, params, cachedResult)
		return
	}

	result := r.matchRoute(req.Method, path)
	if result == nil {
		if r.state.methodNotAllowed.Load() {
			allowed := r.allowedMethods(path)
			if len(allowed) > 0 {
				w.Header().Set("Allow", strings.Join(allowed, ", "))
				_ = contract.WriteError(w, req, contract.NewErrorBuilder().
					Type(contract.TypeMethodNotAllowed).
					Message("method not allowed").
					Build())
				return
			}
		}
		http.NotFound(w, req)
		return
	}

	params := buildParamMap(result.ParamValues, result.ParamKeys)

	r.state.matchCache.Set(cacheKey, result)

	w = responseWriterForMatch(w, req, result)
	r.attachRouteContextAndServe(w, req, params, result)
}

func responseWriterForMatch(w http.ResponseWriter, req *http.Request, result *matchResult) http.ResponseWriter {
	if req != nil && req.Method == http.MethodHead && result != nil {
		return noBodyWriter{w}
	}
	return w
}

func (r *Router) matchRoute(method, path string) *matchResult {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	if path == "/" {
		if method != MethodAny {
			tree := r.state.trees[method]
			if tree != nil && tree.handler != nil {
				return &matchResult{
					Handler:      tree.handler,
					RoutePattern: "/",
					RouteMethod:  method,
					RouteName:    tree.routeName,
				}
			}
		}
		// RFC 7231 §4.3.2: HEAD is identical to GET except no body.
		if method == http.MethodHead {
			if getTree := r.state.trees[http.MethodGet]; getTree != nil && getTree.handler != nil {
				return &matchResult{
					Handler:      getTree.handler,
					RoutePattern: "/",
					RouteMethod:  http.MethodGet,
					RouteName:    getTree.routeName,
				}
			}
		}
		if anyTree := r.state.trees[MethodAny]; anyTree != nil && anyTree.handler != nil {
			return &matchResult{
				Handler:      anyTree.handler,
				RoutePattern: "/",
				RouteMethod:  MethodAny,
				RouteName:    anyTree.routeName,
			}
		}
		return nil
	}

	partsPtr := splitPathToParts(path)
	parts := *partsPtr
	defer putPathParts(partsPtr)

	if method != MethodAny {
		tree := r.state.trees[method]
		if tree != nil {
			matcher := getRouteMatcher(tree)
			result := matcher.Match(parts)
			putRouteMatcher(matcher)
			if result != nil {
				result.RouteMethod = method
				return result
			}
		}
	}

	// RFC 7231 §4.3.2: fall back to GET handler for HEAD requests.
	if method == http.MethodHead {
		if getTree := r.state.trees[http.MethodGet]; getTree != nil {
			getMatcher := getRouteMatcher(getTree)
			result := getMatcher.Match(parts)
			putRouteMatcher(getMatcher)
			if result != nil {
				result.RouteMethod = http.MethodGet
				return result
			}
		}
	}

	// Fall back to ANY handler.
	if anyTree := r.state.trees[MethodAny]; anyTree != nil {
		anyMatcher := getRouteMatcher(anyTree)
		result := anyMatcher.Match(parts)
		putRouteMatcher(anyMatcher)
		if result != nil {
			result.RouteMethod = MethodAny
			return result
		}
	}

	return nil
}

func (r *Router) attachRouteContextAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, result *matchResult) {
	if result == nil {
		return
	}

	ctx := req.Context()
	rc := contract.RequestContext{
		Params:       params,
		RoutePattern: result.RoutePattern,
		RouteName:    result.RouteName,
	}

	ctx = contract.WithRequestContext(ctx, rc)
	reqWithParams := req.WithContext(ctx)

	result.Handler.ServeHTTP(w, reqWithParams)
}

func (r *Router) allowedMethods(path string) []string {
	normalized := path
	if normalized == "" {
		normalized = "/"
	}

	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	allowed := make([]string, 0, len(r.state.trees))
	if normalized == "/" {
		for method, tree := range r.state.trees {
			if method == MethodAny || tree == nil {
				continue
			}
			if tree.handler != nil {
				allowed = append(allowed, method)
			}
		}
	} else {
		partsPtr := splitPathToParts(normalized)
		parts := *partsPtr

		for method, tree := range r.state.trees {
			if method == MethodAny || tree == nil {
				continue
			}
			matcher := getRouteMatcher(tree)
			if matcher.Match(parts) != nil {
				allowed = append(allowed, method)
			}
			putRouteMatcher(matcher)
		}

		putPathParts(partsPtr)
	}

	if len(allowed) > 1 {
		sort.Strings(allowed)
	}
	allowed = appendImplicitHeadMethod(allowed)
	return allowed
}

func appendImplicitHeadMethod(methods []string) []string {
	hasGet := false
	hasHead := false
	for _, method := range methods {
		switch method {
		case http.MethodGet:
			hasGet = true
		case http.MethodHead:
			hasHead = true
		}
	}
	if !hasGet || hasHead {
		return methods
	}
	methods = append(methods, http.MethodHead)
	sort.Strings(methods)
	return methods
}
