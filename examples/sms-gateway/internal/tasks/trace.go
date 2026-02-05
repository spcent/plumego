package tasks

import (
	"context"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/net/mq"
)

const (
	MetaTraceID      = "trace_id"
	MetaSpanID       = "span_id"
	MetaParentSpanID = "parent_span_id"
)

// AttachTrace captures trace identifiers from context into task metadata.
func AttachTrace(ctx context.Context, task *mq.Task) {
	if task == nil {
		return
	}
	if task.Meta == nil {
		task.Meta = make(map[string]string)
	}

	if tc := contract.TraceContextFromContext(ctx); tc != nil {
		if tc.TraceID != "" {
			task.Meta[MetaTraceID] = string(tc.TraceID)
		}
		if tc.SpanID != "" {
			task.Meta[MetaSpanID] = string(tc.SpanID)
		}
		if tc.ParentSpanID != nil && *tc.ParentSpanID != "" {
			task.Meta[MetaParentSpanID] = string(*tc.ParentSpanID)
		}
		return
	}

	if traceID := contract.TraceIDFromContext(ctx); traceID != "" {
		task.Meta[MetaTraceID] = traceID
	}
}

// ContextWithTrace restores trace identifiers from task metadata into context.
func ContextWithTrace(ctx context.Context, task mq.Task) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(task.Meta) == 0 {
		return ctx
	}

	traceID := strings.TrimSpace(task.Meta[MetaTraceID])
	spanID := strings.TrimSpace(task.Meta[MetaSpanID])
	parentID := strings.TrimSpace(task.Meta[MetaParentSpanID])
	if traceID == "" && spanID == "" && parentID == "" {
		return ctx
	}

	if traceID != "" && spanID != "" {
		tc := contract.TraceContext{
			TraceID: contract.TraceID(traceID),
			SpanID:  contract.SpanID(spanID),
		}
		if parentID != "" {
			parentSpan := contract.SpanID(parentID)
			tc.ParentSpanID = &parentSpan
		}
		return contract.ContextWithTraceContext(ctx, tc)
	}

	if traceID != "" {
		return context.WithValue(ctx, contract.TraceIDKey{}, traceID)
	}

	return ctx
}

