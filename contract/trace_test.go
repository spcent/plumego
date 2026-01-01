package contract

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRandomIDGenerator(t *testing.T) {
	generator := NewRandomIDGenerator()

	// Test trace ID generation
	traceID1 := generator.GenerateTraceID()
	traceID2 := generator.GenerateTraceID()

	if traceID1 == traceID2 {
		t.Fatalf("expected different trace IDs")
	}

	if len(traceID1) != 32 {
		t.Fatalf("expected trace ID length 32, got %d", len(traceID1))
	}

	// Test span ID generation
	spanID1 := generator.GenerateSpanID()
	spanID2 := generator.GenerateSpanID()

	if spanID1 == spanID2 {
		t.Fatalf("expected different span IDs")
	}

	if len(spanID1) != 16 {
		t.Fatalf("expected span ID length 16, got %d", len(spanID1))
	}
}

func TestParseTraceID(t *testing.T) {
	// Test valid trace ID
	validID := "1234567890abcdef1234567890abcdef"
	traceID, err := ParseTraceID(validID)

	if err != nil {
		t.Fatalf("expected no error for valid trace ID: %v", err)
	}

	if TraceID(validID) != traceID {
		t.Fatalf("expected parsed trace ID to match")
	}

	// Test invalid lengths
	invalidIDs := []string{
		"123",                                 // Too short
		"1234567890abcdef1234567890abcdef123", // Too long
	}

	for _, invalidID := range invalidIDs {
		_, err := ParseTraceID(invalidID)
		if err == nil {
			t.Fatalf("expected error for invalid ID length: %s", invalidID)
		}
	}

	// Test invalid hex format
	_, err = ParseTraceID("invalid_trace_id_!!!")
	if err == nil {
		t.Fatalf("expected error for invalid hex format")
	}
}

func TestParseSpanID(t *testing.T) {
	// Test valid span ID
	validID := "1234567890abcdef"
	spanID, err := ParseSpanID(validID)

	if err != nil {
		t.Fatalf("expected no error for valid span ID: %v", err)
	}

	if SpanID(validID) != spanID {
		t.Fatalf("expected parsed span ID to match")
	}

	// Test invalid lengths
	invalidIDs := []string{
		"123",                 // Too short
		"1234567890abcdef123", // Too long
	}

	for _, invalidID := range invalidIDs {
		_, err := ParseSpanID(invalidID)
		if err == nil {
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

func TestTracer(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	// Test starting a trace
	ctx, span := tracer.StartTrace(context.Background(), "test-operation")

	if span == nil {
		t.Fatalf("expected span to be created")
	}

	if span.Name != "test-operation" {
		t.Fatalf("expected span name to be test-operation")
	}

	if span.TraceID == "" {
		t.Fatalf("expected trace ID to be set")
	}

	if span.ID == "" {
		t.Fatalf("expected span ID to be set")
	}

	// Test that context contains trace information
	traceContext := TraceContextFromContext(ctx)
	if traceContext == nil {
		t.Fatalf("expected trace context in context")
	}

	if traceContext.TraceID != span.TraceID {
		t.Fatalf("expected trace context to match span trace ID")
	}

	if traceContext.SpanID != span.ID {
		t.Fatalf("expected trace context to match span ID")
	}

	// Test ending the span
	time.Sleep(time.Millisecond) // Ensure some duration
	tracer.EndSpan(span)

	if span.EndTime == nil {
		t.Fatalf("expected end time to be set")
	}

	if span.Duration <= 0 {
		t.Fatalf("expected positive duration")
	}
}

func TestTracerWithOptions(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	attrs := map[string]any{
		"service.name":    "test-service",
		"service.version": "1.0.0",
		"environment":     "test",
	}

	ctx, span := tracer.StartTrace(
		context.Background(),
		"test-operation",
		WithTraceAttributes(attrs),
		WithSpanKind(SpanKindClient),
	)

	if span.Kind != SpanKindClient {
		t.Fatalf("expected span kind to be client")
	}

	// Note: Trace attributes are applied to the trace, not the span
	// In a real implementation, you'd need to store trace attributes separately
	traceContext := TraceContextFromContext(ctx)
	if traceContext == nil {
		t.Fatalf("expected trace context")
	}
}

func TestStartChildSpan(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	// Start parent trace
	_, parentSpan := tracer.StartTrace(context.Background(), "parent-operation")

	// Start child span
	ctx, childSpan := tracer.StartChildSpan(context.Background(), parentSpan, "child-operation")

	if childSpan == nil {
		t.Fatalf("expected child span to be created")
	}

	if childSpan.TraceID != parentSpan.TraceID {
		t.Fatalf("expected child span to have same trace ID as parent")
	}

	if childSpan.ParentSpanID == nil || *childSpan.ParentSpanID != parentSpan.ID {
		t.Fatalf("expected child span to reference parent span ID")
	}

	// Test that context contains child span information
	traceContext := TraceContextFromContext(ctx)
	if traceContext.SpanID != childSpan.ID {
		t.Fatalf("expected context to contain child span ID")
	}

	if traceContext.ParentSpanID == nil || *traceContext.ParentSpanID != parentSpan.ID {
		t.Fatalf("expected context to contain parent span ID")
	}
}

func TestStartChildSpanWithNilParent(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	// Start span with nil parent should create new trace
	ctx, span := tracer.StartChildSpan(context.Background(), nil, "standalone-operation")

	if span == nil {
		t.Fatalf("expected span to be created")
	}

	if span.ParentSpanID != nil {
		t.Fatalf("expected no parent span ID for standalone operation")
	}

	traceContext := TraceContextFromContext(ctx)
	if traceContext.ParentSpanID != nil {
		t.Fatalf("expected no parent span ID in context")
	}
}

func TestRecordSpanEvent(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	_, span := tracer.StartTrace(context.Background(), "test-operation")

	eventAttrs := map[string]any{
		"event.attr1": "value1",
		"event.attr2": 123,
	}

	tracer.RecordSpanEvent(span, "test-event", WithEventAttributes(eventAttrs))

	if len(span.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(span.Events))
	}

	event := span.Events[0]
	if event.Name != "test-event" {
		t.Fatalf("expected event name to be test-event")
	}

	if event.Attributes["event.attr1"] != "value1" {
		t.Fatalf("expected event attribute to be set")
	}

	if event.Attributes["event.attr2"] != 123 {
		t.Fatalf("expected numeric event attribute to be set")
	}
}

func TestRecordError(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	_, span := tracer.StartTrace(context.Background(), "test-operation")

	testErr := errors.New("test error")
	errorAttrs := map[string]any{
		"error.code": "TEST_ERROR",
	}

	tracer.RecordError(span, testErr, WithErrorAttributes(errorAttrs))

	if span.Status.Code != StatusCodeError {
		t.Fatalf("expected span status to be error")
	}

	if span.Status.Description != testErr.Error() {
		t.Fatalf("expected span status description to match error message")
	}

	if span.Attributes["error.type"] == nil {
		t.Fatalf("expected error type to be set")
	}

	if span.Attributes["error.message"] != testErr.Error() {
		t.Fatalf("expected error message to be set")
	}

	if span.Attributes["error.code"] != "TEST_ERROR" {
		t.Fatalf("expected custom error code to be set")
	}
}

func TestSimpleTraceCollector(t *testing.T) {
	collector := NewSimpleTraceCollector()

	// Test collecting traces
	trace := &Trace{
		ID:         "test-trace-id",
		RootSpanID: "root-span-id",
		StartTime:  time.Now(),
		Spans:      []*Span{},
		Attributes: make(map[string]any),
	}

	collector.Collect(trace)

	// Test retrieving trace
	retrieved, exists := collector.GetTrace("test-trace-id")
	if !exists {
		t.Fatalf("expected trace to exist")
	}

	if retrieved.ID != "test-trace-id" {
		t.Fatalf("expected retrieved trace ID to match")
	}

	// Test filtering
	filtered := collector.GetTraces(func(t *Trace) bool {
		return t.ID == "test-trace-id"
	})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered trace, got %d", len(filtered))
	}

	// Test non-existent trace
	_, exists = collector.GetTrace("non-existent")
	if exists {
		t.Fatalf("expected non-existent trace to not exist")
	}
}

func TestProbabilitySampler(t *testing.T) {
	sampler := NewProbabilitySampler(0.5) // 50% sampling rate

	// Generate test trace IDs
	traceIDs := make([]TraceID, 100)
	for i := 0; i < 100; i++ {
		traceIDs[i] = TraceID(generateTestTraceID(i))
	}

	// Debug: check first few trace IDs and their sampling decisions
	for i := 0; i < 5; i++ {
		t.Logf("Trace ID %d: %s, should sample: %v", i, traceIDs[i], sampler.ShouldSample(traceIDs[i]))
	}

	// Count sampled traces
	sampledCount := 0
	for _, traceID := range traceIDs {
		if sampler.ShouldSample(traceID) {
			sampledCount++
		}
	}

	t.Logf("Sampled count: %d/100", sampledCount)

	// With 100 traces and 50% sampling, we expect roughly 50 sampled traces
	// Allow for some variance (40-60 range)
	if sampledCount < 40 || sampledCount > 60 {
		t.Fatalf("expected sampling to be approximately 50%%, got %d/100 (%d%%)",
			sampledCount, sampledCount)
	}
}

func TestProbabilitySamplerZeroProbability(t *testing.T) {
	sampler := NewProbabilitySampler(0.0) // 0% sampling rate

	// All traces should be rejected
	for i := 0; i < 10; i++ {
		traceID := TraceID(generateTestTraceID(i))
		if sampler.ShouldSample(traceID) {
			t.Fatalf("expected no traces to be sampled with 0%% probability")
		}
	}
}

func TestProbabilitySamplerFullProbability(t *testing.T) {
	sampler := NewProbabilitySampler(1.0) // 100% sampling rate

	// All traces should be accepted
	for i := 0; i < 10; i++ {
		traceID := TraceID(generateTestTraceID(i))
		if !sampler.ShouldSample(traceID) {
			t.Fatalf("expected all traces to be sampled with 100%% probability")
		}
	}
}

func TestTracerGetActiveSpan(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	// Start a trace
	ctx, span := tracer.StartTrace(context.Background(), "test-operation")

	// Get active span
	activeSpan := tracer.GetActiveSpan(ctx)
	if activeSpan == nil {
		t.Fatalf("expected active span to be found")
	}

	if activeSpan.ID != span.ID {
		t.Fatalf("expected active span to match original span")
	}

	// End the span
	tracer.EndSpan(span)

	// Active span should no longer be available
	activeSpan = tracer.GetActiveSpan(ctx)
	if activeSpan != nil {
		t.Fatalf("expected active span to be nil after ending")
	}
}

func TestTracerGetTraceStatistics(t *testing.T) {
	config := DefaultTracerConfig()
	config.SamplingRate = 1.0 // 100% sampling for tests
	tracer := NewTracer(config)

	// Create some traces
	for i := 0; i < 5; i++ {
		_, span := tracer.StartTrace(context.Background(), "operation-%d")

		// Record some events
		tracer.RecordSpanEvent(span, "event-1")
		tracer.RecordSpanEvent(span, "event-2")

		// Add some attributes
		span.Attributes["operation.id"] = i

		// End span
		time.Sleep(time.Millisecond)
		tracer.EndSpan(span, WithSpanStatus(SpanStatus{Code: StatusCodeOk}))
	}

	// Get statistics
	stats := tracer.GetTraceStatistics()

	if stats.TotalTraces != 5 {
		t.Fatalf("expected 5 total traces, got %d", stats.TotalTraces)
	}

	if stats.TotalSpans != 5 {
		t.Fatalf("expected 5 total spans, got %d", stats.TotalSpans)
	}

	if stats.CompletedTraces != 5 {
		t.Fatalf("expected 5 completed traces, got %d", stats.CompletedTraces)
	}

	if stats.ActiveTraces != 0 {
		t.Fatalf("expected 0 active traces, got %d", stats.ActiveTraces)
	}

	if stats.ByStatus[StatusCodeOk] != 5 {
		t.Fatalf("expected 5 successful spans, got %d", stats.ByStatus[StatusCodeOk])
	}

	if stats.AverageDuration <= 0 {
		t.Fatalf("expected positive average duration")
	}
}

func TestTraceContextManagement(t *testing.T) {
	// Test adding and retrieving trace context
	originalCtx := context.Background()
	traceContext := TraceContext{
		TraceID: "test-trace",
		SpanID:  "test-span",
		Flags:   TraceFlagsSampled,
		Sampled: true,
		Baggage: map[string]string{
			"user.id":    "123",
			"request.id": "abc",
		},
	}

	ctx := ContextWithTraceContext(originalCtx, traceContext)

	// Retrieve context
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

	if !retrieved.Sampled {
		t.Fatalf("expected sampled flag to be true")
	}

	if retrieved.Baggage["user.id"] != "123" {
		t.Fatalf("expected baggage to be preserved")
	}
}

func TestLegacyTraceIDKeyCompatibility(t *testing.T) {
	originalCtx := context.Background()

	// Test with new trace context
	traceContext := TraceContext{
		TraceID: "legacy-trace-id",
		SpanID:  "legacy-span-id",
	}

	ctx := ContextWithTraceContext(originalCtx, traceContext)

	// Legacy function should return trace ID from new context
	traceID := TraceIDFromContext(ctx)
	if traceID != "legacy-trace-id" {
		t.Fatalf("expected legacy function to return trace ID from new context")
	}

	// Test with old-style context value
	oldCtx := context.WithValue(originalCtx, TraceIDKey{}, "old-trace-id")
	oldTraceID := TraceIDFromContext(oldCtx)
	if oldTraceID != "old-trace-id" {
		t.Fatalf("expected legacy function to work with old-style context")
	}

	// Test with empty context
	emptyCtx := context.Background()
	emptyTraceID := TraceIDFromContext(emptyCtx)
	if emptyTraceID != "" {
		t.Fatalf("expected empty trace ID for empty context")
	}
}

func TestDefaultTracerConfig(t *testing.T) {
	config := DefaultTracerConfig()

	if config.ServiceName != "plumego" {
		t.Fatalf("expected service name to be plumego")
	}

	if config.ServiceVersion != "1.0.0" {
		t.Fatalf("expected service version to be 1.0.0")
	}

	if config.Environment != "production" {
		t.Fatalf("expected environment to be production")
	}

	if config.SamplingRate != 0.1 {
		t.Fatalf("expected sampling rate to be 0.1")
	}

	if config.MaxSpansPerTrace != 1000 {
		t.Fatalf("expected max spans per trace to be 1000")
	}

	if config.MaxTraceAge != 24*time.Hour {
		t.Fatalf("expected max trace age to be 24 hours")
	}
}

func TestNewTracer(t *testing.T) {
	config := TracerConfig{
		ServiceName:      "test-service",
		ServiceVersion:   "2.0.0",
		Environment:      "development",
		SamplingRate:     0.5,
		MaxSpansPerTrace: 500,
		MaxTraceAge:      time.Hour,
	}

	tracer := NewTracer(config)

	if tracer == nil {
		t.Fatalf("expected tracer to be created")
	}

	// Verify that tracer has the expected components
	if tracer.generator == nil {
		t.Fatalf("expected generator to be set")
	}

	if tracer.collector == nil {
		t.Fatalf("expected collector to be set")
	}

	if tracer.sampler == nil {
		t.Fatalf("expected sampler to be set")
	}
}

// Helper function to generate test trace IDs for testing
func generateTestTraceID(i int) string {
	// Generate different deterministic IDs for testing with good distribution
	// Use a seed and multiply by a prime number to get better distribution
	seed := (i * 2654435761) & 0xFFFFFFFF
	return fmt.Sprintf("%016x", seed)
}

// Test utility functions for span options
func TestSpanOptions(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	attrs := map[string]any{
		"test.attr": "value",
	}

	_, span := tracer.StartTrace(
		context.Background(),
		"test-operation",
		func(trace *Trace, span *Span) {
			for k, v := range attrs {
				span.Attributes[k] = v
			}
			span.Kind = SpanKindProducer
		},
	)

	if span.Attributes["test.attr"] != "value" {
		t.Fatalf("expected span attribute to be set")
	}

	if span.Kind != SpanKindProducer {
		t.Fatalf("expected span kind to be producer")
	}

	// Test end options
	time.Sleep(time.Millisecond)
	tracer.EndSpan(
		span,
		WithSpanAttributes(map[string]any{"end.attr": "end-value"}),
		WithSpanStatus(SpanStatus{Code: StatusCodeOk, Description: "completed"}),
	)

	if span.Attributes["end.attr"] != "end-value" {
		t.Fatalf("expected end attribute to be set")
	}

	if span.Status.Code != StatusCodeOk {
		t.Fatalf("expected status code to be set")
	}

	if span.Status.Description != "completed" {
		t.Fatalf("expected status description to be set")
	}
}

func TestEventOptions(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	_, span := tracer.StartTrace(context.Background(), "test-operation")

	attrs := map[string]any{
		"event.custom": "value",
		"event.number": 42,
	}

	tracer.RecordSpanEvent(
		span,
		"test-event",
		WithEventAttributes(attrs),
	)

	if len(span.Events) != 1 {
		t.Fatalf("expected 1 event")
	}

	event := span.Events[0]
	if event.Attributes["event.custom"] != "value" {
		t.Fatalf("expected custom event attribute")
	}

	if event.Attributes["event.number"] != 42 {
		t.Fatalf("expected numeric event attribute")
	}
}

func TestErrorOptions(t *testing.T) {
	config := DefaultTracerConfig()
	tracer := NewTracer(config)

	_, span := tracer.StartTrace(context.Background(), "test-operation")

	err := errors.New("test error")
	attrs := map[string]any{
		"error.custom": "value",
		"retryable":    true,
	}

	tracer.RecordError(
		span,
		err,
		WithErrorAttributes(attrs),
	)

	if span.Attributes["error.custom"] != "value" {
		t.Fatalf("expected custom error attribute")
	}

	if span.Attributes["retryable"] != true {
		t.Fatalf("expected retryable attribute")
	}
}
