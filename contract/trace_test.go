package contract

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestTraceIDParserInternal(t *testing.T) {
	validID := "1234567890abcdef1234567890abcdef"
	traceID, err := parseTraceID(validID)
	if err != nil {
		t.Fatalf("expected no error for valid trace ID: %v", err)
	}
	if validID != traceID {
		t.Fatalf("expected parsed trace ID to match")
	}

	for _, invalidID := range []string{
		"123",
		"1234567890abcdef1234567890abcdef123",
	} {
		if _, err := parseTraceID(invalidID); err == nil {
			t.Fatalf("expected error for invalid ID length: %s", invalidID)
		}
	}

	if _, err := parseTraceID("invalid_trace_id_!!!"); err == nil {
		t.Fatalf("expected error for invalid hex format")
	}
	if _, err := parseTraceID("00000000000000000000000000000000"); err == nil {
		t.Fatalf("expected error for all-zero trace ID")
	}
}

func TestTraceIDParserInternalCanonicalizesLowercase(t *testing.T) {
	traceID, err := parseTraceID("1234567890ABCDEF1234567890ABCDEF")
	if err != nil {
		t.Fatalf("expected uppercase trace ID to parse: %v", err)
	}
	if got, want := string(traceID), "1234567890abcdef1234567890abcdef"; got != want {
		t.Fatalf("expected canonical trace ID %q, got %q", want, got)
	}
}

func TestSpanIDParserInternal(t *testing.T) {
	validID := "1234567890abcdef"
	spanID, err := parseSpanID(validID)
	if err != nil {
		t.Fatalf("expected no error for valid span ID: %v", err)
	}
	if validID != spanID {
		t.Fatalf("expected parsed span ID to match")
	}

	for _, invalidID := range []string{
		"123",
		"1234567890abcdef123",
	} {
		if _, err := parseSpanID(invalidID); err == nil {
			t.Fatalf("expected error for invalid ID length: %s", invalidID)
		}
	}
	if _, err := parseSpanID("0000000000000000"); err == nil {
		t.Fatalf("expected error for all-zero span ID")
	}
}

func TestSpanIDParserInternalCanonicalizesLowercase(t *testing.T) {
	spanID, err := parseSpanID("1234567890ABCDEF")
	if err != nil {
		t.Fatalf("expected uppercase span ID to parse: %v", err)
	}
	if got, want := string(spanID), "1234567890abcdef"; got != want {
		t.Fatalf("expected canonical span ID %q, got %q", want, got)
	}
}

func TestTraceIDValidityInternal(t *testing.T) {
	if !isValidTraceID("1234567890abcdef1234567890abcdef") {
		t.Fatalf("expected valid trace ID to be recognized")
	}
	if isValidTraceID("invalid") {
		t.Fatalf("expected invalid trace ID to be rejected")
	}
	if isValidTraceID("00000000000000000000000000000000") {
		t.Fatalf("expected all-zero trace ID to be rejected")
	}
}

func TestSpanIDValidityInternal(t *testing.T) {
	if !isValidSpanID("1234567890abcdef") {
		t.Fatalf("expected valid span ID to be recognized")
	}
	if isValidSpanID("invalid") {
		t.Fatalf("expected invalid span ID to be rejected")
	}
	if isValidSpanID("0000000000000000") {
		t.Fatalf("expected all-zero span ID to be rejected")
	}
}

func TestTraceContextValidityHelpers(t *testing.T) {
	validTraceID := "1234567890abcdef1234567890abcdef"
	validSpanID := "1234567890abcdef"

	tests := []struct {
		name      string
		tc        TraceContext
		wantTrace bool
		wantSpan  bool
		wantValid bool
	}{
		{
			name:      "full valid context",
			tc:        TraceContext{TraceID: validTraceID, SpanID: validSpanID},
			wantTrace: true,
			wantSpan:  true,
			wantValid: true,
		},
		{
			name:      "trace only",
			tc:        TraceContext{TraceID: validTraceID},
			wantTrace: true,
		},
		{
			name:     "span only",
			tc:       TraceContext{SpanID: validSpanID},
			wantSpan: true,
		},
		{
			name: "malformed ids",
			tc:   TraceContext{TraceID: "not-a-trace", SpanID: "not-span"},
		},
		{
			name: "empty context",
			tc:   TraceContext{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tc.HasTraceID(); got != tt.wantTrace {
				t.Fatalf("HasTraceID() = %v, want %v", got, tt.wantTrace)
			}
			if got := tt.tc.HasSpanID(); got != tt.wantSpan {
				t.Fatalf("HasSpanID() = %v, want %v", got, tt.wantSpan)
			}
			if got := tt.tc.Valid(); got != tt.wantValid {
				t.Fatalf("Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestTraceContextManagement(t *testing.T) {
	var originalCtx context.Context
	traceContext := TraceContext{
		TraceID: "test-trace",
		SpanID:  "test-span",
		Flags:   traceFlagsSampled,
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
	if retrieved.Flags != traceFlagsSampled {
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
	if tc := TraceContextFromContext(nil); tc != nil {
		t.Fatalf("expected nil trace context for nil context, got %#v", tc)
	}
	if tc := TraceContextFromContext(t.Context()); tc != nil {
		t.Fatalf("expected nil trace context for nil context, got %#v", tc)
	}
}

func TestTraceContextUsesDefensiveCopies(t *testing.T) {
	parent := "1111111111111111"
	baggage := map[string]string{"user.id": "123"}
	ctx := WithTraceContext(t.Context(), TraceContext{
		TraceID:      "trace-abc",
		SpanID:       "2222222222222222",
		ParentSpanID: &parent,
		Baggage:      baggage,
	})

	parent = "3333333333333333"
	baggage["user.id"] = "mutated"

	got := TraceContextFromContext(ctx)
	if got == nil {
		t.Fatal("expected TraceContext")
	}
	if got.ParentSpanID == nil || *got.ParentSpanID != "1111111111111111" {
		t.Fatalf("expected stored parent span to be isolated, got %#v", got.ParentSpanID)
	}
	if got.Baggage["user.id"] != "123" {
		t.Fatalf("expected stored baggage to be isolated, got %#v", got.Baggage)
	}

	*got.ParentSpanID = "4444444444444444"
	got.Baggage["user.id"] = "returned-mutated"

	again := TraceContextFromContext(ctx)
	if again == nil {
		t.Fatal("expected TraceContext on second lookup")
	}
	if again.ParentSpanID == nil || *again.ParentSpanID != "1111111111111111" {
		t.Fatalf("expected returned parent span mutation to be isolated, got %#v", again.ParentSpanID)
	}
	if again.Baggage["user.id"] != "123" {
		t.Fatalf("expected returned baggage mutation to be isolated, got %#v", again.Baggage)
	}
}

func TestTraceContextStableCarrierFields(t *testing.T) {
	rt := reflect.TypeOf(TraceContext{})
	want := []string{"TraceID", "SpanID", "ParentSpanID", "Baggage", "Flags"}
	if rt.NumField() != len(want) {
		t.Fatalf("TraceContext field count = %d, want %d", rt.NumField(), len(want))
	}
	for i, name := range want {
		if got := rt.Field(i).Name; got != name {
			t.Fatalf("TraceContext field %d = %s, want %s", i, got, name)
		}
	}
}

func TestWithTraceContextDoesNotValidateCarrierPolicy(t *testing.T) {
	ctx := WithTraceContext(t.Context(), TraceContext{
		TraceID: "caller-provided-trace",
		SpanID:  "caller-provided-span",
		Baggage: map[string]string{
			"":                          "kept",
			"oversized-or-policy-owned": strings.Repeat("x", 128),
		},
	})

	got := TraceContextFromContext(ctx)
	if got == nil {
		t.Fatal("expected TraceContext")
	}
	if got.Valid() {
		t.Fatalf("invalid caller-provided identifiers should remain carrier data, got valid context: %+v", got)
	}
	if got.Baggage[""] != "kept" {
		t.Fatalf("baggage policy should not be enforced by contract, got %+v", got.Baggage)
	}
	if got.Baggage["oversized-or-policy-owned"] != strings.Repeat("x", 128) {
		t.Fatalf("baggage value should round-trip without policy enforcement, got %+v", got.Baggage)
	}
}

func TestTraceContextReadPatternRequiresValid(t *testing.T) {
	invalid := WithTraceContext(t.Context(), TraceContext{
		TraceID: "not-a-trace-id",
		SpanID:  "not-a-span-id",
	})
	if got := TraceContextFromContext(invalid); got == nil {
		t.Fatal("expected invalid carrier data to be returned for caller inspection")
	} else if got.Valid() {
		t.Fatalf("caller must not treat invalid carrier data as valid trace context: %+v", got)
	}

	valid := WithTraceContext(t.Context(), TraceContext{
		TraceID: "1234567890abcdef1234567890abcdef",
		SpanID:  "1234567890abcdef",
	})
	if got := TraceContextFromContext(valid); got == nil || !got.Valid() {
		t.Fatalf("expected valid trace context after caller-provided valid ids, got %+v", got)
	}
}
