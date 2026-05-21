package router

// matchNode traverses the trie to find the best matching route for a given URL path.
// rawPath is the normalized path string from which parts were sliced; it is used
// for zero-allocation wildcard suffix extraction.
//
// Matching strategy:
//  1. Try static path segments first (exact match)
//  2. Try parameter segments (dynamic segments like ":id")
//  3. Try wildcard segments (catch-all segments like "*path")
func matchNode(root *node, parts []string, rawPath string) *matchResult {
	if root == nil {
		return nil
	}

	current := root
	// Use pooled slice for parameter values to avoid per-request allocation
	pvPtr := getParamValues()
	paramValues := *pvPtr

	for i, pathSegment := range parts {
		// Reject empty path segments (e.g., from double slashes /users//123).
		if pathSegment == "" {
			*pvPtr = paramValues
			putParamValues(pvPtr)
			return nil
		}

		// Try to find exact match first
		if child := findStaticChild(current, pathSegment); child != nil {
			current = child
			continue
		}

		// Try param match
		if paramChild := findParamChild(current); paramChild != nil {
			paramValues = append(paramValues, pathSegment)
			current = paramChild
			continue
		}

		// Try wildcard match
		if wildChild := findWildChild(current); wildChild != nil {
			// Compute the byte offset of parts[i] in rawPath via the lengths of
			// preceding segments (each separated by one '/'), then slice directly
			// to avoid the strings.Join allocation.
			off := 1 // skip the leading '/'
			for j := 0; j < i; j++ {
				off += len(parts[j]) + 1 // +1 for the '/' separator
			}
			paramValues = append(paramValues, rawPath[off:])
			current = wildChild
			break
		}

		// No match found
		*pvPtr = paramValues
		putParamValues(pvPtr)
		return nil
	}

	// Check if we found a valid handler
	if current == nil || current.handler == nil {
		*pvPtr = paramValues
		putParamValues(pvPtr)
		return nil
	}

	// Copy param values out of the pooled slice so we can return it to the pool.
	// For zero params this is a nil slice (no allocation).
	var resultParams []string
	if len(paramValues) > 0 {
		resultParams = make([]string, len(paramValues))
		copy(resultParams, paramValues)
	}
	*pvPtr = paramValues
	putParamValues(pvPtr)

	return &matchResult{
		Handler:      current.handler,
		ParamValues:  resultParams,
		ParamKeys:    current.paramKeys,
		RoutePattern: current.fullPath,
		RouteName:    current.routeName,
	}
}
