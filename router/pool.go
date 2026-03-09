package router

import (
	"sync"
)

// Object pools for reducing allocations during request processing.
// All pool helpers are unexported; they are implementation details.

var paramValuesPool = sync.Pool{
	New: func() any {
		s := make([]string, 0, DefaultPoolSliceCap)
		return &s
	},
}

func getParamValues() *[]string {
	return paramValuesPool.Get().(*[]string)
}

func putParamValues(s *[]string) {
	if s == nil {
		return
	}
	*s = (*s)[:0]
	paramValuesPool.Put(s)
}

var routeMatcherPool = sync.Pool{
	New: func() any {
		return &RouteMatcher{}
	},
}

func getRouteMatcher(root *node) *RouteMatcher {
	rm := routeMatcherPool.Get().(*RouteMatcher)
	rm.root = root
	return rm
}

func putRouteMatcher(rm *RouteMatcher) {
	if rm == nil {
		return
	}
	rm.root = nil
	routeMatcherPool.Put(rm)
}

var pathPartsPool = sync.Pool{
	New: func() any {
		s := make([]string, 0, DefaultPathPartsCap)
		return &s
	},
}

func getPathParts() *[]string {
	return pathPartsPool.Get().(*[]string)
}

func putPathParts(s *[]string) {
	if s == nil {
		return
	}
	*s = (*s)[:0]
	pathPartsPool.Put(s)
}

// splitPathToParts splits a path into parts using a pooled slice.
// This preserves empty segments to properly reject paths like /users//123.
// Caller must call putPathParts when done.
func splitPathToParts(path string) *[]string {
	parts := getPathParts()
	if path == "" || path == "/" {
		return parts
	}

	start := 0
	end := len(path)
	if path[0] == '/' {
		start = 1
	}
	if end > start && path[end-1] == '/' {
		end--
	}

	for i := start; i < end; i++ {
		if path[i] == '/' {
			*parts = append(*parts, path[start:i])
			start = i + 1
		}
	}
	if start <= end {
		*parts = append(*parts, path[start:end])
	}

	return parts
}
