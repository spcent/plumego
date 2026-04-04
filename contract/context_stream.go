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

// ErrUnsupportedStreamSource is returned when Stream receives a Source value whose
// type does not match any of the accepted streaming primitives.
var ErrUnsupportedStreamSource = errors.New("unsupported StreamSource type")

// StreamFormat selects the wire format for a streaming response.
type StreamFormat string

const (
	StreamFormatJSON StreamFormat = "json" // newline-delimited JSON (application/x-ndjson)
	StreamFormatText StreamFormat = "text" // plain-text lines (text/plain; charset=utf-8)
	StreamFormatSSE  StreamFormat = "sse"  // Server-Sent Events (text/event-stream)
)

// StreamConfig describes a streaming operation passed to Ctx.Stream.
//
// Source must be one of the following concrete types:
//   - []any            — JSON or Text items streamed from a slice
//   - []string         — Text lines streamed from a slice
//   - []SSEEvent       — SSE events streamed from a slice
//   - <-chan any        — JSON or Text items streamed from a channel
//   - <-chan string     — Text items streamed from a channel
//   - <-chan SSEEvent   — SSE events streamed from a channel
//   - func() (any, error)      — JSON or Text generator
//   - func() (string, error)   — Text generator
//   - func() (SSEEvent, error) — SSE generator
//   - io.Reader                — binary streaming (Format is ignored)
//
// ChunkSize > 0 enables chunked flushing (slice sources only).
// MaxRetry > 0 enables automatic retries on transient errors (generator sources only).
type StreamConfig struct {
	Format     StreamFormat
	Source     any
	ChunkSize  int
	MaxRetry   int
	RetryDelay time.Duration
}

// Stream dispatches a streaming response according to cfg.
// It replaces the family of Stream* methods and supports all format/source combinations.
// A negative ChunkSize returns ErrInvalidChunkSize; zero means no chunking (or default for io.Reader).
func (c *Ctx) Stream(cfg StreamConfig) error {
	if cfg.ChunkSize < 0 {
		return ErrInvalidChunkSize
	}
	switch src := cfg.Source.(type) {
	case io.Reader:
		chunkSize := cfg.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 32 * 1024
		}
		return c.streamBinaryReader(src, chunkSize)

	case []any:
		if cfg.ChunkSize > 0 {
			return c.streamJSONSliceChunked(src, cfg.ChunkSize)
		}
		ctx, err := c.initStream("application/x-ndjson")
		if err != nil {
			return err
		}
		return streamFromSlice(ctx, src, c.jsonWrite())

	case []string:
		if cfg.Format == StreamFormatSSE {
			// treat strings as SSE data lines
			events := make([]SSEEvent, len(src))
			for i, s := range src {
				events[i] = SSEEvent{Data: s}
			}
			if cfg.ChunkSize > 0 {
				return c.streamSSESliceChunked(events, cfg.ChunkSize)
			}
			return c.streamSSESlice(events)
		}
		if cfg.ChunkSize > 0 {
			return c.streamTextSliceChunked(src, cfg.ChunkSize)
		}
		ctx, err := c.initStream("text/plain; charset=utf-8")
		if err != nil {
			return err
		}
		return streamFromSlice(ctx, src, c.textWrite())

	case []SSEEvent:
		if cfg.ChunkSize > 0 {
			return c.streamSSESliceChunked(src, cfg.ChunkSize)
		}
		return c.streamSSESlice(src)

	case <-chan any:
		ctx, err := c.initStream("application/x-ndjson")
		if err != nil {
			return err
		}
		return streamFromChan(ctx, src, c.jsonWrite())

	case <-chan string:
		ctx, err := c.initStream("text/plain; charset=utf-8")
		if err != nil {
			return err
		}
		return streamFromChan(ctx, src, c.textWrite())

	case <-chan SSEEvent:
		ctx := c.streamContext()
		if err := ctx.Err(); err != nil {
			return err
		}
		sw, err := c.RespondWithSSE()
		if err != nil {
			return err
		}
		return streamFromChan(ctx, src, sw.Write)

	case func() (any, error):
		ctx, err := c.initStream("application/x-ndjson")
		if err != nil {
			return err
		}
		if cfg.MaxRetry > 0 {
			return streamFromGenWithRetry(ctx, src, cfg.MaxRetry, cfg.RetryDelay, c.jsonWrite())
		}
		return streamFromGen(ctx, src, c.jsonWrite())

	case func() (string, error):
		ctx, err := c.initStream("text/plain; charset=utf-8")
		if err != nil {
			return err
		}
		if cfg.MaxRetry > 0 {
			return streamFromGenWithRetry(ctx, src, cfg.MaxRetry, cfg.RetryDelay, c.textWrite())
		}
		return streamFromGen(ctx, src, c.textWrite())

	case func() (SSEEvent, error):
		ctx := c.streamContext()
		if err := ctx.Err(); err != nil {
			return err
		}
		sw, err := c.RespondWithSSE()
		if err != nil {
			return err
		}
		if cfg.MaxRetry > 0 {
			return streamFromGenWithRetry(ctx, src, cfg.MaxRetry, cfg.RetryDelay, sw.Write)
		}
		return streamFromGen(ctx, src, sw.Write)

	default:
		return ErrUnsupportedStreamSource
	}
}

// --- shared slice/chunked helpers called by Stream and deprecated wrappers ---

func (c *Ctx) streamBinaryReader(reader io.Reader, chunkSize int) error {
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

func (c *Ctx) streamJSONSliceChunked(items []any, chunkSize int) error {
	ctx, err := c.initStream("application/x-ndjson")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(c.W)
	return streamFromSliceChunked(ctx, items, chunkSize, enc.Encode, c.flush)
}

func (c *Ctx) streamTextSliceChunked(lines []string, chunkSize int) error {
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

func (c *Ctx) streamSSESlice(events []SSEEvent) error {
	ctx := c.streamContext()
	if err := ctx.Err(); err != nil {
		return err
	}
	sw, err := c.RespondWithSSE()
	if err != nil {
		return err
	}
	return streamFromSlice(ctx, events, sw.Write)
}

func (c *Ctx) streamSSESliceChunked(events []SSEEvent, chunkSize int) error {
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
