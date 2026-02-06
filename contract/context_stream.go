package contract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrSSEInvalidField is returned when an SSE event field contains invalid characters.
var ErrSSEInvalidField = errors.New("SSE event field contains invalid characters (newlines not allowed in id/event)")

// sanitizeSSEField strips newline characters from SSE id and event fields
// to prevent protocol injection attacks.
func sanitizeSSEField(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// writeSSEData writes SSE data field, correctly handling multi-line data
// by prefixing each line with "data: " per the SSE specification.
func writeSSEData(w http.ResponseWriter, data string) error {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	// End the event with an extra newline
	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}
	return nil
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	ID    string
	Event string
	Data  string
	Retry time.Duration
}

// SSEWriter is a helper for writing Server-Sent Events.
type SSEWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

// ErrSSENotSupported is returned when the response writer doesn't support SSE.
var ErrSSENotSupported = errors.New("SSE not supported: response writer does not implement http.Flusher")

// NewSSEWriter creates a new SSE writer.
// Returns an error if the response writer doesn't support flushing.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrSSENotSupported
	}
	return &SSEWriter{w: w, f: f}, nil
}

// Write writes an SSE event to the response.
// ID and Event fields are sanitized to prevent protocol injection (newlines stripped).
// Data fields with newlines are correctly split into multiple "data:" lines per the SSE spec.
func (sw *SSEWriter) Write(event SSEEvent) error {
	if event.ID != "" {
		if _, err := fmt.Fprintf(sw.w, "id: %s\n", sanitizeSSEField(event.ID)); err != nil {
			return err
		}
	}
	if event.Event != "" {
		if _, err := fmt.Fprintf(sw.w, "event: %s\n", sanitizeSSEField(event.Event)); err != nil {
			return err
		}
	}
	if event.Retry > 0 {
		if _, err := fmt.Fprintf(sw.w, "retry: %d\n", int(event.Retry.Milliseconds())); err != nil {
			return err
		}
	}
	if event.Data != "" {
		if err := writeSSEData(sw.w, event.Data); err != nil {
			return err
		}
	} else {
		// Empty data still needs the terminating double newline
		if _, err := fmt.Fprint(sw.w, "\n"); err != nil {
			return err
		}
	}
	sw.f.Flush()
	return nil
}

func (c *Ctx) streamContext() context.Context {
	if c == nil || c.R == nil {
		return context.Background()
	}
	return c.R.Context()
}

func checkStreamContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	if ctx == nil {
		time.Sleep(delay)
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func validateChunkSize(chunkSize int) error {
	if chunkSize <= 0 {
		return ErrInvalidChunkSize
	}
	return nil
}

// flush attempts to flush the response writer if it supports http.Flusher.
func (c *Ctx) flush() {
	if f, ok := c.W.(http.Flusher); ok {
		f.Flush()
	}
}

// RespondWithSSE starts a Server-Sent Events response.
// Returns an SSEWriter for sending events, or an error if SSE is not supported.
func (c *Ctx) RespondWithSSE() (*SSEWriter, error) {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
	c.W.WriteHeader(http.StatusOK)

	return NewSSEWriter(c.W)
}

// RespondWithEventSource starts an event stream response.
// This is an alias for RespondWithSSE.
func (c *Ctx) RespondWithEventSource() (*SSEWriter, error) {
	return c.RespondWithSSE()
}

// StreamJSON streams JSON data line by line.
func (c *Ctx) StreamJSON(items []any) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for _, item := range items {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := encoder.Encode(item); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamText streams text data line by line.
func (c *Ctx) StreamText(lines []string) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for _, line := range lines {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamBinary streams binary data in chunks.
func (c *Ctx) StreamBinary(reader io.Reader, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.WriteHeader(http.StatusOK)

	buf := make([]byte, chunkSize)
	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if n > 0 {
			if err := checkStreamContext(ctx); err != nil {
				return err
			}
			if _, err := c.W.Write(buf[:n]); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	return nil
}

// StreamJSONWithChannel streams JSON data from a channel.
func (c *Ctx) StreamJSONWithChannel(itemChan <-chan any) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-itemChan:
			if !ok {
				return nil
			}
			if err := encoder.Encode(item); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// StreamTextWithChannel streams text data from a channel.
func (c *Ctx) StreamTextWithChannel(lineChan <-chan string) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, ok := <-lineChan:
			if !ok {
				return nil
			}
			if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// StreamSSEWithChannel streams Server-Sent Events from a channel.
func (c *Ctx) StreamSSEWithChannel(eventChan <-chan SSEEvent) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter, err := c.RespondWithSSE()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}
			if err := sseWriter.Write(event); err != nil {
				return err
			}
		}
	}
}

// StreamJSONWithGenerator streams JSON data using a generator function.
func (c *Ctx) StreamJSONWithGenerator(generator func() (any, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		item, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := encoder.Encode(item); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamTextWithGenerator streams text data using a generator function.
func (c *Ctx) StreamTextWithGenerator(generator func() (string, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		line, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamSSEWithGenerator streams Server-Sent Events using a generator function.
func (c *Ctx) StreamSSEWithGenerator(generator func() (SSEEvent, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter, err := c.RespondWithSSE()
	if err != nil {
		return err
	}

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		event, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := sseWriter.Write(event); err != nil {
			return err
		}
	}

	return nil
}

// IsSSESupported checks if the response writer supports SSE.
func (c *Ctx) IsSSESupported() bool {
	_, ok := c.W.(http.Flusher)
	return ok
}

// SetSSEHeaders sets the standard SSE headers.
func (c *Ctx) SetSSEHeaders() {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
}

// SetEventSourceHeaders sets the standard event source headers.
func (c *Ctx) SetEventSourceHeaders() {
	c.SetSSEHeaders()
}

// WriteSSE writes a single SSE event and flushes.
func (c *Ctx) WriteSSE(event SSEEvent) error {
	sseWriter, err := NewSSEWriter(c.W)
	if err != nil {
		return err
	}
	return sseWriter.Write(event)
}

// WriteEventSource writes a single event source event and flushes.
func (c *Ctx) WriteEventSource(event SSEEvent) error {
	return c.WriteSSE(event)
}

// StreamJSONChunked streams JSON data in chunks with custom chunk size.
func (c *Ctx) StreamJSONChunked(items []any, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for i, item := range items {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := encoder.Encode(item); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			c.flush()
		}
	}

	// Final flush
	c.flush()

	return nil
}

// StreamTextChunked streams text data in chunks with custom chunk size.
func (c *Ctx) StreamTextChunked(lines []string, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for i, line := range lines {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			c.flush()
		}
	}

	// Final flush
	c.flush()

	return nil
}

// StreamSSEChunked streams Server-Sent Events in chunks with custom chunk size.
func (c *Ctx) StreamSSEChunked(events []SSEEvent, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter, err := c.RespondWithSSE()
	if err != nil {
		return err
	}

	for i, event := range events {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := sseWriter.Write(event); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			c.flush()
		}
	}

	// Final flush
	c.flush()

	return nil
}

// StreamJSONWithRetry streams JSON data with retry logic on error.
func (c *Ctx) StreamJSONWithRetry(generator func() (any, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		item, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if err := encoder.Encode(item); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamTextWithRetry streams text data with retry logic on error.
func (c *Ctx) StreamTextWithRetry(generator func() (string, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		line, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		c.flush()
	}

	return nil
}

// StreamSSEWithRetry streams Server-Sent Events with retry logic on error.
func (c *Ctx) StreamSSEWithRetry(generator func() (SSEEvent, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter, err := c.RespondWithSSE()
	if err != nil {
		return err
	}

	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		event, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if err := sseWriter.Write(event); err != nil {
			return err
		}
	}

	return nil
}
