package tracer

import (
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

func TestRandomIDGenerator(t *testing.T) {
	gen := NewRandomIDGenerator()

	id1 := gen.GenerateTraceID()
	id2 := gen.GenerateTraceID()
	if id1 == id2 {
		t.Fatalf("expected different trace IDs")
	}
	if len(id1) != contract.TraceIDLength {
		t.Fatalf("expected trace ID length %d, got %d", contract.TraceIDLength, len(id1))
	}

	sid1 := gen.GenerateSpanID()
	sid2 := gen.GenerateSpanID()
	if sid1 == sid2 {
		t.Fatalf("expected different span IDs")
	}
	if len(sid1) != contract.SpanIDLength {
		t.Fatalf("expected span ID length %d, got %d", contract.SpanIDLength, len(sid1))
	}
}

func TestTracer(t *testing.T) {
	tracer := NewTracer(DefaultTracerConfig())

	ctx, span := tracer.StartTrace(t.Context(), "test-operation")

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

	traceCtx := contract.TraceContextFromContext(ctx)
	if traceCtx == nil {
		t.Fatalf("expected trace context in context")
	}
	if traceCtx.TraceID != span.TraceID {
		t.Fatalf("expected trace context to match span trace ID")
	}
	if traceCtx.SpanID != span.ID {
		t.Fatalf("expected trace context to match span ID")
	}

	time.Sleep(time.Millisecond)
	tracer.EndSpan(span)

	if span.EndTime == nil {
		t.Fatalf("expected end time to be set")
	}
	if span.Duration <= 0 {
		t.Fatalf("expected positive duration")
	}
}

func TestTracerWithOptions(t *testing.T) {
	tracer := NewTracer(DefaultTracerConfig())
	attrs := map[string]any{"service.name": "test-service"}

	ctx, span := tracer.StartTrace(t.Context(), "test-operation",
		WithTraceAttributes(attrs),
		WithSpanKind(SpanKindClient),
	)

	if span.Kind != SpanKindClient {
		t.Fatalf("expected span kind to be client")
	}
	if contract.TraceContextFromContext(ctx) == nil {
		t.Fatalf("expected trace context")
	}
}

func TestStartChildSpan(t *testing.T) {
	tracer := NewTracer(DefaultTracerConfig())
	_, parentSpan := tracer.StartTrace(t.Context(), "parent-operation")
	ctx, childSpan := tracer.StartChildSpan(t.Context(), parentSpan, "child-operation")

	if childSpan == nil {
		t.Fatalf("expected child span to be created")
	}
	if childSpan.TraceID != parentSpan.TraceID {
		t.Fatalf("expected child span to have same trace ID as parent")
	}
	if childSpan.ParentSpanID == nil || *childSpan.ParentSpanID != parentSpan.ID {
		t.Fatalf("expected child span to reference parent span ID")
	}

	traceCtx := contract.TraceContextFromContext(ctx)
	if traceCtx.SpanID != childSpan.ID {
		t.Fatalf("expected context to contain child span ID")
	}
	if traceCtx.ParentSpanID == nil || *traceCtx.ParentSpanID != parentSpan.ID {
		t.Fatalf("expected context to contain parent span ID")
	}
}

func TestStartChildSpanWithNilParent(t *testing.T) {
	tracer := NewTracer(DefaultTracerConfig())
	ctx, span := tracer.StartChildSpan(t.Context(), nil, "standalone-operation")

	if span == nil {
		t.Fatalf("expected span to be created")
	}
	if span.ParentSpanID != nil {
		t.Fatalf("expected no parent span ID for standalone operation")
	}
	if contract.TraceContextFromContext(ctx).ParentSpanID != nil {
		t.Fatalf("expected no parent span ID in context")
	}
}

func TestTracerMaxSpansPerTrace(t *testing.T) {
	config := DefaultTracerConfig()
	config.SamplingRate = 1.0
	config.MaxSpansPerTrace = 1

	tracer := NewTracer(config)
	ctx, root := tracer.StartTrace(t.Context(), "root")
	_, child := tracer.StartChildSpan(ctx, root, "child")
	tracer.EndSpan(child)
	tracer.EndSpan(root)

	trace, exists := tracer.collector.GetTrace(root.TraceID)
	if !exists {
		t.Fatalf("expected trace to be collected")
	}
	if len(trace.Spans) != 1 {
		t.Fatalf("expected 1 span collected, got %d", len(trace.Spans))
	}
}

func TestDefaultTracerConfig(t *testing.T) {
	cfg := DefaultTracerConfig()
	if cfg.ServiceName != "plumego" {
		t.Fatalf("expected service name plumego, got %s", cfg.ServiceName)
	}
	if cfg.SamplingRate != 0.1 {
		t.Fatalf("expected sampling rate 0.1, got %f", cfg.SamplingRate)
	}
	if cfg.MaxSpansPerTrace != 1000 {
		t.Fatalf("expected max spans 1000, got %d", cfg.MaxSpansPerTrace)
	}
}
