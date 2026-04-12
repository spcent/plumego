package contract

import (
	"context"
	"testing"
)

func TestParseTraceID(t *testing.T) {
	validID := "1234567890abcdef1234567890abcdef"
	traceID, err := ParseTraceID(validID)
	if err != nil {
		t.Fatalf("expected no error for valid trace ID: %v", err)
	}
	if TraceID(validID) != traceID {
		t.Fatalf("expected parsed trace ID to match")
	}

	for _, invalidID := range []string{
		"123",
		"1234567890abcdef1234567890abcdef123",
	} {
		if _, err := ParseTraceID(invalidID); err == nil {
			t.Fatalf("expected error for invalid ID length: %s", invalidID)
		}
	}

	if _, err := ParseTraceID("invalid_trace_id_!!!"); err == nil {
		t.Fatalf("expected error for invalid hex format")
	}
}

func TestParseSpanID(t *testing.T) {
	validID := "1234567890abcdef"
	spanID, err := ParseSpanID(validID)
	if err != nil {
		t.Fatalf("expected no error for valid span ID: %v", err)
	}
	if SpanID(validID) != spanID {
		t.Fatalf("expected parsed span ID to match")
	}

	for _, invalidID := range []string{
		"123",
		"1234567890abcdef123",
	} {
		if _, err := ParseSpanID(invalidID); err == nil {
			t.Fatalf("expected error for invalid ID length: %s", invalidID)
		}
	}
}

func TestIsValidTraceID(t *testing.T) {
	if !IsValidTraceID("1234567890abcdef1234567890abcdef") {
		t.Fatalf("expected valid trace ID to be recognized")
	}
	if IsValidTraceID("invalid") {
		t.Fatalf("expected invalid trace ID to be rejected")
	}
}

func TestIsValidSpanID(t *testing.T) {
	if !IsValidSpanID("1234567890abcdef") {
		t.Fatalf("expected valid span ID to be recognized")
	}
	if IsValidSpanID("invalid") {
		t.Fatalf("expected invalid span ID to be rejected")
	}
}

func TestTraceContextManagement(t *testing.T) {
	var originalCtx context.Context
	traceContext := TraceContext{
		TraceID: "test-trace",
		SpanID:  "test-span",
		Flags:   TraceFlagsSampled,
		Baggage: map[string]string{
			"user.id":    "123",
			"request.id": "abc",
		},
	}

	ctx := WithTraceContext(originalCtx, traceContext)
	retrieved := TraceContextFromContext(ctx)
	if retrieved == nil {
		t.Fatalf("expected trace context to be retrieved")
	}
	if retrieved.TraceID != "test-trace" {
		t.Fatalf("expected trace ID to match")
	}
	if retrieved.SpanID != "test-span" {
		t.Fatalf("expected span ID to match")
	}
	if retrieved.Flags != TraceFlagsSampled {
		t.Fatalf("expected flags to match")
	}
	if !retrieved.IsSampled() {
		t.Fatalf("expected sampled flag to be true")
	}
	if retrieved.Baggage["user.id"] != "123" {
		t.Fatalf("expected baggage to be preserved")
	}
}

func TestTraceContextFromContextNilSafe(t *testing.T) {
	if tc := TraceContextFromContext(t.Context()); tc != nil {
		t.Fatalf("expected nil trace context for nil context, got %#v", tc)
	}
}

func TestWithSpanIDStringPreservesExistingTraceContext(t *testing.T) {
	base := WithTraceContext(t.Context(), TraceContext{
		TraceID: "trace-abc",
		Baggage: map[string]string{"user.id": "123"},
	})
	updated := WithSpanIDString(base, "1234567890abcdef")
	tc := TraceContextFromContext(updated)
	if tc == nil {
		t.Fatal("expected TraceContext to be set")
	}
	if string(tc.TraceID) != "trace-abc" {
		t.Fatalf("expected TraceID %q, got %q", "trace-abc", tc.TraceID)
	}
	if string(tc.SpanID) != "1234567890abcdef" {
		t.Fatalf("expected SpanID %q, got %q", "1234567890abcdef", tc.SpanID)
	}
	if tc.Baggage["user.id"] != "123" {
		t.Fatalf("expected baggage to be preserved, got %#v", tc.Baggage)
	}
}

func TestWithSpanIDStringIgnoresInvalidSpanID(t *testing.T) {
	base := WithTraceContext(t.Context(), TraceContext{
		TraceID: "trace-abc",
		SpanID:  "1111111111111111",
		Baggage: map[string]string{"user.id": "123"},
	})

	updated := WithSpanIDString(base, "span-new")
	tc := TraceContextFromContext(updated)
	if tc == nil {
		t.Fatal("expected TraceContext to be set")
	}
	if string(tc.SpanID) != "1111111111111111" {
		t.Fatalf("expected invalid span id to be ignored, got %q", tc.SpanID)
	}
	if tc.Baggage["user.id"] != "123" {
		t.Fatalf("expected baggage to be preserved, got %#v", tc.Baggage)
	}
}
