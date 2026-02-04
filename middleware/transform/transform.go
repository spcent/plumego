// Package transform provides request/response transformation middleware
//
// This package allows modification of HTTP requests and responses without
// protocol-specific knowledge. Transformations can include:
//   - Header manipulation (add, remove, rename)
//   - JSON field renaming and restructuring
//   - Body content modification
//   - Query parameter manipulation
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/middleware/transform"
//		"github.com/spcent/plumego/core"
//	)
//
//	app := core.New()
//
//	// Simple transformations
//	app.Use(transform.Middleware(transform.Config{
//		RequestTransformers: []transform.RequestTransformer{
//			transform.AddRequestHeader("X-Custom-Header", "value"),
//			transform.RemoveRequestHeader("X-Internal-Header"),
//		},
//		ResponseTransformers: []transform.ResponseTransformer{
//			transform.AddResponseHeader("X-API-Version", "2"),
//		},
//	}))
package transform

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// RequestTransformer modifies an HTTP request
type RequestTransformer func(*http.Request) error

// ResponseTransformer modifies an HTTP response
type ResponseTransformer func(*http.Response) error

// Config holds transformation middleware configuration
type Config struct {
	// RequestTransformers are applied to incoming requests
	RequestTransformers []RequestTransformer

	// ResponseTransformers are applied to outgoing responses
	ResponseTransformers []ResponseTransformer

	// OnError is called when transformation fails (optional)
	OnError func(err error)
}

// Middleware creates a transformation middleware
func Middleware(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply request transformers
			for _, transformer := range config.RequestTransformers {
				if err := transformer(r); err != nil {
					if config.OnError != nil {
						config.OnError(err)
					}
					http.Error(w, "Request transformation failed", http.StatusBadRequest)
					return
				}
			}

			// If no response transformers, just call next
			if len(config.ResponseTransformers) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Create response recorder for transformation
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				header:         make(http.Header),
				body:           &bytes.Buffer{},
			}

			// Call next handler
			next.ServeHTTP(recorder, r)

			// Create http.Response for transformation
			resp := &http.Response{
				StatusCode: recorder.statusCode,
				Header:     recorder.header,
				Body:       io.NopCloser(bytes.NewReader(recorder.body.Bytes())),
			}

			// Apply response transformers
			for _, transformer := range config.ResponseTransformers {
				if err := transformer(resp); err != nil {
					if config.OnError != nil {
						config.OnError(err)
					}
					http.Error(w, "Response transformation failed", http.StatusInternalServerError)
					return
				}
			}

			// Read transformed body
			var transformedBody []byte
			if resp.Body != nil {
				transformedBody, _ = io.ReadAll(resp.Body)
				resp.Body.Close()
			}

			// Write transformed response
			for key, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(resp.StatusCode)
			if len(transformedBody) > 0 {
				w.Write(transformedBody)
			}
		})
	}
}

// responseRecorder captures HTTP response for transformation
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
	body       *bytes.Buffer
	written    bool
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		r.written = true
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(b)
}

// ========================================
// Request Transformers
// ========================================

// AddRequestHeader adds a header to the request
func AddRequestHeader(key, value string) RequestTransformer {
	return func(r *http.Request) error {
		r.Header.Set(key, value)
		return nil
	}
}

// RemoveRequestHeader removes a header from the request
func RemoveRequestHeader(key string) RequestTransformer {
	return func(r *http.Request) error {
		r.Header.Del(key)
		return nil
	}
}

// RenameRequestHeader renames a request header
func RenameRequestHeader(from, to string) RequestTransformer {
	return func(r *http.Request) error {
		if value := r.Header.Get(from); value != "" {
			r.Header.Set(to, value)
			r.Header.Del(from)
		}
		return nil
	}
}

// SetRequestMethod changes the HTTP method
func SetRequestMethod(method string) RequestTransformer {
	return func(r *http.Request) error {
		r.Method = method
		return nil
	}
}

// AddQueryParam adds a query parameter to the request
func AddQueryParam(key, value string) RequestTransformer {
	return func(r *http.Request) error {
		q := r.URL.Query()
		q.Add(key, value)
		r.URL.RawQuery = q.Encode()
		return nil
	}
}

// RemoveQueryParam removes a query parameter from the request
func RemoveQueryParam(key string) RequestTransformer {
	return func(r *http.Request) error {
		q := r.URL.Query()
		q.Del(key)
		r.URL.RawQuery = q.Encode()
		return nil
	}
}

// RenameJSONRequestField renames a JSON field in request body
func RenameJSONRequestField(from, to string) RequestTransformer {
	return func(r *http.Request) error {
		// Only process JSON content
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		// Parse JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// Not JSON object, restore body
			r.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		// Rename field
		if val, exists := data[from]; exists {
			data[to] = val
			delete(data, from)
		}

		// Marshal back to JSON
		newBody, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// Update request
		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		return nil
	}
}

// ModifyJSONRequest applies a custom function to modify JSON request body
func ModifyJSONRequest(modifier func(map[string]interface{}) error) RequestTransformer {
	return func(r *http.Request) error {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		// Apply modifier
		if err := modifier(data); err != nil {
			return err
		}

		newBody, err := json.Marshal(data)
		if err != nil {
			return err
		}

		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		return nil
	}
}

// ========================================
// Response Transformers
// ========================================

// AddResponseHeader adds a header to the response
func AddResponseHeader(key, value string) ResponseTransformer {
	return func(r *http.Response) error {
		r.Header.Set(key, value)
		return nil
	}
}

// RemoveResponseHeader removes a header from the response
func RemoveResponseHeader(key string) ResponseTransformer {
	return func(r *http.Response) error {
		r.Header.Del(key)
		return nil
	}
}

// RenameResponseHeader renames a response header
func RenameResponseHeader(from, to string) ResponseTransformer {
	return func(r *http.Response) error {
		if value := r.Header.Get(from); value != "" {
			r.Header.Set(to, value)
			r.Header.Del(from)
		}
		return nil
	}
}

// RenameJSONResponseField renames a JSON field in response body
func RenameJSONResponseField(from, to string) ResponseTransformer {
	return func(r *http.Response) error {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		// Rename field
		if val, exists := data[from]; exists {
			data[to] = val
			delete(data, from)
		}

		newBody, err := json.Marshal(data)
		if err != nil {
			return err
		}

		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.Header.Set("Content-Length", string(rune(len(newBody))))

		return nil
	}
}

// ModifyJSONResponse applies a custom function to modify JSON response body
func ModifyJSONResponse(modifier func(map[string]interface{}) error) ResponseTransformer {
	return func(r *http.Response) error {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		if err := modifier(data); err != nil {
			return err
		}

		newBody, err := json.Marshal(data)
		if err != nil {
			return err
		}

		r.Body = io.NopCloser(bytes.NewReader(newBody))

		return nil
	}
}

// WrapJSONResponse wraps response JSON in a standard envelope
func WrapJSONResponse(wrapperKey string) ResponseTransformer {
	return func(r *http.Response) error {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		// Wrap data
		wrapped := map[string]interface{}{
			wrapperKey: data,
		}

		newBody, err := json.Marshal(wrapped)
		if err != nil {
			return err
		}

		r.Body = io.NopCloser(bytes.NewReader(newBody))

		return nil
	}
}

// SetResponseStatus changes the response status code
func SetResponseStatus(statusCode int) ResponseTransformer {
	return func(r *http.Response) error {
		r.StatusCode = statusCode
		return nil
	}
}

// Chain combines multiple transformers into one
func ChainRequest(transformers ...RequestTransformer) RequestTransformer {
	return func(r *http.Request) error {
		for _, transformer := range transformers {
			if err := transformer(r); err != nil {
				return err
			}
		}
		return nil
	}
}

// ChainResponse combines multiple response transformers into one
func ChainResponse(transformers ...ResponseTransformer) ResponseTransformer {
	return func(r *http.Response) error {
		for _, transformer := range transformers {
			if err := transformer(r); err != nil {
				return err
			}
		}
		return nil
	}
}
