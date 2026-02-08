package router

import (
	"strings"
	"sync"
)

// Optimized path normalization without strings.Trim allocation.
// Returns the path with leading/trailing slashes removed.
// For the root path "/" it returns "/".
func fastNormalizePath(path string) string {
	if len(path) == 0 || path == "/" {
		return "/"
	}

	start := 0
	end := len(path)

	if path[0] == '/' {
		start = 1
	}
	if end > start && path[end-1] == '/' {
		end--
	}

	if start == 0 && end == len(path) {
		return path
	}
	return path[start:end]
}

// fastBuildCacheKey builds a cache key from method and path without allocation
// by using a pooled strings.Builder.
var cacheKeyPool = sync.Pool{
	New: func() any {
		b := &strings.Builder{}
		b.Grow(64)
		return b
	},
}

func fastBuildCacheKey(method, path string) string {
	b := cacheKeyPool.Get().(*strings.Builder)
	b.Reset()
	b.WriteString(method)
	b.WriteByte(':')
	b.WriteString(path)
	s := b.String()
	cacheKeyPool.Put(b)
	return s
}

// precompiledPattern stores a pre-split pattern for fast matching
// without repeated string splitting on every cache lookup.
type precompiledPattern struct {
	pattern      string   // Original pattern string
	parts        []string // Pre-split pattern parts
	paramIndices []int    // Indices of parameter parts (starting with ':')
	wildcardIdx  int      // Index of wildcard part (-1 if none)
	numParts     int      // Length of parts slice
}

func newPrecompiledPattern(pattern string) precompiledPattern {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	pp := precompiledPattern{
		pattern:  pattern,
		parts:    parts,
		numParts: len(parts),
	}

	pp.wildcardIdx = -1
	for i, part := range parts {
		if len(part) > 0 && part[0] == '*' {
			pp.wildcardIdx = i
			break
		}
		if len(part) > 0 && part[0] == ':' {
			pp.paramIndices = append(pp.paramIndices, i)
		}
	}

	return pp
}

// matchPrecompiled matches path parts against a precompiled pattern,
// extracting parameter values into the provided slice to avoid allocation.
// Returns the number of extracted params and whether the match succeeded.
func matchPrecompiled(pp *precompiledPattern, pathParts []string, paramBuf []string) (int, bool) {
	if pp.wildcardIdx < 0 {
		// No wildcard: path must have exact same segment count
		if len(pathParts) != pp.numParts {
			return 0, false
		}
	} else {
		// Wildcard present: path must have at least wildcardIdx segments
		if len(pathParts) < pp.wildcardIdx {
			return 0, false
		}
	}

	paramCount := 0
	for i := 0; i < pp.numParts; i++ {
		part := pp.parts[i]

		if len(part) > 0 && part[0] == '*' {
			// Wildcard: capture remaining path
			if paramCount < len(paramBuf) {
				paramBuf[paramCount] = strings.Join(pathParts[i:], "/")
				paramCount++
			}
			return paramCount, true
		}

		if i >= len(pathParts) {
			return 0, false
		}

		if len(part) > 0 && part[0] == ':' {
			// Parameter: capture value
			if paramCount < len(paramBuf) {
				paramBuf[paramCount] = pathParts[i]
				paramCount++
			}
		} else if part != pathParts[i] {
			// Static: must match exactly
			return 0, false
		}
	}

	return paramCount, true
}

// matchParamBufPool provides reusable buffers for parameter extraction
// during pattern cache lookups.
var matchParamBufPool = sync.Pool{
	New: func() any {
		buf := make([]string, 8)
		return &buf
	},
}

// fastSplitPath splits a path by '/' without allocating a new slice,
// writing results into the provided buffer. Returns the number of parts written.
func fastSplitPath(path string, buf []string) int {
	if len(path) == 0 || path == "/" {
		return 0
	}

	start := 0
	end := len(path)
	if path[0] == '/' {
		start = 1
	}
	if end > start && path[end-1] == '/' {
		end--
	}

	count := 0
	segStart := start
	for i := start; i < end; i++ {
		if path[i] == '/' {
			if count < len(buf) {
				buf[count] = path[segStart:i]
				count++
			}
			segStart = i + 1
		}
	}
	if segStart <= end && count < len(buf) {
		buf[count] = path[segStart:end]
		count++
	}
	return count
}

// pathBufPool provides reusable path segment buffers for splitting.
var pathBufPool = sync.Pool{
	New: func() any {
		buf := make([]string, 16)
		return &buf
	},
}

// buildParamMapPooled creates a parameter map using a pooled map
// to reduce allocations on the hot path. The caller does NOT need
// to return the map to the pool — it is handed to context and GC'd normally.
func buildParamMapPooled(paramValues []string, paramKeys []string) map[string]string {
	if len(paramValues) == 0 || len(paramKeys) == 0 {
		return nil
	}

	params := GetParamsMap()

	minLen := len(paramValues)
	if len(paramKeys) < minLen {
		minLen = len(paramKeys)
	}

	for i := 0; i < minLen; i++ {
		params[paramKeys[i]] = paramValues[i]
	}

	return params
}
