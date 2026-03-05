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

func validateChunkSize(chunkSize int) error {
	if chunkSize <= 0 {
		return ErrInvalidChunkSize
	}
	return nil
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

// flush attempts to flush the response writer if it supports http.Flusher.
func (c *Ctx) flush() {
	if f, ok := c.W.(http.Flusher); ok {
		f.Flush()
	}
}

// --- Generic internal streaming helpers ---

// initStream sets content-type headers, writes 200 OK, and returns the request context.
// It is the common setup step for all non-SSE streaming methods.
func (c *Ctx) initStream(contentType string) (context.Context, error) {
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.W.Header().Set("Content-Type", contentType)
	c.W.WriteHeader(http.StatusOK)
	return ctx, nil
}

// streamFromSlice iterates a slice, calling write for each element.
// Context cancellation is checked before every write.
func streamFromSlice[T any](ctx context.Context, items []T, write func(T) error) error {
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := write(item); err != nil {
			return err
		}
	}
	return nil
}

// streamFromChan reads from ch until it is closed or ctx is cancelled.
func streamFromChan[T any](ctx context.Context, ch <-chan T, write func(T) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-ch:
			if !ok {
				return nil
			}
			if err := write(item); err != nil {
				return err
			}
		}
	}
}

// streamFromGen calls gen in a loop until it returns io.EOF or ctx is cancelled.
func streamFromGen[T any](ctx context.Context, gen func() (T, error), write func(T) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		item, err := gen()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := write(item); err != nil {
			return err
		}
	}
}

// streamFromSliceChunked writes items and calls flush every chunkSize elements and at the end.
func streamFromSliceChunked[T any](ctx context.Context, items []T, chunkSize int, write func(T) error, flush func()) error {
	for i, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := write(item); err != nil {
			return err
		}
		if (i+1)%chunkSize == 0 {
			flush()
		}
	}
	flush()
	return nil
}

// streamFromGenWithRetry calls gen in a loop, retrying up to maxRetries times on transient errors.
func streamFromGenWithRetry[T any](ctx context.Context, gen func() (T, error), maxRetries int, delay time.Duration, write func(T) error) error {
	retries := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		item, err := gen()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if retries < maxRetries {
				if err := sleepWithContext(ctx, delay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}
		retries = 0
		if err := write(item); err != nil {
			return err
		}
	}
}

// jsonWrite returns a write function that encodes items as JSON and flushes immediately.
func (c *Ctx) jsonWrite() func(any) error {
	enc := json.NewEncoder(c.W)
	return func(v any) error {
		if err := enc.Encode(v); err != nil {
			return err
		}
		c.flush()
		return nil
	}
}

// textWrite returns a write function that writes a text line and flushes immediately.
func (c *Ctx) textWrite() func(string) error {
	return func(line string) error {
		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		c.flush()
		return nil
	}
}

// --- SSE API ---

// RespondWithSSE starts a Server-Sent Events response.
// Returns an SSEWriter for sending events, or an error if SSE is not supported.
func (c *Ctx) RespondWithSSE() (*SSEWriter, error) {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
	c.W.WriteHeader(http.StatusOK)
	return NewSSEWriter(c.W)
}

// IsSSESupported reports whether the response writer supports SSE flushing.
func (c *Ctx) IsSSESupported() bool {
	_, ok := c.W.(http.Flusher)
	return ok
}

// SetSSEHeaders sets the standard SSE response headers without writing the status line.
func (c *Ctx) SetSSEHeaders() {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
}

// WriteSSE writes a single SSE event and flushes.
func (c *Ctx) WriteSSE(event SSEEvent) error {
	sw, err := NewSSEWriter(c.W)
	if err != nil {
		return err
	}
	return sw.Write(event)
}

// --- Slice-based streaming ---

// StreamJSON streams items as newline-delimited JSON, flushing after each item.
func (c *Ctx) StreamJSON(items []any) error {
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	return streamFromSlice(ctx, items, c.jsonWrite())
}

// StreamText streams lines of text, flushing after each line.
func (c *Ctx) StreamText(lines []string) error {
	ctx, err := c.initStream("text/plain; charset=utf-8")
	if err != nil {
		return err
	}
	return streamFromSlice(ctx, lines, c.textWrite())
}

// StreamBinary streams binary data in chunks of the given size.
func (c *Ctx) StreamBinary(reader io.Reader, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx, err := c.initStream("application/octet-stream")
	if err != nil {
		return err
	}
	buf := make([]byte, chunkSize)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, err := c.W.Write(buf[:n]); err != nil {
				return err
			}
			c.flush()
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

// --- Channel-based streaming ---

// StreamJSONWithChannel streams JSON items from a channel, flushing after each item.
func (c *Ctx) StreamJSONWithChannel(itemChan <-chan any) error {
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	return streamFromChan(ctx, itemChan, c.jsonWrite())
}

// StreamTextWithChannel streams text lines from a channel, flushing after each line.
func (c *Ctx) StreamTextWithChannel(lineChan <-chan string) error {
	ctx, err := c.initStream("text/plain; charset=utf-8")
	if err != nil {
		return err
	}
	return streamFromChan(ctx, lineChan, c.textWrite())
}

// StreamSSEWithChannel streams Server-Sent Events from a channel.
func (c *Ctx) StreamSSEWithChannel(eventChan <-chan SSEEvent) error {
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return err
	}
	sw, err := c.RespondWithSSE()
	if err != nil {
		return err
	}
	return streamFromChan(ctx, eventChan, sw.Write)
}

// --- Generator-based streaming ---

// StreamJSONWithGenerator streams JSON items from a generator function until it returns io.EOF.
func (c *Ctx) StreamJSONWithGenerator(generator func() (any, error)) error {
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	return streamFromGen(ctx, generator, c.jsonWrite())
}

// StreamTextWithGenerator streams text lines from a generator function until it returns io.EOF.
func (c *Ctx) StreamTextWithGenerator(generator func() (string, error)) error {
	ctx, err := c.initStream("text/plain; charset=utf-8")
	if err != nil {
		return err
	}
	return streamFromGen(ctx, generator, c.textWrite())
}

// StreamSSEWithGenerator streams Server-Sent Events from a generator function until it returns io.EOF.
func (c *Ctx) StreamSSEWithGenerator(generator func() (SSEEvent, error)) error {
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return err
	}
	sw, err := c.RespondWithSSE()
	if err != nil {
		return err
	}
	return streamFromGen(ctx, generator, sw.Write)
}

// --- Chunked streaming ---

// StreamJSONChunked streams JSON items, flushing every chunkSize items.
func (c *Ctx) StreamJSONChunked(items []any, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(c.W)
	return streamFromSliceChunked(ctx, items, chunkSize, enc.Encode, c.flush)
}

// StreamTextChunked streams text lines, flushing every chunkSize lines.
func (c *Ctx) StreamTextChunked(lines []string, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx, err := c.initStream("text/plain; charset=utf-8")
	if err != nil {
		return err
	}
	write := func(line string) error {
		_, err := fmt.Fprintf(c.W, "%s\n", line)
		return err
	}
	return streamFromSliceChunked(ctx, lines, chunkSize, write, c.flush)
}

// StreamSSEChunked streams Server-Sent Events, flushing every chunkSize events.
// Note: SSEWriter.Write already flushes after each event; the chunk flush is an additional guarantee.
func (c *Ctx) StreamSSEChunked(events []SSEEvent, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return err
	}
	sw, err := c.RespondWithSSE()
	if err != nil {
		return err
	}
	return streamFromSliceChunked(ctx, events, chunkSize, sw.Write, c.flush)
}

// --- Streaming with retry ---

// StreamJSONWithRetry streams JSON items from a generator, retrying up to maxRetries times on error.
func (c *Ctx) StreamJSONWithRetry(generator func() (any, error), maxRetries int, retryDelay time.Duration) error {
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	return streamFromGenWithRetry(ctx, generator, maxRetries, retryDelay, c.jsonWrite())
}

// StreamTextWithRetry streams text lines from a generator, retrying up to maxRetries times on error.
func (c *Ctx) StreamTextWithRetry(generator func() (string, error), maxRetries int, retryDelay time.Duration) error {
	ctx, err := c.initStream("text/plain; charset=utf-8")
	if err != nil {
		return err
	}
	return streamFromGenWithRetry(ctx, generator, maxRetries, retryDelay, c.textWrite())
}

// StreamSSEWithRetry streams Server-Sent Events from a generator, retrying up to maxRetries times on error.
func (c *Ctx) StreamSSEWithRetry(generator func() (SSEEvent, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return err
	}
	sw, err := c.RespondWithSSE()
	if err != nil {
		return err
	}
	return streamFromGenWithRetry(ctx, generator, maxRetries, retryDelay, sw.Write)
}
