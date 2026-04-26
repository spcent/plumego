package bodylimit

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	mw "github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

var errRequestTooLarge = errors.New("request body too large")

// BodyLimit enforces a maximum request body size using a protective reader that
// surfaces a structured error to the client instead of the default plaintext
// response from http.MaxBytesReader.
func BodyLimit(maxBytes int64, logger log.StructuredLogger) mw.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			limited := &limitedBodyReader{
				w:        w,
				req:      r,
				r:        r.Body,
				maxBytes: maxBytes,
				logger:   logger,
				now:      time.Now,
			}
			r.Body = limited

			next.ServeHTTP(w, r)
		})
	}
}

type limitedBodyReader struct {
	w        http.ResponseWriter
	req      *http.Request
	r        io.ReadCloser
	maxBytes int64
	used     int64
	exceeded bool
	logger   log.StructuredLogger
	now      func() time.Time
}

func (l *limitedBodyReader) Read(p []byte) (int, error) {
	if l.exceeded {
		return 0, errRequestTooLarge
	}

	remaining := l.maxBytes - l.used
	if remaining <= 0 {
		var probe [1]byte
		n, err := l.r.Read(probe[:])
		if n > 0 {
			l.used += int64(n)
			return l.fail()
		}
		return 0, err
	}

	if int64(len(p)) > remaining {
		p = p[:int(remaining)+1]
	}

	n, err := l.r.Read(p)
	if int64(n) > remaining {
		l.used += int64(n)
		return int(remaining), l.failErr()
	}

	l.used += int64(n)
	return n, err
}

func (l *limitedBodyReader) Close() error {
	return l.r.Close()
}

func (l *limitedBodyReader) fail() (int, error) {
	return 0, l.failErr()
}

func (l *limitedBodyReader) failErr() error {
	if !l.exceeded {
		l.exceeded = true
		mw.WriteTransportError(l.w, l.req, http.StatusRequestEntityTooLarge, contract.CodeRequestBodyTooLarge, "request body exceeds configured limit", contract.CategoryClient, map[string]any{
			"max_bytes":  l.maxBytes,
			"seen_bytes": l.used,
			"at":         l.now().UTC(),
		})
		if l.logger != nil {
			fields := internalobs.MiddlewareLogFields(l.req, http.StatusRequestEntityTooLarge, 0)
			fields["max_bytes"] = l.maxBytes
			fields["seen_bytes"] = l.used
			l.logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Warn("request body too large")
		}
	}

	return errRequestTooLarge
}
