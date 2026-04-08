package router

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// noBodyWriter wraps http.ResponseWriter and discards all body writes.
// It is used when HEAD request is served via a GET handler (RFC 7231 §4.3.2).
type noBodyWriter struct {
	http.ResponseWriter
}

func (noBodyWriter) Write([]byte) (int, error) { return 0, nil }
func (noBodyWriter) ReadFrom(io.Reader) (int64, error) {
	return 0, nil
}

// ServeHTTP implements http.Handler and handles incoming HTTP requests.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := r.normalizePath(req.URL.Path)
	cachePath := path
	if path != "/" {
		cachePath = "/" + path
	}

	cacheKey := fastBuildCacheKey(req.Method, cachePath)
	if cachedResult, paramValues, exists := r.state.matchCache.Lookup(req.Method, cachePath, cacheKey); exists {
		// For HEAD requests served by a GET handler, suppress the body.
		effectiveW := w
		if req.Method == HEAD && cachedResult.RouteMethod == GET {
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
				contract.WriteError(w, req, contract.NewErrorBuilder().
					Status(http.StatusMethodNotAllowed).
					Code("METHOD_NOT_ALLOWED").
					Message(http.StatusText(http.StatusMethodNotAllowed)).
					Category(contract.CategoryClient).
					Build())
				return
			}
		}
		http.NotFound(w, req)
		return
	}

	// For HEAD requests auto-served via the GET handler, suppress the body.
	if req.Method == HEAD && result.RouteMethod == GET {
		w = noBodyWriter{w}
	}

	if result.RouteMethod == "" {
		if matchedAny {
			result.RouteMethod = ANY
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
	if params != nil && r.writeValidationError(w, req, result.Validation, params) {
		return
	}

	if result.RoutePattern != "" && isParameterized(result.RoutePattern) {
		r.state.matchCache.SetPattern(req.Method, result.RoutePattern, result)
	} else {
		r.state.matchCache.Set(cacheKey, result)
	}

	r.applyMiddlewareAndServe(w, req, params, result)
}

func (r *Router) matchRoute(method, path string) (*MatchResult, bool) {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	if path == "/" {
		tree := r.state.trees[method]
		if tree != nil && tree.handler != nil {
			return &MatchResult{
				Handler:          tree.handler,
				RouteMiddlewares: tree.middlewares,
				RoutePattern:     "/",
				RouteMethod:      method,
				Validation:       tree.validation,
			}, false
		}
		// RFC 7231 §4.3.2: HEAD is identical to GET except no body.
		if method == HEAD {
			if getTree := r.state.trees[GET]; getTree != nil && getTree.handler != nil {
				return &MatchResult{
					Handler:          getTree.handler,
					RouteMiddlewares: getTree.middlewares,
					RoutePattern:     "/",
					RouteMethod:      GET,
					Validation:       getTree.validation,
				}, false
			}
		}
		if method != ANY {
			if anyTree := r.state.trees[ANY]; anyTree != nil && anyTree.handler != nil {
				return &MatchResult{
					Handler:          anyTree.handler,
					RouteMiddlewares: anyTree.middlewares,
					RoutePattern:     "/",
					RouteMethod:      ANY,
					Validation:       anyTree.validation,
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
			return result, false
		}
	}

	// RFC 7231 §4.3.2: fall back to GET handler for HEAD requests.
	if method == HEAD {
		if getTree := r.state.trees[GET]; getTree != nil {
			getMatcher := getRouteMatcher(getTree)
			result := getMatcher.Match(parts)
			putRouteMatcher(getMatcher)
			if result != nil {
				return result, false
			}
		}
	}

	// Fall back to ANY handler.
	if method != ANY {
		if anyTree := r.state.trees[ANY]; anyTree != nil {
			anyMatcher := getRouteMatcher(anyTree)
			result := anyMatcher.Match(parts)
			putRouteMatcher(anyMatcher)
			if result != nil {
				return result, true
			}
		}
	}

	return nil, false
}

func (r *Router) normalizePath(path string) string {
	return fastNormalizePath(path)
}

func (r *Router) serveCachedMatch(w http.ResponseWriter, req *http.Request, result *MatchResult, paramValues []string) {
	if paramValues == nil {
		paramValues = result.ParamValues
	}

	params := r.buildParamMap(paramValues, result.ParamKeys)
	if params != nil && r.writeValidationError(w, req, result.Validation, params) {
		return
	}

	r.applyMiddlewareAndServe(w, req, params, result)
}

func (r *Router) writeValidationError(w http.ResponseWriter, req *http.Request, validation *RouteValidation, params map[string]string) bool {
	if validation == nil {
		return false
	}
	if err := validation.Validate(params); err != nil {
		contract.WriteError(w, req, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code("VALIDATION_ERROR").
			Message(err.Error()).
			Category(contract.CategoryValidation).
			Build())
		return true
	}
	return false
}

func (r *Router) buildParamMap(paramValues []string, paramKeys []string) map[string]string {
	return buildParamMapPooled(paramValues, paramKeys)
}

func (r *Router) applyMiddlewareAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, result *MatchResult) {
	if result == nil {
		return
	}

	ctx := req.Context()
	existingRC := contract.RequestContextFromContext(ctx)

	if len(params) > 0 {
		existingRC.Params = params
	}
	if result.RoutePattern != "" {
		existingRC.RoutePattern = result.RoutePattern
	}
	meta := r.metaFor(result.RouteMethod, result.RoutePattern)
	if meta.Name != "" {
		existingRC.RouteName = meta.Name
	}

	// Merge params into context using a single WithValue call.
	ctx = contract.WithRequestContext(ctx, existingRC)
	reqWithParams := req.WithContext(ctx)

	// Obtain the handler (possibly from middleware-chain cache).
	version := r.middlewareManager.getVersion()
	var handler http.Handler
	if cached, ok := result.loadCached(version); ok {
		handler = cached
	} else {
		allMiddlewares := r.middlewareManager.getMiddlewares()
		combined := make([]middleware.Middleware, 0, len(allMiddlewares)+len(result.RouteMiddlewares))
		combined = append(combined, allMiddlewares...)
		combined = append(combined, result.RouteMiddlewares...)

		if len(combined) == 0 {
			handler = result.Handler
		} else {
			chain := middleware.NewChain(combined...)
			handler = chain.Build(result.Handler)
			result.storeCached(version, handler)
		}
	}

	handler.ServeHTTP(w, reqWithParams)
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
			if method == ANY || tree == nil {
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
			if method == ANY || tree == nil {
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
	return allowed
}

func (r *Router) findRouteNodeLocked(method, path string) *node {
	tree := r.state.trees[method]
	if tree == nil {
		return nil
	}
	if path == "/" {
		if tree.handler == nil {
			return nil
		}
		return tree
	}

	current := tree
	for _, seg := range compilePathSegments(path) {
		switch {
		case seg.isParam:
			current = r.findParamChild(current)
		case seg.isWild:
			current = r.findWildChild(current)
		default:
			current = r.findChild(current, seg.raw)
		}
		if current == nil {
			return nil
		}
	}
	if current.handler == nil {
		return nil
	}
	return current
}
