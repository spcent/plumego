package log

import "context"

// requestIDKey is an unexported zero-value context key declared at the call
// site, following the project's context-accessor convention
// (With{Type} + {Type}FromContext).
type requestIDKey struct{}

// RequestIDField is the structured field name under which a request ID carried
// by the context is logged by the *Ctx methods.
const RequestIDField = "request_id"

// WithRequestID returns a context that carries id for the *Ctx logging methods.
//
// The carrier is self-contained: the log package cannot import contract, which
// owns the transport-level request ID accessors, so applications that want the
// request ID attached to log output should populate this carrier explicitly
// (for example from middleware/requestid).
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request ID stored by WithRequestID, if any.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}

// ctxFields extracts the log package's own context-carried fields for the *Ctx
// methods. It returns nil when the context carries no loggable values.
func ctxFields(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}
	if id, ok := RequestIDFromContext(ctx); ok && id != "" {
		return Fields{RequestIDField: id}
	}
	return nil
}
