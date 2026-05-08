package router

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

func fastBuildCacheKey(method, path string) string {
	return method + ":" + path
}

// buildParamMap creates a parameter map for the given keys and values.
// The map is handed to the request context and GC'd with the request.
func buildParamMap(paramValues []string, paramKeys []string) map[string]string {
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
