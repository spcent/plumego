package compression

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/spcent/plumego/middleware"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

// Gzip compresses HTTP responses when the client supports it via Accept-Encoding.
// It intelligently skips compression for:
//   - WebSocket upgrades
//   - Server-Sent Events (SSE)
//   - Already compressed content
//   - Binary content (images, videos, etc.)
//   - Large streaming responses (to avoid memory spikes)
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/compression"
//
//	// Use explicit configuration
//	config := compression.GzipConfig{
//		MaxBufferBytes: 5 << 20, // 5MB max buffer
//	}
//	handler := compression.Gzip(config)(myHandler)
//
// The middleware adds the following headers:
//   - Content-Encoding: gzip (when compression is used)
//   - Vary: Accept-Encoding (to indicate response varies by encoding)
//
// Note: Compression is skipped for error responses (status >= 400) to avoid
// compressing small error messages that don't benefit from compression.
// GzipConfig configures Gzip middleware behavior.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/compression"
//
//	config := compression.GzipConfig{
//		MaxBufferBytes: 5 << 20, // 5MB max buffer
//	}
//	handler := compression.Gzip(config)(myHandler)
type GzipConfig struct {
	// MaxBufferBytes is the maximum response size to buffer for compression.
	// Responses larger than this will bypass compression to avoid memory spikes.
	// Default: 10MB (10 << 20)
	MaxBufferBytes int
}

// Gzip creates a Gzip middleware with explicit configuration.
func Gzip(cfg GzipConfig) middleware.Middleware {
	if cfg.MaxBufferBytes <= 0 {
		cfg.MaxBufferBytes = 10 << 20
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if client doesn't support gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Skip WebSocket upgrades
			connection := strings.ToLower(r.Header.Get("Connection"))
			if strings.Contains(connection, "upgrade") || r.Header.Get("Upgrade") != "" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip SSE (Server-Sent Events)
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "text/event-stream") {
				next.ServeHTTP(w, r)
				return
			}

			// Create wrapper
			gw := &gzipResponseWriter{
				ResponseWriter: w,
				cfg:            cfg,
			}

			next.ServeHTTP(gw, r)

			// Finalize compression if used
			if gw.compressionUsed {
				if gw.gz != nil {
					gw.gz.Close()
				} else {
					gw.flushHeaders()
					if gw.buffer != nil && gw.buffer.Len() > 0 {
						_, _ = internaltransport.SafeWrite(w, gw.buffer.Body())
					}
				}
			}
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	cfg             GzipConfig
	gz              *gzip.Writer
	compressionUsed bool
	wroteHeader     bool
	buffer          *internaltransport.BufferedResponse
	headersFlushed  bool
}

func (w *gzipResponseWriter) Header() http.Header {
	w.ensureBuffer()
	return w.buffer.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Determine if we should compress
	shouldCompress := w.shouldCompress(statusCode)

	if !shouldCompress {
		w.compressionUsed = false
		w.ensureBuffer()
		w.buffer.WriteHeader(statusCode)
		w.flushHeaders()
		return
	}

	// Start buffering
	w.ensureBuffer()
	w.buffer.WriteHeader(statusCode)
	w.compressionUsed = true
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// If not compressing, write directly
	if !w.compressionUsed {
		w.flushHeaders()
		return internaltransport.SafeWrite(w.ResponseWriter, p)
	}

	// Buffer the data
	w.ensureBuffer()

	// If buffer exceeds max, switch to bypass mode
	if w.cfg.MaxBufferBytes > 0 && w.buffer.Len()+len(p) > w.cfg.MaxBufferBytes {
		// Write buffered data as-is (no compression)
		w.compressionUsed = false

		// Remove compression headers
		w.buffer.Header().Del("Content-Encoding")
		w.flushHeaders()

		// Write buffered data in chunks to avoid memory spikes
		const chunkSize = 8192 // 8KB chunks
		totalWritten := 0
		body := w.buffer.Body()
		for totalWritten < len(body) {
			end := totalWritten + chunkSize
			if end > len(body) {
				end = len(body)
			}

			// SECURITY NOTE: This Write method only compresses/buffers response data.
			// The 'w.bodyBuffer' contains response data from upstream handlers,
			// not user input. This middleware does not modify response content
			// and therefore does not introduce XSS vulnerabilities.
			// XSS protection should be implemented in handlers that generate HTML content.
			n, err := internaltransport.SafeWrite(w.ResponseWriter, body[totalWritten:end])
			if err != nil {
				w.buffer.ClearBody()
				return totalWritten, err
			}
			totalWritten += n
		}

		w.buffer.ClearBody()
		n, err := internaltransport.SafeWrite(w.ResponseWriter, p)
		return totalWritten + n, err
	}

	_, err := w.buffer.Write(p)
	if err != nil {
		return 0, err
	}

	// If we have enough data, start compressing
	if w.gz == nil && w.buffer.Len() > 0 {
		// Update headers for compression
		w.buffer.Header().Del("Content-Length")
		w.buffer.Header().Add("Vary", "Accept-Encoding")
		w.buffer.Header().Set("Content-Encoding", "gzip")
		w.flushHeaders()
		w.gz = gzip.NewWriter(w.ResponseWriter)

		// Write buffered data
		if _, err := w.gz.Write(w.buffer.Body()); err != nil {
			w.buffer.ClearBody()
			return 0, err
		}

		// Clear buffer
		w.buffer.ClearBody()
		return len(p), nil
	}

	// Already compressing
	if w.gz != nil {
		return w.gz.Write(p)
	}

	return len(p), nil
}

func (w *gzipResponseWriter) shouldCompress(statusCode int) bool {
	// Don't compress error responses
	if statusCode >= 400 {
		return false
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		return false
	}

	// Skip already compressed content
	if w.Header().Get("Content-Encoding") != "" {
		return false
	}

	// Skip streaming content types
	lowerContentType := strings.ToLower(contentType)
	if strings.Contains(lowerContentType, "stream") ||
		strings.Contains(lowerContentType, "event-stream") ||
		strings.Contains(lowerContentType, "websocket") {
		return false
	}

	// Skip binary content
	compressedTypes := []string{
		"image/", "video/", "audio/",
		"application/zip", "application/gzip",
		"application/pdf", "application/octet-stream",
	}
	for _, ct := range compressedTypes {
		if strings.HasPrefix(lowerContentType, ct) {
			return false
		}
	}

	return true
}

func (w *gzipResponseWriter) ensureBuffer() {
	if w.buffer == nil {
		w.buffer = internaltransport.NewBufferedResponse(w.cfg.MaxBufferBytes)
	}
}

func (w *gzipResponseWriter) flushHeaders() {
	if w.buffer == nil || w.ResponseWriter == nil {
		return
	}
	if w.headersFlushed {
		return
	}
	for k, values := range w.buffer.Header() {
		for _, v := range values {
			w.ResponseWriter.Header().Add(k, v)
		}
	}
	internaltransport.EnsureNoSniff(w.ResponseWriter.Header())
	w.ResponseWriter.WriteHeader(w.buffer.StatusCode())
	w.headersFlushed = true
}

func (w *gzipResponseWriter) Flush() {
	if !w.compressionUsed {
		w.flushHeaders()
	}
	if w.gz != nil {
		w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
