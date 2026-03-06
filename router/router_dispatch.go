package router

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// ServeHTTP implements http.Handler and handles incoming HTTP requests.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := r.normalizePath(req.URL.Path)
	cachePath := path
	if path != "/" {
		cachePath = "/" + path
	}

	cacheKey := fastBuildCacheKey(req.Method, cachePath)
	if cachedResult, paramValues, exists := r.state.routeCache.Lookup(req.Method, cachePath, cacheKey); exists {
		r.serveCachedMatch(w, req, cachedResult, paramValues)
		return
	}

	result, matchedAny := r.matchRoute(req.Method, path)
	if result == nil {
		r.state.mu.RLock()
		methodNotAllowed := r.state.methodNotAllowed
		r.state.mu.RUnlock()

		if methodNotAllowed {
			allowed := r.allowedMethods(path)
			if len(allowed) > 0 {
				w.Header().Set("Allow", strings.Join(allowed, ", "))
				contract.WriteError(w, req, contract.APIError{
					Status:   http.StatusMethodNotAllowed,
					Code:     "METHOD_NOT_ALLOWED",
					Message:  http.StatusText(http.StatusMethodNotAllowed),
					Category: contract.CategoryClient,
				})
				return
			}
		}
		http.NotFound(w, req)
		return
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

	if result.RoutePattern != "" && IsParameterized(result.RoutePattern) {
		r.state.routeCache.SetPattern(req.Method, result.RoutePattern, result)
	} else {
		r.state.routeCache.Set(cacheKey, result)
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

	tree := r.state.trees[method]
	if tree == nil {
		tree = r.state.trees[ANY]
		if tree == nil {
			return nil, false
		}
	}

	partsPtr := SplitPathToParts(path)
	parts := *partsPtr
	defer PutPathParts(partsPtr)

	matcher := GetRouteMatcher(tree)
	result := matcher.Match(parts)
	PutRouteMatcher(matcher)

	matchedAny := false
	if result == nil && method != ANY {
		if anyTree := r.state.trees[ANY]; anyTree != nil {
			anyMatcher := GetRouteMatcher(anyTree)
			result = anyMatcher.Match(parts)
			PutRouteMatcher(anyMatcher)
			if result != nil {
				matchedAny = true
			}
		}
	}

	return result, matchedAny
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
		contract.WriteError(w, req, contract.APIError{
			Status:   http.StatusBadRequest,
			Code:     "VALIDATION_ERROR",
			Message:  err.Error(),
			Category: contract.CategoryValidation,
		})
		return true
	}
	return false
}

func (r *Router) buildParamMap(paramValues []string, paramKeys []string) map[string]string {
	return buildParamMapPooled(paramValues, paramKeys)
}

func (r *Router) applyMiddlewareAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, result *MatchResult) {
	reqWithParams := req
	ctx := req.Context()

	existingRC, ok := ctx.Value(contract.RequestContextKey{}).(contract.RequestContext)
	if !ok {
		existingRC = contract.RequestContext{}
	}

	if len(params) > 0 {
		existingRC.Params = params
	}
	if result != nil {
		if result.RoutePattern != "" {
			existingRC.RoutePattern = result.RoutePattern
		}
		meta := r.metaFor(result.RouteMethod, result.RoutePattern)
		if meta.Name != "" {
			existingRC.RouteName = meta.Name
		}
	}

	if len(params) > 0 {
		ctx = context.WithValue(ctx, contract.ParamsContextKey{}, params)
	}
	ctx = context.WithValue(ctx, contract.RequestContextKey{}, existingRC)
	if ctx != req.Context() {
		reqWithParams = req.WithContext(ctx)
	}

	if result == nil {
		return
	}

	version := r.middlewareManager.Version()
	if cached, ok := result.loadCached(version); ok {
		cached.ServeHTTP(w, reqWithParams)
		return
	}

	allMiddlewares := r.middlewareManager.GetMiddlewares()
	combined := make([]middleware.Middleware, 0, len(allMiddlewares)+len(result.RouteMiddlewares))
	combined = append(combined, allMiddlewares...)
	combined = append(combined, result.RouteMiddlewares...)

	if len(combined) == 0 {
		result.Handler.ServeHTTP(w, reqWithParams)
		return
	}

	chain := middleware.NewChain(combined...)
	wrappedHandler := chain.Apply(result.Handler)
	result.storeCached(version, wrappedHandler)
	wrappedHandler.ServeHTTP(w, reqWithParams)
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
		partsPtr := SplitPathToParts(normalized)
		parts := *partsPtr

		for method, tree := range r.state.trees {
			if method == ANY || tree == nil {
				continue
			}
			matcher := GetRouteMatcher(tree)
			if matcher.Match(parts) != nil {
				allowed = append(allowed, method)
			}
			PutRouteMatcher(matcher)
		}

		PutPathParts(partsPtr)
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
