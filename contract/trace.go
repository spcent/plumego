package contract

import (
	"context"
	"encoding/hex"
	"fmt"
)

// TraceID represents a unique identifier for a trace.
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
	TraceID      TraceID           `json:"trace_id"`
	SpanID       SpanID            `json:"span_id"`
	ParentSpanID *SpanID           `json:"parent_span_id,omitempty"`
	Baggage      map[string]string `json:"baggage,omitempty"`
	Flags        TraceFlags        `json:"flags"`
	Sampled      bool              `json:"sampled"`
}

type traceContextKey struct{}

// WithTraceContext adds trace context to a context.
func WithTraceContext(ctx context.Context, traceContext TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKey{}, &traceContext)
}

// ContextWithTraceContext adds trace context to a context.
// Deprecated: Use WithTraceContext instead.
func ContextWithTraceContext(ctx context.Context, traceContext TraceContext) context.Context {
	return WithTraceContext(ctx, traceContext)
}

// TraceContextFromContext retrieves trace context from a context.
func TraceContextFromContext(ctx context.Context) *TraceContext {
	if v := ctx.Value(traceContextKey{}); v != nil {
		if tc, ok := v.(*TraceContext); ok {
			return tc
		}
	}
	return nil
}

// TraceIDFromContext extracts the trace id string from the context.
func TraceIDFromContext(ctx context.Context) string {
	if tc := TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID)
	}
	return ""
}

// WithTraceIDString stores the given trace ID string in the context.
// If a TraceContext already exists, only its TraceID field is updated so that
// span and baggage information are preserved.
func WithTraceIDString(ctx context.Context, id string) context.Context {
	if existing := TraceContextFromContext(ctx); existing != nil {
		updated := *existing
		updated.TraceID = TraceID(id)
		return WithTraceContext(ctx, updated)
	}
	return WithTraceContext(ctx, TraceContext{TraceID: TraceID(id)})
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
