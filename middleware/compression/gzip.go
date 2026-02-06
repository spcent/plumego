package compression

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/utils"
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
//	// Use default configuration
//	handler := compression.Gzip()(myHandler)
//
//	// Or with custom configuration
//	config := compression.GzipConfig{
//		MaxBufferBytes: 5 << 20, // 5MB max buffer
//	}
//	handler := compression.GzipWithConfig(config)(myHandler)
//
// The middleware adds the following headers:
//   - Content-Encoding: gzip (when compression is used)
//   - Vary: Accept-Encoding (to indicate response varies by encoding)
//
// Note: Compression is skipped for error responses (status >= 400) to avoid
// compressing small error messages that don't benefit from compression.
func Gzip() middleware.Middleware {
	return GzipWithConfig(GzipConfig{
		MaxBufferBytes: 10 << 20, // 10MB max buffer
	})
}

// GzipConfig configures Gzip middleware behavior.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/compression"
//
//	config := compression.GzipConfig{
//		MaxBufferBytes: 5 << 20, // 5MB max buffer
//	}
//	handler := compression.GzipWithConfig(config)(myHandler)
type GzipConfig struct {
	// MaxBufferBytes is the maximum response size to buffer for compression.
	// Responses larger than this will bypass compression to avoid memory spikes.
	// Default: 10MB (10 << 20)
	MaxBufferBytes int
}

// GzipWithConfig creates a Gzip middleware with custom configuration.
func GzipWithConfig(cfg GzipConfig) middleware.Middleware {
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
			if gw.compressionUsed && gw.gz != nil {
				gw.gz.Close()
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
	bodyBuffer      []byte
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
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
		utils.EnsureNoSniff(w.ResponseWriter.Header())
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	// Start buffering
	w.bodyBuffer = make([]byte, 0, 1024) // Start with small buffer
	w.compressionUsed = true
	utils.EnsureNoSniff(w.ResponseWriter.Header())
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// If not compressing, write directly
	if !w.compressionUsed {
		return utils.SafeWrite(w.ResponseWriter, p)
	}

	// Buffer the data
	w.bodyBuffer = append(w.bodyBuffer, p...)

	// If buffer exceeds max, switch to bypass mode
	if len(w.bodyBuffer) > w.cfg.MaxBufferBytes {
		// Write buffered data as-is (no compression)
		w.compressionUsed = false

		// Remove compression headers
		w.Header().Del("Content-Encoding")

		// Write buffered data in chunks to avoid memory spikes
		const chunkSize = 8192 // 8KB chunks
		totalWritten := 0
		for totalWritten < len(w.bodyBuffer) {
			end := totalWritten + chunkSize
			if end > len(w.bodyBuffer) {
				end = len(w.bodyBuffer)
			}

			// SECURITY NOTE: This Write method only compresses/buffers response data.
			// The 'w.bodyBuffer' contains response data from upstream handlers,
			// not user input. This middleware does not modify response content
			// and therefore does not introduce XSS vulnerabilities.
			// XSS protection should be implemented in handlers that generate HTML content.
			n, err := utils.SafeWrite(w.ResponseWriter, w.bodyBuffer[totalWritten:end])
			if err != nil {
				// Ensure buffer is cleared even on error
				w.bodyBuffer = nil
				return totalWritten, err
			}
			totalWritten += n
		}

		// Clear buffer
		w.bodyBuffer = nil

		return totalWritten, nil
	}

	// If we have enough data, start compressing
	if w.gz == nil && len(w.bodyBuffer) > 0 {
		// Update headers for compression
		w.Header().Del("Content-Length")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")

		w.gz = gzip.NewWriter(w.ResponseWriter)

		// Write buffered data
		if _, err := w.gz.Write(w.bodyBuffer); err != nil {
			// Ensure buffer is cleared even on error
			w.bodyBuffer = nil
			return 0, err
		}

		// Clear buffer
		w.bodyBuffer = nil
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

func (w *gzipResponseWriter) Flush() {
	if w.gz != nil {
		w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
