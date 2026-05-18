package client

import (
	"context"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	obstracer "github.com/spcent/plumego/x/observability/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TraceStarter interface {
	StartTrace(context.Context, string, ...obstracer.TraceOption) (context.Context, *obstracer.Span)
	EndSpan(*obstracer.Span, ...obstracer.SpanOption)
	RecordError(*obstracer.Span, error, ...obstracer.ErrorOption)
}

func LoggingInterceptor(logger plumelog.StructuredLogger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if logger != nil {
			code := status.Code(err)
			logger.InfoCtx(ctx, "rpc client call", plumelog.Fields{
				"method":      method,
				"duration_ms": time.Since(start).Milliseconds(),
				"status_code": int(code),
				"status":      code.String(),
			})
		}
		return err
	}
}

func RetryInterceptor(maxAttempts int, retryCodes ...codes.Code) grpc.UnaryClientInterceptor {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	retryable := make(map[codes.Code]struct{}, len(retryCodes))
	for _, code := range retryCodes {
		retryable[code] = struct{}{}
	}
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			if attempt == maxAttempts {
				return err
			}
			if _, ok := retryable[status.Code(err)]; !ok {
				return err
			}
		}
		return err
	}
}

func TracingInterceptor(tracer TraceStarter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var span *obstracer.Span
		if tracer != nil {
			ctx, span = tracer.StartTrace(ctx, method, obstracer.WithSpanKind(obstracer.SpanKindClient))
		}
		ctx = contextWithTraceMetadata(ctx)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if tracer != nil {
			if err != nil {
				tracer.RecordError(span, err)
			}
			tracer.EndSpan(span, obstracer.WithSpanStatus(obstracer.SpanStatus{
				Code:        spanStatusCode(err),
				Description: status.Code(err).String(),
			}))
		}
		return err
	}
}

func contextWithTraceMetadata(ctx context.Context) context.Context {
	traceCtx := contract.TraceContextFromContext(ctx)
	if traceCtx == nil || !traceCtx.Valid() {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx,
		"x-trace-id", traceCtx.TraceID,
		"x-span-id", traceCtx.SpanID,
	)
}

func spanStatusCode(err error) obstracer.StatusCode {
	if err != nil {
		return obstracer.StatusCodeError
	}
	return obstracer.StatusCodeOk
}
