package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// Gzip compresses HTTP responses when the client supports it via Accept-Encoding.
func Gzip() Middleware {
	return func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Accept-Encoding")
			w.Header().Set("Content-Encoding", "gzip")

			gz := gzip.NewWriter(w)
			defer gz.Close()

			gw := &gzipResponseWriter{
				ResponseWriter: w,
				writer:         gz,
			}

			next.ServeHTTP(gw, r)
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(p))
	}
	return w.writer.Write(p)
}
