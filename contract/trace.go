package contract

import (
	"context"
	"encoding/hex"
	"fmt"
)

// TraceID represents a unique identifier for tracing context.
type TraceID string

// SpanID represents a unique identifier for a span.
type SpanID string

// TraceFlags are flags that control tracing behavior.
type TraceFlags uint8

const (
	// TraceFlagsSampled indicates that the trace should be sampled.
	TraceFlagsSampled TraceFlags = 0x01

	// TraceIDLength is the expected length of a trace ID in hexadecimal format (32 hex chars = 16 bytes).
	TraceIDLength = 32

	// SpanIDLength is the expected length of a span ID in hexadecimal format (16 hex chars = 8 bytes).
	SpanIDLength = 16
)

// TraceContext is the minimal tracing context propagated via HTTP and stored in context.Context.
// Full tracing infrastructure (Tracer, Span, Collector, Sampler) lives in x/observability/tracer.
type TraceContext struct {
	TraceID      TraceID `json:"trace_id"`
	SpanID       SpanID  `json:"span_id"`
	ParentSpanID *SpanID `json:"parent_span_id,omitempty"`
	// Baggage carries W3C baggage key-value pairs.
	// Extraction from HTTP headers and injection into outgoing requests is not
	// implemented in this package; use x/observability for full propagation support.
	Baggage map[string]string `json:"baggage,omitempty"`
	Flags   TraceFlags        `json:"flags"`
}

// IsSampled reports whether the W3C sampled flag is set in Flags.
func (tc TraceContext) IsSampled() bool {
	return tc.Flags&TraceFlagsSampled != 0
}

type traceContextKey struct{}

// WithTraceContext adds trace context to a context.
func WithTraceContext(ctx context.Context, traceContext TraceContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceContextKey{}, &traceContext)
}

// TraceContextFromContext retrieves trace context from a context.
func TraceContextFromContext(ctx context.Context) *TraceContext {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(traceContextKey{}); v != nil {
		if tc, ok := v.(*TraceContext); ok {
			return tc
		}
	}
	return nil
}

// WithSpanIDString stores the given span ID string in the context.
// If a TraceContext already exists, only its SpanID field is updated so that
// trace and baggage information are preserved.
// Invalid values are ignored so the transport carrier never stores malformed
// span ids.
func WithSpanIDString(ctx context.Context, id string) context.Context {
	parsed, err := ParseSpanID(id)
	if err != nil {
		return ctx
	}
	if existing := TraceContextFromContext(ctx); existing != nil {
		updated := *existing
		updated.SpanID = parsed
		return WithTraceContext(ctx, updated)
	}
	return WithTraceContext(ctx, TraceContext{SpanID: parsed})
}

// ParseTraceID parses and validates a trace ID string.
func ParseTraceID(id string) (TraceID, error) {
	if len(id) != TraceIDLength {
		return "", fmt.Errorf("invalid trace ID length: expected %d, got %d", TraceIDLength, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid trace ID format: %v", err)
	}
	return TraceID(id), nil
}

// ParseSpanID parses and validates a span ID string.
func ParseSpanID(id string) (SpanID, error) {
	if len(id) != SpanIDLength {
		return "", fmt.Errorf("invalid span ID length: expected %d, got %d", SpanIDLength, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid span ID format: %v", err)
	}
	return SpanID(id), nil
}

// IsValidTraceID reports whether traceID is a valid hex-encoded 16-byte trace ID.
func IsValidTraceID(traceID TraceID) bool {
	_, err := ParseTraceID(string(traceID))
	return err == nil
}

// IsValidSpanID reports whether spanID is a valid hex-encoded 8-byte span ID.
func IsValidSpanID(spanID SpanID) bool {
	_, err := ParseSpanID(string(spanID))
	return err == nil
}
