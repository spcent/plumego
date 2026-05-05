package router

import (
	"strings"
	"sync"
)

// This file contains low-level path and cache helpers used by trie matching.

// Optimized path normalization without strings.Trim allocation.
// Returns the path with leading and all trailing slashes removed.
// For the root path "/" it returns "/".
func fastNormalizePath(path string) string {
	if len(path) == 0 || path == "/" {
		return "/"
	}

	start := 0
	end := len(path)

	for start < end && path[start] == '/' {
		start++
	}
	if start == end {
		return "/"
	}
	for end > start && path[end-1] == '/' {
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

// buildParamMapPooled creates a parameter map for the given keys and values.
// The map is handed to the request context and GC'd with the request.
func buildParamMapPooled(paramValues []string, paramKeys []string) map[string]string {
	if len(paramValues) == 0 || len(paramKeys) == 0 {
		return nil
	}

	minLen := len(paramValues)
	if len(paramKeys) < minLen {
		minLen = len(paramKeys)
	}

	params := make(map[string]string, minLen)
	for i := 0; i < minLen; i++ {
		params[paramKeys[i]] = paramValues[i]
	}

	return params
}
