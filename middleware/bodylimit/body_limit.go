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
		return l.fail()
	}

	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := l.r.Read(p)
	l.used += int64(n)

	if l.used > l.maxBytes {
		return l.fail()
	}

	return n, err
}

func (l *limitedBodyReader) Close() error {
	return l.r.Close()
}

func (l *limitedBodyReader) fail() (int, error) {
	if !l.exceeded {
		l.exceeded = true
		mw.WriteTransportError(l.w, nil, http.StatusRequestEntityTooLarge, mw.CodeRequestBodyTooLarge, "request body exceeds configured limit", contract.CategoryClient, map[string]any{
			"max_bytes":  l.maxBytes,
			"seen_bytes": l.used,
			"at":         l.now().UTC(),
		})
		if l.logger != nil {
			fields := internalobs.MiddlewareLogFields(nil, http.StatusRequestEntityTooLarge, 0)
			fields["max_bytes"] = l.maxBytes
			fields["seen_bytes"] = l.used
			l.logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Warn("request body too large")
		}
	}

	return 0, errRequestTooLarge
}
