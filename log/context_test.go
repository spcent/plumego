package glog

import (
	"context"
	"testing"
)

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "test-trace-123"

	ctx = WithTraceID(ctx, traceID)
	extracted := TraceIDFromContext(ctx)

	if extracted != traceID {
		t.Errorf("expected trace ID %q, got %q", traceID, extracted)
	}
}

func TestTraceIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	extracted := TraceIDFromContext(ctx)

	if extracted != "" {
		t.Errorf("expected empty trace ID, got %q", extracted)
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewGLogger()

	ctx = WithLogger(ctx, logger)
	extracted := LoggerFromContext(ctx)

	if extracted != logger {
		t.Error("expected same logger instance")
	}
}

func TestLoggerFromContext_Default(t *testing.T) {
	ctx := context.Background()
	logger := LoggerFromContext(ctx)
	logger2 := LoggerFromContext(ctx)

	if logger == nil {
		t.Error("expected default logger, got nil")
	}
	if logger != logger2 {
		t.Error("expected default logger instance to be reused")
	}

	// Should return a gLogger instance
	if _, ok := logger.(*gLogger); !ok {
		t.Errorf("expected *gLogger, got %T", logger)
	}
}

func TestLoggerFromContextOrNew(t *testing.T) {
	t.Run("with existing logger", func(t *testing.T) {
		ctx := context.Background()
		existingLogger := NewGLogger()
		ctx = WithLogger(ctx, existingLogger)

		logger, newCtx := LoggerFromContextOrNew(ctx)

		if logger != existingLogger {
			t.Error("expected existing logger to be returned")
		}
		if newCtx != ctx {
			t.Error("expected same context to be returned")
		}
	})

	t.Run("without existing logger", func(t *testing.T) {
		ctx := context.Background()

		logger, newCtx := LoggerFromContextOrNew(ctx)

		if logger == nil {
			t.Error("expected logger to be created")
		}

		// Should have generated a trace ID
		traceID := TraceIDFromContext(newCtx)
		if traceID == "" {
			t.Error("expected trace ID to be generated")
		}

		// Should have attached the logger
		extractedLogger := LoggerFromContext(newCtx)
		if extractedLogger != logger {
			t.Error("expected logger to be attached to context")
		}
	})

	t.Run("with trace ID but no logger", func(t *testing.T) {
		ctx := context.Background()
		existingTraceID := "existing-trace"
		ctx = WithTraceID(ctx, existingTraceID)

		logger, newCtx := LoggerFromContextOrNew(ctx)

		if logger == nil {
			t.Error("expected logger to be created")
		}

		// Should preserve existing trace ID
		traceID := TraceIDFromContext(newCtx)
		if traceID != existingTraceID {
			t.Errorf("expected trace ID %q, got %q", existingTraceID, traceID)
		}
	})
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()
	traceID := "chain-trace-123"
	logger := NewGLogger()

	// Chain context operations
	ctx = WithTraceID(ctx, traceID)
	ctx = WithLogger(ctx, logger)

	// Verify both values are preserved
	extractedTraceID := TraceIDFromContext(ctx)
	extractedLogger := LoggerFromContext(ctx)

	if extractedTraceID != traceID {
		t.Errorf("expected trace ID %q, got %q", traceID, extractedTraceID)
	}

	if extractedLogger != logger {
		t.Error("expected same logger instance")
	}
}

func TestNilContextSafety(t *testing.T) {
	ctx := WithTraceID(nil, "trace-nil")
	if got := TraceIDFromContext(ctx); got != "trace-nil" {
		t.Fatalf("expected trace id from nil context wrapper, got %q", got)
	}

	logger := NewGLogger()
	ctx = WithLogger(nil, logger)
	if got := LoggerFromContext(ctx); got != logger {
		t.Fatalf("expected logger from nil context wrapper")
	}

	if got := TraceIDFromContext(nil); got != "" {
		t.Fatalf("expected empty trace id for nil context, got %q", got)
	}

	fromNil := LoggerFromContext(nil)
	if fromNil == nil {
		t.Fatalf("expected default logger from nil context")
	}

	newLogger, newCtx := LoggerFromContextOrNew(nil)
	if newLogger == nil {
		t.Fatalf("expected logger from nil context")
	}
	if TraceIDFromContext(newCtx) == "" {
		t.Fatalf("expected generated trace id from nil context")
	}
}
