package bodylimit

import (
	"bufio"
	"errors"
	"io"
	"net"
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

			bw := &bodyLimitResponseWriter{ResponseWriter: w}
			limited := &limitedBodyReader{
				w:        bw,
				req:      r,
				r:        r.Body,
				maxBytes: maxBytes,
				logger:   logger,
				now:      time.Now,
			}
			r.Body = limited

			next.ServeHTTP(bw, r)
		})
	}
}

type bodyLimitResponseWriter struct {
	http.ResponseWriter
	started bool
	blocked bool
}

func (w *bodyLimitResponseWriter) WriteHeader(statusCode int) {
	if w.blocked {
		return
	}
	w.started = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *bodyLimitResponseWriter) Write(p []byte) (int, error) {
	if w.blocked {
		return len(p), nil
	}
	w.started = true
	return w.ResponseWriter.Write(p)
}

func (w *bodyLimitResponseWriter) Flush() {
	if w.blocked {
		return
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *bodyLimitResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	w.started = true
	return hijacker.Hijack()
}

func (w *bodyLimitResponseWriter) writeLimitError(r *http.Request, maxBytes, seenBytes int64, at time.Time) {
	if w.blocked {
		return
	}
	if !w.started {
		mw.WriteTransportError(w.ResponseWriter, r, http.StatusRequestEntityTooLarge, contract.CodeRequestBodyTooLarge, "request body exceeds configured limit", contract.CategoryClient, map[string]any{
			"max_bytes":  maxBytes,
			"seen_bytes": seenBytes,
			"at":         at.UTC(),
		})
	}
	w.blocked = true
}

type limitedBodyReader struct {
	w        *bodyLimitResponseWriter
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
		l.w.writeLimitError(l.req, l.maxBytes, l.used, l.now())
		if l.logger != nil {
			fields := internalobs.MiddlewareLogFields(l.req, http.StatusRequestEntityTooLarge, 0)
			fields["max_bytes"] = l.maxBytes
			fields["seen_bytes"] = l.used
			l.logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Warn("request body too large")
		}
	}

	return errRequestTooLarge
}
