package router

import "net/http"

func mustAddRoute(r *Router, method, path string, handler http.Handler, opts ...RouteOption) {
	if err := r.AddRoute(method, path, handler, opts...); err != nil {
		panic(err)
	}
}
