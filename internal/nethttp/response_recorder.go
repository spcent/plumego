package nethttp

import (
	stdhttp "net/http"

	httputil "github.com/spcent/plumego/internal/httputil"
)

// ResponseRecorder captures response data while still writing to the underlying writer.
type ResponseRecorder = httputil.ResponseRecorder

// NewResponseRecorder creates a response recorder with sane defaults.
func NewResponseRecorder(w stdhttp.ResponseWriter) *ResponseRecorder {
	return httputil.NewResponseRecorder(w)
}
