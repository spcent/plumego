package contract

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
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

// TraceContext is the minimal trace/span metadata carrier stored in
// context.Context. Full tracing infrastructure (Tracer, Span, Collector,
// Sampler) lives in x/observability/tracer.
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

// HasTraceID reports whether TraceID is present and valid.
func (tc TraceContext) HasTraceID() bool {
	return isValidTraceID(tc.TraceID)
}

// HasSpanID reports whether SpanID is present and valid.
func (tc TraceContext) HasSpanID() bool {
	return isValidSpanID(tc.SpanID)
}

// Valid reports whether TraceContext carries both required W3C identifiers.
// Callers that read a TraceContext from context must check Valid before
// treating the carrier as a usable propagation context.
func (tc TraceContext) Valid() bool {
	return tc.HasTraceID() && tc.HasSpanID()
}

type traceContextKey struct{}

// WithTraceContext adds trace context to a context.
// It stores a defensive copy and does not validate, extract, or inject trace
// propagation headers; owning observability packages handle propagation policy.
func WithTraceContext(ctx context.Context, traceContext TraceContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	traceContext = cloneTraceContext(traceContext)
	return context.WithValue(ctx, traceContextKey{}, &traceContext)
}

// TraceContextFromContext retrieves trace context from a context. The returned
// carrier may be invalid or span-only; callers must use Valid when they require
// a complete propagation context.
func TraceContextFromContext(ctx context.Context) *TraceContext {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(traceContextKey{}); v != nil {
		if tc, ok := v.(*TraceContext); ok {
			cloned := cloneTraceContext(*tc)
			return &cloned
		}
	}
	return nil
}

func cloneTraceContext(traceContext TraceContext) TraceContext {
	if traceContext.ParentSpanID != nil {
		parentSpanID := *traceContext.ParentSpanID
		traceContext.ParentSpanID = &parentSpanID
	}
	if traceContext.Baggage != nil {
		baggage := make(map[string]string, len(traceContext.Baggage))
		for k, v := range traceContext.Baggage {
			baggage[k] = v
		}
		traceContext.Baggage = baggage
	}
	return traceContext
}

func parseTraceID(id string) (TraceID, error) {
	id = strings.ToLower(id)
	if len(id) != TraceIDLength {
		return "", fmt.Errorf("invalid trace ID length: expected %d, got %d", TraceIDLength, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid trace ID format: %w", err)
	}
	if isAllZeroHex(id) {
		return "", fmt.Errorf("invalid trace ID format: all zero")
	}
	return TraceID(id), nil
}

func parseSpanID(id string) (SpanID, error) {
	id = strings.ToLower(id)
	if len(id) != SpanIDLength {
		return "", fmt.Errorf("invalid span ID length: expected %d, got %d", SpanIDLength, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid span ID format: %w", err)
	}
	if isAllZeroHex(id) {
		return "", fmt.Errorf("invalid span ID format: all zero")
	}
	return SpanID(id), nil
}

func isAllZeroHex(id string) bool {
	for _, ch := range id {
		if ch != '0' {
			return false
		}
	}
	return true
}

func isValidTraceID(traceID TraceID) bool {
	_, err := parseTraceID(string(traceID))
	return err == nil
}

func isValidSpanID(spanID SpanID) bool {
	_, err := parseSpanID(string(spanID))
	return err == nil
}
