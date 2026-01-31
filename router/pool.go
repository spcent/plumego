package router

import (
	"sync"
)

// Object pools for reducing allocations during request processing

// paramValuesPool is a pool for []string slices used to store parameter values
var paramValuesPool = sync.Pool{
	New: func() any {
		// Pre-allocate with common capacity (most routes have < 4 params)
		s := make([]string, 0, 4)
		return &s
	},
}

// GetParamValues retrieves a string slice from the pool
func GetParamValues() *[]string {
	return paramValuesPool.Get().(*[]string)
}

// PutParamValues returns a string slice to the pool after resetting it
func PutParamValues(s *[]string) {
	if s == nil {
		return
	}
	// Reset slice length but keep capacity
	*s = (*s)[:0]
	paramValuesPool.Put(s)
}

// paramsMapPool is a pool for map[string]string used to store route parameters
var paramsMapPool = sync.Pool{
	New: func() any {
		return make(map[string]string, 4)
	},
}

// GetParamsMap retrieves a params map from the pool
func GetParamsMap() map[string]string {
	return paramsMapPool.Get().(map[string]string)
}

// PutParamsMap returns a params map to the pool after clearing it
func PutParamsMap(m map[string]string) {
	if m == nil {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	paramsMapPool.Put(m)
}

// routeMatcherPool is a pool for RouteMatcher instances
var routeMatcherPool = sync.Pool{
	New: func() any {
		return &RouteMatcher{}
	},
}

// GetRouteMatcher retrieves a RouteMatcher from the pool
func GetRouteMatcher(root *node) *RouteMatcher {
	rm := routeMatcherPool.Get().(*RouteMatcher)
	rm.root = root
	return rm
}

// PutRouteMatcher returns a RouteMatcher to the pool
func PutRouteMatcher(rm *RouteMatcher) {
	if rm == nil {
		return
	}
	rm.root = nil
	routeMatcherPool.Put(rm)
}

// pathPartsPool is a pool for []string slices used to store split path parts
var pathPartsPool = sync.Pool{
	New: func() any {
		// Pre-allocate with common capacity (most paths have < 8 segments)
		s := make([]string, 0, 8)
		return &s
	},
}

// GetPathParts retrieves a path parts slice from the pool
func GetPathParts() *[]string {
	return pathPartsPool.Get().(*[]string)
}

// PutPathParts returns a path parts slice to the pool after resetting it
func PutPathParts(s *[]string) {
	if s == nil {
		return
	}
	*s = (*s)[:0]
	pathPartsPool.Put(s)
}

// SplitPathToParts splits a path into parts using a pooled slice
// This preserves empty segments to properly reject paths like /users//123
// Caller must call PutPathParts when done
func SplitPathToParts(path string) *[]string {
	parts := GetPathParts()
	if path == "" || path == "/" {
		return parts
	}

	// Trim leading/trailing slashes
	start := 0
	end := len(path)
	if path[0] == '/' {
		start = 1
	}
	if end > start && path[end-1] == '/' {
		end--
	}

	// Split by '/', preserving empty segments
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
