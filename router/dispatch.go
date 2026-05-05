package router

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/spcent/plumego/contract"
)

// noBodyWriter wraps http.ResponseWriter and discards all body writes.
// It is used when HEAD request is served via a GET handler (RFC 7231 §4.3.2).
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

	path := r.normalizePath(req.URL.Path)
	cachePath := path
	if path != "/" {
		cachePath = "/" + path
	}

	cacheKey := fastBuildCacheKey(req.Method, cachePath)
	if cachedResult, paramValues, exists := r.state.matchCache.Lookup(cacheKey); exists {
		// For HEAD requests served by a GET handler, suppress the body.
		effectiveW := w
		if req.Method == http.MethodHead && cachedResult.RouteMethod == http.MethodGet {
			effectiveW = noBodyWriter{w}
		}
		r.serveCachedMatch(effectiveW, req, cachedResult, paramValues)
		return
	}

	result, matchedAny := r.matchRoute(req.Method, path)
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

	// For HEAD requests auto-served via the GET handler, suppress the body.
	if req.Method == http.MethodHead && result.RouteMethod == http.MethodGet {
		w = noBodyWriter{w}
	}

	if result.RouteMethod == "" {
		if matchedAny {
			result.RouteMethod = MethodAny
		} else {
			result.RouteMethod = req.Method
		}
	}
	if result.RoutePattern == "" {
		if path == "/" {
			result.RoutePattern = "/"
		} else {
			result.RoutePattern = "/" + path
		}
	}

	params := r.buildParamMap(result.ParamValues, result.ParamKeys)

	r.state.matchCache.Set(cacheKey, result)

	r.attachRouteContextAndServe(w, req, params, result)
}

func (r *Router) matchRoute(method, path string) (*matchResult, bool) {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	if path == "/" {
		tree := r.state.trees[method]
		if tree != nil && tree.handler != nil {
			return &matchResult{
				Handler:      tree.handler,
				RoutePattern: "/",
				RouteMethod:  method,
			}, false
		}
		// RFC 7231 §4.3.2: HEAD is identical to GET except no body.
		if method == http.MethodHead {
			if getTree := r.state.trees[http.MethodGet]; getTree != nil && getTree.handler != nil {
				return &matchResult{
					Handler:      getTree.handler,
					RoutePattern: "/",
					RouteMethod:  http.MethodGet,
				}, false
			}
		}
		if method != MethodAny {
			if anyTree := r.state.trees[MethodAny]; anyTree != nil && anyTree.handler != nil {
				return &matchResult{
					Handler:      anyTree.handler,
					RoutePattern: "/",
					RouteMethod:  MethodAny,
				}, true
			}
		}
		return nil, false
	}

	partsPtr := splitPathToParts(path)
	parts := *partsPtr
	defer putPathParts(partsPtr)

	tree := r.state.trees[method]
	if tree != nil {
		matcher := getRouteMatcher(tree)
		result := matcher.Match(parts)
		putRouteMatcher(matcher)
		if result != nil {
			result.RouteMethod = method
			return result, false
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
				return result, false
			}
		}
	}

	// Fall back to ANY handler.
	if method != MethodAny {
		if anyTree := r.state.trees[MethodAny]; anyTree != nil {
			anyMatcher := getRouteMatcher(anyTree)
			result := anyMatcher.Match(parts)
			putRouteMatcher(anyMatcher)
			if result != nil {
				result.RouteMethod = MethodAny
				return result, true
			}
		}
	}

	return nil, false
}

func (r *Router) normalizePath(path string) string {
	return fastNormalizePath(path)
}

func (r *Router) serveCachedMatch(w http.ResponseWriter, req *http.Request, result *matchResult, paramValues []string) {
	if paramValues == nil {
		paramValues = result.ParamValues
	}

	params := r.buildParamMap(paramValues, result.ParamKeys)

	r.attachRouteContextAndServe(w, req, params, result)
}

func (r *Router) buildParamMap(paramValues []string, paramKeys []string) map[string]string {
	return buildParamMapPooled(paramValues, paramKeys)
}

func (r *Router) attachRouteContextAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, result *matchResult) {
	if result == nil {
		return
	}

	ctx := req.Context()
	existingRC := contract.RequestContextFromContext(ctx)

	existingRC.Params = params
	existingRC.RoutePattern = result.RoutePattern
	existingRC.RouteName = ""
	meta := r.metaFor(result.RouteMethod, result.RoutePattern)
	if meta.Name != "" {
		existingRC.RouteName = meta.Name
	}

	// Merge params into context using a single WithValue call.
	ctx = contract.WithRequestContext(ctx, existingRC)
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
