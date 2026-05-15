package transport

import httputil "github.com/spcent/plumego/internal/httputil"

// BufferedResponse captures headers, status, and body in memory.
// It does not write to an underlying ResponseWriter.
type BufferedResponse = httputil.BufferedResponseRecorder

// NewBufferedResponse creates a buffered response with an optional max size.
// A maxBytes of 0 disables the size limit.
func NewBufferedResponse(maxBytes int) *BufferedResponse {
	return httputil.NewBufferedResponse(maxBytes)
}
