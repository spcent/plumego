package router

import (
	"net/http"
	"sync"
)

// metricsResponseWriter wraps http.ResponseWriter to capture the status code
// and total bytes written for metrics reporting.
type metricsResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	if m.status == 0 {
		m.status = code
	}
	m.ResponseWriter.WriteHeader(code)
}

func (m *metricsResponseWriter) Write(b []byte) (int, error) {
	if m.status == 0 {
		m.status = http.StatusOK
	}
	n, err := m.ResponseWriter.Write(b)
	m.bytes += n
	return n, err
}

func (m *metricsResponseWriter) reset(w http.ResponseWriter) {
	m.ResponseWriter = w
	m.status = 0
	m.bytes = 0
}

var metricsWriterPool = sync.Pool{
	New: func() any { return &metricsResponseWriter{} },
}

func getMetricsWriter(w http.ResponseWriter) *metricsResponseWriter {
	mw := metricsWriterPool.Get().(*metricsResponseWriter)
	mw.reset(w)
	return mw
}

func putMetricsWriter(mw *metricsResponseWriter) {
	mw.ResponseWriter = nil
	metricsWriterPool.Put(mw)
}

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
